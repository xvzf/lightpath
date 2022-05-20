package translations

import (
	"errors"
	"fmt"

	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
)

// mapKubeEndpointToEnvoyLbEndpoint maps a Kubernetes discoveryv1.Endpoint to an Envoy specific LBEndpoint.
func mapKubeEndpointToEnvoyLbEndpoint(protocol v1.Protocol, port int32, kubeEndpoint *discoveryv1.Endpoint) (*endpoint.LbEndpoint, error) {
	if kubeEndpoint == nil {
		return nil, errors.New("kubeEndpoint cannot be nil")
	}
	if len(kubeEndpoint.Addresses) < 1 {
		return nil, errors.New("at least one address has to be present")
	}

	// Map healthstatus to envoy
	var envoyHealthStatus core.HealthStatus
	if kubeEndpoint.Conditions.Serving != nil && kubeEndpoint.Conditions.Terminating != nil && *kubeEndpoint.Conditions.Serving {
		if *kubeEndpoint.Conditions.Terminating {
			envoyHealthStatus = core.HealthStatus_DRAINING
		} else {
			envoyHealthStatus = core.HealthStatus_HEALTHY
		}
	} else {
		envoyHealthStatus = core.HealthStatus_UNKNOWN
	}

	// Return envoy LbEndpoint
	return &endpoint.LbEndpoint{
		HealthStatus: envoyHealthStatus,
		HostIdentifier: &endpoint.LbEndpoint_Endpoint{
			Endpoint: &endpoint.Endpoint{
				Address: &core.Address{
					Address: &core.Address_SocketAddress{
						SocketAddress: &core.SocketAddress{
							Protocol: mapKubeProtocolToEnvoyProtocol(protocol),
							Address:  kubeEndpoint.Addresses[0],
							PortSpecifier: &core.SocketAddress_PortValue{
								PortValue: uint32(port),
							},
						},
					},
				},
			},
		},
	}, nil
}

// mapKubeEndpointsToEnvoyEndpoints maps kubernetes endpoints to envoy endpoints.
// Returns a map of zone -> []endpoint.LbEndpoints.
func mapKubeEndpointsToEnvoyEndpoints(protocol v1.Protocol, port int32, kubeEndpoints *[]discoveryv1.Endpoint) (map[string][]*endpoint.LbEndpoint, error) {
	res := make(map[string][]*endpoint.LbEndpoint)
	if kubeEndpoints == nil {
		klog.Warning("kubeEndpoints is nil")
		return res, nil
	}

	for _, kubeEndpoint := range *kubeEndpoints {
		// Extract zone info
		zone := pointer.StringDeref(kubeEndpoint.Zone, "default")
		// convert endpoint
		envoyEndpoint, err := mapKubeEndpointToEnvoyLbEndpoint(protocol, port, &kubeEndpoint) //nolint
		if err != nil {
			return nil, err
		}

		if _, ok := res[zone]; !ok {
			res[zone] = []*endpoint.LbEndpoint{envoyEndpoint}
		} else {
			res[zone] = append(res[zone], envoyEndpoint)
		}
	}

	return res, nil
}

func mergeZoneMappedEndpoints(a, b map[string][]*endpoint.LbEndpoint) map[string][]*endpoint.LbEndpoint {
	merged := a
	for k, v := range b {
		if _, ok := merged[k]; ok {
			merged[k] = append(merged[k], v...)
		} else {
			merged[k] = b[k]
		}
	}
	return merged
}

func transformZoneMappedEndpointsToLocalityLbEndpoints(zoneMappedEndpoints map[string][]*endpoint.LbEndpoint) []*endpoint.LocalityLbEndpoints {
	res := make([]*endpoint.LocalityLbEndpoints, 0, len(zoneMappedEndpoints))
	for zone, endpoints := range zoneMappedEndpoints {
		// Create LocalityLbEndpoints for each specified zone
		res = append(res, &endpoint.LocalityLbEndpoints{
			Locality: &core.Locality{
				Zone: zone,
			},
			LbEndpoints: endpoints,
		})
	}
	return res
}

func mapKubeEndpointSlicesToMappendLocalityEndpoints(endpointSlices []*discoveryv1.EndpointSlice) (map[string][]*endpoint.LocalityLbEndpoints, error) {
	protocolPortLbEndpointMap := map[string]map[string][]*endpoint.LbEndpoint{}

	// Iterate across all ports
	for _, endpointslice := range endpointSlices {
		// Check if supported
		if endpointslice.AddressType == discoveryv1.AddressTypeFQDN {
			klog.Warning("addresstype FQDN not supported", "endpointslice", fmt.Sprintf("%s/%s", endpointslice.Namespace, endpointslice.Name))
			continue
		}

		for _, endpointSlicePort := range endpointslice.Ports {
			idx := idxFamilyProtocolPort(v1.IPFamily(endpointslice.AddressType), *endpointSlicePort.Protocol, *endpointSlicePort.Port)

			newZoneMappedEndpoints, err := mapKubeEndpointsToEnvoyEndpoints(
				v1.Protocol(pointer.StringDeref((*string)(endpointSlicePort.Protocol), string(v1.ProtocolTCP))),
				pointer.Int32Deref(endpointSlicePort.Port, 0),
				&endpointslice.Endpoints,
			)
			if err != nil {
				return nil, err
			}

			if existing, ok := protocolPortLbEndpointMap[idx]; ok {
				protocolPortLbEndpointMap[idx] = mergeZoneMappedEndpoints(existing, newZoneMappedEndpoints)
			} else {
				protocolPortLbEndpointMap[idx] = newZoneMappedEndpoints
			}
		}
	}

	// Transform nested structure to simple array
	res := make(map[string][]*endpoint.LocalityLbEndpoints, len(protocolPortLbEndpointMap))
	for idx, zoneMappedEndpoints := range protocolPortLbEndpointMap {
		mapped := transformZoneMappedEndpointsToLocalityLbEndpoints(zoneMappedEndpoints)
		if mappedRes, ok := res[idx]; ok {
			res[idx] = append(mappedRes, mapped...)
		} else {
			res[idx] = mapped
		}
	}
	return res, nil
}
