package translations

import (
	"strings"
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
		clusterName := strings.Join([]string{string(svcSnap.Obj.GetUID()), idx}, "-")
		res = append(res, &cluster.Cluster{
			Name:                 clusterName,
			ConnectTimeout:       durationpb.New(DEFAULT_CONNECT_TIMEOUT),          // FIXME make configureable with annotation
			ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_EDS}, // retrieve endpoints from EDS
			LbPolicy:             cluster.Cluster_ROUND_ROBIN,                      // FIXME make configureable
			LoadAssignment: &endpoint.ClusterLoadAssignment{
				ClusterName: clusterName,
				Endpoints:   endpoints,
			},
		})
	}

	return res, nil
}
