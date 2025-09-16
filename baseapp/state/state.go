package state

import (
	"sync"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// State represents a single execution state with its associated context and multi-store.
// It provides thread-safe access to the context and store for a specific execution mode.
type State struct {
	MultiStore storetypes.CacheMultiStore // Cached multi-store for this execution state

	mtx sync.RWMutex // Protects concurrent access to the context
	ctx sdk.Context  // SDK context containing execution metadata
}

// NewState creates a new State instance with the provided context and multi-store.
func NewState(ctx sdk.Context, ms storetypes.CacheMultiStore) *State {
	return &State{
		MultiStore: ms,
		ctx:        ctx,
	}
}

// CacheMultiStore creates and returns a new cached multi-store from the state's
// underlying CacheMultiStore. This allows for nested caching of store operations.
func (st *State) CacheMultiStore() storetypes.CacheMultiStore {
	return st.MultiStore.CacheMultiStore()
}

// SetContext updates the state's context to the provided context.
// This method is thread-safe and should be used when the context needs to be modified.
func (st *State) SetContext(ctx sdk.Context) {
	st.mtx.Lock()
	defer st.mtx.Unlock()
	st.ctx = ctx
}

// Context returns the current context of the state.
// This method is thread-safe and provides read-only access to the context.
func (st *State) Context() sdk.Context {
	st.mtx.RLock()
	defer st.mtx.RUnlock()
	return st.ctx
}
