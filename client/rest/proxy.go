package rest

import (
	"github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/rpc/core"
	rpc "github.com/tendermint/tendermint/rpc/lib/server"
)

func Routes(c client.Client) map[string]*rpc.RPCFunc {
	return map[string]*rpc.RPCFunc{
		// subscribe/unsubscribe are reserved for websocket events.
		// We can just the core Tendermint implementation, which uses
		// the EventSwitch that we registered in NewWebsocketManager above.
		"subscribe":   rpc.NewWSRPCFunc(core.Subscribe, "event"),
		"unsubscribe": rpc.NewWSRPCFunc(core.Unsubscribe, "event"),

		// info API
		"status":     rpc.NewRPCFunc(c.Status, ""),
		"blockchain": rpc.NewRPCFunc(c.BlockchainInfo, "minHeight,maxHeight"),
		"genesis":    rpc.NewRPCFunc(c.Genesis, ""),
		"block":      rpc.NewRPCFunc(c.Block, "height"),
		"commit":     rpc.NewRPCFunc(c.Commit, "height"),
		"tx":         rpc.NewRPCFunc(c.Tx, "hash.prove"),
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
