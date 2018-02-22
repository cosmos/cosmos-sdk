package client

import rpcclient "github.com/tendermint/tendermint/rpc/client"

// GetNode prepares a simple rpc.Client from the flags
func GetNode(uri string) rpcclient.Client {
	return rpcclient.NewHTTP(uri, "/websocket")
}
