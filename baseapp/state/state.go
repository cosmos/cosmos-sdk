package state

import (
	"sync"

	"go.opentelemetry.io/otel/trace"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type State struct {
	MultiStore storetypes.CommitBranch

	mtx sync.RWMutex
	ctx sdk.Context

	span trace.Span
}

func NewState(ctx sdk.Context, ms storetypes.CommitBranch) *State {
	return &State{
		MultiStore: ms,
		ctx:        ctx,
	}
}

// CacheMultiStore calls and returns a CacheMultiStore on the state's underlying
// CacheMultiStore.
func (st *State) CacheMultiStore() storetypes.MultiStore {
	return st.MultiStore.CacheMultiStore()
}

// SetContext updates the state's context to the context provided.
func (st *State) SetContext(ctx sdk.Context) {
	st.mtx.Lock()
	defer st.mtx.Unlock()
	st.ctx = ctx
}

// Context returns the Context of the state.
func (st *State) Context() sdk.Context {
	st.mtx.RLock()
	defer st.mtx.RUnlock()
	return st.ctx
}
