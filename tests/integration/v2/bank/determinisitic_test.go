package bank

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktestutil "cosmossdk.io/x/bank/testutil"
	banktypes "cosmossdk.io/x/bank/types"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtestutil "github.com/cosmos/cosmos-sdk/x/auth/testutil"
)

var denomRegex = `[a-zA-Z][a-zA-Z0-9/:._-]{2,127}`

type deterministicFixture struct {
	*testing.T
	ctx        context.Context
	app        *integration.App
	bankKeeper bankkeeper.Keeper
}

func (f *deterministicFixture) QueryBalance(
	ctx context.Context, req *banktypes.QueryBalanceRequest,
) (*banktypes.QueryBalanceResponse, error) {
	res, err := f.app.Query(ctx, 0, req)
	return res.(*banktypes.QueryBalanceResponse), err
}

func fundAccount(f *deterministicFixture, addr sdk.AccAddress, coin ...sdk.Coin) {
	err := banktestutil.FundAccount(f.ctx, f.bankKeeper, addr, sdk.NewCoins(coin...))
	require.NoError(f.T, err)
}

func getCoin(rt *rapid.T) sdk.Coin {
	return sdk.NewCoin(
		rapid.StringMatching(denomRegex).Draw(rt, "denom"),
		math.NewInt(rapid.Int64Min(1).Draw(rt, "amount")),
	)
}

func initDeterministicFixture(t *testing.T) *deterministicFixture {
	t.Helper()

	ctrl := gomock.NewController(t)
	acctsModKeeper := authtestutil.NewMockAccountsModKeeper(ctrl)
	accNum := uint64(0)
	acctsModKeeper.EXPECT().NextAccountNumber(gomock.Any()).AnyTimes().DoAndReturn(func(ctx context.Context) (
		uint64, error,
	) {
		currentNum := accNum
		accNum++
		return currentNum, nil
	})

	startupConfig := integration.DefaultStartUpConfig(t)
	startupConfig.GenesisBehavior = integration.Genesis_SKIP
	diConfig := configurator.NewAppV2Config(
		configurator.TxModule(),
		configurator.AuthModule(),
		configurator.BankModule(),
	)

	var bankKeeper bankkeeper.Keeper
	diConfig = depinject.Configs(diConfig, depinject.Supply(acctsModKeeper, log.NewNopLogger()))
	app, err := integration.NewApp(diConfig, startupConfig, &bankKeeper)
	require.NoError(t, err)
	require.NotNil(t, app)
	return &deterministicFixture{app: app, bankKeeper: bankKeeper, T: t}
}

func TestGRPCQueryBalance(t *testing.T) {
	f := initDeterministicFixture(t)
	f.ctx = f.app.StateLatestContext(t)

	rapid.Check(t, func(rt *rapid.T) {
		addr := testdata.AddressGenerator(rt).Draw(rt, "address")
		coin := getCoin(rt)
		fundAccount(f, addr, coin)

		addrStr, err := codectestutil.CodecOptions{}.GetAddressCodec().BytesToString(addr)
		require.NoError(t, err)

		req := banktypes.NewQueryBalanceRequest(addrStr, coin.GetDenom())

		testdata.DeterministicIterationsV2(t, f.ctx, nil, req, f.QueryBalance, 0, true)
	})
}
