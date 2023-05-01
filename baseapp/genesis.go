package baseapp

import (
	"github.com/cometbft/cometbft/abci/types"
)

// GenesisState allows modules to define a set of state transitions
// that will initialize the chain's state at genesis.
type GenesisState interface {
	// SetState sets the genesis state.
	// This should be called in a order define by the application developer
	SetState([]byte) types.ResponseDeliverTx
}

var _ GenesisState = (*BaseApp)(nil)

// SetState implements genesis.GenesisState from
// cosmossdk.io/core/genesis to set initial state in genesis
func (ba BaseApp) SetState(tx []byte) types.ResponseDeliverTx {
	return ba.DeliverTx(types.RequestDeliverTx{Tx: tx})
}
