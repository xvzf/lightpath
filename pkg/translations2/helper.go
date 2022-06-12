package translations2

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
)

func ipStringToIpFamily(ip string) v1.IPFamily {
	if strings.Contains(ip, ":") {
		return v1.IPv6Protocol
	}
	return v1.IPv4Protocol
}

func getClusterName(namespace, name, ipFamily string, port int32) string {
	return fmt.Sprintf("%s|%s|%s|%d", ipFamily, namespace, name, port)
}

func getListenerName(namespace, name, ip string, port int32) string {
	return fmt.Sprintf("%s|%s|%s|%s|%d", ipStringToIpFamily(ip), namespace, name, ip, port)
}
