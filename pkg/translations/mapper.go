package translations

import (
	"time"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/xvzf/lightpath/pkg/state/snapshot"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// EnvoySnaphsotFromKubeSnapshot generates an Envoy configuration snapshot servable to xDS APIs from Kubernetes Service/EndpointSlices.
func (km *KubeMapper) EnvoySnapshotFromKubeSnapshot(snap *snapshot.Snapshot) (*cache.Snapshot, error) {

	clusters := []types.Resource{}
	listeners := []types.Resource{}
	endpoints := []types.Resource{}

	for _, svc := range snap.Services {
		for _, port := range svc.Obj.Spec.Ports {
			if port.Protocol != v1.ProtocolTCP {
				klog.Warning("Protocol %s not supported, skipping", port.Protocol)
				continue
			}
			clusters = append(clusters, km.MapServicePortToClusters(svc.Obj, &port)...)
			listeners = append(listeners, km.MapServicePortToListeners(svc.Obj, &port)...)
		}
		// Map endpounts
		endpoints = append(endpoints, km.MapEndpointSliceToLocalityEndpoints(svc.Obj, svc.EndpointSlices)...)
	}

	envoySnap, err := cache.NewSnapshot(time.Now().Format(time.RFC3339Nano),
		map[resource.Type][]types.Resource{
			resource.ListenerType: listeners,
			resource.ClusterType:  clusters,
			resource.EndpointType: endpoints,
		},
	)

	return envoySnap, err
}
