package main

import (
	"context"
	"flag"
	"os"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/envoyproxy/go-control-plane/pkg/test/v3"
	"github.com/xvzf/lightpath/internal/version"
	"github.com/xvzf/lightpath/pkg/logger"
	"github.com/xvzf/lightpath/pkg/resource"
	"github.com/xvzf/lightpath/pkg/server"
	"k8s.io/klog/v2"
)

var (
	l      logger.Logger
	port   uint
	nodeID string
)

func init() {
	// The port that this xDS server listens on
	flag.UintVar(&port, "port", 18000, "xDS management server port")

	// Tell Envoy to use this Node ID
	flag.StringVar(&nodeID, "nodeID", "test-id", "Node ID")

	// Configure logging flags
	klog.InitFlags(nil)
	flag.Set("v", "3")
}

func main() {
	flag.Parse()

	// Configure log level
	l = logger.New("lightpath")
	defer klog.Flush()

	// Print version string
	l.Infof(version.GetVersion())

	// Create a cache
	cache := cache.NewSnapshotCache(false, cache.IDHash{}, l)

	// Create the snapshot that we'll serve to Envoy
	snapshot := resource.GenerateSnapshot()
	if err := snapshot.Consistent(); err != nil {
		l.Errorf("snapshot inconsistency: %+v\n%+v", snapshot, err)
		os.Exit(1)
	}
	l.Debugf("will serve snapshot %+v", snapshot)

	// Add the snapshot to the cache
	if err := cache.SetSnapshot(context.Background(), nodeID, snapshot); err != nil {
		l.Errorf("snapshot error %q for %+v", err, snapshot)
		os.Exit(1)
	}

	// Run the xDS server
	ctx := context.Background()
	cb := &test.Callbacks{Debug: true}
	srv := envoy_server.NewServer(ctx, cache, cb)
	server.RunServer(ctx, srv, port)
}
