package invariants

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// register all bank invariants
func RegisterInvariants(ak auth.AccountKeeper, invarRoutes sdk.InvarRoutes) InvarRoutes {
	(&invarRoutes).Register("bank/nonnegative-balance",
		NonnegativeBalanceInvariant(ak))
	return invarRoutes
}

// NonnegativeBalanceInvariant checks that all accounts in the application have non-negative balances
func NonnegativeBalanceInvariant(ak auth.AccountKeeper) sdk.Invariant {
	return func(ctx sdk.Context) error {
		accts := ak.GetAllAccounts(ctx)
		for _, acc := range accts {
			coins := acc.GetCoins()
			if coins.IsAnyNegative() {
				return fmt.Errorf("%s has a negative denomination of %s",
					acc.GetAddress().String(),
					coins.String())
			}
		}
		return nil
	}
}
