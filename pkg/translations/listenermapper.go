package translations

import (
	"fmt"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	tcpproxy "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/xvzf/lightpath/pkg/state/snapshot"
)

func getTcpProxyFilterConfig(svcObj *v1.Service, ipFamily v1.IPFamily, protocol v1.Protocol, port int32) *anypb.Any {
	proxy := &tcpproxy.TcpProxy{
		StatPrefix:       "destination",
		ClusterSpecifier: &tcpproxy.TcpProxy_Cluster{},
	}

	pbst, err := anypb.New(proxy)
	if err != nil {
		panic(err) // should never happen
	}

	return pbst
}

func getTCPListener(svcObj *v1.Service, servicePort v1.ServicePort) []*listener.Listener {
	res := []*listener.Listener{}

	// ClusterIP(s)
	for _, clusterIP := range svcObj.Spec.ClusterIPs {
		if clusterIP == v1.ClusterIPNone {
			continue
		}

		listener := &listener.Listener{
			Name: fmt.Sprintf("%s-%s-%d", clusterIP, servicePort.Protocol, servicePort.Port),
			Address: &core.Address{
				Address: &core.Address_SocketAddress{
					SocketAddress: &core.SocketAddress{
						Protocol: mapKubeProtocolToEnvoyProtocol(servicePort.Protocol),
						Address:  clusterIP,
					},
				},
			},
			TrafficDirection: core.TrafficDirection_OUTBOUND,
			BindToPort:       wrapperspb.Bool(false), // Proxied by IPtables
			// We just have one filter for the TCPProxy here
			FilterChains: []*listener.FilterChain{{
				Filters: []*listener.Filter{{
					Name: wellknown.TCPProxy,
					ConfigType: &listener.Filter_TypedConfig{
						TypedConfig: getTcpProxyFilterConfig(
							svcObj,
							ipStringToIpFamily(clusterIP),
							servicePort.Protocol,
							int32(servicePort.TargetPort.IntValue()),
						),
					},
				}},
			}},
		}
		res = append(res, listener)
	}

	// FIXME add NodePort support

	return res
}

func getListenerFromServiceSnapshpt(svcSnap *snapshot.Service) []*listener.Listener {
	listeners := []*listener.Listener{}

	for _, servicePort := range svcSnap.Obj.Spec.Ports {
		// multi-IPFamily support is based on e.g. ClusterIP/NodePort
		switch servicePort.Protocol {
		case v1.ProtocolTCP:
			listeners = append(listeners, getTCPListener(svcSnap.Obj, servicePort)...)
		default:
			klog.Warning("Not implemented", "protocol", servicePort.Protocol)
		}
	}

	return listeners
}
