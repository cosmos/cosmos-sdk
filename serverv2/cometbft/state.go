package cometbft

import (
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	execMode uint8

	// StoreLoader defines a customizable function to control how we load the
	// CommitMultiStore from disk. This is useful for state migration, when
	// loading a datastore written with an older version of the software. In
	// particular, if a module changed the substore key name (or removed a substore)
	// between two versions of the software.
	StoreLoader func(ms storetypes.CommitMultiStore) error
)

const (
	execModeCheck           execMode = iota // Check a transaction
	execModeReCheck                         // Recheck a (pending) transaction after a commit
	execModeSimulate                        // Simulate a transaction
	execModePrepareProposal                 // Prepare a block proposal
	execModeProcessProposal                 // Process a block proposal
	execModeVoteExtension                   // Extend or verify a pre-commit vote
	execModeFinalize                        // Finalize a block proposal
)

type state struct {
	ms  storetypes.CacheMultiStore
	ctx sdk.Context
}

// CacheMultiStore calls and returns a CacheMultiStore on the state's underling
// CacheMultiStore.
func (st *state) CacheMultiStore() storetypes.CacheMultiStore {
	return st.ms.CacheMultiStore()
}

// Context returns the Context of the state.
func (st *state) Context() sdk.Context {
	return st.ctx
}
