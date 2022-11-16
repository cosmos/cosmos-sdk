package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-plugin"

	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/streaming/plugins/abci/v1"
	abci "github.com/tendermint/tendermint/abci/types"
)

// ABCIListener is the implementation of the baseapp.ABCIListener interface
// For Go plugins this is all that is required to process data sent over gRPC.
type ABCIListener struct{}

func (a ABCIListener) ListenBeginBlock(ctx context.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) error {
	// process begin block messages (i.e: sent to external system)
	fmt.Printf("listen-begin-block: req=%v res=%v", res, res)
	return nil
}

func (a ABCIListener) ListenEndBlock(ctx context.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) error {
	// process end block messages (i.e: sent to external system)
	fmt.Printf("listen-end-block: req=%v res=%v", res, res)
	return nil
}

func (a ABCIListener) ListenDeliverTx(ctx context.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) error {
	// process tx messages (i.e: sent to external system)
	fmt.Printf("listen-deliver-tx: req=%v res=%v", res, res)
	return nil
}

func (a ABCIListener) ListenCommit(ctx context.Context, res abci.ResponseCommit, changeSet []*store.StoreKVPair) error {
	// process block commit messages (i.e: sent to external system)
	fmt.Printf("listen-commit: block_height=%d data=%v", res.RetainHeight, changeSet)
	return nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: v1.Handshake,
		Plugins: map[string]plugin.Plugin{
			"abci_v1": &v1.ABCIListenerGRPCPlugin{Impl: &ABCIListener{}},
		},

		// A non-nil value here enables gRPC serving for this streaming...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
