package baseapp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type state struct {
	ms  sdk.CacheRootStore
	ctx sdk.Context
}

// CacheRootStore calls and returns a CacheRootStore on the state's underling
// CacheRootStore.
func (st *state) CacheRootStore() sdk.CacheRootStore {
	return st.ms.CacheRootStore()
}

// Context returns the Context of the state.
func (st *state) Context() sdk.Context {
	return st.ctx
}
