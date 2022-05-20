package translations

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
)

// create a unique index based on IPFamily, protocol & port.
func idxFamilyProtocolPort(ipFamily v1.IPFamily, protocol v1.Protocol, port int32) string {
	return fmt.Sprintf("%s-%s-%d", ipFamily, protocol, port)
}

func mapKubeProtocolToEnvoyProtocol(protocol v1.Protocol) core.SocketAddress_Protocol {
	if protocol == v1.ProtocolUDP {
		return core.SocketAddress_UDP
	}
	return core.SocketAddress_TCP
}

func ipStringToIpFamily(ip string) v1.IPFamily {
	if strings.Contains(ip, ":") {
		return v1.IPv6Protocol
	}
	return v1.IPv4Protocol
}
