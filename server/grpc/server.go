package grpc

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/server/grpc/gogoreflection"
	reflection "github.com/cosmos/cosmos-sdk/server/grpc/reflection/v2alpha1"
	"github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino" // Import amino.proto file for reflection
)

// NewGRPCServer returns a correctly configured and initialized gRPC server.
// Note, the caller is responsible for starting the server. See StartGRPCServer.
func NewGRPCServer(clientCtx client.Context, app types.Application, cfg config.GRPCConfig) (*grpc.Server, error) {
	srv, _, err := NewGRPCServerAndContext(clientCtx, app, cfg, log.NewNopLogger())
	return srv, err
}

// NewGRPCServerAndContext returns a correctly configured and initialized gRPC server
// along with an updated client context that may include historical gRPC connections.
func NewGRPCServerAndContext(clientCtx client.Context, app types.Application, cfg config.GRPCConfig, logger log.Logger) (*grpc.Server, client.Context, error) {
	maxSendMsgSize := cfg.MaxSendMsgSize
	if maxSendMsgSize == 0 {
		maxSendMsgSize = config.DefaultGRPCMaxSendMsgSize
	}

	maxRecvMsgSize := cfg.MaxRecvMsgSize
	if maxRecvMsgSize == 0 {
		maxRecvMsgSize = config.DefaultGRPCMaxRecvMsgSize
	}

	// Setup historical gRPC connections if configured
	if len(cfg.HistoricalGRPCAddressBlockRange) > 0 {
		updatedCtx, err := setupHistoricalGRPCConnections(
			clientCtx,
			cfg.HistoricalGRPCAddressBlockRange,
			maxRecvMsgSize,
			maxSendMsgSize,
			logger,
		)
		if err != nil {
			return nil, clientCtx, fmt.Errorf("failed to setup historical gRPC connections: %w", err)
		}
		clientCtx = updatedCtx
	}

	grpcSrv := grpc.NewServer(
		grpc.ForceServerCodec(codec.NewProtoCodec(clientCtx.InterfaceRegistry).GRPCCodec()),
		grpc.MaxSendMsgSize(maxSendMsgSize),
		grpc.MaxRecvMsgSize(maxRecvMsgSize),
	)

	app.RegisterGRPCServerWithSkipCheckHeader(grpcSrv, cfg.SkipCheckHeader)

	// Reflection allows consumers to build dynamic clients that can write to any
	// Cosmos SDK application without relying on application packages at compile
	// time.
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
		SdkConfig:         sdk.GetConfig(),
		InterfaceRegistry: clientCtx.InterfaceRegistry,
	})
	if err != nil {
		return nil, clientCtx, fmt.Errorf("failed to register reflection service: %w", err)
	}

	// Reflection allows external clients to see what services and methods
	// the gRPC server exposes.
	gogoreflection.Register(grpcSrv)

	return grpcSrv, clientCtx, nil
}

// setupHistoricalGRPCConnections creates historical gRPC connections based on the configuration.
func setupHistoricalGRPCConnections(
	clientCtx client.Context,
	historicalAddresses map[config.BlockRange]string,
	maxRecvMsgSize, maxSendMsgSize int,
	logger log.Logger,
) (client.Context, error) {
	if len(historicalAddresses) == 0 {
		return clientCtx, nil
	}

	historicalConns := make(config.HistoricalGRPCConnections)
	for blockRange, address := range historicalAddresses {
		conn, err := grpc.NewClient(
			address,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(
				grpc.ForceCodec(codec.NewProtoCodec(clientCtx.InterfaceRegistry).GRPCCodec()),
				grpc.MaxCallRecvMsgSize(maxRecvMsgSize),
				grpc.MaxCallSendMsgSize(maxSendMsgSize),
			),
		)
		if err != nil {
			return clientCtx, fmt.Errorf("failed to create historical gRPC connection for %s: %w", address, err)
		}
		historicalConns[blockRange] = conn
	}

	// Get the default connection from the clientCtx
	defaultConn := clientCtx.GRPCClient
	if defaultConn == nil {
		return clientCtx, fmt.Errorf("default gRPC client not set in clientCtx")
	}

	provider := client.NewGRPCConnProvider(defaultConn, historicalConns)
	clientCtx = clientCtx.WithGRPCConnProvider(provider)

	logger.Info("historical gRPC connections configured", "count", len(historicalConns))
	return clientCtx, nil
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
		logger.Error("failed to start gRPC server", "err", err)
		return err
	}
}
