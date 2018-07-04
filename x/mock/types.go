package mock

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
)

type (
	// TestAndRunTx produces a random transaction, and check the state
	// transition. It returns a descriptive message "action" about what this
	// random tx actually did, for ease of debugging.
	TestAndRunTx func(
		t *testing.T, r *rand.Rand, app *App, ctx sdk.Context,
		privKeys []crypto.PrivKey, log string,
	) (action string, err sdk.Error)

	// RandSetup performs the random setup the mock module needs.
	RandSetup func(r *rand.Rand, privKeys []crypto.PrivKey)

	// AssertInvariants asserts invariants for the mock module. It will print
	// out the log when failing.
	AssertInvariants func(t *testing.T, app *App, log string)
)

// AuthInvariant enforces an invariant for the Auth module.
// TODO: Does this belong here?
func AuthInvariant(t *testing.T, app *App, log string) {
	ctx := app.BaseApp.NewContext(false, abci.Header{})
	totalCoins := sdk.Coins{}

	chkAccount := func(acc auth.Account) bool {
		coins := acc.GetCoins()
		totalCoins = totalCoins.Plus(coins)

		for _, coin := range coins {
			require.True(t, coin.Amount.GT(sdk.ZeroInt()), log)
		}
		return false
	}

	app.AccountMapper.IterateAccounts(ctx, chkAccount)
	require.Equal(t, app.TotalCoinsSupply, totalCoins, log)
}

// PeriodicInvariant returns an AssertInvariants function closure that asserts
// a given invariant if the mock application's last block modulo the given
// period is congruent to the given offset.
func PeriodicInvariant(invariant AssertInvariants, period int, offset int) AssertInvariants {
	return func(t *testing.T, app *App, log string) {
		if int(app.LastBlockHeight())%period == offset {
			invariant(t, app, log)
		}
	}
}
