package grpc

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/server/grpc/gogoreflection"
	reflection "github.com/cosmos/cosmos-sdk/server/grpc/reflection/v2alpha1"
	"github.com/cosmos/cosmos-sdk/server/types"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino" // Import amino.proto file for reflection
)

// StartGRPCServer starts a gRPC server on the given address.
func StartGRPCServer(clientCtx client.Context, app types.Application, address string) (*grpc.Server, error) {
	grpcSrv := grpc.NewServer()
	app.RegisterGRPCServer(grpcSrv)
	// Respond to health checks.
	grpc_health_v1.RegisterHealthServer(grpcSrv, health.NewServer())
	// reflection allows consumers to build dynamic clients that can write
	// to any cosmos-sdk application without relying on application packages at compile time
	err := reflection.Register(grpcSrv, reflection.Config{
		SigningModes: func() map[string]int32 {
			supportedModes := clientCtx.TxConfig.SignModeHandler().SupportedModes()
			modes := make(map[string]int32, len(supportedModes))
			for _, m := range supportedModes {
				modes[m.String()] = (int32)(m)
			}

			return modes
		}(),
		ChainID:           clientCtx.ChainID,
		InterfaceRegistry: clientCtx.InterfaceRegistry,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to register reflection service: %w", err)
	}

	// Reflection allows external clients to see what services and methods
	// the gRPC server exposes.
	gogoreflection.Register(grpcSrv)

	return grpcSrv, nil
}

// StartGRPCServer starts the provided gRPC server on the address specified in cfg.
//
// Note, this creates a blocking process if the server is started successfully.
// Otherwise, an error is returned. The caller is expected to provide a Context
// that is properly canceled or closed to indicate the server should be stopped.
func StartGRPCServer(ctx context.Context, logger log.Logger, cfg config.GRPCConfig, grpcSrv *grpc.Server) error {
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return fmt.Errorf("failed to listen on address %s: %w", cfg.Address, err)
	}

	errCh := make(chan error)

	// Start the gRPC in an external goroutine as Serve is blocking and will return
	// an error upon failure, which we'll send on the error channel that will be
	// consumed by the for block below.
	go func() {
		logger.Info("starting gRPC server...", "address", cfg.Address)
		errCh <- grpcSrv.Serve(listener)
	}()

	// Start a blocking select to wait for an indication to stop the server or that
	// the server failed to start properly.
	select {
	case <-ctx.Done():
		// The calling process canceled or closed the provided context, so we must
		// gracefully stop the gRPC server.
		logger.Info("stopping gRPC server...", "address", cfg.Address)
		grpcSrv.GracefulStop()

		return nil

	case err := <-errCh:
		return nil, err
	case <-time.After(types.ServerStartTime): // assume server started successfully
		return grpcSrv, nil
	}
}
