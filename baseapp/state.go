package baseapp

import (
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type state struct {
	lock sync.RWMutex
	ms   sdk.CacheMultiStore
	ctx  sdk.Context
}

// CacheMultiStore calls and returns a CacheMultiStore on the state's underling
// CacheMultiStore.
func (st *state) CacheMultiStore() sdk.CacheMultiStore {
	return st.ms.CacheMultiStore()
}

// Context returns the Context of the state.
func (st *state) Context() sdk.Context {
	defer st.lock.RUnlock()
	st.lock.RLock()

	return st.ctx
}

// WithContext update context of the state
func (st *state) WithContext(ctx sdk.Context) {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.ctx = ctx
}
