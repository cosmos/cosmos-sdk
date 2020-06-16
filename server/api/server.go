package api

import (
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"github.com/tendermint/tendermint/libs/log"
	tmrpcserver "github.com/tendermint/tendermint/rpc/jsonrpc/server"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/config"

	// unnamed import of statik for swagger UI support
	_ "github.com/cosmos/cosmos-sdk/client/docs/statik"
)

// Server defines the server's API interface.
type Server struct {
	Router    *mux.Router
	ClientCtx client.Context

	logger   log.Logger
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
func (s *Server) Start(cfg config.APIConfig) error {
	if cfg.Swagger {
		s.registerSwaggerUI()
	}

	tmCfg := tmrpcserver.DefaultConfig()
	tmCfg.MaxOpenConnections = int(cfg.MaxOpenConnections)
	tmCfg.ReadTimeout = time.Duration(cfg.RPCReadTimeout) * time.Second
	tmCfg.WriteTimeout = time.Duration(cfg.RPCWriteTimeout) * time.Second
	tmCfg.MaxBodyBytes = int64(cfg.RPCMaxBodyBytes)

	listener, err := tmrpcserver.Listen(cfg.Address, tmCfg)
	if err != nil {
		return err
	}

	s.listener = listener
	var h http.Handler = s.Router

	if cfg.EnableUnsafeCORS {
		return tmrpcserver.Serve(s.listener, handlers.CORS()(h), s.logger, tmCfg)
	}

	return tmrpcserver.Serve(s.listener, s.Router, s.logger, tmCfg)
}

func (s *Server) registerSwaggerUI() {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	s.Router.PathPrefix("/").Handler(staticServer)
}
