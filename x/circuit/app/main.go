package main

import (
	"context"
	"time"

	appmanager "cosmossdk.io/core/app"
	"cosmossdk.io/x/circuit/app/deps"
)

var app = deps.NewStf()

// main is required for the `wasi` target, even if it isn't used.
func main() {
	_, _, _ = app.DeliverBlock(context.Background(), &appmanager.BlockRequest[deps.Tx]{
		Height:            0,
		Time:              time.Time{},
		Hash:              nil,
		ChainId:           "",
		AppHash:           nil,
		Txs:               nil,
		ConsensusMessages: nil,
	},
		nil, // TODO: state
	)
}
