package translations

import (
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"google.golang.org/protobuf/types/known/durationpb"
	v1 "k8s.io/api/core/v1"
)

// MapServicePortToClusters creates a cluster for each service and ipfamily
func (km *KubeMapper) MapServicePortToClusters(svc *v1.Service, port *v1.ServicePort) []types.Resource {
	buf := []types.Resource{}

	portSettings := getPortSetings(svc, port)

	// map both IPv4&IPv6
	for _, ipFamily := range svc.Spec.IPFamilies {
		// Port
		clusterName := getClusterName(svc.Namespace, svc.Name, string(ipFamily), port.TargetPort.IntVal)

		cluster := &cluster.Cluster{
			Name:           clusterName,
			ConnectTimeout: durationpb.New(portSettings.UpstreamConnectTimeout),

			ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_EDS},
			EdsClusterConfig: &cluster.Cluster_EdsClusterConfig{
				EdsConfig: &core.ConfigSource{
					ResourceApiVersion:    core.ApiVersion_V3,
					ConfigSourceSpecifier: &core.ConfigSource_Ads{},
				},
			},
			LbPolicy: cluster.Cluster_LEAST_REQUEST, // FIXME make configureable
		}

		// HTTP cluster specific settings
		// FIXME implement sane defaults coming from https://www.envoyproxy.io/docs/envoy/v1.23.0/configuration/best_practices/edge
		/*
			if protocol == PROTOCOL_HTTP {
				cluster.TypedExtensionProtocolOptions = map[string]*anypb.Any{
					// FIXME find correct wrapper
					InitialStreamWindowSize:     wrapperspb.UInt32(65536),   //64KiB
					InitialConnectionWindowSize: wrapperspb.UInt32(1048576), // 1MiB

				}
			}
		*/
		buf = append(buf, cluster)
	}

	return buf
}
