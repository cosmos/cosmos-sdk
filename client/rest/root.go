package rest

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	httpswagger "github.com/swaggo/http-swagger"
	"github.com/tendermint/tendermint/libs/log"
	rpcserver "github.com/tendermint/tendermint/rpc/lib/server"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/rest/docs"
	"github.com/cosmos/cosmos-sdk/codec"
	keybase "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/server"
)

// Swagger main API variables set at build-time. Applications may choose to
// populate these values via ldflags or manually setting them.
var (
	SwaggerHost        string
	SwaggerVersion     string
	SwaggerBasePath    string
	SwaggerTitle       string
	SwaggerDescription string
)

// RestServer represents the Light Client Rest server
type RestServer struct {
	Mux     *mux.Router
	CliCtx  context.CLIContext
	KeyBase keybase.Keybase

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

// Start starts the rest server
func (rs *RestServer) Start(listenAddr string, maxOpen int, readTimeout, writeTimeout uint) (err error) {
	server.TrapSignal(func() {
		err := rs.listener.Close()
		rs.log.Error("error closing listener", "err", err)
	})

	cfg := rpcserver.DefaultConfig()
	cfg.MaxOpenConnections = maxOpen
	cfg.ReadTimeout = time.Duration(readTimeout) * time.Second
	cfg.WriteTimeout = time.Duration(writeTimeout) * time.Second

	rs.listener, err = rpcserver.Listen(listenAddr, cfg)
	if err != nil {
		return
	}
	rs.log.Info(
		fmt.Sprintf(
			"Starting application REST service (chain-id: %q)...",
			viper.GetString(flags.FlagChainID),
		),
	)

	return rpcserver.StartHTTPServer(rs.listener, rs.Mux, rs.log, cfg)
}

func (rs *RestServer) registerSwaggerDocs() {
	// set Swagger main API values from build variables
	docs.SwaggerInfo.Host = SwaggerHost
	docs.SwaggerInfo.Version = SwaggerVersion
	docs.SwaggerInfo.BasePath = SwaggerBasePath
	docs.SwaggerInfo.Title = SwaggerTitle
	docs.SwaggerInfo.Description = SwaggerDescription

	rs.Mux.PathPrefix("/swagger/").Handler(httpswagger.WrapHandler)
}

// ServeCommand will start the application REST service as a blocking process. It
// takes a codec to create a RestServer object and a function to register all
// necessary routes.
func ServeCommand(cdc *codec.Codec, registerRoutesFn func(*RestServer)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rest-server",
		Short: "Start a local REST server",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			rs := NewRestServer(cdc)

			// mount client provided handlers
			registerRoutesFn(rs)

			// mount pre-compiled REST (Swagger) documentation
			rs.registerSwaggerDocs()

			// Start the rest server and return error if one exists
			err = rs.Start(
				viper.GetString(flags.FlagListenAddr),
				viper.GetInt(flags.FlagMaxOpenConnections),
				uint(viper.GetInt(flags.FlagRPCReadTimeout)),
				uint(viper.GetInt(flags.FlagRPCWriteTimeout)),
			)

			return err
		},
	}

	return flags.RegisterRestServerFlags(cmd)
}
