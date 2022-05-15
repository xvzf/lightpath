package configprovider

import (
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/xvzf/lightpath/pkg/state/snapshot"
	"k8s.io/klog/v2"
)

const (
	DEFAULT_CONNECT_TIMEOUT = 5 * time.Second
)

func EnvoySnapshotFromKubeSnapshot(snap *snapshot.Snapshot) (*cache.Snapshot, error) {
	// Vars
	clusters := []*cluster.Cluster{}

	for _, svc := range snap.Services {
		// Derive Clusters (& Endpoints)
		svcClusters, err := getClustersFromServiceSnapshot(svc)
		if err != nil {
			klog.ErrorS(err, "failed to derive clusters, continuing")
			continue
		}
		clusters = append(clusters, svcClusters...)
	}

	return nil, nil // FIXME
}
