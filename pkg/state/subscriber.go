package state

import (
	"context"
	"sync"

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

	ReceiveEvent(ctx context.Context) *snapshot.Snapshot
	Snapshot() *snapshot.Snapshot
}

type ServiceStateSubscriberOpts struct {
	Log logger.Logger
}

type serviceStateSubscriber struct {
	log logr.Logger

	// Periodic even tchannel
	eventLostLock sync.Mutex
	eventLost     bool
	eventChan     chan struct{}

	clusterState *clusterServiceStateCache
}

func New(opts ServiceStateSubscriberOpts) *serviceStateSubscriber {
	return &serviceStateSubscriber{
		log: opts.Log.GetLogger(),

		eventChan: make(chan struct{}),
		eventLost: false,
		// Init datastructures
		clusterState: NewClusterServiceStateCache(ClusterServiceStateOpts{
			Log: opts.Log.GetLogger().WithValues("sub-component", "cluster-state"),
		}),
	}
}

// ReceiveEvent blocks until an event occurs; this is an atomic operation.
func (cs *serviceStateSubscriber) ReceiveEvent(ctx context.Context) *snapshot.Snapshot {
	cs.eventLostLock.Lock()
	defer cs.eventLostLock.Unlock()
	if cs.eventLost {
		cs.eventLost = false
		return cs.Snapshot()
	}

	// Block until any of the events occur or context deadline exceeds
	select {
	case <-ctx.Done():
		return nil
	case <-cs.eventChan:
		return cs.Snapshot()
	}
}

func (cs *serviceStateSubscriber) dispatchEvent() {
	// Try to dispatch event to eventChan; if not working, mark a stale lost event
	// this effectively is a "fire and forget" so we're deduplicating lost events
	select {
	case cs.eventChan <- struct{}{}:
	default:
		cs.eventLostLock.Lock()
		defer cs.eventLostLock.Unlock()
		cs.eventLost = true
	}
}

func (cs *serviceStateSubscriber) OnServiceAdd(svc *v1.Service) {
	if cs.clusterState.UpdateServices(svc, false) {
		cs.dispatchEvent()
	}
}

func (cs *serviceStateSubscriber) OnServiceUpdate(_, svc *v1.Service) {
	if cs.clusterState.UpdateServices(svc, false) {
		cs.dispatchEvent()
	}
}

func (cs *serviceStateSubscriber) OnServiceDelete(svc *v1.Service) {
	if cs.clusterState.UpdateServices(svc, true) {
		cs.dispatchEvent()
	}
}

func (cs *serviceStateSubscriber) OnServiceSynced() {} // noop

func (cs *serviceStateSubscriber) OnEndpointSliceAdd(es *discoveryv1.EndpointSlice) {
	if cs.clusterState.UpdatEndpointSlice(es, false) {
		cs.dispatchEvent()
	}
}

func (cs *serviceStateSubscriber) OnEndpointSliceUpdate(_, es *discoveryv1.EndpointSlice) {
	if cs.clusterState.UpdatEndpointSlice(es, false) {
		cs.dispatchEvent()
	}
}

func (cs *serviceStateSubscriber) OnEndpointSliceDelete(es *discoveryv1.EndpointSlice) {
	if cs.clusterState.UpdatEndpointSlice(es, true) {
		cs.dispatchEvent()
	}
}

func (cs *serviceStateSubscriber) OnEndpointSlicesSynced() {} // noop

func (cs *serviceStateSubscriber) Snapshot() *snapshot.Snapshot {
	return cs.clusterState.Snapshot()
}
