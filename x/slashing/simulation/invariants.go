package simulation

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	abci "github.com/tendermint/tendermint/abci/types"
)

// TODO Any invariants to check here?
// AllInvariants tests all slashing invariants
func AllInvariants() simulation.Invariant {
	return func(_ *baseapp.BaseApp, _ abci.Header) error {
		return nil
	}
}
