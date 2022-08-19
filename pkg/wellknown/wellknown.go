package wellknown

import "time"

const (
	LightpathProxyName         = "lightpath.cloud"       // same as in deploy/webhook/rego/labelinject.rego
	LightpathProxyEnabledLabel = "lightpath.cloud/proxy" // same as in deploy/webhook/rego/labelinject.rego, enabled or disabled
	LightpathRedirectIptables  = "iptables"
	LightpathEnvoyPort         = 1666

	// Per port configuration options
	PortConfigAnnotationPrefix = "config.lightpath.cloud/" // matches `<prefix>/<port-name>/...`

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

	// Application protocols supported by lightpath
	AppProtocolTCP   = "tcp"
	AppProtocolHTTP  = "http"
	AppProtocolRedis = "lightpath.cloud/redis"
)
