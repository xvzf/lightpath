package state

import (
	"errors"
	"sync"

	"github.com/go-logr/logr"
	"github.com/xvzf/lightpath/pkg/state/snapshot"
	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
)

type ClusterServiceStateOpts struct {
	Log logr.Logger
}

type clusterServiceStateCache struct {
	m        sync.Mutex
	opts     ClusterServiceStateOpts
	log      logr.Logger
	services map[string]*serviceContainer
}

func NewClusterServiceStateCache(opts ClusterServiceStateOpts) *clusterServiceStateCache {
	return &clusterServiceStateCache{
		opts:     opts,
		log:      opts.Log,
		services: make(map[string]*serviceContainer),
	}
}

// UpdateServices handles service update events origining form a Kubernetes informer.
func (c *clusterServiceStateCache) UpdateServices(service *v1.Service, delete bool) {
	c.m.Lock()
	defer c.m.Unlock()

	if service == nil {
		c.log.V(0).Error(errors.New("nullptr"), "Requested update operation with invalid reference")
		return
	}

	if delete {
		c.deleteService(service)
		return
	}

	c.updateService(service)
}

// updateService handles addition or in-place updates of new services.
func (c *clusterServiceStateCache) updateService(new *v1.Service) {
	c.log.V(5).Info("Updated service call started", "name", new.Name)
	existing, ok := c.services[new.Name]

	// Check if we have to update
	if ok && existing.obj.ResourceVersion >= new.ResourceVersion {
		c.log.V(4).Info("Service ResourceVersion did not change",
			"name", new.Name,
			"new_resource_version", new.ResourceVersion,
			"existing_resource_version", existing.obj.ResourceVersion,
		)
		return
	}

	if ok {
		// Update an existing service in-place
		err := existing.Update(new)
		if err != nil {
			c.log.V(0).Error(err, "Service updated failed")
			return
		}
		c.log.V(2).Info("Updated service", "name", new.Name)
	} else {
		// Create a new service container for the received service
		c.services[new.Name] = NewServiceContainer(new)
		c.log.V(2).Info("Added service", "name", new.Name)
	}
	c.log.V(5).Info("Updated service call completed", "name", new.Name)
}

// deleteService handles deletion of an existing service
func (c *clusterServiceStateCache) deleteService(remove *v1.Service) {
	c.log.V(5).Info("Delete service call started", "name", remove.Name)
	if _, ok := c.services[remove.Name]; ok {
		delete(c.services, remove.Name)
		c.log.V(2).Info("Deleted service", "name", remove.Name)
	}
	c.log.V(1).Info("Deleting service failed (not in-memory)", remove.Name)
	c.log.V(5).Info("Delete service call completed", "name", remove.Name)
}

func (c *clusterServiceStateCache) UpdatEndpointSlice(endpointslice *discoveryv1.EndpointSlice, delete bool) {
	if endpointslice == nil {
		c.log.V(0).Error(errors.New("nullptr"), "Requested update operation with invalid reference")
		return
	}
	svcName, ok := endpointslice.Labels[discoveryv1.LabelServiceName]
	if !ok {
		c.log.V(1).Info("Endpointslice has no service label attached", "name", endpointslice.Name)
		return
	}
	c.m.Lock()
	svc, ok := c.services[svcName]
	c.m.Unlock()

	if ok {
		res := svc.UpdatEndpointslices(endpointslice, delete)
		if res > ENDPOINTSLICE_OPERATION_STATUS_DELETED {
			c.log.V(1).Info(res.String(), "name", endpointslice.Name)
		} else if res > ENDPOINTSLICE_OPERATION_STATUS_NOOP {
			c.log.V(2).Info(res.String(), "name", endpointslice.Name)
		}
	} else {
		c.log.V(1).Info("Received EndpointSlice for non-registered service")
	}
}

func (c *clusterServiceStateCache) DeepCopy() *clusterServiceStateCache {
	c.m.Lock()
	defer c.m.Unlock()
	newCopy := &clusterServiceStateCache{
		opts:     c.opts,
		log:      c.log,
		services: make(map[string]*serviceContainer),
	}
	for k, v := range c.services {
		newCopy.services[k] = v.DeepCopy()
	}
	return newCopy
}

func (orig *clusterServiceStateCache) Snapshot() *snapshot.Snapshot {
	// Deep copy and don't block -> updates can go in again
	orig.m.Lock()
	c := orig.DeepCopy()
	orig.m.Unlock()

	snap := &snapshot.Snapshot{
		Services: make([]*snapshot.Service, 0, len(c.services)),
	}

	// Computate Snapshot from k8s resources
	for _, svc := range c.services {
		snap.Services = append(snap.Services, svc.Snapshot())
	}

	return snap
}
