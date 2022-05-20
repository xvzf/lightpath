package state

import (
	"errors"
	"sync"

	"github.com/xvzf/lightpath/pkg/state/snapshot"
	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/klog"
)

// serviceContainers map to clusters (CDS), listeners (LDS) , routes (RDS)
// EndpointSlices/IPs map to Endpoints (EDS)

const (
	ENDPOINTSLICE_OPERATION_STATUS_NOOP = iota
	ENDPOINTSLICE_OPERATION_STATUS_UPDATED
	ENDPOINTSLICE_OPERATION_STATUS_DELETED
	ENDPOINTSLICE_OPERATION_STATUS_INVALID_SERVICE_LABEL
	ENDPOINTSLICE_OPERATION_STATUS_ERROR
)

type EndpointSliceOperationStatus int

func (eos EndpointSliceOperationStatus) String() string {
	switch eos {
	case ENDPOINTSLICE_OPERATION_STATUS_DELETED:
		return "EndpointSlice deleted"
	case ENDPOINTSLICE_OPERATION_STATUS_UPDATED:
		return "EndpointSlice updated"
	case ENDPOINTSLICE_OPERATION_STATUS_INVALID_SERVICE_LABEL:
		return "EndpointSlice update failed, invalid service label"
	case ENDPOINTSLICE_OPERATION_STATUS_ERROR:
		return "EndpointSlice update failed, invalid reference"
	default:
		return "EndpointSlice updated not required"
	}
}

type serviceOpts struct {
	// noop right now
}

// NewServiceOpts parses options coming from svc annotations.
func NewServiceOpts(svc *v1.Service) *serviceOpts {
	so := &serviceOpts{}
	so.Update(svc)
	return so
}

// Update updates the service options based on a k8s service object.
func (so *serviceOpts) Update(svc *v1.Service) {
	// noop
}

// serviceContainer maps a kubernetes serviceContainer loosely to an internal representation used for envoy
type serviceContainer struct {
	m   sync.Mutex // Allow update operations
	obj *v1.Service

	opts           *serviceOpts // serviceContainer options
	endpointSlices map[string]*discoveryv1.EndpointSlice
}

// NewServiceContainer creates a new service container
func NewServiceContainer(svc *v1.Service) *serviceContainer {
	return &serviceContainer{
		obj:            svc,                 // Pass received k8s obj
		opts:           NewServiceOpts(svc), // FIXME
		endpointSlices: make(map[string]*discoveryv1.EndpointSlice),
	}
}

// DeepCopy creates a deep copy of the serviceContainer object
func (s *serviceContainer) DeepCopy() *serviceContainer {
	s.m.Lock()
	defer s.m.Unlock()

	copy := &serviceContainer{
		obj:            s.obj.DeepCopy(),
		opts:           s.opts,
		endpointSlices: make(map[string]*discoveryv1.EndpointSlice),
	}

	for k, v := range s.endpointSlices {
		copy.endpointSlices[k] = v.DeepCopy()
	}

	return copy
}

func (s *serviceContainer) Update(new *v1.Service) error {
	if new == nil {
		return errors.New("nullptr")
	}

	s.m.Lock()
	defer s.m.Unlock()

	// Update reference & opts
	s.obj = new
	s.opts.Update(new)

	return nil
}

// UpdateEndpointslices handles updates to underlying endpointslice objects.
func (s *serviceContainer) UpdatEndpointslices(endpointslice *discoveryv1.EndpointSlice, delete bool) EndpointSliceOperationStatus {
	if endpointslice == nil {
		return ENDPOINTSLICE_OPERATION_STATUS_ERROR
	}

	s.m.Lock()
	defer s.m.Unlock()

	// Check preconditions
	if svcName, ok := endpointslice.ObjectMeta.Labels[discoveryv1.LabelServiceName]; !ok || svcName != s.obj.Name {
		return ENDPOINTSLICE_OPERATION_STATUS_INVALID_SERVICE_LABEL
	}

	// Handle delete
	if delete {
		return s.deleteEndpointSlice(endpointslice)
	}

	// Handle update
	return s.updateEndpointSlice(endpointslice)

}

// updateEndpointSlice handles insertions/updates of existing ones
func (s *serviceContainer) updateEndpointSlice(new *discoveryv1.EndpointSlice) EndpointSliceOperationStatus {
	existing, ok := s.endpointSlices[new.Name]

	// Existing and needs to be updated
	if ok && existing.ResourceVersion >= new.ResourceVersion {
		return ENDPOINTSLICE_OPERATION_STATUS_NOOP
	}

	s.endpointSlices[new.Name] = new
	return ENDPOINTSLICE_OPERATION_STATUS_UPDATED
}

// deleteEndpointSlice deletes a single endpointslice
func (s *serviceContainer) deleteEndpointSlice(remove *discoveryv1.EndpointSlice) EndpointSliceOperationStatus {
	if _, ok := s.endpointSlices[remove.Name]; ok {
		delete(s.endpointSlices, remove.Name)
		return ENDPOINTSLICE_OPERATION_STATUS_DELETED
	}

	return ENDPOINTSLICE_OPERATION_STATUS_NOOP
}

func (s *serviceContainer) Snapshot() *snapshot.Service {
	klog.V(9).Info("Aquiring lock for serviceContailer")
	s.m.Lock()
	defer klog.V(9).Info("Releasing lock for serviceContailer")
	defer s.m.Unlock()

	svc := &snapshot.Service{
		Obj:            s.obj,
		EndpointSlices: make([]*discoveryv1.EndpointSlice, 0, len(s.endpointSlices)),
	}

	for _, endpointslice := range s.endpointSlices {
		svc.EndpointSlices = append(svc.EndpointSlices, endpointslice)
	}

	return svc
}
