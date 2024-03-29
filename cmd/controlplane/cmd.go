package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/xvzf/lightpath/internal/version"
	"github.com/xvzf/lightpath/pkg/logger"
	"k8s.io/klog/v2"
)

var (
	l          logger.Logger
	metricPort uint
	nodeID     string

	host       string
	kubeconfig string
)

func init() {
	// Metric endpoint
	flag.UintVar(&metricPort, "metric-port", 9000, "metric endpoint")
	// Envoy controlplane node-id
	flag.StringVar(&nodeID, "nodeID", "k8s", "Node ID (kubernetes node name)")
	flag.StringVar(&host, "host", "127.0.0.1:18000", "host for xDS server")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Kubeconfig")

	// pass logging flags from klog
	klog.InitFlags(nil)
}

func main() {
	flag.Parse()

	// Configure log level
	l = logger.New("lightpath")
	defer klog.Flush()

	// Print version string
	klog.Info(version.GetVersion())

	ctx, cancel := context.WithCancel(context.Background())

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
	}()

	// Run main entrypoint
	err := run(ctx)
	if err != nil {
		klog.Error(err, "exited with error")
	}
}
