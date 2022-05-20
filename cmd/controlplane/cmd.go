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
	port       uint
	metricPort uint
	nodeID     string

	host string
)

func init() {
	// The port that this xDS server listens on
	flag.UintVar(&port, "port", 18000, "xDS management server port")
	// Metric endpoint
	flag.UintVar(&metricPort, "metric-port", 9000, "metric endpoint")
	// Envoy controlplane node-id
	flag.StringVar(&nodeID, "nodeID", "k8s", "Node ID (kubernetes node name)")
	flag.StringVar(&host, "host", ":12000", "host for xDS server")

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
