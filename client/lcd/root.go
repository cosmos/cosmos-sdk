package lcd

import (
	"errors"
	"net"
	"net/http"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	auth "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	bank "github.com/cosmos/cosmos-sdk/x/bank/client/rest"
	gov "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
	slashing "github.com/cosmos/cosmos-sdk/x/slashing/client/rest"
	stake "github.com/cosmos/cosmos-sdk/x/stake/client/rest"
	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"
	tmserver "github.com/tendermint/tendermint/rpc/lib/server"
)

const (
	flagListenAddr         = "laddr"
	flagCORS               = "cors"
	flagMaxOpenConnections = "max-open"
	flagInsecure           = "insecure"
	flagSSLHosts           = "ssl-hosts"
	flagSSLCertFile        = "ssl-certfile"
	flagSSLKeyFile         = "ssl-keyfile"
)

// ServeCommand will generate a long-running rest server
// (aka Light Client Daemon) that exposes functionality similar
// to the cli, but over rest
func ServeCommand(cdc *codec.Codec) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "rest-server",
		Short: "Start LCD (light-client daemon), a local REST server",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			listenAddr := viper.GetString(flagListenAddr)
			handler := createHandler(cdc)
			logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "rest-server")
			maxOpen := viper.GetInt(flagMaxOpenConnections)
			sslHosts := viper.GetString(flagSSLHosts)
			certFile := viper.GetString(flagSSLCertFile)
			keyFile := viper.GetString(flagSSLKeyFile)
			cleanupFunc := func() {}

			var listener net.Listener
			var fingerprint string
			if viper.GetBool(flagInsecure) {
				listener, err = tmserver.StartHTTPServer(
					listenAddr, handler, logger,
					tmserver.Config{MaxOpenConnections: maxOpen},
				)
				if err != nil {
					return
				}
			} else {
				if certFile != "" {
					// validateCertKeyFiles() is needed to work around tendermint/tendermint#2460
					err = validateCertKeyFiles(certFile, keyFile)
					if err != nil {
						return err
					}
					//  cert/key pair is provided, read the fingerprint
					fingerprint, err = fingerprintFromFile(certFile)
					if err != nil {
						return err
					}
				} else {
					// if certificate is not supplied, generate a self-signed one
					certFile, keyFile, fingerprint, err = genCertKeyFilesAndReturnFingerprint(sslHosts)
					if err != nil {
						return err
					}
					cleanupFunc = func() {
						os.Remove(certFile)
						os.Remove(keyFile)
					}
					defer cleanupFunc()
				}
				listener, err = tmserver.StartHTTPAndTLSServer(
					listenAddr, handler,
					certFile, keyFile,
					logger,
					tmserver.Config{MaxOpenConnections: maxOpen},
				)
				if err != nil {
					return
				}
				logger.Info(fingerprint)
			}
			logger.Info("REST server started")

			// wait forever and cleanup
			cmn.TrapSignal(func() {
				defer cleanupFunc()
				err := listener.Close()
				logger.Error("error closing listener", "err", err)
			})

			return nil
		},
	}

	cmd.Flags().String(flagListenAddr, "tcp://localhost:1317", "The address for the server to listen on")
	cmd.Flags().Bool(flagInsecure, false, "Do not set up SSL/TLS layer")
	cmd.Flags().String(flagSSLHosts, "", "Comma-separated hostnames and IPs to generate a certificate for")
	cmd.Flags().String(flagSSLCertFile, "", "Path to a SSL certificate file. If not supplied, a self-signed certificate will be generated.")
	cmd.Flags().String(flagSSLKeyFile, "", "Path to a key file; ignored if a certificate file is not supplied.")
	cmd.Flags().String(flagCORS, "", "Set the domains that can make CORS requests (* for all)")
	cmd.Flags().String(client.FlagChainID, "", "Chain ID of Tendermint node")
	cmd.Flags().String(client.FlagNode, "tcp://localhost:26657", "Address of the node to connect to")
	cmd.Flags().Int(flagMaxOpenConnections, 1000, "The number of maximum open connections")
	cmd.Flags().Bool(client.FlagTrustNode, false, "Trust connected full node (don't verify proofs for responses)")
	viper.BindPFlag(client.FlagTrustNode, cmd.Flags().Lookup(client.FlagTrustNode))
	viper.BindPFlag(client.FlagChainID, cmd.Flags().Lookup(client.FlagChainID))
	viper.BindPFlag(client.FlagNode, cmd.Flags().Lookup(client.FlagNode))

	return cmd
}

func createHandler(cdc *codec.Codec) http.Handler {
	r := mux.NewRouter()

	kb, err := keys.GetKeyBase() //XXX
	if err != nil {
		panic(err)
	}

	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}
	staticServer := http.FileServer(statikFS)
	r.PathPrefix("/swagger-ui/").Handler(http.StripPrefix("/swagger-ui/", staticServer))

	cliCtx := context.NewCLIContext().WithCodec(cdc)


	// TODO: make more functional? aka r = keys.RegisterRoutes(r)
	r.HandleFunc("/version", CLIVersionRequestHandler).Methods("GET")
	r.HandleFunc("/node_version", NodeVersionRequestHandler(cliCtx)).Methods("GET")

	keys.RegisterRoutes(r)
	rpc.RegisterRoutes(cliCtx, r)
	tx.RegisterRoutes(cliCtx, r, cdc)
	auth.RegisterRoutes(cliCtx, r, cdc, "acc")
	bank.RegisterRoutes(cliCtx, r, cdc, kb)
	stake.RegisterRoutes(cliCtx, r, cdc, kb)
	slashing.RegisterRoutes(cliCtx, r, cdc, kb)
	gov.RegisterRoutes(cliCtx, r, cdc)

	return r
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
