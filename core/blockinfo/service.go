// Package blockinfo provides an API for app modules to get basic
// information about blocks that is available against any underlying Tendermint
// core version (or other consensus layer that could be used in the future).
package blockinfo

import (
	"context"
	"time"
)

// Service is a type which retrieves basic block info from a context independent
// of any specific Tendermint core version. Modules which need a specific
// Tendermint header should use a different service and should expect a need
// to update whenever Tendermint makes any changes. blockinfo.Service is a
// core API type that should be provided by the runtime module being used to
// build an app via depinject.
type Service interface {
	// GetBlockInfo returns the current block info for the context.
	GetBlockInfo(ctx context.Context) BlockInfo
}

// BlockInfo represents basic block info independent of any specific Tendermint
// core version.
type BlockInfo struct {
	// ChainID is the chain ID.
	ChainID string

	// Height is the current block height.
	Height int64

	// Time is the current block timestamp.
	Time time.Time

	// Hash is the current block hash.
	Hash []byte
}
