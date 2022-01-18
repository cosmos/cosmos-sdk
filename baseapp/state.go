package baseapp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type state struct {
	ms  sdk.CacheMultiStore
	ctx sdk.Context
}

// CacheMultiStore calls and returns a CacheMultiStore on the state's underling
// CacheMultiStore.
func (st *state) CacheWrap() sdk.CacheMultiStore {
	return st.ms.CacheWrap()
}

// Context returns the Context of the state.
func (st *state) Context() sdk.Context {
	return st.ctx
}
