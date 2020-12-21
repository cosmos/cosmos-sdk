package api

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gogo/gateway"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/tendermint/tendermint/libs/log"
	tmrpcserver "github.com/tendermint/tendermint/rpc/jsonrpc/server"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/telemetry"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/types/rest"

	// unnamed import of statik for swagger UI support
	_ "github.com/cosmos/cosmos-sdk/client/docs/statik"
)

type Server interface {
	BaseServer() *BaseServer

	Start(config.ServerConfig) error
	Close() error
}

var _ Server = &BaseServer{}

// BaseServer defines the SDK server's API interface.
type BaseServer struct {
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
func New(clientCtx client.Context, logger log.Logger) *BaseServer {
	// The default JSON marshaller used by the gRPC-Gateway is unable to marshal non-nullable non-scalar fields.
	// Using the gogo/gateway package with the gRPC-Gateway WithMarshaler option fixes the scalar field marshalling issue.
	marshalerOption := &gateway.JSONPb{
		EmitDefaults: true,
		Indent:       "  ",
		OrigName:     true,
		AnyResolver:  clientCtx.InterfaceRegistry,
	}

	return &BaseServer{
		Router:    mux.NewRouter(),
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

// BaseServer implements the Server interface.
func (s *BaseServer) BaseServer() *BaseServer { return s }

// Start starts the API server. Internally, the API server leverages Tendermint's
// JSON RPC server. Configuration options are provided via config.APIConfig
// and are delegated to the Tendermint JSON RPC server. The process is
// non-blocking, so an external signal handler must be used.
func (s *BaseServer) Start(cfg config.ServerConfig) error {
	sdkCfg := cfg.GetSDKConfig()

	if sdkCfg.Telemetry.Enabled {
		m, err := telemetry.New(sdkCfg.Telemetry)
		if err != nil {
			return err
		}

		s.metrics = m
		s.registerMetrics()
	}

	tmCfg := tmrpcserver.DefaultConfig()
	tmCfg.MaxOpenConnections = int(sdkCfg.API.MaxOpenConnections)
	tmCfg.ReadTimeout = time.Duration(sdkCfg.API.RPCReadTimeout) * time.Second
	tmCfg.WriteTimeout = time.Duration(sdkCfg.API.RPCWriteTimeout) * time.Second
	tmCfg.MaxBodyBytes = int64(sdkCfg.API.RPCMaxBodyBytes)

	listener, err := tmrpcserver.Listen(sdkCfg.API.Address, tmCfg)
	if err != nil {
		return err
	}

	s.registerGRPCGatewayRoutes()

	s.listener = listener
	var h http.Handler = s.Router

	if sdkCfg.API.EnableUnsafeCORS {
		allowAllCORS := handlers.CORS(handlers.AllowedHeaders([]string{"Content-Type"}))
		return tmrpcserver.Serve(s.listener, allowAllCORS(h), s.logger, tmCfg)
	}

	s.logger.Info("starting API server...")
	return tmrpcserver.Serve(s.listener, s.Router, s.logger, tmCfg)
}

// Close closes the API server.
func (s *BaseServer) Close() error {
	return s.listener.Close()
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
