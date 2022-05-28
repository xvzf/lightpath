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
	udpproxy "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/udp/udp_proxy/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/xvzf/lightpath/pkg/state/snapshot"
)

func getTCPProxyFilterConfig(svcObj *v1.Service, ipFamily v1.IPFamily, protocol v1.Protocol, port int32) *anypb.Any {
	proxy := &tcpproxy.TcpProxy{
		StatPrefix: "destination_tcp",
		ClusterSpecifier: &tcpproxy.TcpProxy_Cluster{
			Cluster: "fixme", // FIXME !!!
		},
	}

	pbst, err := anypb.New(proxy)
	if err != nil {
		panic(err) // should never happen
	}

	return pbst
}

func getUDPProxyFilterConfig(svcObj *v1.Service, ipFamily v1.IPFamily, protocol v1.Protocol, port int32) *anypb.Any {
	proxy := &udpproxy.UdpProxyConfig{
		StatPrefix: "destination_udp",
		RouteSpecifier: &udpproxy.UdpProxyConfig_Cluster{
			Cluster: "fixme", // FIXME
		},
		// UseOriginalSrcIp: false, // DRS
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
						PortSpecifier: &core.SocketAddress_PortValue{
							PortValue: uint32(servicePort.Port),
						},
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
						TypedConfig: getTCPProxyFilterConfig(
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

func getUDPListener(svcObj *v1.Service, servicePort v1.ServicePort) []*listener.Listener {
	res := []*listener.Listener{}

	// ClusterIP(s)
	for _, clusterIP := range svcObj.Spec.ClusterIPs {
		if clusterIP == v1.ClusterIPNone {
			continue
		}

		listener := &listener.Listener{
			Name: fmt.Sprintf("%s-%s-%d", clusterIP, servicePort.Protocol, servicePort.Port),
			UdpListenerConfig: &listener.UdpListenerConfig{
				DownstreamSocketConfig: &core.UdpSocketConfig{
					MaxRxDatagramSize: wrapperspb.UInt64(1500),
					PreferGro:         wrapperspb.Bool(true),
				},
				QuicOptions: nil, // this UDP listener is not related to HTTP/3
			},
			Address: &core.Address{
				Address: &core.Address_SocketAddress{
					SocketAddress: &core.SocketAddress{
						Protocol: mapKubeProtocolToEnvoyProtocol(servicePort.Protocol),
						Address:  clusterIP,
						PortSpecifier: &core.SocketAddress_PortValue{
							PortValue: uint32(servicePort.Port),
						},
					},
				},
			},
			TrafficDirection: core.TrafficDirection_OUTBOUND,
			BindToPort:       wrapperspb.Bool(false),
			// Listenerfilter, no filter chains
			/*
				ListenerFilters: []*listener.ListenerFilter{{
					Name: "envoy.filters.udp_listener.udp_proxy",
					ConfigType: &listener.ListenerFilter_TypedConfig{
						TypedConfig: getUDPProxyFilterConfig(
							svcObj,
							ipStringToIpFamily(clusterIP),
							servicePort.Protocol,
							int32(servicePort.TargetPort.IntValue()),
						),
					},
				}},*/
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
			// listeners = append(listeners, getTCPListener(svcSnap.Obj, servicePort)...)
		case v1.ProtocolUDP:
			listeners = append(listeners, getUDPListener(svcSnap.Obj, servicePort)...)
		default:
			klog.Warning("Not implemented", "protocol", servicePort.Protocol)
		}
	}

	return listeners
}
