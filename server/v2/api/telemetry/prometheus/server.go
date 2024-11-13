package prometheus

import (
	"context"
	"cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/server/v2/api/telemetry"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

const ServerName = "telemetry"

type Server[T transaction.Tx] struct {
	config     *telemetry.Config
	logger     log.Logger
	httpServer *http.Server
}

func New[T transaction.Tx](cfg server.ConfigMap, logger log.Logger) (*Server[T], error) {
	srv := &Server[T]{}
	serverCfg := srv.Config().(*telemetry.Config)
	if len(cfg) > 0 {
		if err := serverv2.UnmarshalSubConfig(cfg, srv.Name(), &serverCfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}
	srv.logger = logger.With(log.ModuleKey, srv.Name())
	srv.config = serverCfg
	return srv, nil
}

func (s *Server[T]) Name() string {
	return ServerName
}

func (s *Server[T]) Config() any {
	if s.config == nil || s.config.Address == "" {
		return telemetry.DefaultConfig()
	}

	return s.config
}

func (s *Server[T]) Start(context.Context) error {
	if !s.config.Enable {
		s.logger.Info(fmt.Sprintf("%s server is disabled via config", s.Name()))
		return nil
	}
	s.httpServer = &http.Server{Addr: ":2112"}
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := s.httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error(fmt.Sprintf("ListenAndServe(): %v", err))
		}
	}()
	return nil
}

func (s *Server[T]) Stop(ctx context.Context) error {
	if !s.config.Enable || s.httpServer == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}
