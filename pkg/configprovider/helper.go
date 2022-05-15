package configprovider

import (
	"fmt"

	"github.com/xvzf/lightpath/pkg/state/snapshot"
	v1 "k8s.io/api/core/v1"
)

// create a unique index based on IPFamily, protocol & port
func idxFamilyProtocolPort(ipFamily v1.IPFamily, protocol v1.Protocol, port int32) string {
	return fmt.Sprintf("%s-%s-%d", ipFamily, protocol, port)
}

// getEnvoyClusterName builds a deterministic envoy cluster name based on svc UID, IP Family & port/targetPort
func getEnvoyClusterName(svcSnap *snapshot.Service, ipFamily v1.IPFamily, protocol v1.Protocol, port int32) string {
	return fmt.Sprintf("%s-%s", svcSnap.Obj.GetUID(), idxFamilyProtocolPort(ipFamily, protocol, port))
}

// getEnvoyClusterName builds a deterministic envoy cluster name based on svc UID, IP Family & port/targetPort
