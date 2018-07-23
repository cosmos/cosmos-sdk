package simulation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NonnegativeBalanceInvariant checks that all accounts in the application have non-negative balances
func NonnegativeBalanceInvariant(mapper auth.AccountMapper) simulation.Invariant {
	return func(t *testing.T, app *baseapp.BaseApp, log string) {
		ctx := app.NewContext(false, abci.Header{})
		accts := mock.GetAllAccounts(mapper, ctx)
		for _, acc := range accts {
			coins := acc.GetCoins()
			require.True(t, coins.IsNotNegative(),
				fmt.Sprintf("%s has a negative denomination of %s\n%s",
					acc.GetAddress().String(),
					coins.String(),
					log),
			)
		}
	}
}

// TotalCoinsInvariant checks that the sum of the coins across all accounts
// is what is expected
func TotalCoinsInvariant(mapper auth.AccountMapper, totalSupplyFn func() sdk.Coins) simulation.Invariant {
	return func(t *testing.T, app *baseapp.BaseApp, log string) {
		ctx := app.NewContext(false, abci.Header{})
		totalCoins := sdk.Coins{}

		chkAccount := func(acc auth.Account) bool {
			coins := acc.GetCoins()
			totalCoins = totalCoins.Plus(coins)
			return false
		}

		mapper.IterateAccounts(ctx, chkAccount)
		require.Equal(t, totalSupplyFn(), totalCoins, log)
	}
}
