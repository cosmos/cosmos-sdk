package grpc

import (
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/server/types"
)

var _ types.Service = &Service{}

// Service defines the gRPC service used on the Cosmos SDK.
type Service struct {
	// standard gRPC server
	grpcServer *grpc.Server
	// listen address
	address string
}

// NewService creates a new gRPC server instance with a defined listener address.
func NewService(address string) *Service {
	return &Service{
		grpcServer: grpc.NewServer(),
		address:    address,
	}
}

// Name returns the gRPC service name
func (Service) Name() string {
	return "gRPC"
}

// RegisterRoutes registers the gRPC server to the application.
func (s *Service) RegisterRoutes() bool {
	// TODO: register on app
	// app.RegisterGRPCServer(s.grpcServer)

	// Reflection allows external clients to see what services and methods
	// the gRPC server exposes.
	reflection.Register(s.grpcServer)
	return true
}

// Start starts the gRPC server on the registered address
func (s *Service) Start(cfg config.ServerConfig) error {
	if !cfg.GetSDKConfig().GRPC.Enable {
		return nil
	}

	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	errCh := make(chan error)
	go func() {
		err = s.grpcServer.Serve(listener)
		if err != nil {
			errCh <- fmt.Errorf("failed to serve: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-time.After(5 * time.Second): // assume server started successfully
		return nil
	}
}

// Stop stops the gRPC service gracefully. It stops the server from accepting new connections and
// RPCs and blocks until all the pending RPCs are finished.
func (s *Service) Stop() error {
	s.grpcServer.GracefulStop()
	return nil
}
