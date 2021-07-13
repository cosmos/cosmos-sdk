package grpc

import (
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/server/grpc/gogoreflection"
	pb "github.com/cosmos/cosmos-sdk/store/streaming/file/server/v1beta"
)

// StartGRPCServer starts a gRPC server on the given address.
func StartGRPCServer(handler Handler, address string) (*grpc.Server, error) {
	grpcSrv := grpc.NewServer()
	pb.RegisterStateFileServer(grpcSrv, handler)
	gogoreflection.Register(grpcSrv)
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
