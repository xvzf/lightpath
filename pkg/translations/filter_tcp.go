package translations

import (
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	tcp_proxy "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// genTCPListener creates a new listener with a name, ip address, port and targetCluster.
func (km *KubeMapper) genTCPFilterChain(portSettings *PortSettings, targetClusterName string) []*listener.FilterChain {

	tcpProxy, err := anypb.New(&tcp_proxy.TcpProxy{
		StatPrefix: "source_tcp",
		ClusterSpecifier: &tcp_proxy.TcpProxy_Cluster{
			Cluster: targetClusterName,
		},
		IdleTimeout:         durationpb.New(portSettings.IdleTimeout),
		UpstreamIdleTimeout: durationpb.New(portSettings.UpstreamIdleTimeout),
		MaxConnectAttempts:  wrapperspb.UInt32(portSettings.NumRetries),
	})
	if err != nil {
		panic(err) // Should never happen
	}

	return []*listener.FilterChain{{
		Filters: []*listener.Filter{{
			Name: wellknown.TCPProxy,
			// Proxy config
			ConfigType: &listener.Filter_TypedConfig{
				TypedConfig: tcpProxy,
			},
		}},
	}}
}
