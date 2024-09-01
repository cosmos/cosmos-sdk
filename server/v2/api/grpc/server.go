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

	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

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
func New[T transaction.Tx](cfgOptions ...CfgOption) *Server[T] {
	return &Server[T]{
		cfgOptions: cfgOptions,
	}
}

// Init returns a correctly configured and initialized gRPC server.
// Note, the caller is responsible for starting the server.
func (s *Server[T]) Init(appI serverv2.AppI[T], v *viper.Viper, logger log.Logger) error {
	cfg := s.Config().(*Config)
	if v != nil {
		if err := serverv2.UnmarshalSubConfig(v, s.Name(), &cfg); err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}
	methodsMap := appI.GetGPRCMethodsToMessageMap()

	grpcSrv := grpc.NewServer(
		grpc.ForceServerCodec(newProtoCodec(appI.InterfaceRegistry()).GRPCCodec()),
		grpc.MaxSendMsgSize(cfg.MaxSendMsgSize),
		grpc.MaxRecvMsgSize(cfg.MaxRecvMsgSize),
		grpc.UnknownServiceHandler(
			makeUnknownServiceHandler(methodsMap, appI.GetAppManager()),
		),
	)

	// Reflection allows external clients to see what services and methods the gRPC server exposes.
	gogoreflection.Register(grpcSrv, slices.Collect(maps.Keys(methodsMap)), logger.With("sub-module", "grpc-reflection"))

	s.grpcSrv = grpcSrv
	s.config = cfg
	s.logger = logger.With(log.ModuleKey, s.Name())

	return nil
}

func (s *Server[T]) StartCmdFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet(s.Name(), pflag.ExitOnError)
	flags.String(FlagAddress, "localhost:9090", "Listen address")
	return flags
}

func makeUnknownServiceHandler(messageMap map[string]func() proto.Message, querier interface {
	Query(ctx context.Context, version uint64, msg proto.Message) (proto.Message, error)
},
) grpc.StreamHandler {
	return func(srv any, stream grpc.ServerStream) error {
		method, ok := grpc.MethodFromServerStream(stream)
		if !ok {
			return status.Error(codes.InvalidArgument, "unable to get method")
		}
		makeMsg, exists := messageMap[method]
		if !exists {
			return status.Errorf(codes.Unimplemented, "gRPC method %s is not handled", method)
		}
		for {
			req := makeMsg()
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
			resp, err := querier.Query(ctx, height, req)
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
	if s.config == nil || s.config == (&Config{}) {
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
		return nil
	}

	listener, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		return fmt.Errorf("failed to listen on address %s: %w", s.config.Address, err)
	}

	errCh := make(chan error)

	// Start the gRPC in an external goroutine as Serve is blocking and will return
	// an error upon failure, which we'll send on the error channel that will be
	// consumed by the for block below.
	go func() {
		s.logger.Info("starting gRPC server...", "address", s.config.Address)
		errCh <- s.grpcSrv.Serve(listener)
	}()

	// Start a blocking select to wait for an indication to stop the server or that
	// the server failed to start properly.
	err = <-errCh
	if err != nil {
		s.logger.Error("failed to start gRPC server", "err", err)
	}

	return err
}

func (s *Server[T]) Stop(ctx context.Context) error {
	if !s.config.Enable {
		return nil
	}

	s.logger.Info("stopping gRPC server...", "address", s.config.Address)
	s.grpcSrv.GracefulStop()

	return nil
}
