package state

import (
	"sync"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type State struct {
	ms storetypes.CacheMultiStore

	mtx sync.RWMutex
	ctx sdk.Context
}

// CacheMultiStore calls and returns a CacheMultiStore on the State's underling
// CacheMultiStore.
func (st *State) CacheMultiStore() storetypes.CacheMultiStore {
	return st.ms.CacheMultiStore()
}

// SetContext updates the State's context to the context provided.
func (st *State) SetContext(ctx sdk.Context) {
	st.mtx.Lock()
	defer st.mtx.Unlock()
	st.ctx = ctx
}

// Context returns the Context of the State.
func (st *State) Context() sdk.Context {
	st.mtx.RLock()
	defer st.mtx.RUnlock()
	return st.ctx
}

func (st *State) SetMultiStore(ms storetypes.CacheMultiStore) {
	st.ms = ms
}
