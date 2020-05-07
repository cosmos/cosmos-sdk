package lcd

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/log"
	rpcserver "github.com/tendermint/tendermint/rpc/lib/server"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"

	// unnamed import of statik for swagger UI support
	_ "github.com/cosmos/cosmos-sdk/client/lcd/statik"
)

// RestServer represents the Light Client Rest server
type RestServer struct {
	Mux    *mux.Router
	CliCtx context.CLIContext

	log      log.Logger
	listener net.Listener
}

// NewRestServer creates a new rest server instance
func NewRestServer(cdc *codec.Codec) *RestServer {
	r := mux.NewRouter()
	cliCtx := context.NewCLIContext().WithCodec(cdc)
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "rest-server")

	return &RestServer{
		Mux:    r,
		CliCtx: cliCtx,
		log:    logger,
	}
}

// StartWithConfig starts the REST server that listens on the provided listenAddr.
// It will use the provided RPC configuration.
func (rs *RestServer) StartWithConfig(listenAddr string, cors bool, cfg *rpcserver.Config) error {
	server.TrapSignal(func() {
		err := rs.listener.Close()
		rs.log.Error("error closing listener", "err", err)
	})

	listener, err := rpcserver.Listen(listenAddr, cfg)
	if err != nil {
		return err
	}

	rs.listener = listener

	rs.log.Info(
		fmt.Sprintf("Starting application REST service (chain-id: %q)...", viper.GetString(flags.FlagChainID)),
	)

	var h http.Handler = rs.Mux

	if cors {
		return rpcserver.StartHTTPServer(rs.listener, handlers.CORS()(h), rs.log, cfg)
	}

	return rpcserver.StartHTTPServer(rs.listener, rs.Mux, rs.log, cfg)
}

// Start starts the REST server that listens on the provided listenAddr. The REST
// service will use Tendermint's default RPC configuration, where the R/W timeout
// and max open connections are overridden.
func (rs *RestServer) Start(listenAddr string, maxOpen int, readTimeout, writeTimeout uint, cors bool) error {
	cfg := rpcserver.DefaultConfig()
	cfg.MaxOpenConnections = maxOpen
	cfg.ReadTimeout = time.Duration(readTimeout) * time.Second
	cfg.WriteTimeout = time.Duration(writeTimeout) * time.Second

	return rs.StartWithConfig(listenAddr, cors, cfg)
}

// ServeCommand will start the application REST service as a blocking process. It
// takes a codec to create a RestServer object and a function to register all
// necessary routes.
func ServeCommand(cdc *codec.Codec, registerRoutesFn func(*RestServer)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rest-server",
		Short: "Start LCD (light-client daemon), a local REST server",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			rs := NewRestServer(cdc)

			registerRoutesFn(rs)
			rs.registerSwaggerUI()

			cfg := rpcserver.DefaultConfig()
			cfg.MaxOpenConnections = viper.GetInt(flags.FlagMaxOpenConnections)
			cfg.ReadTimeout = time.Duration(viper.GetInt64(flags.FlagRPCReadTimeout)) * time.Second
			cfg.WriteTimeout = time.Duration(viper.GetInt64(flags.FlagRPCWriteTimeout)) * time.Second
			cfg.MaxBodyBytes = viper.GetInt64(flags.FlagRPCMaxBodyBytes)

			// start the rest server and return error if one exists
			return rs.StartWithConfig(
				viper.GetString(flags.FlagListenAddr),
				viper.GetBool(flags.FlagUnsafeCORS),
				cfg,
			)
		},
	}

	return flags.RegisterRestServerFlags(cmd)
}

func (rs *RestServer) registerSwaggerUI() {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rs.Mux.PathPrefix("/").Handler(staticServer)
}
