package snapshot

import (
	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
)

type Service struct {
	Obj            *v1.Service
	EndpointSlices []*discoveryv1.EndpointSlice
}

type Snapshot struct {
	Services []*Service
}
