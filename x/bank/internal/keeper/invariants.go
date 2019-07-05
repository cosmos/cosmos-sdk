package keeper

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/bank/internal/types"
)

// register bank invariants
func RegisterInvariants(ir sdk.InvariantRegistry, ak types.AccountKeeper) {
	ir.RegisterRoute(types.ModuleName, "nonnegative-outstanding",
		NonnegativeBalanceInvariant(ak))
}

// NonnegativeBalanceInvariant checks that all accounts in the application have non-negative balances
func NonnegativeBalanceInvariant(ak types.AccountKeeper) sdk.Invariant {
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

// TotalCoinsInvariant checks that the sum of the coins across all accounts
// is what is expected
func TotalCoinsInvariant(ak types.AccountKeeper, totalSupplyFn func() sdk.Coins) sdk.Invariant {
	return func(ctx sdk.Context) error {
		totalCoins := sdk.NewCoins()

		chkAccount := func(acc exported.Account) bool {
			coins := acc.GetCoins()
			totalCoins = totalCoins.Add(coins)
			return false
		}

		ak.IterateAccounts(ctx, chkAccount)
		if !totalSupplyFn().IsEqual(totalCoins) {
			return errors.New("total calculated coins doesn't equal expected coins")
		}
		return nil
	}
}
