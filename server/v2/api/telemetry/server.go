package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
)

var (
	_ serverv2.ServerComponent[transaction.Tx] = (*Server[transaction.Tx])(nil)
	_ serverv2.HasConfig                       = (*Server[transaction.Tx])(nil)
)

const ServerName = "telemetry"

type Server[T transaction.Tx] struct {
	logger     log.Logger
	config     *Config
	cfgOptions []CfgOption
	server     *http.Server
	metrics    *Metrics
}

// New creates a new telemetry server.
func New[T transaction.Tx](cfg server.ConfigMap, logger log.Logger, enableTelemetry func(), cfgOptions ...CfgOption) (*Server[T], error) {
	srv := &Server[T]{}
	serverCfg := srv.Config().(*Config)
	if len(cfg) > 0 {
		if err := serverv2.UnmarshalSubConfig(cfg, srv.Name(), &serverCfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}
	srv.config = serverCfg
	srv.cfgOptions = cfgOptions
	srv.logger = logger.With(log.ModuleKey, srv.Name())

	if enableTelemetry == nil {
		panic("enableTelemetry must be provided")
	}

	if srv.config.Enable {
		enableTelemetry()
	}

	metrics, err := NewMetrics(srv.config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}
	srv.metrics = metrics
	mux := http.NewServeMux()
	// /metrics is the default standard path for Prometheus metrics.
	mux.HandleFunc("/metrics", srv.metricsHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/metrics", http.StatusMovedPermanently)
	})

	srv.server = &http.Server{
		Addr:    srv.config.Address,
		Handler: mux,
	}
	return srv, nil
}

// NewWithConfigOptions creates a new telemetry server with the provided config options.
// It is *not* a fully functional server (since it has been created without dependencies)
// The returned server should only be used to get and set configuration.
func NewWithConfigOptions[T transaction.Tx](opts ...CfgOption) *Server[T] {
	return &Server[T]{
		cfgOptions: opts,
	}
}

// Name returns the server name.
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

	s.logger.Info("starting telemetry server...", "address", s.config.Address)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start telemetry server: %w", err)
	}

	return nil
}

func (s *Server[T]) Stop(ctx context.Context) error {
	if !s.config.Enable || s.server == nil {
		return nil
	}

	s.logger.Info("stopping telemetry server...", "address", s.config.Address)
	return s.server.Shutdown(ctx)
}

func (s *Server[T]) metricsHandler(w http.ResponseWriter, r *http.Request) {
	format := strings.TrimSpace(r.FormValue("format"))

	// errorResponse defines the attributes of a JSON error response.
	type errorResponse struct {
		Code  int    `json:"code,omitempty"`
		Error string `json:"error"`
	}

	gr, err := s.metrics.Gather(format)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		bz, err := json.Marshal(errorResponse{Code: 400, Error: fmt.Sprintf("failed to gather metrics: %s", err)})
		if err != nil {
			return
		}
		_, _ = w.Write(bz)

		return
	}

	w.Header().Set("Content-Type", gr.ContentType)
	_, _ = w.Write(gr.Metrics)
}
