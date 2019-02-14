package lcd

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/log"
	rpcserver "github.com/tendermint/tendermint/rpc/lib/server"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	keybase "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/server"

	// Import statik for light client stuff
	_ "github.com/cosmos/cosmos-sdk/client/lcd/statik"
)

// RestServer represents the Light Client Rest server
type RestServer struct {
	Router  *mux.Router
	CliCtx  context.CLIContext
	KeyBase keybase.Keybase
	Cdc     *codec.Codec

	log         log.Logger
	listener    net.Listener
	fingerprint string
}

// NewRestServer creates a new rest server instance
func NewRestServer(cdc *codec.Codec) *RestServer {
	router := mux.NewRouter()
	cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "rest-server")

	return &RestServer{
		Router: router,
		CliCtx: cliCtx,
		Cdc:    cdc,
		log:    logger,
	}
}

// Start starts the rest server.
func (rs *RestServer) Start(
	listenAddr string, sslHosts string, certFile string, keyFile string, maxOpen int, secure bool,
) (err error) {

	server.TrapSignal(func() {
		err := rs.listener.Close()
		rs.log.Error("error closing listener", "err", err)
	})

	rs.listener, err = rpcserver.Listen(
		listenAddr,
		rpcserver.Config{MaxOpenConnections: maxOpen},
	)
	if err != nil {
		return
	}

	rs.log.Info(
		fmt.Sprintf(
			"Starting Gaia REST service (chain-id: %s)...",
			viper.GetString(client.FlagChainID),
		),
	)

	// launch rest-server in insecure mode
	if !secure {
		return rpcserver.StartHTTPServer(rs.listener, rs.Router, rs.log)
	}

	// handle certificates
	if certFile != "" {
		// validateCertKeyFiles() is needed to work around tendermint/tendermint#2460
		if err := validateCertKeyFiles(certFile, keyFile); err != nil {
			return err
		}

		// cert/key pair is provided, read the fingerprint
		rs.fingerprint, err = fingerprintFromFile(certFile)
		if err != nil {
			return err
		}
	} else {
		// if certificate is not supplied, generate a self-signed one
		certFile, keyFile, rs.fingerprint, err = genCertKeyFilesAndReturnFingerprint(sslHosts)
		if err != nil {
			return err
		}

		defer func() {
			os.Remove(certFile)
			os.Remove(keyFile)
		}()
	}

	rs.log.Info(rs.fingerprint)
	return rpcserver.StartHTTPAndTLSServer(
		rs.listener,
		rs.Router,
		certFile, keyFile,
		rs.log,
	)
}

// ServeCommand will start a Gaia Lite REST service as a blocking process. It
// takes a codec to create a RestServer object and a function to register all
// necessary routes.
func ServeCommand(cdc *codec.Codec, registerRoutesFn func(*RestServer)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rest-server",
		Short: "Start a local client REST server",
		Long: `Start a local client REST server listening on "--laddr". The REST client
will forward requests to a Tendermint RPC node defined by "--node". By default the
REST client will not expose unsafe routes, which are enabled by "--allow-unsafe".
Unsafe routes are defined as routes that sign txs or that expose any addresses.

The REST client may also accept a SSL/TLS certificate enabling a secure connection.
By default, the REST server will not trust the full node it is connected to, requiring
the client to verify proofs. This can be bypassed via the "--trust-node" flag.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			rs := NewRestServer(cdc)

			registerRoutesFn(rs)

			// start the rest server and return error if one exists
			return rs.Start(
				viper.GetString(client.FlagListenAddr),
				viper.GetString(client.FlagSSLHosts),
				viper.GetString(client.FlagSSLCertFile),
				viper.GetString(client.FlagSSLKeyFile),
				viper.GetInt(client.FlagMaxOpenConnections),
				viper.GetBool(client.FlagTLS),
			)
		},
	}

	return client.RegisterRestServerFlags(cmd)
}

func (rs *RestServer) registerSwaggerUI() {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}
	staticServer := http.FileServer(statikFS)
	rs.Router.PathPrefix("/swagger-ui/").Handler(http.StripPrefix("/swagger-ui/", staticServer))
}

func validateCertKeyFiles(certFile, keyFile string) error {
	if keyFile == "" {
		return errors.New("a key file is required")
	}
	if _, err := os.Stat(certFile); err != nil {
		return err
	}
	if _, err := os.Stat(keyFile); err != nil {
		return err
	}
	return nil
}
