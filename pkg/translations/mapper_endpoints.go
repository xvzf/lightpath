package translations

import (
	"fmt"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/utils/pointer"
)

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

		// create cluster key if not existing
		if _, ok := loadAssignmentsMap[clusterName]; !ok {
			loadAssignmentsMap[clusterName] = &endpoint.ClusterLoadAssignment{
				ClusterName: clusterName,
				Endpoints:   []*endpoint.LocalityLbEndpoints{},
			}
		}

		// make sure DS are ready
		if _, ok := endpoints[clusterName]; !ok {
			endpoints[clusterName] = make(map[locality][]*endpoint.LbEndpoint)
		}
		if _, ok := endpoints[clusterName][meta.locality]; !ok {
			endpoints[clusterName][meta.locality] = make([]*endpoint.LbEndpoint, 0, 1)
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
