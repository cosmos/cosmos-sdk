package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-plugin"

	"cosmossdk.io/server/v2/streaming"
)

// StdoutPlugin is the implementation of the ABCIListener interface
// For Go plugins this is all that is required to process data sent over gRPC.
type StdoutPlugin struct {
	BlockHeight int64
}

func (a *StdoutPlugin) ListenDeliverBlock(ctx context.Context, req streaming.ListenDeliverBlockRequest) error {
	a.BlockHeight = req.BlockHeight
	// process tx messages (i.e: sent to external system)
	fmt.Printf("listen-finalize-block: block-height=%d req=%v res=%v", a.BlockHeight, req, nil)
	return nil
}

func (a *StdoutPlugin) ListenStateChanges(ctx context.Context, changeSet []*streaming.StoreKVPair) error {
	// process block commit messages (i.e: sent to external system)
	fmt.Printf("listen-commit: block_height=%d res=%v data=%v", a.BlockHeight, changeSet, nil)
	return nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: streaming.Handshake,
		Plugins: map[string]plugin.Plugin{
			"abci": &streaming.ListenerGRPCPlugin{Impl: &StdoutPlugin{}},
		},

		// A non-nil value here enables gRPC serving for this streaming...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
