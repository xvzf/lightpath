package redirect

import (
	"fmt"
	"net"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

func getComment(svc *corev1.Service) string {
	return fmt.Sprintf("lightpath:%s/%s", svc.Namespace, svc.Name)
}

// svcDiff compares two services and returns a list of IP addresses to add and remove
func svcIPDiff(old, new *corev1.Service) ([]net.IP, []net.IP) {
	add := []net.IP{}
	remove := []net.IP{}

	oldIPs := map[string]interface{}{}
	newIPs := map[string]interface{}{}

	// Build a map of existing IPs
	for _, ip := range append(old.Spec.ClusterIPs, old.Spec.ExternalIPs...) {
		oldIPs[ip] = struct{}{}
	}

	// Calculate diff
	for _, ip := range append(new.Spec.ClusterIPs, new.Spec.ExternalIPs...) {
		newIPs[ip] = struct{}{}
		// Update IPs to add
		if _, ok := oldIPs[ip]; !ok {
			add = append(add, net.ParseIP(ip))
		}
	}
	for ip := range oldIPs {
		// Update IPs to remove
		if _, ok := newIPs[ip]; !ok {
			remove = append(remove, net.ParseIP(ip))
		}
	}

	return add, remove
}

// NewEventHandler wraps arount a redirect interfaces allowing it to be consumed by a client-go informer
func NewEventHandler(impl Redirect) cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		// AddFunc (-> when a service is added)
		AddFunc: func(obj interface{}) {
			svc := obj.(*corev1.Service)
			comment := getComment(svc)
			for _, ip := range append(svc.Spec.ClusterIPs, svc.Spec.ExternalIPs...) {
				err := impl.AddIP(comment, net.ParseIP(ip))
				if err != nil {
					klog.ErrorS(err, "failed to add IP", "ip", ip)
				} else {
					klog.V(3).InfoS("added", "ip", ip)
				}
			}
		},

		// UpdateFunc (-> when a service is updated)
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldSvc := oldObj.(*corev1.Service)
			newSvc := newObj.(*corev1.Service)
			comment := getComment(newSvc)

			add, remove := svcIPDiff(oldSvc, newSvc)

			for _, ip := range add {
				err := impl.AddIP(comment, ip)
				if err != nil {
					klog.ErrorS(err, "failed to add IP", "ip", fmt.Sprint(ip))
				} else {
					klog.V(3).InfoS("added", "ip", ip)
				}
			}

			for _, ip := range remove {
				err := impl.RemoveIP(comment, ip)
				if err != nil {
					klog.ErrorS(err, "failed to remove IP", "ip", fmt.Sprint(ip))
				} else {
					klog.V(3).InfoS("removed", "ip", ip)
				}
			}
		},

		// DeleteFunc (-> when a service is deleted)
		DeleteFunc: func(obj interface{}) {
			svc := obj.(*corev1.Service)
			comment := getComment(svc)
			for _, ip := range append(svc.Spec.ClusterIPs, svc.Spec.ExternalIPs...) {
				err := impl.RemoveIP(comment, net.ParseIP(ip))
				if err != nil {
					klog.ErrorS(err, "failed to remove IP", "ip", ip)
				} else {
					klog.V(3).InfoS("removed", "ip", ip)
				}
			}
		},
	}
}
