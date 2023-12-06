package grpc

import (
	"context"
	"fmt"
	"net"

	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/serverv2"
	"github.com/cosmos/cosmos-sdk/serverv2/grpc/gogoreflection"
	reflection "github.com/cosmos/cosmos-sdk/serverv2/grpc/reflection/v2alpha1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino" // Import amino.proto file for reflection
)

var _ serverv2.Module = &Server{}

type Server struct {
	grpcSrv *grpc.Server
	logger  log.Logger
	config  Config
}

type GRPCService interface {
	// RegisterGRPCServer registers gRPC services directly with the gRPC server.
	RegisterGRPCServer(gogogrpc.Server)
}

// NewServer returns a correctly configured and initialized gRPC server.
// Note, the caller is responsible for starting the server.
func NewServer(logger log.Logger, interfaceRegistry codectypes.InterfaceRegistry, txConfig client.TxConfig, app GRPCService, cfg Config, chainID string) (Server, error) {
	maxSendMsgSize := cfg.MaxSendMsgSize
	if maxSendMsgSize == 0 {
		maxSendMsgSize = DefaultGRPCMaxSendMsgSize
	}

	maxRecvMsgSize := cfg.MaxRecvMsgSize
	if maxRecvMsgSize == 0 {
		maxRecvMsgSize = DefaultGRPCMaxRecvMsgSize
	}

	grpcSrv := grpc.NewServer(
		grpc.ForceServerCodec(codec.NewProtoCodec(interfaceRegistry).GRPCCodec()),
		grpc.MaxSendMsgSize(maxSendMsgSize),
		grpc.MaxRecvMsgSize(maxRecvMsgSize),
	)

	app.RegisterGRPCServer(grpcSrv)

	// Reflection allows consumers to build dynamic clients that can write to any
	// Cosmos SDK application without relying on application packages at compile
	// time.
	err := reflection.Register(grpcSrv, reflection.Config{
		SigningModes: func() map[string]int32 {
			supportedModes := txConfig.SignModeHandler().SupportedModes()
			modes := make(map[string]int32, len(supportedModes))
			for _, m := range supportedModes {
				modes[m.String()] = (int32)(m)
			}

			return modes
		}(),
		ChainID:                    chainID,
		Bech32AccountAddressPrefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
		InterfaceRegistry:          interfaceRegistry,
	})
	if err != nil {
		return Server{}, fmt.Errorf("failed to register reflection service: %w", err)
	}

	// Reflection allows external clients to see what services and methods
	// the gRPC server exposes.
	gogoreflection.Register(grpcSrv)

	return Server{
		grpcSrv: grpcSrv,
		config:  cfg,
		logger:  logger.With("module", "grpc-server"),
	}, nil
}

func (g Server) Name() string {
	return "grpc-server"
}

func (g Server) Start(context.Context) error {
	listener, err := net.Listen("tcp", g.config.Address)
	if err != nil {
		return fmt.Errorf("failed to listen on address %s: %w", g.config.Address, err)
	}

	errCh := make(chan error)

	// Start the gRPC in an external goroutine as Serve is blocking and will return
	// an error upon failure, which we'll send on the error channel that will be
	// consumed by the for block below.
	go func() {
		g.logger.Info("starting gRPC server...", "address", g.config.Address)
		errCh <- g.grpcSrv.Serve(listener)
	}()

	// Start a blocking select to wait for an indication to stop the server or that
	// the server failed to start properly.
	err = <-errCh
	g.logger.Error("failed to start gRPC server", "err", err)

	return err
}

func (g Server) Stop(context.Context) error {
	g.logger.Info("stopping gRPC server...", "address", g.config.Address)
	g.grpcSrv.GracefulStop()

	return nil
}

func (g Server) Config() (any, *viper.Viper) {
	v := viper.New()

	// TODO
	v.Set("enable", true)
	v.Set("address", g.config.Address)
	v.Set("max_send_msg_size", g.config.MaxSendMsgSize)
	v.Set("max_recv_msg_size", g.config.MaxRecvMsgSize)

	return g.config, v
}
