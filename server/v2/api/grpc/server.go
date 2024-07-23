package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/server/v2/api/grpc/gogoreflection"
)

type GRPCServer[T transaction.Tx] struct {
	logger     log.Logger
	config     *Config
	cfgOptions []CfgOption

	grpcSrv *grpc.Server
}

// New creates a new grpc server.
func New[T transaction.Tx](cfgOptions ...CfgOption) *GRPCServer[T] {
	return &GRPCServer[T]{
		cfgOptions: cfgOptions,
	}
}

// Init returns a correctly configured and initialized gRPC server.
// Note, the caller is responsible for starting the server.
func (s *GRPCServer[T]) Init(appI serverv2.AppI[T], v *viper.Viper, logger log.Logger) error {
	cfg := s.Config().(*Config)
	if v != nil {
		if err := serverv2.UnmarshalSubConfig(v, s.Name(), &cfg); err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	grpcSrv := grpc.NewServer(
		grpc.ForceServerCodec(newProtoCodec(appI.InterfaceRegistry()).GRPCCodec()),
		grpc.MaxSendMsgSize(cfg.MaxSendMsgSize),
		grpc.MaxRecvMsgSize(cfg.MaxRecvMsgSize),
		grpc.UnknownServiceHandler(makeUnknownServiceHandler(appI.GetGPRCMethodsToMessageMap(), appI.GetAppManager())),
	)

	// Reflection allows external clients to see what services and methods the gRPC server exposes.
	gogoreflection.Register(grpcSrv)

	s.grpcSrv = grpcSrv
	s.config = cfg
	s.logger = logger.With(log.ModuleKey, s.Name())

	return nil
}

func makeUnknownServiceHandler(messageMap map[string]func() proto.Message, querier interface {
	Query(ctx context.Context, version uint64, msg proto.Message) (proto.Message, error)
}) grpc.StreamHandler {
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
			height := uint64(0)

			resp, err := querier.Query(stream.Context(), height, req)
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

func (s *GRPCServer[T]) Name() string {
	return "grpc"
}

func (s *GRPCServer[T]) Config() any {
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

func (s *GRPCServer[T]) Start(ctx context.Context) error {
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
	s.logger.Error("failed to start gRPC server", "err", err)
	return err
}

func (s *GRPCServer[T]) Stop(ctx context.Context) error {
	if !s.config.Enable {
		return nil
	}

	s.logger.Info("stopping gRPC server...", "address", s.config.Address)
	s.grpcSrv.GracefulStop()

	return nil
}
