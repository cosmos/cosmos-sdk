package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
	config  *Config
	logger  log.Logger
	server  *http.Server
	metrics *Metrics
}

// New creates a new telemetry server.
func New[T transaction.Tx]() *Server[T] {
	return &Server[T]{}
}

// Name returns the server name.
func (s *Server[T]) Name() string {
	return ServerName
}

func (s *Server[T]) Config() any {
	if s.config == nil || s.config.Address == "" {
		return DefaultConfig()
	}

	return s.config
}

// Init implements serverv2.ServerComponent.
func (s *Server[T]) Init(appI serverv2.AppI[T], cfg map[string]any, logger log.Logger) error {
	serverCfg := s.Config().(*Config)
	if len(cfg) > 0 {
		if err := serverv2.UnmarshalSubConfig(cfg, s.Name(), &serverCfg); err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}
	s.config = serverCfg
	s.logger = logger.With(log.ModuleKey, s.Name())

	metrics, err := NewMetrics(s.config)
	if err != nil {
		return fmt.Errorf("failed to initialize metrics: %w", err)
	}
	s.metrics = metrics

	return nil
}

func (s *Server[T]) Start(ctx context.Context) error {
	if !s.config.Enable {
		s.logger.Info(fmt.Sprintf("%s server is disabled via config", s.Name()))
		return nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.metricsHandler)
	// keeping /metrics for backwards compatibility
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	})

	s.server = &http.Server{
		Addr:    s.config.Address,
		Handler: mux,
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
