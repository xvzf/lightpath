package translations

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xvzf/lightpath/pkg/wellknown"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
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

func getPortSettingsAnnotationsKey(svc *v1.Service, port *v1.ServicePort, setting string) string {
	return fmt.Sprintf("%s%s-%s", wellknown.PortConfigAnnotationPrefix, port.Name, setting)
}

func getUint32Config(svc *v1.Service, port *v1.ServicePort, setting string, defaultValue uint32) uint32 {
	key := getPortSettingsAnnotationsKey(svc, port, setting)

	if orig, ok := svc.Annotations[key]; ok {
		if res, err := strconv.ParseInt(orig, 10, 32); err != nil {
			return uint32(res)
		} else {
			klog.Warningf("Invalid %s=%s on service %s/%s", key, orig, svc.Namespace, svc.Name)
		}
	}
	return defaultValue
}

func getDurationConfig(svc *v1.Service, port *v1.ServicePort, setting string, defaultValue time.Duration) time.Duration {
	res := getUint32Config(svc, port, setting, uint32(defaultValue.Seconds()))
	return time.Duration(res) * time.Second
}

func getBoolConfig(svc *v1.Service, port *v1.ServicePort, setting string, defaultValue bool) bool {
	key := getPortSettingsAnnotationsKey(svc, port, setting)

	if orig, ok := svc.Annotations[key]; ok {
		if res, err := strconv.ParseBool(orig); err != nil {
			return res
		} else {
			klog.Warningf("Invalid for %s on service %s/%s", key, svc.Namespace, svc.Name)
		}
	}
	return defaultValue
}

func getStringConfig(svc *v1.Service, port *v1.ServicePort, setting string, defaultValue string) string {
	key := getPortSettingsAnnotationsKey(svc, port, setting)

	if orig, ok := svc.Annotations[key]; ok {
		return orig
	}
	return defaultValue
}
