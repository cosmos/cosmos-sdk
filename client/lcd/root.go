package lcd

import (
	"net/http"
	"os"

	"fmt"
	client "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	keys "github.com/cosmos/cosmos-sdk/client/keys"
	rpc "github.com/cosmos/cosmos-sdk/client/rpc"
	tx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/wire"
	auth "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	bank "github.com/cosmos/cosmos-sdk/x/bank/client/rest"
	gov "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
	ibc "github.com/cosmos/cosmos-sdk/x/ibc/client/rest"
	slashing "github.com/cosmos/cosmos-sdk/x/slashing/client/rest"
	stake "github.com/cosmos/cosmos-sdk/x/stake/client/rest"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"github.com/tendermint/tendermint/libs/cli"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"
	tendermintLiteProxy "github.com/tendermint/tendermint/lite/proxy"
	tmserver "github.com/tendermint/tendermint/rpc/lib/server"
	"strings"
)

// ServeCommand will generate a long-running rest server
// (aka Light Client Daemon) that exposes functionality similar
// to the cli, but over rest
func ServeCommand(cdc *wire.Codec) *cobra.Command {
	flagListenAddr := "laddr"
	flagCORS := "cors"
	flagMaxOpenConnections := "max-open"

	cmd := &cobra.Command{
		Use:   "rest-server",
		Short: "Start LCD (light-client daemon), a local REST server",
		RunE: func(cmd *cobra.Command, args []string) error {
			listenAddr := viper.GetString(flagListenAddr)
			handler := createHandler(cdc)
			logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "rest-server")
			maxOpen := viper.GetInt(flagMaxOpenConnections)

			listener, err := tmserver.StartHTTPServer(
				listenAddr, handler, logger,
				tmserver.Config{MaxOpenConnections: maxOpen},
			)
			if err != nil {
				return err
			}

			logger.Info("REST server started")

			// wait forever and cleanup
			cmn.TrapSignal(func() {
				err := listener.Close()
				logger.Error("error closing listener", "err", err)
			})

			return nil
		},
	}

	cmd.Flags().String(flagListenAddr, "tcp://localhost:1317", "The address for the server to listen on")
	cmd.Flags().String(flagCORS, "", "Set the domains that can make CORS requests (* for all)")
	cmd.Flags().String(client.FlagChainID, "", "The chain ID to connect to")
	cmd.Flags().String(client.FlagNode, "tcp://localhost:26657", "Address of the node to connect to")
	cmd.Flags().Int(flagMaxOpenConnections, 1000, "The number of maximum open connections")

	return cmd
}

func createHandler(cdc *wire.Codec) http.Handler {
	r := mux.NewRouter()

	kb, err := keys.GetKeyBase() //XXX
	if err != nil {
		panic(err)
	}

	cliCtx := context.NewCLIContext().WithCodec(cdc).WithLogger(os.Stdout)

	// TODO: make more functional? aka r = keys.RegisterRoutes(r)
	r.HandleFunc("/version", CLIVersionRequestHandler).Methods("GET")
	r.HandleFunc("/node_version", NodeVersionRequestHandler(cliCtx)).Methods("GET")

	keys.RegisterRoutes(r)
	rpc.RegisterRoutes(cliCtx, r)
	tx.RegisterRoutes(cliCtx, r, cdc)
	auth.RegisterRoutes(cliCtx, r, cdc, "acc")
	bank.RegisterRoutes(cliCtx, r, cdc, kb)
	ibc.RegisterRoutes(cliCtx, r, cdc, kb)
	stake.RegisterRoutes(cliCtx, r, cdc, kb)
	slashing.RegisterRoutes(cliCtx, r, cdc, kb)
	gov.RegisterRoutes(cliCtx, r, cdc)

	return r
}

// ServeSwaggerCommand will generate a long-running rest server
// that exposes functionality similar to the ServeCommand, but it provide swagger-ui
// Which is much friendly for further development
func ServeSwaggerCommand(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lite-server",
		Short: "Start Gaia-lite (gaia light client daemon), a local REST server with swagger-ui, default url: http://localhost:1317/swagger/index.html",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).
				With("module", "lite-server")
			listenAddr := viper.GetString(client.FlagListenAddr)
			//Create rest server
			server := gin.New()
			createSwaggerHandler(server, cdc)
			go server.Run(listenAddr)

			logger.Info("REST server started")

			// Wait forever and cleanup
			cmn.TrapSignal(func() {
				logger.Info("Closing rest server...")
			})

			return nil
		},
	}

	cmd.Flags().String(client.FlagListenAddr, "localhost:1317", "Address for server to listen on.")
	cmd.Flags().String(client.FlagNodeList, "tcp://localhost:26657", "Node list to connect to, example: \"tcp://10.10.10.10:26657,tcp://20.20.20.20:26657\".")
	cmd.Flags().String(client.FlagChainID, "", "ID of chain we connect to, must be specified.")
	cmd.Flags().String(client.FlagSwaggerHostIP, "localhost", "The host IP of the Gaia-lite server, swagger-ui will send request to this host.")
	cmd.Flags().String(client.FlagModules, "general,key,bank", "Enabled modules.")
	cmd.Flags().Bool(client.FlagTrustNode, false, "Trust full nodes or not.")

	return cmd
}

func createSwaggerHandler(server *gin.Engine, cdc *wire.Codec) {
	rootDir := viper.GetString(cli.HomeFlag)
	nodeAddrs := viper.GetString(client.FlagNodeList)
	chainID := viper.GetString(client.FlagChainID)
	modules := viper.GetString(client.FlagModules)
	//Get key store
	kb, err := keys.GetKeyBase()
	if err != nil {
		panic(err)
	}
	//Split the node list string into multi full node URIs
	nodeAddrArray := strings.Split(nodeAddrs, ",")
	if len(nodeAddrArray) < 1 {
		panic(fmt.Errorf("missing node URLs"))
	}
	//Tendermint certifier can only connect to one full node. Here we assign the first full node to it
	certifier, err := tendermintLiteProxy.GetCertifier(chainID, rootDir, nodeAddrArray[0])
	if err != nil {
		panic(err)
	}
	//Create load balancing engine
	clientManager, err := context.NewClientManager(nodeAddrs)
	if err != nil {
		panic(err)
	}
	//Assign tendermint certifier and load balancing engine to ctx
	ctx := context.NewCLIContext().WithCodec(cdc).WithLogger(os.Stdout).WithCertifier(certifier).WithClientManager(clientManager)

	server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	moduleArray := strings.Split(modules, ",")
	if utils.StringArrayContains(moduleArray, "general") {
		server.GET("/version", CLIVersionRequest)
		server.GET("/node_version", NodeVersionRequest(ctx))
	}

	// This API belongs to ICS20, later it will be moved under key management category
	auth.RegisterSwaggerRoutes(server.Group("/"), ctx, cdc, "acc")

	if utils.StringArrayContains(moduleArray, "transaction") {
		tx.RegisterSwaggerRoutes(server.Group("/"), ctx, cdc)
	}

	if utils.StringArrayContains(moduleArray, "bank") {
		bank.RegisterSwaggerRoutes(server.Group("/"), ctx, cdc, kb)
	}
}
