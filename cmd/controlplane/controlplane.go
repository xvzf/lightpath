package main

import (
	"context"
	"time"

	"github.com/xvzf/lightpath/internal/lightpath/state"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/informers"
	proxyapis "k8s.io/kubernetes/pkg/proxy/apis"
	proxyconfig "k8s.io/kubernetes/pkg/proxy/config"
)

const (
	RESYNC_INTERVAL = 4 * time.Second
)

func run(ctx context.Context) error {
	// Create clients
	client, _, err := createClients("/Users/xvzf/.kube/config") // FIXME
	if err != nil {
		return err
	}

	// We're not interested in explicitly ignored services
	noProxyName, err := labels.NewRequirement(proxyapis.LabelServiceProxyName, selection.DoesNotExist, nil)
	if err != nil {
		return err
	}
	// We're also not interested in headless services (so far, might be interesting in the future)
	noHeadlessEndpoints, err := labels.NewRequirement(v1.IsHeadlessService, selection.DoesNotExist, nil)
	if err != nil {
		return err
	}

	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(*noProxyName, *noHeadlessEndpoints)

	// pass labelselector options to informer factory
	informerFactory := informers.NewSharedInformerFactoryWithOptions(
		client,
		RESYNC_INTERVAL,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = labelSelector.String()
		}),
	)

	// Subscribe to updates on services & endpoints
	serviceConfig := proxyconfig.NewServiceConfig(informerFactory.Core().V1().Services(), RESYNC_INTERVAL)
	endpointSliceConfig := proxyconfig.NewEndpointSliceConfig(informerFactory.Discovery().V1().EndpointSlices(), RESYNC_INTERVAL)

	stateSubscriber := state.New(state.ServiceStateSubscriberOpts{
		Log: l.WithValues("component", "state-suscriber"),
	})

	serviceConfig.RegisterEventHandler(stateSubscriber)
	endpointSliceConfig.RegisterEventHandler(stateSubscriber)

	informerFactory.Start(ctx.Done())

	<-ctx.Done()
	return nil
}
