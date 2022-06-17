package main

import (
	"context"
	"sync"
	"time"

	"github.com/xvzf/lightpath/pkg/server"
	"github.com/xvzf/lightpath/pkg/state"
	"github.com/xvzf/lightpath/pkg/translations"
	"github.com/xvzf/lightpath/pkg/wellknown"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/informers"
	"k8s.io/klog/v2"
	proxyapis "k8s.io/kubernetes/pkg/proxy/apis"
	proxyconfig "k8s.io/kubernetes/pkg/proxy/config"
)

const (
	RESYNC_INTERVAL           = 4 * time.Second
	NEW_SNAPSHOT_WAIT_TIMEOUT = 10 * time.Second

	GRPC_KEEPALIVE_TIME         = 10 * time.Second
	GRPC_KEEPALIVE_TIMEOUT      = 5 * time.Second
	GRPC_KEEPALIVE_MIN_TIME     = 5 * time.Second
	GRPC_MAX_CONCURRENT_STREAMS = 1000000
)

func updateSnapshot(ctx context.Context, server server.XdsServer, subscriber state.ServiceStateSubscriber) error {
	// Update snapshots on new events
	recvCtx, recvCtxCancel := context.WithTimeout(ctx, NEW_SNAPSHOT_WAIT_TIMEOUT)
	defer recvCtxCancel()

	if snap := subscriber.ReceiveEvent(recvCtx); snap != nil {
		mapper := translations.KubeMapper{}
		envoySnap, err := mapper.EnvoySnapshotFromKubeSnapshot(snap)
		if err != nil {
			return err
		}
		// Snapshot extraction successful, try to ingest
		err = server.UpdateSnapshot(ctx, envoySnap)
		if err != nil {
			return err
		}
		klog.InfoS("Updated envoy snapshot")
	} else {
		// No snapshot received, continuing
		klog.Info("No update required to envoy snapshot")
	}

	return nil
}

func run(parentCtx context.Context) error {
	ctx, cancelCtx := context.WithCancel(parentCtx)
	defer cancelCtx()

	// Create clients
	client, _, err := createClients(kubeconfig)
	if err != nil {
		return err
	}

	// Create xDS server
	xdsServer, err := server.New(server.XdsServerOpts{
		// GRPC Server settings
		GrpcKeepaliveTime:        GRPC_KEEPALIVE_TIME,
		GrpcKeepaliveTimeout:     GRPC_KEEPALIVE_TIMEOUT,
		GrpcKeepaliveMinTime:     GRPC_KEEPALIVE_MIN_TIME,
		GrpcMaxConcurrentStreams: GRPC_MAX_CONCURRENT_STREAMS,

		Logger: l.WithValues("component", "xds-server"),
		NodeID: nodeID,
		Host:   host,
	})
	if err != nil {
		return err
	}

	// We're implementing an alternative proxy for lightpath -> let's
	lightPathProxy, err := labels.NewRequirement(proxyapis.LabelServiceProxyName, selection.Equals, []string{wellknown.LightpathProxyName})
	if err != nil {
		return err
	}
	// We're also not interested in headless services
	noHeadlessEndpoints, err := labels.NewRequirement(v1.IsHeadlessService, selection.DoesNotExist, nil)
	if err != nil {
		return err
	}

	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(*lightPathProxy, *noHeadlessEndpoints)

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
		Log: l.WithValues("component", "state-subscriber"),
	})

	serviceConfig.RegisterEventHandler(stateSubscriber)
	endpointSliceConfig.RegisterEventHandler(stateSubscriber)

	// Waitgroup for asunc tasks
	var wg sync.WaitGroup

	// Start informer
	wg.Add(1)
	go func() {
		defer wg.Done()
		informerFactory.Start(ctx.Done())
		<-ctx.Done()
	}()

	// Start GRPC Server
	wg.Add(1)
	go func() {
		defer wg.Done()
		klog.Info("Starting xDS Server")
		err := xdsServer.Start(ctx)
		if err != nil {
			klog.Error(err, "xDS Server failure")
			cancelCtx()
		}
		klog.Info("GRPC Server exited")
	}()

	// Listen until we have
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				// Context is done -> don't enter another iteration
				return
			default:
				err := updateSnapshot(ctx, xdsServer, stateSubscriber)
				if err != nil {
					klog.ErrorS(err, "Failed to update snapshot")
				}
			}
		}
	}()

	wg.Wait()
	return nil
}
