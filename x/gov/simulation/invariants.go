package simulation

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
)

// AllInvariants tests all governance invariants
func AllInvariants() simulation.Invariant {
	return func(app *baseapp.BaseApp) error {
		// TODO Add some invariants!
		// Checking proposal queues, no passed-but-unexecuted proposals, etc.
		return nil
	}
}
