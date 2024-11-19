package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"slices"
	"strconv"
	"strings"
	"sync"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/server/v2/api/grpc/gogoreflection"
)

const (
	ServerName = "grpc"

	BlockHeightHeader = "x-cosmos-block-height"
)

type Server[T transaction.Tx] struct {
	logger     log.Logger
	config     *Config
	cfgOptions []CfgOption

	grpcSrv *grpc.Server
}

// New creates a new grpc server.
func New[T transaction.Tx](
	logger log.Logger,
	interfaceRegistry server.InterfaceRegistry,
	queryHandlers map[string]appmodulev2.Handler,
	queryable func(ctx context.Context, version uint64, msg transaction.Msg) (transaction.Msg, error),
	cfg server.ConfigMap,
	cfgOptions ...CfgOption,
) (*Server[T], error) {
	srv := &Server[T]{
		cfgOptions: cfgOptions,
	}
	serverCfg := srv.Config().(*Config)
	if len(cfg) > 0 {
		if err := serverv2.UnmarshalSubConfig(cfg, srv.Name(), &serverCfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	grpcSrv := grpc.NewServer(
		grpc.ForceServerCodec(newProtoCodec(interfaceRegistry).GRPCCodec()),
		grpc.MaxSendMsgSize(serverCfg.MaxSendMsgSize),
		grpc.MaxRecvMsgSize(serverCfg.MaxRecvMsgSize),
		grpc.UnknownServiceHandler(makeUnknownServiceHandler(queryHandlers, queryable)),
	)

	// Reflection allows external clients to see what services and methods the gRPC server exposes.
	gogoreflection.Register(grpcSrv, slices.Collect(maps.Keys(queryHandlers)), logger.With("sub-module", "grpc-reflection"))

	// Register V2 grpc handlers
	RegisterServiceServer(grpcSrv, &v2Service{queryHandlers, queryable})

	srv.grpcSrv = grpcSrv
	srv.config = serverCfg
	srv.logger = logger.With(log.ModuleKey, srv.Name())

	return srv, nil
}

// NewWithConfigOptions creates a new GRPC server with the provided config options.
// It is *not* a fully functional server (since it has been created without dependencies)
// The returned server should only be used to get and set configuration.
func NewWithConfigOptions[T transaction.Tx](opts ...CfgOption) *Server[T] {
	return &Server[T]{
		cfgOptions: opts,
	}
}

func (s *Server[T]) StartCmdFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet(s.Name(), pflag.ExitOnError)
	flags.String(FlagAddress, "localhost:9090", "Listen address")
	return flags
}

func makeUnknownServiceHandler(
	handlers map[string]appmodulev2.Handler,
	queryable func(ctx context.Context, version uint64, msg transaction.Msg) (transaction.Msg, error),
) grpc.StreamHandler {
	getRegistry := sync.OnceValues(gogoproto.MergedRegistry)

	return func(srv any, stream grpc.ServerStream) error {
		method, ok := grpc.MethodFromServerStream(stream)
		if !ok {
			return status.Error(codes.InvalidArgument, "unable to get method")
		}
		// if this fails we cannot serve queries anymore...
		registry, err := getRegistry()
		if err != nil {
			return fmt.Errorf("failed to get registry: %w", err)
		}

		method = strings.TrimPrefix(method, "/")
		fullName := protoreflect.FullName(strings.ReplaceAll(method, "/", "."))
		// get descriptor from the invoke method
		desc, err := registry.FindDescriptorByName(fullName)
		if err != nil {
			return fmt.Errorf("failed to find descriptor %s: %w", method, err)
		}
		md, ok := desc.(protoreflect.MethodDescriptor)
		if !ok {
			return fmt.Errorf("%s is not a method", method)
		}
		// find handler
		handler, exists := handlers[string(md.Input().FullName())]
		if !exists {
			return status.Errorf(codes.Unimplemented, "gRPC method %s is not handled", method)
		}

		for {
			req := handler.MakeMsg()
			err := stream.RecvMsg(req)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return err
			}

			// extract height header
			ctx := stream.Context()
			height, err := getHeightFromCtx(ctx)
			if err != nil {
				return status.Errorf(codes.InvalidArgument, "invalid get height from context: %v", err)
			}
			resp, err := queryable(ctx, height, req)
			if err != nil {
				return err
			}
			err = stream.SendMsg(resp)
			if err != nil {
				return err
			}
		}
	}
}

func getHeightFromCtx(ctx context.Context) (uint64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, nil
	}
	values := md.Get(BlockHeightHeader)
	if len(values) == 0 {
		return 0, nil
	}
	if len(values) != 1 {
		return 0, fmt.Errorf("gRPC height metadata must be of length 1, got: %d", len(values))
	}

	heightStr := values[0]
	height, err := strconv.ParseUint(heightStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("unable to parse height string from gRPC metadata %s: %w", heightStr, err)
	}

	return height, nil
}

func (s *Server[T]) Name() string {
	return ServerName
}

func (s *Server[T]) Config() any {
	if s.config == nil || s.config.Address == "" {
		cfg := DefaultConfig()
		// overwrite the default config with the provided options
		for _, opt := range s.cfgOptions {
			opt(cfg)
		}

		return cfg
	}

	return s.config
}

func (s *Server[T]) Start(ctx context.Context) error {
	if !s.config.Enable {
		s.logger.Info(fmt.Sprintf("%s server is disabled via config", s.Name()))
		return nil
	}

	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", s.config.Address)
	if err != nil {
		return fmt.Errorf("failed to listen on address %s: %w", s.config.Address, err)
	}

	s.logger.Info("starting gRPC server...", "address", s.config.Address)
	if err := s.grpcSrv.Serve(listener); err != nil {
		return fmt.Errorf("failed to start gRPC server: %w", err)
	}

	return nil
}

func (s *Server[T]) Stop(ctx context.Context) error {
	if !s.config.Enable {
		return nil
	}

	s.logger.Info("stopping gRPC server...", "address", s.config.Address)
	s.grpcSrv.GracefulStop()

	return nil
}

// GetGRPCServer returns the underlying gRPC server.
func (s *Server[T]) GetGRPCServer() *grpc.Server {
	return s.grpcSrv
}
