package translations

import (
	"time"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/xvzf/lightpath/pkg/state/snapshot"
	"k8s.io/klog/v2"
)

// EnvoySnaphsotFromKubeSnapshot generates an Envoy configuration snapshot servable to xDS APIs from Kubernetes Service/EndpointSlices.
func EnvoySnapshotFromKubeSnapshot(snap *snapshot.Snapshot) (*cache.Snapshot, error) {
	// Vars
	clusters := []types.Resource{}
	listeners := []types.Resource{}

	for _, svc := range snap.Services {
		// Derive Clusters (& Endpoints)
		svcClusters, err := getClustersFromServiceSnapshot(svc)
		if err != nil {
			klog.ErrorS(err, "failed to derive clusters, continuing")
			continue
		}
		// Nested for loop otherwise struct -> interface conversion doesn't work
		for _, svcCluster := range svcClusters {
			clusters = append(clusters, svcCluster)
		}
		// Derive listeners (&filters)
		svcListeners := getListenerFromServiceSnapshpt(svc)
		// Nested for loop otherwise struct -> interface conversion doesn't work
		for _, svcListener := range svcListeners {
			listeners = append(listeners, svcListener)
		}
	}

	envoySnap, err := cache.NewSnapshot(time.Now().Format(time.RFC3339Nano),
		map[resource.Type][]types.Resource{
			resource.ListenerType: listeners,
			resource.ClusterType:  clusters, // Includes endpoints due to nested datastructure
		},
	)

	return envoySnap, err
}
