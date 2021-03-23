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
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/rpc/client/local"
	tmrpcserver "github.com/tendermint/tendermint/rpc/jsonrpc/server"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	keybase "github.com/cosmos/cosmos-sdk/crypto/keys"

	// unnamed import of statik for swagger UI support
	_ "github.com/cosmos/cosmos-sdk/client/lcd/statik"
)

// RestServer represents the Light Client Rest server
type RestServer struct {
	Mux     *mux.Router
	CliCtx  context.CLIContext
	KeyBase keybase.Keybase
	Cdc     *codec.Codec

	log      log.Logger
	listener net.Listener
}

// NewRestServer creates a new rest server instance
func NewRestServer(cdc *codec.Codec, tmNode *node.Node) *RestServer {
	rootRouter := mux.NewRouter()
	cliCtx := context.NewCLIContext().WithCodec(cdc)
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "rest-server")
	if tmNode != nil {
		cliCtx.Client = local.New(tmNode)
		logger = tmNode.Logger.With("module", "rest-server")
	}

	cliCtx.TrustNode = true

	return &RestServer{
		Mux:    rootRouter,
		CliCtx: cliCtx,
		Cdc:    cdc,

		log: logger,
	}
}

func (rs *RestServer) Logger() log.Logger {
	return rs.log
}

// Start starts the rest server
func (rs *RestServer) Start(listenAddr string, maxOpen int, readTimeout, writeTimeout uint, cors bool) (err error) {
	//trapSignal(func() {
	//	err := rs.listener.Close()
	//	rs.log.Error("error closing listener", "err", err)
	//})

	cfg := tmrpcserver.DefaultConfig()
	cfg.MaxOpenConnections = maxOpen
	cfg.ReadTimeout = time.Duration(readTimeout) * time.Second
	cfg.WriteTimeout = time.Duration(writeTimeout) * time.Second

	rs.listener, err = tmrpcserver.Listen(listenAddr, cfg)
	if err != nil {
		return
	}
	rs.log.Info(
		fmt.Sprintf(
			"Starting application REST service (chain-id: %q)...",
			viper.GetString(flags.FlagChainID),
		),
	)

	var h http.Handler = rs.Mux
	if cors {
		allowAllCORS := handlers.CORS(handlers.AllowedHeaders([]string{"Content-Type"}))
		h = allowAllCORS(h)
	}

	return tmrpcserver.Serve(rs.listener, h, rs.log, cfg)
}

// ServeCommand will start the application REST service as a blocking process. It
// takes a codec to create a RestServer object and a function to register all
// necessary routes.
func ServeCommand(cdc *codec.Codec, registerRoutesFn func(*RestServer)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rest-server",
		Short: "Start LCD (light-client daemon), a local REST server",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			rs := NewRestServer(cdc, nil)

			registerRoutesFn(rs)
			rs.registerSwaggerUI()

			// Start the rest server and return error if one exists
			err = rs.Start(
				viper.GetString(flags.FlagListenAddr),
				viper.GetInt(flags.FlagMaxOpenConnections),
				uint(viper.GetInt(flags.FlagRPCReadTimeout)),
				uint(viper.GetInt(flags.FlagRPCWriteTimeout)),
				viper.GetBool(flags.FlagUnsafeCORS),
			)

			return err
		},
	}

	return flags.RegisterRestServerFlags(cmd)
}

func StartRestServer(cdc *codec.Codec, registerRoutesFn func(*RestServer), tmNode *node.Node, addr string) error {
	rs := NewRestServer(cdc, tmNode)

	registerRoutesFn(rs)
	rs.registerSwaggerUI()
	rs.log.Info("start rest server")
	// Start the rest server and return error if one exists
	err := rs.Start(
		addr,
		viper.GetInt(flags.FlagMaxOpenConnections),
		uint(viper.GetInt(flags.FlagRPCReadTimeout)),
		uint(viper.GetInt(flags.FlagRPCWriteTimeout)),
		viper.GetBool(flags.FlagUnsafeCORS),
	)

	return err
}

func (rs *RestServer) registerSwaggerUI() {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}
	staticServer := http.FileServer(statikFS)
	rs.Mux.PathPrefix("/swagger-ui/").Handler(http.StripPrefix("/swagger-ui/", staticServer))
}
