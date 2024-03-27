package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"

	gogogrpc "github.com/cosmos/gogoproto/grpc"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/proto"

	_ "cosmossdk.io/api/amino" // Import amino.proto file for reflection
	appmanager "cosmossdk.io/core/app"
	"cosmossdk.io/log"
	"cosmossdk.io/server/v2/api/grpc/gogoreflection"
)

const serverName = "grpc-server"

type GRPCServer struct {
	logger log.Logger

	grpcSrv *grpc.Server
	config  *Config
}

type GRPCService interface {
	// RegisterGRPCServer registers gRPC services directly with the gRPC server.
	RegisterGRPCServer(gogogrpc.Server)
}

// New returns a correctly configured and initialized gRPC server.
// Note, the caller is responsible for starting the server.
func New(logger log.Logger, v *viper.Viper, interfaceRegistry appmanager.InterfaceRegistry, app GRPCService) (GRPCServer, error) {
	cfg := DefaultConfig()
	if v != nil {
		if err := v.Sub(serverName).Unmarshal(&cfg); err != nil {
			return GRPCServer{}, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	grpcSrv := grpc.NewServer(
		grpc.ForceServerCodec(newProtoCodec(interfaceRegistry).GRPCCodec()),
		grpc.MaxSendMsgSize(cfg.MaxSendMsgSize),
		grpc.MaxRecvMsgSize(cfg.MaxRecvMsgSize),
	)

	app.RegisterGRPCServer(grpcSrv)

	// Reflection allows external clients to see what services and methods
	// the gRPC server exposes.
	gogoreflection.Register(grpcSrv)

	return GRPCServer{
		grpcSrv: grpcSrv,
		config:  cfg,
		logger:  logger.With(log.ModuleKey, serverName),
	}, nil
}

func (g GRPCServer) Name() string {
	return serverName
}

func (g GRPCServer) Start(ctx context.Context) error {
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

func (g GRPCServer) Stop(ctx context.Context) error {
	g.logger.Info("stopping gRPC server...", "address", g.config.Address)
	g.grpcSrv.GracefulStop()

	return nil
}

func (g GRPCServer) Config() any {
	if g.config == nil || g.config == (&Config{}) {
		return DefaultConfig()
	}

	return g.config
}

type protoCodec struct {
	interfaceRegistry appmanager.InterfaceRegistry
}

// newProtoCodec returns a reference to a new ProtoCodec
func newProtoCodec(interfaceRegistry appmanager.InterfaceRegistry) *protoCodec {
	return &protoCodec{
		interfaceRegistry: interfaceRegistry,
	}
}

// Marshal implements BinaryMarshaler.Marshal method.
// NOTE: this function must be used with a concrete type which
// implements proto.Message. For interface please use the codec.MarshalInterface
func (pc *protoCodec) Marshal(o gogoproto.Message) ([]byte, error) {
	// Size() check can catch the typed nil value.
	if o == nil || gogoproto.Size(o) == 0 {
		// return empty bytes instead of nil, because nil has special meaning in places like store.Set
		return []byte{}, nil
	}

	return gogoproto.Marshal(o)
}

// Unmarshal implements BinaryMarshaler.Unmarshal method.
// NOTE: this function must be used with a concrete type which
// implements proto.Message. For interface please use the codec.UnmarshalInterface
func (pc *protoCodec) Unmarshal(bz []byte, ptr gogoproto.Message) error {
	err := gogoproto.Unmarshal(bz, ptr)
	if err != nil {
		return err
	}
	// err = codectypes.UnpackInterfaces(ptr, pc.interfaceRegistry) // TODO: identify if needed for grpc
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (pc *protoCodec) Name() string {
	return "cosmos-sdk-grpc-codec"
}

// GRPCCodec returns the gRPC Codec for this specific ProtoCodec
func (pc *protoCodec) GRPCCodec() encoding.Codec {
	return &grpcProtoCodec{cdc: pc}
}

// grpcProtoCodec is the implementation of the gRPC proto codec.
type grpcProtoCodec struct {
	cdc appmanager.ProtoCodec
}

var errUnknownProtoType = errors.New("codec: unknown proto type") // sentinel error

func (g grpcProtoCodec) Marshal(v any) ([]byte, error) {
	switch m := v.(type) {
	case proto.Message:
		protov2MarshalOpts := proto.MarshalOptions{Deterministic: true}
		return protov2MarshalOpts.Marshal(m)
	case gogoproto.Message:
		return g.cdc.Marshal(m)
	default:
		return nil, fmt.Errorf("%w: cannot marshal type %T", errUnknownProtoType, v)
	}
}

func (g grpcProtoCodec) Unmarshal(data []byte, v any) error {
	switch m := v.(type) {
	case proto.Message:
		return proto.Unmarshal(data, m)
	case gogoproto.Message:
		return g.cdc.Unmarshal(data, m)
	default:
		return fmt.Errorf("%w: cannot unmarshal type %T", errUnknownProtoType, v)
	}
}

func (g grpcProtoCodec) Name() string {
	return "cosmos-sdk-grpc-codec"
}
