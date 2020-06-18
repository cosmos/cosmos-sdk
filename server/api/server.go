package api

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"github.com/tendermint/tendermint/libs/log"
	tmrpcserver "github.com/tendermint/tendermint/rpc/jsonrpc/server"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/cosmos/cosmos-sdk/types/rest"

	// unnamed import of statik for swagger UI support
	_ "github.com/cosmos/cosmos-sdk/client/docs/statik"
)

// Server defines the server's API interface.
type Server struct {
	Router    *mux.Router
	ClientCtx client.Context

	logger   log.Logger
	metrics  *telemetry.Metrics
	listener net.Listener
}

func New(clientCtx client.Context) *Server {
	return &Server{
		Router:    mux.NewRouter(),
		ClientCtx: clientCtx,
		logger:    log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "api-server"),
	}
}

// Start starts the API server. Internally, the API server leverages Tendermint's
// JSON RPC server. Configuration options are provided via config.APIConfig
// and are delegated to the Tendermint JSON RPC server. The process is
// non-blocking, so an external signal handler must be used.
func (s *Server) Start(cfg config.Config) error {
	if cfg.API.Swagger {
		s.registerSwaggerUI()
	}

	if cfg.Telemetry.Enabled {
		m, err := telemetry.New(cfg.Telemetry)
		if err != nil {
			return err
		}

		s.metrics = m
		s.registerMetrics()
	}

	tmCfg := tmrpcserver.DefaultConfig()
	tmCfg.MaxOpenConnections = int(cfg.API.MaxOpenConnections)
	tmCfg.ReadTimeout = time.Duration(cfg.API.RPCReadTimeout) * time.Second
	tmCfg.WriteTimeout = time.Duration(cfg.API.RPCWriteTimeout) * time.Second
	tmCfg.MaxBodyBytes = int64(cfg.API.RPCMaxBodyBytes)

	listener, err := tmrpcserver.Listen(cfg.API.Address, tmCfg)
	if err != nil {
		return err
	}

	s.listener = listener
	var h http.Handler = s.Router

	if cfg.API.EnableUnsafeCORS {
		return tmrpcserver.Serve(s.listener, handlers.CORS()(h), s.logger, tmCfg)
	}

	return tmrpcserver.Serve(s.listener, s.Router, s.logger, tmCfg)
}

// Close closes the API server.
func (s *Server) Close() error {
	return s.listener.Close()
}

func (s *Server) registerSwaggerUI() {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	s.Router.PathPrefix("/").Handler(staticServer)
}

func (s *Server) registerMetrics() {
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
