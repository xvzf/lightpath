package main

import (
	"context"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	proxyapis "k8s.io/kubernetes/pkg/proxy/apis"
	proxyconfig "k8s.io/kubernetes/pkg/proxy/config"
)

func createClients(kubeconfig string) (clientset.Interface, corev1.EventsGetter, error) {
	var kubeConfig *rest.Config
	var err error

	if len(kubeconfig) == 0 {
		klog.InfoS("Neither kubeconfig file nor master URL was specified, falling back to in-cluster config")
		kubeConfig, err = rest.InClusterConfig()
	} else {
		kubeConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
			nil,
		).ClientConfig()
	}
	if err != nil {
		return nil, nil, err
	}

	// FIXME needs to be tuned
	/*
		kubeConfig.AcceptContentTypes = config.AcceptContentTypes
		kubeConfig.ContentType = config.ContentType
		kubeConfig.QPS = config.QPS
		kubeConfig.Burst = int(config.Burst)
	*/

	client, err := clientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil, nil, err
	}

	eventClient, err := clientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil, nil, err
	}

	return client, eventClient.CoreV1(), nil
}

type Proxier interface {
	Run(ctx context.Context) error
}

type Proxy struct {
	Client           clientset.Interface
	EventClient      corev1.EventsGetter
	ConfigSyncPeriod time.Duration
}

// New creates a new proxier.
func New() (Proxier, error) {
	client, eventClient, err := createClients("/Users/xvzf/.kube/config") // FIXME make configureable
	if err != nil {
		return nil, err
	}
	return &Proxy{
		Client:           client,
		EventClient:      eventClient,
		ConfigSyncPeriod: 5 * time.Second, // FIXME make configureable
	}, nil
}

// Run runs the control loop.
func (p *Proxy) Run(ctx context.Context) error {

	var errCh chan error
	errCh = make(chan error)

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
	informerFactory := informers.NewSharedInformerFactoryWithOptions(p.Client, p.ConfigSyncPeriod,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = labelSelector.String()
		}))

	// Subscribe to updates on services & endpoints
	serviceConfig := proxyconfig.NewServiceConfig(informerFactory.Core().V1().Services(), p.ConfigSyncPeriod)
	// endpointSliceConfig := proxyconfig.NewEndpointSliceConfig(informerFactory.Discovery().V1().EndpointSlices(), p.ConfigSyncPeriod)

	serviceConfig.RegisterEventHandler(&ConfigServerImpl{})

	// Create configs (i.e. Watches for Services and Endpoints or EndpointSlices)
	// Note: RegisterHandler() calls need to happen before creation of Sources because sources
	// only notify on changes, and the initial update (on process start) may be lost if no handlers
	// are registered yet.
	/*
		serviceConfig := config.NewServiceConfig(informerFactory.Core().V1().Services(), s.ConfigSyncPeriod)
		serviceConfig.RegisterEventHandler(s.Proxier)
		go serviceConfig.Run(wait.NeverStop)

		if endpointsHandler, ok := s.Proxier.(config.EndpointsHandler); ok && !s.UseEndpointSlices {
			endpointsConfig := config.NewEndpointsConfig(informerFactory.Core().V1().Endpoints(), s.ConfigSyncPeriod)
			endpointsConfig.RegisterEventHandler(endpointsHandler)
			go endpointsConfig.Run(wait.NeverStop)
		} else {
			endpointSliceConfig := config.NewEndpointSliceConfig(informerFactory.Discovery().V1().EndpointSlices(), s.ConfigSyncPeriod)
			endpointSliceConfig.RegisterEventHandler(s.Proxier)
			go endpointSliceConfig.Run(wait.NeverStop)
		}
	*/

	// This has to start after the calls to NewServiceConfig and NewEndpointsConfig because those
	// functions must configure their shared informer event handlers first.
	informerFactory.Start(wait.NeverStop)

	return <-errCh
}

func main() {
	cp, err := New()
	if err != nil {
		panic(err)
	}
	cp.Run(context.Background())
}
