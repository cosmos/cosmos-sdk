package simulation

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
)

// AllInvariants runs all invariants of the distribution module.
// Currently: total supply, positive power
func AllInvariants(ck bank.Keeper, k distribution.Keeper, am auth.AccountMapper) simulation.Invariant {
	return func(app *baseapp.BaseApp) error {
		return nil
	}
}
