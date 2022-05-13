package state

import (
	"github.com/go-logr/logr"
	"github.com/xvzf/lightpath/pkg/logger"
	"github.com/xvzf/lightpath/pkg/state/snapshot"
	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	proxyconfig "k8s.io/kubernetes/pkg/proxy/config"
)

type ServiceStateSubscriber interface {
	proxyconfig.EndpointSliceHandler
	proxyconfig.ServiceHandler
}

type ServiceStateSubscriberOpts struct {
	Log logger.Logger
}

type serviceStateSubscriber struct {
	log logr.Logger

	clusterState *clusterServiceStateCache
}

func New(opts ServiceStateSubscriberOpts) ServiceStateSubscriber {
	return &serviceStateSubscriber{
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

func (cs *serviceStateSubscriber) Snapshot() *snapshot.Snapshot {
	return nil
}
