package appmanager

import (
	"context"
	"encoding/json"
	"io"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/store"
)

type (
	// InitGenesis is a function that will run at application genesis, it will be called with
	// the following arguments:
	// - ctx: the context of the genesis operation
	// - src: the source containing the raw genesis state
	// - txHandler: a function capable of decoding a json tx, will be run for each genesis
	//   transaction
	//
	// It must return a map of the dirty state after the genesis operation.
	InitGenesis func(
		ctx context.Context,
		src io.Reader,
		txHandler func(json.RawMessage) error,
	) (store.WriterMap, []appmodulev2.ValidatorUpdate, error)

	// ExportGenesis is a function type that represents the export of the genesis state.
	ExportGenesis func(ctx context.Context, version uint64) ([]byte, error)
)
