package server

import (
	"context"
	"net"
	"time"

	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointservice "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/xvzf/lightpath/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"k8s.io/klog/v2"
)

// ServerOpts configure the GRPC server.
type XdsServerOpts struct {
	GrpcKeepaliveTime        time.Duration
	GrpcKeepaliveTimeout     time.Duration
	GrpcKeepAliveMinTime     time.Duration
	GrpcMaxConcurrentStreams uint32

	Logger logger.Logger
	NodeID string

	Host string
}

// XdsServer provides a GRPC config server based on Envoys xDS protocol.
type XdsServer interface {
	Start(context.Context) error
	UpdateSnapshot(context.Context, *cache.Snapshot) error
}

type xdsServer struct {
	nodeID string
	cache  cache.SnapshotCache
	listen net.Listener

	server *grpc.Server
}

// New creates a new XDS Server.
func New(opts XdsServerOpts) (XdsServer, error) {
	listen, err := net.Listen("tcp", opts.Host)
	if err != nil {
		return nil, err
	}
	return &xdsServer{
		// Listener
		listen: listen,
		// NodeID
		nodeID: opts.NodeID,
		// cache
		cache: cache.NewSnapshotCache(true, cache.IDHash{}, opts.Logger),
		// Construct GRPC Server
		server: grpc.NewServer(
			grpc.MaxConcurrentStreams(opts.GrpcMaxConcurrentStreams),
			grpc.KeepaliveParams(keepalive.ServerParameters{
				Time:    opts.GrpcKeepaliveTime,
				Timeout: opts.GrpcKeepaliveTimeout,
			}),
			grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
				MinTime:             opts.GrpcKeepAliveMinTime,
				PermitWithoutStream: true,
			}),
		),
	}, nil
}

func (s *xdsServer) UpdateSnapshot(ctx context.Context, snapshot *cache.Snapshot) error {
	if err := snapshot.Consistent(); err != nil {
		klog.ErrorS(err, "Inconsistent snapshot, skipping serving")
		return err
	}
	klog.Info("Updating xDS snapshot")
	return s.cache.SetSnapshot(ctx, s.nodeID, snapshot)
}

func (s *xdsServer) Start(ctx context.Context) error {
	// create xDS server based on snapshot cache
	xdsServer := xds.NewServer(ctx, s.cache, nil)
	// Register xDS endpoints
	discoverygrpc.RegisterAggregatedDiscoveryServiceServer(s.server, xdsServer)
	endpointservice.RegisterEndpointDiscoveryServiceServer(s.server, xdsServer)
	clusterservice.RegisterClusterDiscoveryServiceServer(s.server, xdsServer)
	routeservice.RegisterRouteDiscoveryServiceServer(s.server, xdsServer)
	listenerservice.RegisterListenerDiscoveryServiceServer(s.server, xdsServer)

	errChan := make(chan error)
	// Start GRPC server async
	go func() {
		klog.Info("Starting xDS GRPC Server")
		errChan <- s.server.Serve(s.listen)
	}()

	select {
	// Handle graceful termination
	case <-ctx.Done():
		klog.Info("Stopping xDS GRPC Server")
		s.server.GracefulStop()
		klog.Info("Terminated xDS GRPC Server")
		return nil
	// Handle error case
	case err := <-errChan:
		return err
	}
}
