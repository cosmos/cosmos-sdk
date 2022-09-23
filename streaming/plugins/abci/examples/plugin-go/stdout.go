package main

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/streaming/plugins/abci"
	"github.com/hashicorp/go-plugin"
)

// ABCIListener is the implementation of the abci.ABCIListener interface
type ABCIListener struct{}

func (A ABCIListener) ListenBeginBlock(blockHeight int64, req []byte, res []byte) error {
	fmt.Printf("begin-block: block_height=%d req=%b res=%s", blockHeight, req, res)
	return nil
}

func (A ABCIListener) ListenEndBlock(blockHeight int64, req []byte, res []byte) error {
	fmt.Printf("end-block: block_height=%d req=%b res=%s", blockHeight, req, res)
	return nil
}

func (A ABCIListener) ListenDeliverTx(blockHeight int64, req []byte, res []byte) error {
	fmt.Printf("deliver-tx: block_height=%d req=%b res=%s", blockHeight, req, res)
	return nil
}

func (A ABCIListener) ListenStoreKVPair(blockHeight int64, data []byte) error {
	fmt.Printf("store-kv-pair: block_height=%d req=%b", blockHeight, data)
	return nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: abci.Handshake,
		Plugins: map[string]plugin.Plugin{
			"abci": &abci.ListenerGRPCPlugin{Impl: &ABCIListener{}},
		},

		// A non-nil value here enables gRPC serving for this streaming...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
