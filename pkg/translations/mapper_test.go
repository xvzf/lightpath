package translations

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"testing"

	"github.com/xvzf/lightpath/pkg/state/snapshot"
	"github.com/xvzf/lightpath/pkg/wellknown"
	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	proxyapis "k8s.io/kubernetes/pkg/proxy/apis"
	"k8s.io/utils/pointer"
)

func randIPv4() string {
	buf := make([]byte, 4)

	ip := rand.Uint32()

	binary.LittleEndian.PutUint32(buf, ip)
	return fmt.Sprintf("%s", net.IP(buf))
}

func mockKubeSnapshot(numServices, numEndpoints, numZones int, numHosts int) *snapshot.Snapshot {
	res := &snapshot.Snapshot{
		Services: make([]*snapshot.Service, 0, numServices),
	}

	endpointsPerService := numEndpoints / numServices
	endpointSlicesPerService := endpointsPerService / 1000 // max of 1000 endpoints per EndpointSlice
	// hostsPerZone := numHosts / numZones

	// Generate N services
	for svcId := 0; svcId < numServices; svcId++ {
		svcName := fmt.Sprintf("svc-%d", svcId)
		service := &snapshot.Service{
			Obj: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      svcName,
					Namespace: "bench",
					Labels: map[string]string{
						proxyapis.LabelServiceProxyName: wellknown.LightpathProxyName,
					},
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{{
						Name:        "http",
						Protocol:    v1.ProtocolTCP,
						AppProtocol: pointer.StringPtr("http"),
						TargetPort:  intstr.FromInt(8080),
					}},
				},
			},
			EndpointSlices: make([]*discoveryv1.EndpointSlice, 0, endpointSlicesPerService),
		}

		endpointId := 0
		for esId := 0; esId < endpointSlicesPerService; esId++ {
			endpointSlice := discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("endpointslice-%s-%d", svcName, esId),
					Namespace: "bench",
					Labels: map[string]string{
						discoveryv1.LabelServiceName: svcName,
					},
				},
			}
			// Add Endpoints
			for i := 0; i < 1000 && endpointId < numEndpoints; i++ {
				hostId := endpointId % numHosts
				zoneId := hostId % numZones
				endpoint := discoveryv1.Endpoint{
					Addresses: []string{randIPv4()},
					Conditions: discoveryv1.EndpointConditions{
						Ready:       pointer.BoolPtr(true),
						Serving:     pointer.BoolPtr(true),
						Terminating: pointer.BoolPtr(false),
					},
					Hostname: pointer.StringPtr(fmt.Sprintf("pod-%d", endpointId)),
					NodeName: pointer.StringPtr(fmt.Sprintf("hist-%d", hostId)),
					Zone:     pointer.StringPtr(fmt.Sprintf("zone-%d", zoneId)),
				}

				endpointSlice.Endpoints = append(endpointSlice.Endpoints, endpoint)
				endpointId++
			}

			// Add EndpointSlice
			service.EndpointSlices = append(service.EndpointSlices, &endpointSlice)
		}

		// add service
		res.Services = append(res.Services, service)
	}

	return res
}

func benchmarkTranslations(svc, pods, nodes int, b *testing.B) {
	testSnapshot := mockKubeSnapshot(svc, pods, 3, nodes)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		km := KubeMapper{}
		_, err := km.EnvoySnapshotFromKubeSnapshot(testSnapshot)
		if err != nil {
			b.Error(err)
			b.Fail()
		}
	}
}

func BenchmarkSvc1Pods1000Nodes10(b *testing.B) {
	benchmarkTranslations(1, 1000, 10, b)
}

func BenchmarkSvc10Pods1000Nodes10(b *testing.B) {
	benchmarkTranslations(10, 1000, 10, b)
}
func BenchmarkSvc100Pods1000Nodes10(b *testing.B) {
	benchmarkTranslations(100, 1000, 10, b)
}
func BenchmarkSvc100Pods10000Nodes100(b *testing.B) {
	benchmarkTranslations(100, 10000, 100, b)
}
func BenchmarkSvc1000Pods10000Nodes100(b *testing.B) {
	benchmarkTranslations(1000, 10000, 100, b)
}
func BenchmarkSvc10000Pods100000Nodes1000(b *testing.B) {
	benchmarkTranslations(10000, 100000, 1000, b)
}

func BenchmarkSvc100Pods150000Nodes5000(b *testing.B) {
	benchmarkTranslations(1000, 150000, 5000, b)
}
func BenchmarkSvc1000Pods150000Nodes5000(b *testing.B) {
	benchmarkTranslations(1000, 150000, 5000, b)
}

func BenchmarkSvc10000Pods150000Nodes5000(b *testing.B) {
	benchmarkTranslations(10000, 150000, 5000, b)
}
