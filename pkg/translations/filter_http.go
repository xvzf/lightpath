package translations

import (
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// genTCPListener creates a new listener with a name, ip address, port and targetCluster.
func (km *KubeMapper) genHTTPFilterChain(portSettings *PortSettings, targetClusterName string) []*listener.FilterChain {

	// Bootstrap router config
	routerConfig, err := anypb.New(&router.Router{})
	if err != nil {
		panic(err) // this should never happen!
	}

	tcpProxy, err := anypb.New(&hcm.HttpConnectionManager{
		StatPrefix: "source_http",
		// Configure sane defaults
		CommonHttpProtocolOptions: &corev3.HttpProtocolOptions{
			IdleTimeout:                  durationpb.New(1 * time.Hour),
			HeadersWithUnderscoresAction: corev3.HttpProtocolOptions_REJECT_REQUEST,
		},
		Http2ProtocolOptions: &corev3.Http2ProtocolOptions{
			// Configuration options based on the best-practises defined here https://www.envoyproxy.io/docs/envoy/v1.23.0/configuration/best_practices/edge
			MaxConcurrentStreams:        wrapperspb.UInt32(100),
			InitialStreamWindowSize:     wrapperspb.UInt32(65536),   //64KiB
			InitialConnectionWindowSize: wrapperspb.UInt32(1048576), // 1MiB
		},
		// Those should be configureable
		StreamIdleTimeout: durationpb.New(portSettings.IdleTimeout),
		RequestTimeout:    durationpb.New(portSettings.RequestTimeout),
		// FIXME assess if it is neccessary to enable preserving the http header case
		// CommonHttpProtocolOptions: &corev3.HttpProtocolOptions{ },
		HttpFilters: []*hcm.HttpFilter{{
			Name: wellknown.Router,
			ConfigType: &hcm.HttpFilter_TypedConfig{
				TypedConfig: routerConfig,
			},
		}},
		// Route configuration; we just have one default route here.
		RouteSpecifier: &hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: &route.RouteConfiguration{
				Name: "default",
				// we have one anycast virtual host here
				VirtualHosts: []*route.VirtualHost{{
					Name: "anycast",
					// the purpose of lightpath is not to perform routing actions but to
					// increase traffic awareness and resilience (!)
					Domains:                    []string{"*"},
					IncludeRequestAttemptCount: true,
					Routes: []*route.Route{{
						// catchall route (prefix / -> all paths are matched)
						Name:  "catchall",
						Match: &route.RouteMatch{PathSpecifier: &route.RouteMatch_Prefix{Prefix: "/"}},
						// Route to the target Cluster; load distribution will be performed on the cluster
						Action: &route.Route_Route{Route: &route.RouteAction{
							ClusterSpecifier: &route.RouteAction_Cluster{
								Cluster: targetClusterName,
							},
							// Upstream timeout control
							IdleTimeout: durationpb.New(portSettings.UpstreamIdleTimeout),
							Timeout:     durationpb.New(portSettings.UpstreamRequestTimeout),
						}}, // Action
					}},
					// Configure retries for the rotue
					RetryPolicy: &route.RetryPolicy{},
				},
				},
			}, // route.RouteConfiguratio
		}, // hcm.HttpConnectionManager_RouteConfig
	})
	if err != nil {
		panic(err) // Should never happen
	}

	return []*listener.FilterChain{{
		Filters: []*listener.Filter{{
			Name: wellknown.HTTPConnectionManager,
			// Proxy config
			ConfigType: &listener.Filter_TypedConfig{
				TypedConfig: tcpProxy,
			},
		}},
	}}
}
