package translations

import (
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"github.com/xvzf/lightpath/pkg/wellknown"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
)

// getProtocol retrievs the protocol of a given port.
func getProtocol(port *v1.ServicePort) Protocol {
	appProtocolOrPortName := pointer.StringDeref(port.AppProtocol, port.Name)

	// Try to match to app protocol
	switch appProtocolOrPortName {
	case wellknown.AppProtocolRedis:
		return PROTOCOL_REDIS
	case wellknown.AppProtocolHTTP:
		return PROTOCOL_HTTP
	case wellknown.AppProtocolTCP:
		return PROTOCOL_TCP
	}

	// if this doesn't work, fallback to TCP protocol
	return PROTOCOL_TCP
}

func getLbPolicy(lbPolicy string) cluster.Cluster_LbPolicy {
	switch lbPolicy {
	case "LEAST_REQUEST":
		return cluster.Cluster_LEAST_REQUEST
	case "ROUND_ROBIN":
		return cluster.Cluster_ROUND_ROBIN
	case "MAGLEV":
		return cluster.Cluster_MAGLEV
	case "RANDOM":
		return cluster.Cluster_RANDOM
	case "RING_HASH":
		return cluster.Cluster_RING_HASH
	}
	klog.Warning("LB Policy unsupported", "policy", lbPolicy)
	return cluster.Cluster_LEAST_REQUEST
}

type PortSettings struct {
	// Port protocol
	Protocol Protocol

	// Listener timeout
	IdleTimeout    time.Duration
	RequestTimeout time.Duration

	// Upstream timeouts
	UpstreamRequestTimeout time.Duration
	UpstreamIdleTimeout    time.Duration
	UpstreamConnectTimeout time.Duration

	// Retires
	EnableRetry bool
	RetryOn     string
	NumRetries  uint32

	// Circuit Breaking
	CircuitBreakerDefaultMaxConnections     uint32
	CircuitBreakerDefaultMaxPendingRequests uint32
	CircuitBreakerDefaultMaxRequests        uint32
	CircuitBreakerDefaultMaxRetries         uint32
	CircuitBreakerDefaultTrackRemaining     bool
	CircuitBreakerHighMaxConnections        uint32
	CircuitBreakerHighMaxPendingRequests    uint32
	CircuitBreakerHighMaxRequests           uint32
	CircuitBreakerHighMaxRetries            uint32
	CircuitBreakerHighTrackRemaining        bool

	// AccessLog
	AccessLog bool

	// Loadbalancing algorithm
	LoadBalancingPolicy cluster.Cluster_LbPolicy

	// FIXME add circuit breaking defaults
}

func getPortSetings(svc *v1.Service, port *v1.ServicePort) *PortSettings {
	protocol := getProtocol(port)
	lbPolicy := getLbPolicy(
		getStringConfig(svc, port, wellknown.LoadBalancingPolicy, wellknown.LoadBalancingPolicyDefault),
	)

	return &PortSettings{
		Protocol: protocol,
		// Listener Timeouts
		IdleTimeout:    getDurationConfig(svc, port, wellknown.PortIdleTimeout, wellknown.PortIdleTimeoutDefault),
		RequestTimeout: getDurationConfig(svc, port, wellknown.PortRequestTimeout, wellknown.PortRequestTimeoutDefault),

		// Upstream Timeouts
		UpstreamRequestTimeout: getDurationConfig(svc, port, wellknown.PortUpstreamRequestTimeout, wellknown.PortUpstreamRequestTimeoutDefault),
		UpstreamIdleTimeout:    getDurationConfig(svc, port, wellknown.PortUpstreamIdleTimeout, wellknown.PortUpstreamIdleTimeoutDefault),
		UpstreamConnectTimeout: getDurationConfig(svc, port, wellknown.PortUpstreamConnectTimeout, wellknown.PortUpstreamConnectTimeoutDefault),

		// (Proxy/HttpConnectionManager) Retry policy
		EnableRetry: getBoolConfig(svc, port, wellknown.PortRetryEnabled, wellknown.PortRetryEnabledDefault),
		RetryOn:     getStringConfig(svc, port, wellknown.PortRetryOn, wellknown.PortRetryOnDefault), // HTTP only, ignored for TCP; there just the connection attempts will be mapped
		NumRetries:  getUint32Config(svc, port, wellknown.PortNumRetries, wellknown.PortNumRetriesDefault),

		// CircuitBreaker
		CircuitBreakerDefaultMaxConnections:     getUint32Config(svc, port, wellknown.CircuitBreakerDefaultMaxConnections, wellknown.CircuitBreakerDefaultMaxConnectionsDefault),
		CircuitBreakerDefaultMaxPendingRequests: getUint32Config(svc, port, wellknown.CircuitBreakerDefaultMaxPendingRequests, wellknown.CircuitBreakerDefaultMaxPendingRequestsDefault),
		CircuitBreakerDefaultMaxRequests:        getUint32Config(svc, port, wellknown.CircuitBreakerDefaultMaxRequests, wellknown.CircuitBreakerDefaultMaxRequestsDefault),
		CircuitBreakerDefaultMaxRetries:         getUint32Config(svc, port, wellknown.CircuitBreakerDefaultMaxRetries, wellknown.CircuitBreakerDefaultMaxRetriesDefault),
		CircuitBreakerDefaultTrackRemaining:     getBoolConfig(svc, port, wellknown.CircuitBreakerDefaultTrackRemaining, wellknown.CircuitBreakerDefaultTrackRemainingDefault),
		CircuitBreakerHighMaxConnections:        getUint32Config(svc, port, wellknown.CircuitBreakerHighMaxConnections, wellknown.CircuitBreakerHighMaxConnectionsDefault),
		CircuitBreakerHighMaxPendingRequests:    getUint32Config(svc, port, wellknown.CircuitBreakerHighMaxPendingRequests, wellknown.CircuitBreakerHighMaxPendingRequestsDefault),
		CircuitBreakerHighMaxRequests:           getUint32Config(svc, port, wellknown.CircuitBreakerHighMaxRequests, wellknown.CircuitBreakerHighMaxRequestsDefault),
		CircuitBreakerHighMaxRetries:            getUint32Config(svc, port, wellknown.CircuitBreakerHighMaxRetries, wellknown.CircuitBreakerHighMaxRetriesDefault),
		CircuitBreakerHighTrackRemaining:        getBoolConfig(svc, port, wellknown.CircuitBreakerHighTrackRemaining, wellknown.CircuitBreakerHighTrackRemainingDefault),

		// LB config
		LoadBalancingPolicy: lbPolicy,
		// AccessLog
		AccessLog: getBoolConfig(svc, port, wellknown.AccessLog, wellknown.AccessLogDefault),
	}
}
