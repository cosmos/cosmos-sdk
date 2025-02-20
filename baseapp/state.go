package baseapp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

type state struct {
	ms  sdk.CacheMultiStore
	ctx sdk.Context
	// eventHistory accumulates events returned by DeliverTx throughout a block.
	// [AGORIC] The accumulated events are passed to the EndBlocker in its context's
	// EventManager ABCI event history, and the state's eventHistory is then cleared.
	// Not used for modes or states other than delivery.
	eventHistory []abci.Event
}

// CacheMultiStore calls and returns a CacheMultiStore on the state's underling
// CacheMultiStore.
func (st *state) CacheMultiStore() sdk.CacheMultiStore {
	return st.ms.CacheMultiStore()
}

// Context returns the Context of the state.
func (st *state) Context() sdk.Context {
	return st.ctx
}
