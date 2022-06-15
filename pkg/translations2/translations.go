package translations2

import (
	"fmt"
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	tcp_proxy "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/xvzf/lightpath/pkg/state/snapshot"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
)

const (
	DEFAULT_CONNECT_TIMEOUT = 5 * time.Second
)

type KubeMapper struct {
}

// EnvoySnaphsotFromKubeSnapshot generates an Envoy configuration snapshot servable to xDS APIs from Kubernetes Service/EndpointSlices.
func (km *KubeMapper) EnvoySnapshotFromKubeSnapshot(snap *snapshot.Snapshot) (*cache.Snapshot, error) {

	clusters := []types.Resource{}
	listeners := []types.Resource{}
	endpoints := []types.Resource{}

	for _, svc := range snap.Services {
		for _, port := range svc.Obj.Spec.Ports {
			if port.Protocol != v1.ProtocolTCP {
				klog.Warning("Protocol %s not supported, skipping", port.Protocol)
				continue
			}
			clusters = append(clusters, km.MapServicePortToClusters(svc.Obj, &port)...)
			listeners = append(listeners, km.MapServicePortToListeners(svc.Obj, &port)...)
		}
		// Map endpounts
		endpoints = append(endpoints, km.MapEndpointSliceToLocalityEndpoints(svc.Obj, svc.EndpointSlices)...)
	}

	envoySnap, err := cache.NewSnapshot(time.Now().Format(time.RFC3339Nano),
		map[resource.Type][]types.Resource{
			resource.ListenerType: listeners,
			resource.ClusterType:  clusters,
			resource.EndpointType: endpoints,
		},
	)

	return envoySnap, err
}

// MapServicePortToClusters creates a cluster for each service and ipfamily
func (km *KubeMapper) MapServicePortToClusters(svc *v1.Service, port *v1.ServicePort) []types.Resource {
	buf := []types.Resource{}

	// map both IPv4&IPv6
	for _, ipFamily := range svc.Spec.IPFamilies {
		// Port
		clusterName := getClusterName(svc.Namespace, svc.Name, string(ipFamily), port.TargetPort.IntVal)
		buf = append(buf, &cluster.Cluster{
			Name:                 clusterName,
			ConnectTimeout:       durationpb.New(DEFAULT_CONNECT_TIMEOUT), // FIXME make configurable with annotation
			ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_EDS},
			EdsClusterConfig: &cluster.Cluster_EdsClusterConfig{
				EdsConfig: &core.ConfigSource{
					ResourceApiVersion:    core.ApiVersion_V3,
					ConfigSourceSpecifier: &core.ConfigSource_Ads{},
				},
			},
			LbPolicy: cluster.Cluster_ROUND_ROBIN, // FIXME make configurable
		})
	}

	return buf
}

// genTCPListener creates a new listener with a name, ip address, port and targetCluster.
func (km *KubeMapper) genTCPListener(listenerName, ip string, port int32, bind bool, targetClusterName string) *listener.Listener {

	tcpProxy, err := anypb.New(&tcp_proxy.TcpProxy{
		StatPrefix: "source_tcp",
		ClusterSpecifier: &tcp_proxy.TcpProxy_Cluster{
			Cluster: targetClusterName,
		},
	})
	if err != nil {
		panic(err) // Should never happen
	}

	return &listener.Listener{
		Name: listenerName,
		// Listen address
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  ip,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: uint32(port),
					},
				},
			},
		},
		BindToPort: wrapperspb.Bool(bind), // proxied here by IPTables/eBPF
		// Just a TCP Proxy for now -> more protocols added later
		FilterChains: []*listener.FilterChain{{
			Filters: []*listener.Filter{{
				Name: wellknown.TCPProxy,
				// Proxy config
				ConfigType: &listener.Filter_TypedConfig{
					TypedConfig: tcpProxy,
				},
			}},
		}},
	}
}

// MapServicePortToListeners creats a virtual listener for each IP:port combination
func (km *KubeMapper) MapServicePortToListeners(svc *v1.Service, port *v1.ServicePort) []types.Resource {
	buf := []types.Resource{}

	// map ClusterIP
	for _, ip := range svc.Spec.ClusterIPs {
		targetClusterName := getClusterName(svc.Namespace, svc.Name, string(ipStringToIpFamily(ip)), port.TargetPort.IntVal)
		listenerName := getListenerName(svc.Namespace, svc.Name, ip, port.Port)

		listener := km.genTCPListener(listenerName, ip, port.Port, false, targetClusterName)
		buf = append(buf, listener)
	}

	// map NodePort
	for _, ipFamily := range svc.Spec.IPFamilies {
		targetClusterName := getClusterName(svc.Namespace, svc.Name, string(ipFamily), port.TargetPort.IntVal)
		ip := "0.0.0.0"
		if ipFamily == v1.IPv6Protocol {
			ip = "::"
		}
		listenerName := getListenerName(svc.Namespace, svc.Name, ip, port.NodePort)

		listener := km.genTCPListener(listenerName, ip, port.NodePort, true, targetClusterName)
		buf = append(buf, listener)
	}

	return buf
}

type locality struct {
	Zone    string
	SubZone string
}

type endpointMeta struct {
	ipFamily v1.IPFamily
	ip       string
	port     int32

	locality locality
}

type endpointState struct {
	ready       bool
	serving     bool
	terminating bool
}

func (km *KubeMapper) extractEndpointMetaAndState(endpointslices []*discoveryv1.EndpointSlice) map[endpointMeta]endpointState {

	res := make(map[endpointMeta]endpointState)

	// extract unique {ip, port, zone, node} tuples
	for _, endpointslice := range endpointslices {
		ipFamily := v1.IPv4Protocol
		if endpointslice.AddressType == discoveryv1.AddressTypeIPv6 {
			ipFamily = v1.IPv6Protocol
		}

		for _, port := range endpointslice.Ports {
			for _, endpoint := range endpointslice.Endpoints {
				res[endpointMeta{
					ipFamily: ipFamily,
					ip:       endpoint.Addresses[0],
					port:     pointer.Int32Deref(port.Port, 0),
					locality: locality{
						Zone:    pointer.StringDeref(endpoint.Zone, "None"),
						SubZone: pointer.StringDeref(endpoint.NodeName, "None"),
					},
				}] = endpointState{
					ready:       pointer.BoolDeref(endpoint.Conditions.Ready, true),
					serving:     pointer.BoolDeref(endpoint.Conditions.Serving, true),
					terminating: pointer.BoolDeref(endpoint.Conditions.Terminating, false),
				}
			}
		}
	}

	return res
}

func (km *KubeMapper) mapEndpointMetaAndStateToEndpoint(meta endpointMeta, state endpointState) *endpoint.LbEndpoint {

	var healthStatus core.HealthStatus
	if state.terminating {
		healthStatus = core.HealthStatus_DRAINING
	} else {
		healthStatus = core.HealthStatus_HEALTHY
	}

	return &endpoint.LbEndpoint{
		HealthStatus: healthStatus,
		HostIdentifier: &endpoint.LbEndpoint_Endpoint{
			Endpoint: &endpoint.Endpoint{
				Hostname: fmt.Sprintf("%s-%d", meta.ip, meta.port),
				Address: &core.Address{
					Address: &core.Address_SocketAddress{
						SocketAddress: &core.SocketAddress{
							Protocol: core.SocketAddress_TCP,
							Address:  meta.ip,
							PortSpecifier: &core.SocketAddress_PortValue{
								PortValue: uint32(meta.port),
							},
						},
					},
				},
			},
		},
	}
}

// MapEndpointSliceToLocalityEndpoints maps endpointslices to a cluster load assignment.
func (km *KubeMapper) MapEndpointSliceToLocalityEndpoints(svc *v1.Service, endpointslices []*discoveryv1.EndpointSlice) []types.Resource {
	loadAssignmentsMap := make(map[string]*endpoint.ClusterLoadAssignment)

	endpoints := make(map[string]map[locality][]*endpoint.LbEndpoint)

	for meta, state := range km.extractEndpointMetaAndState(endpointslices) {
		clusterName := getClusterName(svc.Namespace, svc.Name, string(meta.ipFamily), meta.port)

		// create cluster if not existing
		if _, ok := loadAssignmentsMap[clusterName]; !ok {
			loadAssignmentsMap[clusterName] = &endpoint.ClusterLoadAssignment{
				ClusterName: clusterName,
				Endpoints:   []*endpoint.LocalityLbEndpoints{},
			}
		}

		// map endpoints
		if _, ok := endpoints[clusterName][meta.locality]; !ok {
			endpoints[clusterName] = make(map[locality][]*endpoint.LbEndpoint)
			endpoints[clusterName][meta.locality] = make([]*endpoint.LbEndpoint, 0)
		}
		endpoints[clusterName][meta.locality] = append(endpoints[clusterName][meta.locality], km.mapEndpointMetaAndStateToEndpoint(meta, state))
	}

	// Assign endpoints to loadAssignments
	for clusterName, loadAssignment := range loadAssignmentsMap {
		for locality, endpoints := range endpoints[clusterName] {
			loadAssignment.Endpoints = append(loadAssignment.Endpoints, &endpoint.LocalityLbEndpoints{
				Locality: &core.Locality{
					Zone:    locality.Zone,
					SubZone: locality.SubZone,
				},
				LbEndpoints: endpoints,
			})
		}
	}

	loadAssignments := make([]*endpoint.ClusterLoadAssignment, 0, len(loadAssignmentsMap))
	for _, value := range loadAssignmentsMap {
		loadAssignments = append(loadAssignments, value)
	}

	// Map to proper type
	resources := make([]types.Resource, len(loadAssignments))
	for idx := range loadAssignments {
		resources[idx] = loadAssignments[idx]
	}

	return resources
}
