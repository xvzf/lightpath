package main

import (
	"encoding/json"
	"os"

	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	proxyconfig "k8s.io/kubernetes/pkg/proxy/config"
)

type ConfigServer interface {
	proxyconfig.EndpointSliceHandler
	proxyconfig.ServiceHandler

	// Sync immediately syncs the proxy state
	Sync()

	// SyncLoop() runs continously
	SyncLoop()
}

type ConfigServerImpl struct {
}

func (cs *ConfigServerImpl) OnServiceAdd(service *v1.Service) {
	json.NewEncoder(os.Stdout).Encode(service)
}

func (cs *ConfigServerImpl) OnServiceUpdate(oldservice, service *v1.Service) {
	// json.NewEncoder(os.Stdout).Encode(oldservice)
	// json.NewEncoder(os.Stdout).Encode(service)
	// json.NewEncoder(os.Stdout).Encode(oldservice)
	// change, _ := diff.Diff(oldservice, service)
	// json.NewEncoder(os.Stdout).Encode(change)
}

func (cs *ConfigServerImpl) OnServiceDelete(service *v1.Service) {
	// json.NewEncoder(os.Stdout).Encode(service)
}

func (cs *ConfigServerImpl) OnServiceSynced() {

}

func (cs *ConfigServerImpl) OnEndpointSliceAdd(service *discoveryv1.EndpointSlice) {

}

func (cs *ConfigServerImpl) OnEndpointSliceUpdate(service *discoveryv1.EndpointSlice) {

}

func (cs *ConfigServerImpl) OnEndpointSliceDelete(service *discoveryv1.EndpointSlice) {

}
func (cs *ConfigServerImpl) OnEndpointSliceSynced() {

}
