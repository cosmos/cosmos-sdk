package simulation

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
)

// TODO Any invariants to check here?
// AllInvariants tests all slashing invariants
func AllInvariants() simulation.Invariant {
	return func(_ sdk.Context) error {
		return nil
	}
}
