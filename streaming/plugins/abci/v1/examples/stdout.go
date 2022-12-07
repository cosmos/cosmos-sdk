package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-plugin"

	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/streaming/plugins/abci/v1"
	abci "github.com/tendermint/tendermint/abci/types"
)

// StdoutPlugin is the implementation of the baseapp.ABCIListener interface
// For Go plugins this is all that is required to process data sent over gRPC.
type StdoutPlugin struct {
	BlockHeight int64
}

func (a StdoutPlugin) ListenBeginBlock(ctx context.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) error {
	a.BlockHeight = req.Header.Height
	// process tx messages (i.e: sent to external system)
	fmt.Printf("listen-begin-block: block-height=%d req=%v res=%v", a.BlockHeight, req, res)
	return nil
}

func (a StdoutPlugin) ListenEndBlock(ctx context.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) error {
	// process end block messages (i.e: sent to external system)
	fmt.Printf("listen-end-block: block-height=%d req=%v res=%v", a.BlockHeight, req, res)
	return nil
}

func (a StdoutPlugin) ListenDeliverTx(ctx context.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) error {
	// process tx messages (i.e: sent to external system)
	fmt.Printf("listen-deliver-tx: block-height=%d req=%v res=%v", a.BlockHeight, req, res)
	return nil
}

func (a StdoutPlugin) ListenCommit(ctx context.Context, res abci.ResponseCommit, changeSet []*store.StoreKVPair) error {
	// process block commit messages (i.e: sent to external system)
	fmt.Printf("listen-commit: block_height=%d res=%v data=%v", a.BlockHeight, res, changeSet)
	return nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: v1.Handshake,
		Plugins: map[string]plugin.Plugin{
			"abci_v1": &v1.ABCIListenerGRPCPlugin{Impl: &StdoutPlugin{}},
		},

		// A non-nil value here enables gRPC serving for this streaming...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
