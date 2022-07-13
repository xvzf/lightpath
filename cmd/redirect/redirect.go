package main

import (
	"context"
	"fmt"
	"time"

	"github.com/xvzf/lightpath/internal/utils"
	"github.com/xvzf/lightpath/pkg/redirect"
	"github.com/xvzf/lightpath/pkg/wellknown"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/informers"
	"k8s.io/klog/v2"
	proxyapis "k8s.io/kubernetes/pkg/proxy/apis"
)

const (
	RESYNC_INTERVAL = 4 * time.Second
)

func run(parentCtx context.Context) error {
	ctx, cancelCtx := context.WithCancel(parentCtx)
	defer cancelCtx()

	// Create clients
	client, _, err := utils.CreateClients(kubeconfig)
	if err != nil {
		return err
	}

	// We're implementing an alternative proxy for lightpath -> let's
	lightPathProxy, err := labels.NewRequirement(proxyapis.LabelServiceProxyName, selection.Equals, []string{wellknown.LightpathProxyName})
	if err != nil {
		return err
	}
	// We're also not interested in headless services
	noHeadlessEndpoints, err := labels.NewRequirement(corev1.IsHeadlessService, selection.DoesNotExist, nil)
	if err != nil {
		return err
	}

	// Construct informer factory
	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(*lightPathProxy, *noHeadlessEndpoints)
	informerFactory := informers.NewSharedInformerFactoryWithOptions(
		client,
		RESYNC_INTERVAL,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = labelSelector.String()
		}),
	)

	// Construct Redirect service
	var redirectImpl redirect.Redirect
	switch redirectMode {
	case wellknown.LightpathRedirectIptables:
		redirectImpl = redirect.NewIptablesRedirect(wellknown.LightpathEnvoyPort)
	default:
		return fmt.Errorf("invalid redirect mode: %s", redirectMode)
	}

	// Attach service informer to redirect handler
	svcInformer := informerFactory.Core().V1().Services()

	svcInformer.Informer().AddEventHandler(redirect.NewEventHandler(redirectImpl))

	// Gogogo
	redirectImpl.Prereqs()
	informerFactory.Start(ctx.Done())

	// Wait for cache sync
	for informerType, ok := range informerFactory.WaitForCacheSync(ctx.Done()) {
		if !ok {
			cancelCtx()
			return fmt.Errorf("failed to sync cache for %v", informerType)
		}
		klog.Info("Synced cache for ", informerType)
	}

	// Wait for context to be done & cleanup
	<-ctx.Done()
	return redirectImpl.Cleanup()
}
