package translations

import (
	"time"

	accesslogv3 "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	streamaccessloggerv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/stream/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	previoushostv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/retry/host/previous_hosts/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Access log config (JSON)
var stdoutAccessLogConfig = &streamaccessloggerv3.StdoutAccessLog{
	AccessLogFormat: &streamaccessloggerv3.StdoutAccessLog_LogFormat{
		LogFormat: &corev3.SubstitutionFormatString{
			Format: &corev3.SubstitutionFormatString_JsonFormat{
				JsonFormat: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						// Based on the default log line
						"start_time":                     {Kind: &structpb.Value_StringValue{StringValue: "%START_TIME%"}},
						"method":                         {Kind: &structpb.Value_StringValue{StringValue: "%REQ(:METHOD)%"}},
						"authority":                      {Kind: &structpb.Value_StringValue{StringValue: "%REQ(:AUTHORITY)%"}},
						"path":                           {Kind: &structpb.Value_StringValue{StringValue: "%REQ(X-ENVOY-ORIGINAL-PATH?:PATH)%"}},
						"protocol":                       {Kind: &structpb.Value_StringValue{StringValue: "%PROTOCOL%"}},
						"response_code":                  {Kind: &structpb.Value_StringValue{StringValue: "%RESPONSE_CODE%"}},
						"response_flags":                 {Kind: &structpb.Value_StringValue{StringValue: "%RESPONSE_FLAGS%"}},
						"bytes_received":                 {Kind: &structpb.Value_StringValue{StringValue: "%BYTES_RECEIVED%"}},
						"bytes_sent":                     {Kind: &structpb.Value_StringValue{StringValue: "%BYTES_SENT%"}},
						"duration":                       {Kind: &structpb.Value_StringValue{StringValue: "%DURATION%"}},
						"request_id":                     {Kind: &structpb.Value_StringValue{StringValue: "%REQ(X-REQUEST-ID)%"}},
						"upstream_host":                  {Kind: &structpb.Value_StringValue{StringValue: "%UPSTREAM_HOST%"}},
						"upstream_request_attempt_count": {Kind: &structpb.Value_StringValue{StringValue: "%UPSTREAM_REQUEST_ATTEMPT_COUNT%"}},
						"downstream_remote_address":      {Kind: &structpb.Value_StringValue{StringValue: "%DOWNSTREAM_REMOTE_ADDRESS%"}},
					},
				},
			},
		},
	},
}

// genTCPListener creates a new listener with a name, ip address, port and targetCluster.
func (km *KubeMapper) genHTTPFilterChain(portSettings *PortSettings, targetClusterName string) []*listener.FilterChain {

	accessLogConfig, err := anypb.New(stdoutAccessLogConfig)
	if err != nil {
		panic(err) // this should never happen!
	}

	// Bootstrap router config
	routerConfig, err := anypb.New(&router.Router{})
	if err != nil {
		panic(err) // this should never happen!
	}

	// Previous host retry predicate
	retryPreviousHostPredicate, err := anypb.New(&previoushostv3.PreviousHostsPredicate{})
	if err != nil {
		panic(err) // this should never happen!
	}

	hcm := &hcm.HttpConnectionManager{
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
					RetryPolicy: &route.RetryPolicy{
						RetryOn:           portSettings.RetryOn,
						NumRetries:        wrapperspb.UInt32(portSettings.NumRetries),
						PerTryTimeout:     durationpb.New(portSettings.UpstreamRequestTimeout),
						PerTryIdleTimeout: durationpb.New(portSettings.UpstreamIdleTimeout),
						RetryHostPredicate: []*route.RetryPolicy_RetryHostPredicate{
							&route.RetryPolicy_RetryHostPredicate{
								Name: "envoy.retry_host_predicates.previous_hosts", // FIXME should be a well-known
								ConfigType: &route.RetryPolicy_RetryHostPredicate_TypedConfig{
									TypedConfig: retryPreviousHostPredicate,
								},
							},
						},
					},
				},
				},
			}, // route.RouteConfiguratio
		}, // hcm.HttpConnectionManager_RouteConfig
	}

	// Add AccessLog configuration in case it is enabled
	if portSettings.AccessLog {
		hcm.AccessLog = []*accesslogv3.AccessLog{{
			Name: "envoy.access_loggers.stdout", // FIXME this should be a well-known
			ConfigType: &accesslogv3.AccessLog_TypedConfig{
				TypedConfig: accessLogConfig,
			},
		}}
	}

	tcpProxy, err := anypb.New(hcm)
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
