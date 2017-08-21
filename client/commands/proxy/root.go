package proxy

import (
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	certclient "github.com/tendermint/light-client/certifiers/client"
	"github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/rpc/core"
	rpc "github.com/tendermint/tendermint/rpc/lib/server"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/client/commands"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Run proxy server, verifying tendermint rpc",
	Long: `This node will run a secure proxy to a tendermint rpc server.

All calls that can be tracked back to a block header by a proof
will be verified before passing them back to the caller. Other that
that it will present the same interface as a full tendermint node,
just with added trust and running locally.`,
	RunE:         commands.RequireInit(runProxy),
	SilenceUsage: true,
}

const (
	bindFlag   = "serve"
	wsEndpoint = "/websocket"
)

func init() {
	RootCmd.Flags().String(bindFlag, ":8888", "Serve the proxy on the given port")
}

// TODO: pass in a proper logger
var logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))

func init() {
	logger = logger.With("module", "main")
	logger = log.NewFilter(logger, log.AllowInfo())
}

func runProxy(cmd *cobra.Command, args []string) error {
	// First, connect a client
	c := commands.GetNode()
	cert, err := commands.GetCertifier()
	if err != nil {
		return err
	}
	sc := certclient.Wrap(c, cert)
	sc.Start()
	r := routes(sc)

	// build the handler...
	mux := http.NewServeMux()
	rpc.RegisterRPCFuncs(mux, r, logger)
	wm := rpc.NewWebsocketManager(r, c)
	wm.SetLogger(logger)
	core.SetLogger(logger)
	mux.HandleFunc(wsEndpoint, wm.WebsocketHandler)

	_, err = rpc.StartHTTPServer(viper.GetString(bindFlag), mux, logger)
	if err != nil {
		return err
	}

	cmn.TrapSignal(func() {
		// TODO: close up shop
	})

	return nil
}

// First step, proxy with no checks....
func routes(c client.Client) map[string]*rpc.RPCFunc {

	return map[string]*rpc.RPCFunc{
		// Subscribe/unsubscribe are reserved for websocket events.
		// We can just use the core tendermint impl, which uses the
		// EventSwitch we registered in NewWebsocketManager above
		"subscribe":   rpc.NewWSRPCFunc(core.Subscribe, "event"),
		"unsubscribe": rpc.NewWSRPCFunc(core.Unsubscribe, "event"),

		// info API
		"status":     rpc.NewRPCFunc(c.Status, ""),
		"blockchain": rpc.NewRPCFunc(c.BlockchainInfo, "minHeight,maxHeight"),
		"genesis":    rpc.NewRPCFunc(c.Genesis, ""),
		"block":      rpc.NewRPCFunc(c.Block, "height"),
		"commit":     rpc.NewRPCFunc(c.Commit, "height"),
		"tx":         rpc.NewRPCFunc(c.Tx, "hash,prove"),
		"validators": rpc.NewRPCFunc(c.Validators, ""),

		// broadcast API
		"broadcast_tx_commit": rpc.NewRPCFunc(c.BroadcastTxCommit, "tx"),
		"broadcast_tx_sync":   rpc.NewRPCFunc(c.BroadcastTxSync, "tx"),
		"broadcast_tx_async":  rpc.NewRPCFunc(c.BroadcastTxAsync, "tx"),

		// abci API
		"abci_query": rpc.NewRPCFunc(c.ABCIQuery, "path,data,prove"),
		"abci_info":  rpc.NewRPCFunc(c.ABCIInfo, ""),
	}
}
