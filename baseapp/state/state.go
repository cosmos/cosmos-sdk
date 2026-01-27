package state

import (
	"sync"

	"go.opentelemetry.io/otel/trace"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type State struct {
	MultiStore storetypes.MultiStore

	mtx sync.RWMutex
	ctx sdk.Context

	span trace.Span
}

func NewState(ctx sdk.Context, ms storetypes.MultiStore) *State {
	return &State{
		MultiStore: ms,
		ctx:        ctx,
	}
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
