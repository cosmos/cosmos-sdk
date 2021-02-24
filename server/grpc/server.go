package grpc

import (
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/types"
)

// StartGRPCServer starts a gRPC server on the given address.
func StartGRPCServer(clientCtx client.Context, app types.Application, address string) (*grpc.Server, error) {
	grpcSrv := grpc.NewServer()
	app.RegisterGRPCServer(clientCtx, grpcSrv)

	// Reflection allows external clients to see what services and methods
	// the gRPC server exposes.
	reflection.Register(grpcSrv)

	listener, err := net.Listen("tcp", address)
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
