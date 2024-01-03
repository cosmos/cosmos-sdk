package grpc

import (
	"errors"
	"fmt"
	"math"
	"net"

	gogogrpc "github.com/cosmos/gogoproto/grpc"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/proto"

	"cosmossdk.io/log"
	"cosmossdk.io/server/v2/core/appmanager"
	"cosmossdk.io/server/v2/grpc/gogoreflection"
	reflection "cosmossdk.io/server/v2/grpc/reflection/v2alpha1"
	txsigning "cosmossdk.io/x/tx/signing"

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

type ClientContext interface {
	// InterfaceRegistry returns the InterfaceRegistry.
	InterfaceRegistry() appmanager.InterfaceRegistry
	ChainID() string
	TxConfig() TxConfig
}

type TxConfig interface {
	SignModeHandler() *txsigning.HandlerMap
}

// NewGRPCServer returns a correctly configured and initialized gRPC server.
// Note, the caller is responsible for starting the server. See StartGRPCServer.
// TODO: look into removing the clientCtx dependency.
func NewGRPCServer(clientCtx ClientContext, logger log.Logger, app GRPCService, cfg Config) (GRPCServer, error) {
	maxSendMsgSize := cfg.MaxSendMsgSize
	if maxSendMsgSize == 0 {
		maxSendMsgSize = DefaultGRPCMaxSendMsgSize
	}

	maxRecvMsgSize := cfg.MaxRecvMsgSize
	if maxRecvMsgSize == 0 {
		maxRecvMsgSize = DefaultGRPCMaxRecvMsgSize
	}

	grpcSrv := grpc.NewServer(
		grpc.ForceServerCodec(NewProtoCodec(clientCtx.InterfaceRegistry()).GRPCCodec()),
		grpc.MaxSendMsgSize(maxSendMsgSize),
		grpc.MaxRecvMsgSize(maxRecvMsgSize),
	)

	app.RegisterGRPCServer(grpcSrv)

	// Reflection allows consumers to build dynamic clients that can write to any
	// Cosmos SDK application without relying on application packages at compile
	// time.
	err := reflection.Register(grpcSrv, reflection.Config{
		SigningModes: func() map[string]int32 {
			supportedModes := clientCtx.TxConfig().SignModeHandler().SupportedModes()
			modes := make(map[string]int32, len(supportedModes))
			for _, m := range supportedModes {
				modes[m.String()] = (int32)(m)
			}

			return modes
		}(),
		ChainID:           clientCtx.ChainID(),
		InterfaceRegistry: clientCtx.InterfaceRegistry(),
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

type ProtoCodec struct {
	interfaceRegistry appmanager.InterfaceRegistry
}

// NewProtoCodec returns a reference to a new ProtoCodec
func NewProtoCodec(interfaceRegistry appmanager.InterfaceRegistry) *ProtoCodec {
	return &ProtoCodec{
		interfaceRegistry: interfaceRegistry,
	}
}

// Marshal implements BinaryMarshaler.Marshal method.
// NOTE: this function must be used with a concrete type which
// implements proto.Message. For interface please use the codec.MarshalInterface
func (pc *ProtoCodec) Marshal(o gogoproto.Message) ([]byte, error) {
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
func (pc *ProtoCodec) Unmarshal(bz []byte, ptr gogoproto.Message) error {
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

func (pc *ProtoCodec) Name() string {
	return "cosmos-sdk-grpc-codec"
}

// GRPCCodec returns the gRPC Codec for this specific ProtoCodec
func (pc *ProtoCodec) GRPCCodec() encoding.Codec {
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

func assertNotNil(i interface{}) error {
	if i == nil {
		return errors.New("can't marshal <nil> value")
	}
	return nil
}
