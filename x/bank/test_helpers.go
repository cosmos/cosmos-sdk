package bank

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
)

// ModuleInvariants runs all invariants of the bank module.
// Currently only runs non-negative balance invariant
func ModuleInvariants(t *testing.T, app *mock.App, log string) {
	NonnegativeBalanceInvariant(t, app, log)
}

// NonnegativeBalanceInvariant checks that all accounts in the application have non-negative balances
func NonnegativeBalanceInvariant(t *testing.T, app *mock.App, log string) {
	ctx := app.NewContext(false, abci.Header{})
	accts := mock.GetAllAccounts(app.AccountMapper, ctx)
	for _, acc := range accts {
		for _, coin := range acc.GetCoins() {
			assert.True(t, coin.IsNotNegative(), acc.GetAddress().String()+
				" has a negative denomination of "+coin.Denom+"\n"+log)
		}
	}
}
