package main

import (
	"context"
	"time"

	appmanager "cosmossdk.io/core/app"
)

var app = NewStf()

//go:wasm-module stf
//export add
func execute_block() {
	_, _, _ = app.DeliverBlock(context.Background(), &appmanager.BlockRequest[Tx]{
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

// main is required for the `wasi` target, even if it isn't used.
func main() {}
