package appmanager

import (
	"context"
	"encoding/json"
	"io"
)

type (
	// ExportGenesis is a function type that represents the export of the genesis state.
	ExportGenesis func(ctx context.Context, version uint64) ([]byte, error)
	// InitGenesis is a function type that represents the initialization of the genesis state.
	InitGenesis func(ctx context.Context, src io.Reader, txHandler func(json.RawMessage) error) error
)
