package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/xvzf/lightpath/internal/version"
	"k8s.io/klog/v2"
)

var (
	kubeconfig   string
	redirectMode string
)

func init() {
	// Kubeconfig
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Kubeconfig")

	// Redirect mode; right now only iptables supported, in the future also eBPF
	flag.StringVar(&redirectMode, "mode", "iptables", "Redirect mode (right now, only `iptables` is supported)")

	// pass logging flags from klog
	klog.InitFlags(nil)
}

func main() {
	flag.Parse()

	// Configure log level
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
