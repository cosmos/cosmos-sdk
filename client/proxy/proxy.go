package proxy

import (
	"net/http"

	"github.com/tendermint/tmlibs/log"

	rpcclient "github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/rpc/core"
	rpc "github.com/tendermint/tendermint/rpc/lib/server"
)

const (
	wsEndpoint = "/websocket"
)

// StartProxy will start the websocket manager on the client,
// set up the rpc routes to proxy via the given client,
// and start up an http/rpc server on the location given by bind (eg. :1234)
func StartProxy(c rpcclient.Client, listenAddr string, logger log.Logger) error {
	err := c.Start()
	if err != nil {
		return err
	}

	r := RPCRoutes(c)

	// build the handler...
	mux := http.NewServeMux()
	rpc.RegisterRPCFuncs(mux, r, logger)

	wm := rpc.NewWebsocketManager(r, rpc.EventSubscriber(c))
	wm.SetLogger(logger)
	core.SetLogger(logger)
	mux.HandleFunc(wsEndpoint, wm.WebsocketHandler)

	_, err = rpc.StartHTTPServer(listenAddr, mux, logger)

	return err
}

// RPCRoutes just routes everything to the given client, as if it were
// a tendermint fullnode.
//
// if we want security, the client must implement it as a secure client
func RPCRoutes(c rpcclient.Client) map[string]*rpc.RPCFunc {

	return map[string]*rpc.RPCFunc{
		// Subscribe/unsubscribe are reserved for websocket events.
		// We can just use the core tendermint impl, which uses the
		// EventSwitch we registered in NewWebsocketManager above
		"subscribe":   rpc.NewWSRPCFunc(core.Subscribe, "query"),
		"unsubscribe": rpc.NewWSRPCFunc(core.Unsubscribe, "query"),

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
