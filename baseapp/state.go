package baseapp

import (
	"cosmossdk.io/store/v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type state struct {
	brs store.BranchedRootStore
	ctx sdk.Context
}

// BranchRootStore returns a branched RootStore. Note, the underlying RootStore
// is already branched.
func (st *state) BranchRootStore() store.BranchedRootStore {
	return st.brs.Branch()
}

// Context returns the Context of the state.
func (st *state) Context() sdk.Context {
	return st.ctx
}
