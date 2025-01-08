package swagger

import (
    "context"
    "fmt"
    "net/http"

    "cosmossdk.io/core/server"
    "cosmossdk.io/core/transaction"
    "cosmossdk.io/log"
    serverv2 "cosmossdk.io/server/v2"
)

var (
    _ serverv2.ServerComponent[transaction.Tx] = (*Server[transaction.Tx])(nil)
    _ serverv2.HasConfig                       = (*Server[transaction.Tx])(nil)
)

// Server represents a Swagger UI server
type Server[T transaction.Tx] struct {
    logger     log.Logger
    config     *Config
    cfgOptions []CfgOption
    server     *http.Server
}

// New creates a new Swagger UI server
func New[T transaction.Tx](
    logger log.Logger,
    cfg server.ConfigMap,
    cfgOptions ...CfgOption,
) (*Server[T], error) {
    srv := &Server[T]{
        logger:     logger.With(log.ModuleKey, ServerName),
        cfgOptions: cfgOptions,
    }

    serverCfg := DefaultConfig()
    if len(cfg) > 0 {
        if err := serverv2.UnmarshalSubConfig(cfg, ServerName, serverCfg); err != nil {
            return nil, fmt.Errorf("failed to unmarshal config: %w", err)
        }
    }
    for _, opt := range cfgOptions {
        opt(serverCfg)
    }
    srv.config = serverCfg

    if err := srv.config.Validate(); err != nil {
        return nil, err
    }

    mux := http.NewServeMux()
    mux.Handle("/swagger", &swaggerHandler{
        swaggerFS: srv.config.SwaggerUI,
    })

    srv.server = &http.Server{
        Addr:    srv.config.Address,
        Handler: mux,
    }

    return srv, nil
}

// Name returns the server's name
func (s *Server[T]) Name() string {
    return ServerName
}

// Config returns the server configuration
func (s *Server[T]) Config() any {
    if s.config == nil {
        cfg := DefaultConfig()
        for _, opt := range s.cfgOptions {
            opt(cfg)
        }
        return cfg
    }
    return s.config
}

// Start starts the server
func (s *Server[T]) Start(ctx context.Context) error {
    if !s.config.Enable {
        s.logger.Info(fmt.Sprintf("%s server is disabled via config", s.Name()))
        return nil
    }

    s.logger.Info("starting swagger server...", "address", s.config.Address)
    if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        return fmt.Errorf("failed to start swagger server: %w", err)
    }

    return nil
}

// Stop stops the server
func (s *Server[T]) Stop(ctx context.Context) error {
    if !s.config.Enable {
        return nil
    }

    s.logger.Info("stopping swagger server...", "address", s.config.Address)
    return s.server.Shutdown(ctx)
} 
