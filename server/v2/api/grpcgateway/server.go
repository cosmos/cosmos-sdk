package grpcgateway

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	gateway "github.com/cosmos/gogogateway"
	"github.com/cosmos/gogoproto/jsonpb"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
)

var _ serverv2.ServerComponent[
	serverv2.AppI[transaction.Tx], transaction.Tx,
] = (*GRPCGatewayServer[serverv2.AppI[transaction.Tx], transaction.Tx])(nil)

const (
	// GRPCBlockHeightHeader is the gRPC header for block height.
	GRPCBlockHeightHeader = "x-cosmos-block-height"
)

type GRPCGatewayServer[AppT serverv2.AppI[T], T transaction.Tx] struct {
	logger     log.Logger
	config     *Config
	cfgOptions []CfgOption

	GRPCSrv           *grpc.Server
	GRPCGatewayRouter *runtime.ServeMux
}

// New creates a new gRPC-gateway server.
func New[AppT serverv2.AppI[T], T transaction.Tx](grpcSrv *grpc.Server, ir jsonpb.AnyResolver, cfgOptions ...CfgOption) *GRPCGatewayServer[AppT, T] {
	// The default JSON marshaller used by the gRPC-Gateway is unable to marshal non-nullable non-scalar fields.
	// Using the gogo/gateway package with the gRPC-Gateway WithMarshaler option fixes the scalar field marshaling issue.
	marshalerOption := &gateway.JSONPb{
		EmitDefaults: true,
		Indent:       "",
		OrigName:     true,
		AnyResolver:  ir,
	}

	return &GRPCGatewayServer[AppT, T]{
		GRPCSrv: grpcSrv,
		GRPCGatewayRouter: runtime.NewServeMux(
			// Custom marshaler option is required for gogo proto
			runtime.WithMarshalerOption(runtime.MIMEWildcard, marshalerOption),

			// This is necessary to get error details properly
			// marshaled in unary requests.
			runtime.WithProtoErrorHandler(runtime.DefaultHTTPProtoErrorHandler),

			// Custom header matcher for mapping request headers to
			// GRPC metadata
			runtime.WithIncomingHeaderMatcher(CustomGRPCHeaderMatcher),
		),
		cfgOptions: cfgOptions,
	}
}

func (g *GRPCGatewayServer[AppT, T]) Name() string {
	return "grpc-gateway"
}

func (s *GRPCGatewayServer[AppT, T]) Config() any {
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

func (s *GRPCGatewayServer[AppT, T]) Init(appI AppT, v *viper.Viper, logger log.Logger) error {
	cfg := s.Config().(*Config)
	if v != nil {
		if err := v.Sub(s.Name()).Unmarshal(&cfg); err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	// Register the gRPC-Gateway server.
	// appI.RegisterGRPCGatewayRoutes(s.GRPCGatewayRouter, s.GRPCSrv)

	s.logger = logger
	s.config = cfg

	return nil
}

func (s *GRPCGatewayServer[AppT, T]) Start(ctx context.Context) error {
	if !s.Config().(*Config).Enable {
		return nil
	}

	// TODO start a normal Go http server (and do not leverage comet's like https://github.com/cosmos/cosmos-sdk/blob/9df6019de6ee7999fe9864bac836deb2f36dd44a/server/api/server.go#L98)

	return nil
}

func (s *GRPCGatewayServer[AppT, T]) Stop(ctx context.Context) error {
	if !s.Config().(*Config).Enable {
		return nil
	}

	return nil
}

// Register implements registers a grpc-gateway server
func (s *GRPCGatewayServer[AppT, T]) Register(r mux.Router) error {
	// configure grpc-gatway server
	r.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Fall back to grpc gateway server.
		s.GRPCGatewayRouter.ServeHTTP(w, req)
	}))

	return nil
}

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
