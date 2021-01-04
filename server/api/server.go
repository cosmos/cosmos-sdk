package api

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gogo/gateway"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/config"
	grpcservice "github.com/cosmos/cosmos-sdk/server/services/grpc"
	tmservice "github.com/cosmos/cosmos-sdk/server/services/tendermint"
	"github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/types/rest"

	// unnamed import of statik for swagger UI support
	_ "github.com/cosmos/cosmos-sdk/client/docs/statik"
)

var _ types.Server = &BaseServer{}

// BaseServer defines the SDK server's.
type BaseServer struct {
	services map[string]types.Service
	config   config.ServerConfig

	Router            *mux.Router
	GRPCGatewayRouter *runtime.ServeMux
	ClientCtx         client.Context

	logger   log.Logger
	metrics  *telemetry.Metrics
	listener net.Listener
}

// CustomGRPCHeaderMatcher for mapping request headers to
// GRPC metadata.
// HTTP headers that start with 'Grpc-Metadata-' are automatically mapped to
// gRPC metadata after removing prefix 'Grpc-Metadata-'. We can use this
// CustomGRPCHeaderMatcher if headers don't start with `Grpc-Metadata-`
func CustomGRPCHeaderMatcher(key string) (string, bool) {
	switch strings.ToLower(key) {
	case grpctypes.GRPCBlockHeightHeader:
		return grpctypes.GRPCBlockHeightHeader, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

// New creates the default SDK server instance.
func New(clientCtx client.Context, logger log.Logger, cfg config.ServerConfig) *BaseServer {
	// The default JSON marshaller used by the gRPC-Gateway is unable to marshal non-nullable non-scalar fields.
	// Using the gogo/gateway package with the gRPC-Gateway WithMarshaler option fixes the scalar field marshalling issue.
	marshalerOption := &gateway.JSONPb{
		EmitDefaults: true,
		Indent:       "  ",
		OrigName:     true,
		AnyResolver:  clientCtx.InterfaceRegistry,
	}

	router := mux.NewRouter()
	tmsrv := tmservice.NewService(logger, router)
	grpcsrv := grpcservice.NewService("")

	services := make(map[string]types.Service)
	services[tmsrv.Name()] = tmsrv
	services[grpcsrv.Name()] = grpcsrv

	return &BaseServer{
		services:  services,
		config:    cfg,
		Router:    router,
		ClientCtx: clientCtx,
		logger:    logger.With("module", "api-server"),
		GRPCGatewayRouter: runtime.NewServeMux(
			// Custom marshaler option is required for gogo proto
			runtime.WithMarshalerOption(runtime.MIMEWildcard, marshalerOption),

			// This is necessary to get error details properly
			// marshalled in unary requests.
			runtime.WithProtoErrorHandler(runtime.DefaultHTTPProtoErrorHandler),

			// Custom header matcher for mapping request headers to
			// GRPC metadata
			runtime.WithIncomingHeaderMatcher(CustomGRPCHeaderMatcher),
		),
	}
}

// GetService implements the Server interface.
func (s *BaseServer) GetService(name string) types.Service {
	service, _ := s.services[name]
	return service
}

// RegisterServices implements the Server interface.
func (s *BaseServer) RegisterServices() error {
	for name, service := range s.services {
		if !service.RegisterRoutes() {
			return fmt.Errorf("failed to register routes for service %s", name)
		}
	}

	sdkCfg := s.config.GetSDKConfig()
	if sdkCfg.Telemetry.Enabled {
		m, err := telemetry.New(sdkCfg.Telemetry)
		if err != nil {
			return err
		}

		s.metrics = m
		s.registerMetrics()
	}

	s.registerGRPCGatewayRoutes()

	return nil
}

// Start starts the API server. Internally, the API server leverages Tendermint's
// JSON RPC server. Configuration options are provided via config.APIConfig
// and are delegated to the Tendermint JSON RPC server. The process is
// non-blocking, so an external signal handler must be used.
func (s *BaseServer) Start() error {
	for name, service := range s.services {
		if err := service.Start(); err != nil {
			return fmt.Errorf("service %s start failed: %w", name, err)
		}
	}

	return nil
}

// Close closes the API server.
func (s *BaseServer) Close() error {
	for name, service := range s.services {
		if err := service.Stop(); err != nil {
			return fmt.Errorf("service %s stop failed: %w", name, err)
		}
	}

	return nil
}

func (s *BaseServer) registerGRPCGatewayRoutes() {
	s.Router.PathPrefix("/").Handler(s.GRPCGatewayRouter)
}

func (s *BaseServer) registerMetrics() {
	metricsHandler := func(w http.ResponseWriter, r *http.Request) {
		format := strings.TrimSpace(r.FormValue("format"))

		gr, err := s.metrics.Gather(format)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to gather metrics: %s", err))
			return
		}

		w.Header().Set("Content-Type", gr.ContentType)
		_, _ = w.Write(gr.Metrics)
	}

	s.Router.HandleFunc("/metrics", metricsHandler).Methods("GET")
}
