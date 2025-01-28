package grpcgateway

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	gateway "github.com/cosmos/gogogateway"
	"github.com/cosmos/gogoproto/jsonpb"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/server/v2/appmanager"
)

var (
	_ serverv2.ServerComponent[transaction.Tx] = (*Server[transaction.Tx])(nil)
	_ serverv2.HasConfig                       = (*Server[transaction.Tx])(nil)
)

const ServerName = "grpc-gateway"

type Server[T transaction.Tx] struct {
	logger     log.Logger
	config     *Config
	cfgOptions []CfgOption

	server            *http.Server
	GRPCGatewayRouter *runtime.ServeMux
}

// New creates a new gRPC-Gateway server.
func New[T transaction.Tx](
	logger log.Logger,
	config server.ConfigMap,
	ir jsonpb.AnyResolver,
	appManager appmanager.AppManager[T],
	cfgOptions ...CfgOption,
) (*Server[T], error) {
	// The default JSON marshaller used by the gRPC-Gateway is unable to marshal non-nullable non-scalar fields.
	// Using the gogo/gateway package with the gRPC-Gateway WithMarshaler option fixes the scalar field marshaling issue.
	marshalerOption := &gateway.JSONPb{
		EmitDefaults: true,
		Indent:       "",
		OrigName:     true,
		AnyResolver:  ir,
	}

	s := &Server[T]{
		GRPCGatewayRouter: runtime.NewServeMux(
			// Custom marshaler option is required for gogo proto
			runtime.WithMarshalerOption(runtime.MIMEWildcard, marshalerOption),

			// This is necessary to get error details properly
			// marshaled in unary requests.
			runtime.WithProtoErrorHandler(runtime.DefaultHTTPProtoErrorHandler),

			// Custom header uriMatcher for mapping request headers to
			// GRPC metadata
			runtime.WithIncomingHeaderMatcher(CustomGRPCHeaderMatcher),
		),
		cfgOptions: cfgOptions,
	}

	serverCfg := s.Config().(*Config)
	if len(config) > 0 {
		if err := serverv2.UnmarshalSubConfig(config, s.Name(), &serverCfg); err != nil {
			return s, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	s.logger = logger.With(log.ModuleKey, s.Name())
	s.config = serverCfg
	mux := http.NewServeMux()
	err := mountHTTPRoutes(mux, s.GRPCGatewayRouter, appManager)
	if err != nil {
		return nil, fmt.Errorf("failed to register gRPC gateway annotations: %w", err)
	}

	s.server = &http.Server{
		Addr:    s.config.Address,
		Handler: mux,
	}
	return s, nil
}

// NewWithConfigOptions creates a new gRPC-gateway server with the provided config options.
func NewWithConfigOptions[T transaction.Tx](opts ...CfgOption) *Server[T] {
	return &Server[T]{
		cfgOptions: opts,
	}
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

	s.logger.Info("starting gRPC-Gateway server...", "address", s.config.Address)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start gRPC-Gateway server: %w", err)
	}

	return nil
}

func (s *Server[T]) Stop(ctx context.Context) error {
	if !s.config.Enable {
		return nil
	}

	s.logger.Info("stopping gRPC-Gateway server...", "address", s.config.Address)
	return s.server.Shutdown(ctx)
}

// GRPCBlockHeightHeader is the gRPC header for block height.
const GRPCBlockHeightHeader = "x-cosmos-block-height"

// CustomGRPCHeaderMatcher for mapping request headers to
// GRPC metadata.
// HTTP headers that start with 'Grpc-Metadata-' are automatically mapped to
// gRPC metadata after removing prefix 'Grpc-Metadata-'. We can use this
// CustomGRPCHeaderMatcher if headers don't start with `Grpc-Metadata-`
func CustomGRPCHeaderMatcher(key string) (string, bool) {
	switch strings.ToLower(key) {
	case GRPCBlockHeightHeader:
		return GRPCBlockHeightHeader, true

	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}
