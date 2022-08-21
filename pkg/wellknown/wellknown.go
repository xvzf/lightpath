package wellknown

import "time"

const (
	LightpathProxyName         = "lightpath.cloud"       // same as in deploy/webhook/rego/labelinject.rego
	LightpathProxyEnabledLabel = "lightpath.cloud/proxy" // same as in deploy/webhook/rego/labelinject.rego, enabled or disabled
	LightpathRedirectIptables  = "iptables"
	LightpathEnvoyPort         = 1666

	// Per port configuration options
	PortConfigAnnotationPrefix = "config.lightpath.cloud/" // matches `<prefix>/<port-name>-...`

	// Timeouts
	PortIdleTimeout                   = "idle-timeout" // default: 300
	PortIdleTimeoutDefault            = 300 * time.Second
	PortUpstreamIdleTimeout           = "upstream-idle-timeout" // default: 300
	PortUpstreamIdleTimeoutDefault    = 15 * time.Second
	PortRequestTimeout                = "request-timeout" // default: 15s
	PortRequestTimeoutDefault         = 15 * time.Second
	PortUpstreamRequestTimeout        = "upstream-request-timeout" // default: 5s
	PortUpstreamRequestTimeoutDefault = 5 * time.Second
	PortUpstreamConnectTimeout        = "upstream-connect-timeout" // default: 5
	PortUpstreamConnectTimeoutDefault = 5 * time.Second

	// Access Logging
	AccessLog        = "access-log"
	AccessLogDefault = true

	// LoadBalancing Algorithm
	LoadBalancingPolicy        = "lb-policy"
	LoadBalancingPolicyDefault = "LEAST_REQUEST"

	// Retries
	PortRetryEnabled = "retry-enabled" // default: enabled
	PortRetryOn      = "retry-on"      // default HTTP: gateway-error,reset,connect-failure,
	PortNumRetries   = "num-retries"   // default HTTP: gateway-error,reset,connect-failure,
	// Retry defaults
	PortRetryEnabledDefault = true
	PortNumRetriesDefault   = 2
	PortRetryOnDefault      = "5xx,reset,retriable-headers" // default retry options for HTTP

	// Circuit Breaking (per-cluster-level); "disabled" by default
	// There's no option to disable circuit-breaking envoy-side -> set it to very high values
	// in order to pseudo-disable it.
	// Default route priority
	PortCircuitBreakerDefaultMaxConnections            = "circuit-breaker-default-max-connections"
	PortCircuitBreakerDefaultMaxConnectionsDefault     = 1000000000
	PortCircuitBreakerDefaultMaxPendingRequests        = "circuit-breaker-default-max-pending-requests"
	PortCircuitBreakerDefaultMaxPendingRequestsDefault = 1000000000
	PortCircuitBreakerDefaultMaxRequests               = "circuit-breaker-default-max-requests"
	PortCircuitBreakerDefaultMaxRequestsDefault        = 1000000000
	PortCircuitBreakerDefaultMaxRetries                = "circuit-breaker-default-max-retries"
	PortCircuitBreakerDefaultMaxRetriesDefault         = 3
	PortCircuitBreakerDefaultTrackRemaining            = "circuit-breaker-default-track-remaining"
	PortCircuitBreakerDefaultTrackRemainingDefault     = false
	// High route priority
	PortCircuitBreakerHighMaxConnections            = "circuit-breaker-high-max-connections"
	PortCircuitBreakerHighMaxConnectionsDefault     = 1000000000
	PortCircuitBreakerHighMaxPendingRequests        = "circuit-breaker-high-max-pending-requests"
	PortCircuitBreakerHighMaxPendingRequestsDefault = 1000000000
	PortCircuitBreakerHighMaxRequests               = "circuit-breaker-high-max-requests"
	PortCircuitBreakerHighMaxRequestsDefault        = 1000000000
	PortCircuitBreakerHighMaxRetries                = "circuit-breaker-high-max-retries"
	PortCircuitBreakerHighMaxRetriesDefault         = 3
	PortCircuitBreakerHighTrackRemaining            = "circuit-breaker-high-track-remaining"
	PortCircuitBreakerHighTrackRemainingDefault     = false

	// Outlier Detection
	PortOutlierDetectionInterval                         = "outlier-detection-interval"
	PortOutlierDetectionIntervalDefault                  = 10 * time.Second
	PortOutlierDetectionBaseEjectionTime                 = "outlier-detection-base-ejection-time"
	PortOutlierDetectionBaseEjectionTimeDefault          = 30 * time.Second
	PortOutlierDetectionMaxEjectionTime                  = "outlier-detection-max-ejection-time"
	PortOutlierDetectionMaxEjectionTimeDefault           = 300 * time.Second
	PortOutlierDetectionMaxEjectionPercent               = "outlier-detection-max-ejection-percent"
	PortOutlierDetectionMaxEjectionPercentDefault        = 10
	PortOutlierDetectionConsecutive_5Xx                  = "outlier-detection-consecutive-5xx"
	PortOutlierDetectionConsecutive_5XxDefault           = 5
	PortOutlierDetectionConsecutiveGatewayFailure        = "outlier-detection-consecutive-gateway-failure"
	PortOutlierDetectionConsecutiveGatewayFailureDefault = 5

	// Application protocols supported by lightpath
	AppProtocolTCP   = "tcp"
	AppProtocolHTTP  = "http"
	AppProtocolRedis = "lightpath.cloud/redis"
)
