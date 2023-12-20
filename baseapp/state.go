package baseapp

import (
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"sync"
)

type state struct {
	mtx sync.RWMutex
	ctx sdk.Context
	// no need for synchronization on the MultiStore as implementations are thread-safe
	ms  storetypes.CacheMultiStore
}

// CacheMultiStore calls and returns a CacheMultiStore on the state's underling
// CacheMultiStore.
func (st *state) CacheMultiStore() storetypes.CacheMultiStore {
	return st.ms.CacheMultiStore()
}

// UpdateContext updates the state's context to the context provided.
func (st *state) UpdateContext(ctx sdk.Context) {
	st.mtx.Lock()
	defer st.mtx.Unlock()
	st.ctx = ctx
}

// Context returns the Context of the state.
func (st *state) Context() sdk.Context {
	st.mtx.RLock()
	defer st.mtx.RUnlock()
	return st.ctx
}
