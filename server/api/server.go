package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gogo/gateway"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/tendermint/tendermint/libs/log"
	tmrpcserver "github.com/tendermint/tendermint/rpc/jsonrpc/server"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/telemetry"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"

	// unnamed import of statik for swagger UI support
	_ "github.com/cosmos/cosmos-sdk/client/docs/statik"
)

// Server defines the server's API interface.
type Server struct {
	Router            *mux.Router
	GRPCGatewayRouter *runtime.ServeMux
	ClientCtx         client.Context

	logger  log.Logger
	metrics *telemetry.Metrics
	// Start() is blocking and generally called from a separate goroutine.
	// Close() can be called asynchronously and access shared memory
	// via the listener. Therefore, we sync access to Start and Close with
	// this mutex to avoid data races.
	mtx      sync.Mutex
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

func New(clientCtx client.Context, logger log.Logger) *Server {
	// The default JSON marshaller used by the gRPC-Gateway is unable to marshal non-nullable non-scalar fields.
	// Using the gogo/gateway package with the gRPC-Gateway WithMarshaler option fixes the scalar field marshalling issue.
	marshalerOption := &gateway.JSONPb{
		EmitDefaults: true,
		Indent:       "  ",
		OrigName:     true,
		AnyResolver:  clientCtx.InterfaceRegistry,
	}

	return &Server{
		Router:    mux.NewRouter(),
		ClientCtx: clientCtx,
		logger:    logger,
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

// Start starts the API server. Internally, the API server leverages Tendermint's
// JSON RPC server. Configuration options are provided via config.APIConfig
// and are delegated to the Tendermint JSON RPC server.
//
// Note, this creates a blocking process if the server is started successfully.
// Otherwise, an error is returned. The caller is expected to provide a Context
// that is properly canceled or closed to indicate the server should be stopped.
func (s *Server) Start(ctx context.Context, cfg config.APIConfig) error {
	s.mtx.Lock()

	tmCfg := tmrpcserver.DefaultConfig()
	tmCfg.MaxOpenConnections = int(cfg.MaxOpenConnections)
	tmCfg.ReadTimeout = time.Duration(cfg.RPCReadTimeout) * time.Second
	tmCfg.WriteTimeout = time.Duration(cfg.RPCWriteTimeout) * time.Second
	tmCfg.MaxBodyBytes = int64(cfg.RPCMaxBodyBytes)

	listener, err := tmrpcserver.Listen(cfg.Address, tmCfg)
	if err != nil {
		s.mtx.Unlock()
		return err
	}

	s.registerGRPCGatewayRoutes()
	s.listener = listener

	s.mtx.Unlock()

	errCh := make(chan error)

	// Start the API in an external goroutine as Serve is blocking and will return
	// an error upon failure, which we'll send on the error channel that will be
	// consumed by the for block below.
	go func(enableUnsafeCORS bool) {
		s.logger.Info("starting API server...", "address", cfg.Address)

		if enableUnsafeCORS {
			allowAllCORS := handlers.CORS(handlers.AllowedHeaders([]string{"Content-Type"}))
			errCh <- tmrpcserver.Serve(s.listener, allowAllCORS(s.Router), s.logger, tmCfg)
		} else {
			errCh <- tmrpcserver.Serve(s.listener, s.Router, s.logger, tmCfg)
		}
	}(cfg.EnableUnsafeCORS)

	// Start a blocking select to wait for an indication to stop the server or that
	// the server failed to start properly.
	select {
	case <-ctx.Done():
		// The calling process cancelled or closed the provided context, so we must
		// gracefully stop the API server.
		s.logger.Info("stopping API server...", "address", cfg.Address)
		return s.Close()

	case err := <-errCh:
		s.logger.Error("failed to start API server", "err", err)
		return err
	}
}

// Close closes the API server.
func (s *Server) Close() error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.listener.Close()
}

func (s *Server) registerGRPCGatewayRoutes() {
	s.Router.PathPrefix("/").Handler(s.GRPCGatewayRouter)
}

func (s *Server) SetTelemetry(m *telemetry.Metrics) {
	s.mtx.Lock()
	s.metrics = m
	s.registerMetrics()
	s.mtx.Unlock()
}

func (s *Server) registerMetrics() {
	metricsHandler := func(w http.ResponseWriter, r *http.Request) {
		format := strings.TrimSpace(r.FormValue("format"))

		gr, err := s.metrics.Gather(format)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to gather metrics: %s", err))
			return
		}

		w.Header().Set("Content-Type", gr.ContentType)
		_, _ = w.Write(gr.Metrics)
	}

	s.Router.HandleFunc("/metrics", metricsHandler).Methods("GET")
}

// errorResponse defines the attributes of a JSON error response.
type errorResponse struct {
	Code  int    `json:"code,omitempty"`
	Error string `json:"error"`
}

// newErrorResponse creates a new errorResponse instance.
func newErrorResponse(code int, err string) errorResponse {
	return errorResponse{Code: code, Error: err}
}

// writeErrorResponse prepares and writes a HTTP error
// given a status code and an error message.
func writeErrorResponse(w http.ResponseWriter, status int, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(legacy.Cdc.MustMarshalJSON(newErrorResponse(0, err)))
}
