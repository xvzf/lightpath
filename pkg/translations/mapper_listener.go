package translations

import (
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"google.golang.org/protobuf/types/known/wrapperspb"
	v1 "k8s.io/api/core/v1"
)

// MapServicePortToListeners creats a virtual listener for each IP:port combination
func (km *KubeMapper) MapServicePortToListeners(svc *v1.Service, port *v1.ServicePort) []types.Resource {
	buf := []types.Resource{}

	portSettings := getPortSetings(svc, port) // Retrieve port specific settings controlled via service annotations

	// map ClusterIP
	for _, ip := range svc.Spec.ClusterIPs {
		targetClusterName := getClusterName(svc.Namespace, svc.Name, string(ipStringToIpFamily(ip)), port.TargetPort.IntVal)
		listenerName := getListenerName(svc.Namespace, svc.Name, ip, port.Port)

		listener := km.genListener(listenerName, ip, port.Port, false, portSettings, targetClusterName)
		buf = append(buf, listener)
	}

	// map NodePort
	if svc.Spec.Type == "NodePort" {
		for _, ipFamily := range svc.Spec.IPFamilies {
			targetClusterName := getClusterName(svc.Namespace, svc.Name, string(ipFamily), port.TargetPort.IntVal)
			ip := "0.0.0.0"
			if ipFamily == v1.IPv6Protocol {
				ip = "::"
			}
			listenerName := getListenerName(svc.Namespace, svc.Name, ip, port.NodePort)

			listener := km.genListener(listenerName, ip, port.NodePort, true, portSettings, targetClusterName)
			buf = append(buf, listener)
		}
	}

	return buf
}

// genListener bootstraps the listener and attaches the protocol-aware filter chain.
func (km *KubeMapper) genListener(listenerName, ip string, port int32, bind bool, portSettings *PortSettings, targetClusterName string) *listener.Listener {

	var filterChains []*listener.FilterChain
	switch portSettings.Protocol {
	case PROTOCOL_TCP:
		filterChains = km.genTCPFilterChain(portSettings, targetClusterName)
	case PROTOCOL_HTTP:
		filterChains = km.genHTTPFilterChain(portSettings, targetClusterName)
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
		PerConnectionBufferLimitBytes: wrapperspb.UInt32(32768), // 32 KiB, coming from https://www.envoyproxy.io/docs/envoy/v1.23.0/configuration/best_practices/edge
		BindToPort:                    wrapperspb.Bool(bind),    // bind to the port or act on redirect to dummy-ingress port (redirect by IPTables/eBPF)
		// Just a TCP Proxy for now -> more protocols added later
		FilterChains: filterChains,
	}
}
