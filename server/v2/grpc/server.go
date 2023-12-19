package grpc

import (
	"fmt"
	"math"
	"net"

	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"google.golang.org/grpc"

	"cosmossdk.io/log"

	"cosmossdk.io/server/v2/grpc/gogoreflection"
	reflection "cosmossdk.io/server/v2/grpc/reflection/v2alpha1"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino" // Import amino.proto file for reflection
)

const (
	// DefaultGRPCAddress defines the default address to bind the gRPC server to.
	DefaultGRPCAddress = "localhost:9090"

	// DefaultGRPCMaxRecvMsgSize defines the default gRPC max message size in
	// bytes the server can receive.
	DefaultGRPCMaxRecvMsgSize = 1024 * 1024 * 10

	// DefaultGRPCMaxSendMsgSize defines the default gRPC max message size in
	// bytes the server can send.
	DefaultGRPCMaxSendMsgSize = math.MaxInt32
)

type GRPCServer struct {
	grpcSrv *grpc.Server
	logger  log.Logger
	config  Config
}

// GRPCConfig defines configuration for the gRPC server.
type Config struct {
	// Enable defines if the gRPC server should be enabled.
	Enable bool `mapstructure:"enable"`

	// Address defines the API server to listen on
	Address string `mapstructure:"address"`

	// MaxRecvMsgSize defines the max message size in bytes the server can receive.
	// The default value is 10MB.
	MaxRecvMsgSize int `mapstructure:"max-recv-msg-size"`

	// MaxSendMsgSize defines the max message size in bytes the server can send.
	// The default value is math.MaxInt32.
	MaxSendMsgSize int `mapstructure:"max-send-msg-size"`
}

type GRPCService interface {
	// RegisterGRPCServer registers gRPC services directly with the gRPC
	// server.
	RegisterGRPCServer(gogogrpc.Server)
}

// NewGRPCServer returns a correctly configured and initialized gRPC server.
// Note, the caller is responsible for starting the server. See StartGRPCServer.
// TODO: look into removing the clientCtx dependency.
func NewGRPCServer(clientCtx client.Context, logger log.Logger, app GRPCService, cfg Config) (GRPCServer, error) {
	maxSendMsgSize := cfg.MaxSendMsgSize
	if maxSendMsgSize == 0 {
		maxSendMsgSize = DefaultGRPCMaxSendMsgSize
	}

	maxRecvMsgSize := cfg.MaxRecvMsgSize
	if maxRecvMsgSize == 0 {
		maxRecvMsgSize = DefaultGRPCMaxRecvMsgSize
	}

	grpcSrv := grpc.NewServer(
		grpc.ForceServerCodec(codec.NewProtoCodec(clientCtx.InterfaceRegistry).GRPCCodec()),
		grpc.MaxSendMsgSize(maxSendMsgSize),
		grpc.MaxRecvMsgSize(maxRecvMsgSize),
	)

	app.RegisterGRPCServer(grpcSrv)

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
		ChainID:                    clientCtx.ChainID,
		Bech32AccountAddressPrefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
		InterfaceRegistry:          clientCtx.InterfaceRegistry,
	})
	if err != nil {
		return GRPCServer{}, fmt.Errorf("failed to register reflection service: %w", err)
	}

	// Reflection allows external clients to see what services and methods
	// the gRPC server exposes.
	gogoreflection.Register(grpcSrv)

	return GRPCServer{
		grpcSrv: grpcSrv,
		config:  cfg,
		logger:  logger.With("module", "grpc-server"),
	}, nil
}

func (g GRPCServer) Start() error {
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

func (g GRPCServer) Stop() {
	g.logger.Info("stopping gRPC server...", "address", g.config.Address)
	g.grpcSrv.GracefulStop()
}
