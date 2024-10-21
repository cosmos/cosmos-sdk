package appmanager

import (
	"context"
	"encoding/json"
	"io"

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
	) (store.WriterMap, error)

	// ExportGenesis is a function type that represents the export of the genesis state.
	ExportGenesis func(ctx context.Context, version uint64) ([]byte, error)
)
