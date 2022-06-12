package translations2

import (
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	tcp_proxyv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
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
	// endpoints := []types.Resource{}

	for _, svc := range snap.Services {
		for _, port := range svc.Obj.Spec.Ports {
			if port.Protocol != v1.ProtocolTCP {
				klog.Warning("Protocol %s not supported, skipping", port.Protocol)
				continue
			}
			clusters = append(clusters, km.MapServicePortToClusters(svc.Obj, &port)...)
			listeners = append(listeners, km.MapServicePortToListeners(svc.Obj, &port)...)
		}

	}

	envoySnap, err := cache.NewSnapshot(time.Now().Format(time.RFC3339Nano),
		map[resource.Type][]types.Resource{
			resource.ListenerType: listeners,
			resource.ClusterType:  clusters,
			// FIXME add resource.EndpointType
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
			LbPolicy:             cluster.Cluster_ROUND_ROBIN, // FIXME make configurable
		})
	}

	return buf
}

// genTCPListener creates a new listener with a name, ip address, port and targetCluster.
func (km *KubeMapper) genTCPListener(listenerName, ip string, port int32, bind bool, targetClusterName string) *listener.Listener {

	tcpProxy, err := anypb.New(&tcp_proxyv3.TcpProxy{
		StatPrefix: "source_tcp",
		ClusterSpecifier: &tcp_proxyv3.TcpProxy_Cluster{
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

func (km *KubeMapper) MapEndpointSliceToLocalityEndpoints(svc *v1.Service, endpointslice *discoveryv1.EndpointSlice) {

}
