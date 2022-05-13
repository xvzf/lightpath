package main

import (
	clientset "k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
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
