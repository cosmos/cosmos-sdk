package simulation

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
)

// AllInvariants tests all governance invariants
func AllInvariants() simulation.Invariant {
	return func(ctx sdk.Context) error {
		// TODO Add some invariants!
		// Checking proposal queues, no passed-but-unexecuted proposals, etc.
		return nil
	}
}
