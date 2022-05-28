package translations

import (
	"fmt"
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	"github.com/xvzf/lightpath/pkg/state/snapshot"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	DEFAULT_CONNECT_TIMEOUT = 5 * time.Second
)

// getClustersFromServiceSnapshot generates one envoy cluster for each service
// destination port.
func getClustersFromServiceSnapshot(svcSnap *snapshot.Service) ([]*cluster.Cluster, error) {
	mappedLocalityEndpoints, err := mapKubeEndpointSlicesToMappendLocalityEndpoints(svcSnap.EndpointSlices)
	if err != nil {
		return nil, err
	}

	res := make([]*cluster.Cluster, 0, len(mappedLocalityEndpoints))

	// Create service specific clusters
	for idx, endpoints := range mappedLocalityEndpoints {
		clusterName := fmt.Sprintf("%s:%s:%s", svcSnap.Obj.Namespace, svcSnap.Obj.Name, idx)
		res = append(res, &cluster.Cluster{
			Name:                 clusterName,
			ConnectTimeout:       durationpb.New(DEFAULT_CONNECT_TIMEOUT),             // FIXME make configurable with annotation
			ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_STATIC}, // FIXME discover from Endpoint Discovery Service
			LbPolicy:             cluster.Cluster_ROUND_ROBIN,                         // FIXME make configurable
			LoadAssignment: &endpoint.ClusterLoadAssignment{
				ClusterName: clusterName,
				Endpoints:   endpoints,
			},
		})
	}

	// return res, nil
	return []*cluster.Cluster{}, nil
}
