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
		// Retry policy
		EnableRetry: getBoolConfig(svc, port, wellknown.PortRetryEnabled, wellknown.PortRetryEnabledDefault),
		RetryOn:     getStringConfig(svc, port, wellknown.PortRetryOn, wellknown.PortRetryOnDefault), // HTTP only, ignored for TCP; there just the connection attempts will be mapped
		NumRetries:  getUint32Config(svc, port, wellknown.PortNumRetries, wellknown.PortNumRetriesDefault),
		// LB config
		LoadBalancingPolicy: lbPolicy,
	}
}
