//go:build app_v1

package accounts

import (
	"testing"

	"cosmossdk.io/core/header"
	storetypes "cosmossdk.io/store/types"
	counterv1 "cosmossdk.io/x/accounts/testing/counter/v1"
	"cosmossdk.io/x/bank/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// TestDependencies aims to test wiring between different account components,
// inherited from the runtime, specifically:
// - address codec
// - binary codec
// - header service
// - gas service
// - funds
func TestDependencies(t *testing.T) {
	app := setupApp(t)
	ak := app.AccountsKeeper
	ctx := sdk.NewContext(app.CommitMultiStore(), false, app.Logger()).WithHeaderInfo(header.Info{ChainID: "chain-id"})
	ctx = ctx.WithGasMeter(storetypes.NewGasMeter(500_000))

	_, counterAddr, err := ak.Init(ctx, "counter", accCreator, &counterv1.MsgInit{
		InitialValue: 0,
	}, nil)
	require.NoError(t, err)
	// test dependencies
	creatorInitFunds := sdk.NewCoins(sdk.NewInt64Coin("stake", 100_000))
	err = testutil.FundAccount(ctx, app.BankKeeper, accCreator, creatorInitFunds)
	require.NoError(t, err)
	sentFunds := sdk.NewCoins(sdk.NewInt64Coin("stake", 50_000))
	r, err := ak.Execute(
		ctx,
		counterAddr,
		accCreator,
		&counterv1.MsgTestDependencies{},
		sentFunds,
	)
	require.NoError(t, err)
	res := r.(*counterv1.MsgTestDependenciesResponse)

	// test gas
	require.NotZero(t, res.BeforeGas)
	require.NotZero(t, res.AfterGas)
	require.Equal(t, int(uint64(10)), int(res.AfterGas-res.BeforeGas))

	// test header service
	require.Equal(t, ctx.HeaderInfo().ChainID, res.ChainId)

	// test address codec
	wantAddr, err := app.AuthKeeper.AddressCodec().BytesToString(counterAddr)
	require.NoError(t, err)
	require.Equal(t, wantAddr, res.Address)

	// test funds
	creatorFunds := app.BankKeeper.GetAllBalances(ctx, accCreator)
	require.Equal(t, creatorInitFunds.Sub(sentFunds...), creatorFunds)

	accFunds := app.BankKeeper.GetAllBalances(ctx, counterAddr)
	require.Equal(t, sentFunds, accFunds)
}
