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

type Server[T transaction.Tx] struct {
    logger     log.Logger
    config     *Config
    cfgOptions []CfgOption
    server     *http.Server
}

func New[T transaction.Tx](
    logger log.Logger,
    cfg server.ConfigMap,
    cfgOptions ...CfgOption,
) (*Server[T], error) {
    srv := &Server[T]{
        logger:     logger.With(log.ModuleKey, ServerName),
        cfgOptions: cfgOptions,
    }

    serverCfg := srv.Config().(*Config)
    if len(cfg) > 0 {
        if err := serverv2.UnmarshalSubConfig(cfg, srv.Name(), &serverCfg); err != nil {
            return nil, fmt.Errorf("failed to unmarshal config: %w", err)
        }
    }
    srv.config = serverCfg

    mux := http.NewServeMux()
    mux.Handle(srv.config.Path, NewSwaggerHandler())

    srv.server = &http.Server{
        Addr:    srv.config.Address,
        Handler: mux,
    }

    return srv, nil
}

func (s *Server[T]) Name() string {
    return ServerName
}

func (s *Server[T]) Start(ctx context.Context) error {
    if !s.config.Enable {
        s.logger.Info("swagger server is disabled via config")
        return nil
    }

    s.logger.Info("starting swagger server...", "address", s.config.Address)
    if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        return fmt.Errorf("failed to start swagger server: %w", err)
    }

    return nil
}

func (s *Server[T]) Stop(ctx context.Context) error {
    if !s.config.Enable {
        return nil
    }

    s.logger.Info("stopping swagger server...", "address", s.config.Address)
    return s.server.Shutdown(ctx)
}
