package simulation

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NonnegativeBalanceInvariant checks that all accounts in the application have non-negative balances
func NonnegativeBalanceInvariant(mapper auth.AccountKeeper) simulation.Invariant {
	return func(app *baseapp.BaseApp) error {
		ctx := app.NewContext(false, abci.Header{})
		accts := mock.GetAllAccounts(mapper, ctx)
		for _, acc := range accts {
			coins := acc.GetCoins()
			if !coins.IsNotNegative() {
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
func TotalCoinsInvariant(mapper auth.AccountKeeper, totalSupplyFn func() sdk.Coins) simulation.Invariant {
	return func(app *baseapp.BaseApp) error {
		ctx := app.NewContext(false, abci.Header{})
		totalCoins := sdk.Coins{}

		chkAccount := func(acc auth.Account) bool {
			coins := acc.GetCoins()
			totalCoins = totalCoins.Plus(coins)
			return false
		}

		mapper.IterateAccounts(ctx, chkAccount)
		if !totalSupplyFn().IsEqual(totalCoins) {
			return errors.New("total calculated coins doesn't equal expected coins")
		}
		return nil
	}
}
