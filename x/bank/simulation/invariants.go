package simulation

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

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

// TotalCoinsInvariant checks that the sum of the coins across all accounts
// is what is expected
func TotalCoinsInvariant(ak auth.AccountKeeper, totalSupplyFn func() sdk.Coins) sdk.Invariant {
	return func(ctx sdk.Context) error {
		totalCoins := sdk.Coins{}

		chkAccount := func(acc auth.Account) bool {
			coins := acc.GetCoins()
			totalCoins = totalCoins.Plus(coins)
			return false
		}

		ak.IterateAccounts(ctx, chkAccount)
		if !totalSupplyFn().IsEqual(totalCoins) {
			return errors.New("total calculated coins doesn't equal expected coins")
		}
		return nil
	}
}
