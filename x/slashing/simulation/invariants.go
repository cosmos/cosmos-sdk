package simulation

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
)

// AllInvariants tests all slashing invariants
func AllInvariants() simulation.Invariant {
	return func(app *baseapp.BaseApp) error {
		// TODO Any invariants to check here?
		return nil
	}
}
