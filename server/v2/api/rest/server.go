package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/server/v2/appmanager"
)

const (
	ServerName = "rest"
)

type Server[T transaction.Tx] struct {
	logger     log.Logger
	router     *http.ServeMux
	httpServer *http.Server
	config     *Config
	cfgOptions []CfgOption
}

func New[T transaction.Tx](
	logger log.Logger,
	appManager appmanager.AppManager[T],
	cfg server.ConfigMap,
	cfgOptions ...CfgOption,
) (*Server[T], error) {
	srv := &Server[T]{
		logger:     logger.With(log.ModuleKey, ServerName),
		cfgOptions: cfgOptions,
		router:     http.NewServeMux(),
	}

	srv.router.Handle("/", NewDefaultHandler(appManager))

	serverCfg := srv.Config().(*Config)
	if len(cfg) > 0 {
		if err := serverv2.UnmarshalSubConfig(cfg, srv.Name(), &serverCfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}
	srv.config = serverCfg
	srv.httpServer = &http.Server{
		Addr:    srv.config.Address,
		Handler: srv.router,
	}
	return srv, nil
}

// NewWithConfigOptions creates a new REST server with the provided config options.
// It is *not* a fully functional server (since it has been created without dependencies)
// The returned server should only be used to get and set configuration.
func NewWithConfigOptions[T transaction.Tx](opts ...CfgOption) *Server[T] {
	return &Server[T]{
		cfgOptions: opts,
	}
}

func (s *Server[T]) Name() string {
	return ServerName
}

func (s *Server[T]) Start(ctx context.Context) error {
	if !s.config.Enable {
		s.logger.Info(fmt.Sprintf("%s server is disabled via config", s.Name()))
		return nil
	}

	s.logger.Info("starting HTTP server", "address", s.config.Address)
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error("failed to start HTTP server", "error", err)
		return err
	}

	return nil
}

func (s *Server[T]) Stop(ctx context.Context) error {
	if !s.config.Enable {
		return nil
	}

	s.logger.Info("stopping HTTP server")
	return s.httpServer.Shutdown(ctx)
}

func (s *Server[T]) Config() any {
	if s.config == nil || s.config.Address == "" {
		cfg := DefaultConfig()

		for _, opt := range s.cfgOptions {
			opt(cfg)
		}

		return cfg
	}

	return s.config
}
