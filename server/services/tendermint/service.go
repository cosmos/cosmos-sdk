package tendermint

import (
	"net"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/tendermint/tendermint/libs/log"
	tmrpcserver "github.com/tendermint/tendermint/rpc/jsonrpc/server"

	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/server/types"
)

var _ types.Service = &Service{}

type Service struct {
	listener net.Listener
	logger   log.Logger
	router   *mux.Router
}

// NewService returns a new Tendermint RPC service instance
func NewService(logger log.Logger, router *mux.Router) *Service {
	return &Service{
		logger: logger.With("service", "tendermint-rpc"),
		router: router,
	}
}

// Name is the tendermint RPC service name
func (Service) Name() string {
	return "Tendermint RPC"
}

// RegisterRoutes performs a no-op (?)
func (s *Service) RegisterRoutes() bool {
	// app.RegisterAPIRoutes(srv.BaseServer(), sdkCfg.API)
	return true
}

func (s *Service) Start(cfg config.ServerConfig) error {
	sdkCfg := cfg.GetSDKConfig()

	if !sdkCfg.API.Enable {
		return nil
	}

	tmCfg := tmrpcserver.DefaultConfig()
	tmCfg.MaxOpenConnections = int(sdkCfg.API.MaxOpenConnections)
	tmCfg.ReadTimeout = time.Duration(sdkCfg.API.RPCReadTimeout) * time.Second
	tmCfg.WriteTimeout = time.Duration(sdkCfg.API.RPCWriteTimeout) * time.Second
	tmCfg.MaxBodyBytes = int64(sdkCfg.API.RPCMaxBodyBytes)

	var err error
	s.listener, err = tmrpcserver.Listen(sdkCfg.API.Address, tmCfg)
	if err != nil {
		return err
	}

	var handler http.Handler = s.router
	if sdkCfg.API.EnableUnsafeCORS {
		allowAllCORS := handlers.CORS(handlers.AllowedHeaders([]string{"Content-Type"}))
		handler = allowAllCORS(s.router)
	}

	s.logger.Info("starting RPC service...")

	errCh := make(chan error)
	go func() {
		if err := tmrpcserver.Serve(s.listener, handler, s.logger, tmCfg); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-time.After(5 * time.Second): // assume server started successfully
		return nil
	}
}

// Stop closes the Tendermint RPC listener connection.
func (s *Service) Stop() error {
	return s.listener.Close()
}
