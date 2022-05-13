package state

import (
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/go-logr/logr"
	"github.com/xvzf/lightpath/pkg/logger"
	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	proxyconfig "k8s.io/kubernetes/pkg/proxy/config"
)

type ServiceStateSubscriber interface {
	proxyconfig.EndpointSliceHandler
	proxyconfig.ServiceHandler

	// Sync immediately syncs the proxy state
	Sync()

	// SyncLoop() runs continously
	SyncLoop()
}

type ServiceStateSubscriberOpts struct {
	Log logger.Logger
}

type serviceStateSubscriber struct {
	snapshotCache cache.SnapshotCache
	log           logr.Logger

	clusterState *clusterServiceStateCache
}

func New(opts ServiceStateSubscriberOpts) ServiceStateSubscriber {
	return &serviceStateSubscriber{
		snapshotCache: cache.NewSnapshotCache(
			true, // Enable Aggregate discovery service -> xDS combined over one GRPC stream
			cache.IDHash{},
			opts.Log.WithValues("sub-component", "snapshotcache"),
		),
		log: opts.Log.GetLogger(),

		// Init datastructures
		clusterState: NewClusterServiceStateCache(ClusterServiceStateOpts{
			Log: opts.Log.GetLogger().WithValues("sub-component", "cluster-state"),
		}),
	}
}

func (cs *serviceStateSubscriber) OnServiceAdd(new *v1.Service) {
	cs.clusterState.UpdateServices(new, false)
}

func (cs *serviceStateSubscriber) OnServiceUpdate(_, new *v1.Service) {
	cs.clusterState.UpdateServices(new, false)
}

func (cs *serviceStateSubscriber) OnServiceDelete(remove *v1.Service) {
	cs.clusterState.UpdateServices(remove, true)
}

func (cs *serviceStateSubscriber) OnServiceSynced() {} // noop

func (cs *serviceStateSubscriber) OnEndpointSliceAdd(new *discoveryv1.EndpointSlice) {
	cs.clusterState.UpdatEndpointSlice(new, false)
}

func (cs *serviceStateSubscriber) OnEndpointSliceUpdate(_, new *discoveryv1.EndpointSlice) {
	cs.clusterState.UpdatEndpointSlice(new, false)
}

func (cs *serviceStateSubscriber) OnEndpointSliceDelete(remove *discoveryv1.EndpointSlice) {
	cs.clusterState.UpdatEndpointSlice(remove, true)
}

func (cs *serviceStateSubscriber) OnEndpointSlicesSynced() {} // noop

func (cs *serviceStateSubscriber) Sync() {}

func (cs *serviceStateSubscriber) SyncLoop() {}
