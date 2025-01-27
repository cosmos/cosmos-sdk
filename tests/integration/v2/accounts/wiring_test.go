package accounts

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/header"
	stfgas "cosmossdk.io/server/v2/stf/gas"
	counterv1 "cosmossdk.io/x/accounts/testing/counter/v1"
	"cosmossdk.io/x/bank/testutil"

	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestDependencies aims to test wiring between different account components,
// inherited from the runtime, specifically:
// - address codec
// - binary codec
// - header service
// - gas service
// - funds
func TestDependencies(t *testing.T) {
	f := initFixture(t, nil)
	ctx := f.ctx
	ctx = integration.SetHeaderInfo(ctx, header.Info{ChainID: "chain-id"})
	ctx = integration.SetGasMeter(ctx, stfgas.DefaultGasMeter(500_000))

	_, counterAddr, err := f.accountsKeeper.Init(ctx, "counter", accCreator, &counterv1.MsgInit{
		InitialValue: 0,
	}, nil, nil)
	require.NoError(t, err)
	// test dependencies
	creatorInitFunds := sdk.NewCoins(sdk.NewInt64Coin("stake", 100_000))
	err = testutil.FundAccount(ctx, f.bankKeeper, accCreator, creatorInitFunds)
	require.NoError(t, err)
	sentFunds := sdk.NewCoins(sdk.NewInt64Coin("stake", 50_000))
	r, err := f.accountsKeeper.Execute(
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

	headerInfo := integration.HeaderInfoFromContext(ctx)
	// test header service
	require.Equal(t, headerInfo.ChainID, res.ChainId)

	// test address codec
	wantAddr, err := f.authKeeper.AddressCodec().BytesToString(counterAddr)
	require.NoError(t, err)
	require.Equal(t, wantAddr, res.Address)

	// test funds
	creatorFunds := f.bankKeeper.GetAllBalances(ctx, accCreator)
	require.Equal(t, creatorInitFunds.Sub(sentFunds...), creatorFunds)

	accFunds := f.bankKeeper.GetAllBalances(ctx, counterAddr)
	require.Equal(t, sentFunds, accFunds)
}
