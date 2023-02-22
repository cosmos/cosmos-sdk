package main

import (
	"context"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/hashicorp/go-plugin"

	streamingabci "cosmossdk.io/store/streaming/abci"
	store "cosmossdk.io/store/types"
)

// StdoutPlugin is the implementation of the ABCIListener interface
// For Go plugins this is all that is required to process data sent over gRPC.
type StdoutPlugin struct {
	BlockHeight int64
}

func (a *StdoutPlugin) ListenBeginBlock(ctx context.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) error {
	a.BlockHeight = req.Header.Height
	// process tx messages (i.e: sent to external system)
	fmt.Printf("listen-begin-block: block-height=%d req=%v res=%v", a.BlockHeight, req, res)
	return nil
}

func (a *StdoutPlugin) ListenEndBlock(ctx context.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) error {
	// process end block messages (i.e: sent to external system)
	fmt.Printf("listen-end-block: block-height=%d req=%v res=%v", a.BlockHeight, req, res)
	return nil
}

func (a *StdoutPlugin) ListenDeliverTx(ctx context.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) error {
	// process tx messages (i.e: sent to external system)
	fmt.Printf("listen-deliver-tx: block-height=%d req=%v res=%v", a.BlockHeight, req, res)
	return nil
}

func (a *StdoutPlugin) ListenCommit(ctx context.Context, res abci.ResponseCommit, changeSet []*store.StoreKVPair) error {
	// process block commit messages (i.e: sent to external system)
	fmt.Printf("listen-commit: block_height=%d res=%v data=%v", a.BlockHeight, res, changeSet)
	return nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: streamingabci.Handshake,
		Plugins: map[string]plugin.Plugin{
			"abci": &streamingabci.ListenerGRPCPlugin{Impl: &StdoutPlugin{}},
		},

		// A non-nil value here enables gRPC serving for this streaming...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
