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

    serverCfg := srv.Config().(*Config)
    if len(cfg) > 0 {
        if err := serverv2.UnmarshalSubConfig(cfg, srv.Name(), &serverCfg); err != nil {
            return nil, fmt.Errorf("failed to unmarshal config: %w", err)
        }
    }
    srv.config = serverCfg

    if err := srv.config.Validate(); err != nil {
        return nil, err
    }

    mux := http.NewServeMux()
    mux.Handle("/swagger/", &swaggerHandler{
        swaggerFS: srv.config.SwaggerUI,
    })

    srv.server = &http.Server{
        Addr:    srv.config.Address,
        Handler: mux,
    }

    return srv, nil
} 
