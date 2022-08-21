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
	CircuitBreakerDefaultMaxConnections            = "circuit-breaker-default-max-connections"
	CircuitBreakerDefaultMaxConnectionsDefault     = 1000000000
	CircuitBreakerDefaultMaxPendingRequests        = "circuit-breaker-default-max-pending-requests"
	CircuitBreakerDefaultMaxPendingRequestsDefault = 1000000000
	CircuitBreakerDefaultMaxRequests               = "circuit-breaker-default-max-requests"
	CircuitBreakerDefaultMaxRequestsDefault        = 1000000000
	CircuitBreakerDefaultMaxRetries                = "circuit-breaker-default-max-retries"
	CircuitBreakerDefaultMaxRetriesDefault         = 3
	CircuitBreakerDefaultTrackRemaining            = "circuit-breaker-default-track-remaining"
	CircuitBreakerDefaultTrackRemainingDefault     = false
	// High route priority
	CircuitBreakerHighMaxConnections            = "circuit-breaker-high-max-connections"
	CircuitBreakerHighMaxConnectionsDefault     = 1000000000
	CircuitBreakerHighMaxPendingRequests        = "circuit-breaker-high-max-pending-requests"
	CircuitBreakerHighMaxPendingRequestsDefault = 1000000000
	CircuitBreakerHighMaxRequests               = "circuit-breaker-high-max-requests"
	CircuitBreakerHighMaxRequestsDefault        = 1000000000
	CircuitBreakerHighMaxRetries                = "circuit-breaker-high-max-retries"
	CircuitBreakerHighMaxRetriesDefault         = 3
	CircuitBreakerHighTrackRemaining            = "circuit-breaker-high-track-remaining"
	CircuitBreakerHighTrackRemainingDefault     = false

	// Application protocols supported by lightpath
	AppProtocolTCP   = "tcp"
	AppProtocolHTTP  = "http"
	AppProtocolRedis = "lightpath.cloud/redis"
)
