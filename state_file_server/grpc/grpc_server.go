package grpc

import (
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/state_file_server/config"
)

// StartGRPCServer starts a gRPC server on the given address.
func StartGRPCServer(config config.StateServerConfig) (*grpc.Server, error) {
	grpcSrv := grpc.NewServer()
	listener, err := net.Listen("tcp", config.GRPCAddress)
	if err != nil {
		return nil, err
	}

	errCh := make(chan error)
	go func() {
		err = grpcSrv.Serve(listener)
		if err != nil {
			errCh <- fmt.Errorf("failed to serve: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return nil, err
	case <-time.After(5 * time.Second): // assume server started successfully
		return grpcSrv, nil
	}
}
