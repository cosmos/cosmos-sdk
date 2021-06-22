package crisis_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	testModuleName        = "dummy"
	dummyRouteWhichPasses = types.NewInvarRoute(testModuleName, "which-passes", func(_ sdk.Context) (string, bool) { return "", false })
	dummyRouteWhichFails  = types.NewInvarRoute(testModuleName, "which-fails", func(_ sdk.Context) (string, bool) { return "whoops", true })
)

func createTestApp() (*simapp.SimApp, sdk.Context, []sdk.AccAddress) {
	db := dbm.NewMemDB()
	app := simapp.NewSimApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, simapp.DefaultNodeHome, 1, simapp.MakeTestEncodingConfig(), simapp.EmptyAppOptions{})
	ctx := app.NewContext(true, tmproto.Header{})

	constantFee := sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)
	app.CrisisKeeper.SetConstantFee(ctx, constantFee)
	app.StakingKeeper.SetParams(ctx, stakingtypes.DefaultParams())

	app.CrisisKeeper.RegisterRoute(testModuleName, dummyRouteWhichPasses.Route, dummyRouteWhichPasses.Invar)
	app.CrisisKeeper.RegisterRoute(testModuleName, dummyRouteWhichFails.Route, dummyRouteWhichFails.Invar)

	feePool := distrtypes.InitialFeePool()
	feePool.CommunityPool = sdk.NewDecCoinsFromCoins(sdk.NewCoins(constantFee)...)
	app.DistrKeeper.SetFeePool(ctx, feePool)
	app.BankKeeper.SetSupply(ctx, banktypes.NewSupply(sdk.Coins{}))

	addrs := simapp.AddTestAddrs(app, ctx, 1, sdk.NewInt(10000))

	return app, ctx, addrs
}

// ____________________________________________________________________________

func TestHandleMsgVerifyInvariant(t *testing.T) {
	app, ctx, addrs := createTestApp()
	sender := addrs[0]

	cases := []struct {
		name           string
		msg            sdk.Msg
		expectedResult string
	}{
		{"bad invariant route", types.NewMsgVerifyInvariant(sender, testModuleName, "route-that-doesnt-exist"), "fail"},
		{"invariant broken", types.NewMsgVerifyInvariant(sender, testModuleName, dummyRouteWhichFails.Route), "panic"},
		{"invariant passing", types.NewMsgVerifyInvariant(sender, testModuleName, dummyRouteWhichPasses.Route), "pass"},
		{"invalid msg", testdata.NewTestMsg(), "fail"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			h := crisis.NewHandler(app.CrisisKeeper)

			switch tc.expectedResult {
			case "fail":
				res, err := h(ctx, tc.msg)
				require.Error(t, err)
				require.Nil(t, res)

			case "pass":
				res, err := h(ctx, tc.msg)
				require.NoError(t, err)
				require.NotNil(t, res)

			case "panic":
				require.Panics(t, func() {
					h(ctx, tc.msg) // nolint:errcheck
				})
			}
		})
	}
}

func TestHandleMsgVerifyInvariantWithNotEnoughSenderCoins(t *testing.T) {
	app, ctx, addrs := createTestApp()
	sender := addrs[0]
	coin := app.BankKeeper.GetAllBalances(ctx, sender)[0]
	excessCoins := sdk.NewCoin(coin.Denom, coin.Amount.AddRaw(1))
	app.CrisisKeeper.SetConstantFee(ctx, excessCoins)

	h := crisis.NewHandler(app.CrisisKeeper)
	msg := types.NewMsgVerifyInvariant(sender, testModuleName, dummyRouteWhichPasses.Route)

	res, err := h(ctx, msg)
	require.Error(t, err)
	require.Nil(t, res)
}

func TestHandleMsgVerifyInvariantWithInvariantBrokenAndNotEnoughPoolCoins(t *testing.T) {
	app, ctx, addrs := createTestApp()
	sender := addrs[0]

	// set the community pool to empty
	feePool := app.DistrKeeper.GetFeePool(ctx)
	feePool.CommunityPool = sdk.DecCoins{}
	app.DistrKeeper.SetFeePool(ctx, feePool)

	h := crisis.NewHandler(app.CrisisKeeper)
	msg := types.NewMsgVerifyInvariant(sender, testModuleName, dummyRouteWhichFails.Route)

	var res *sdk.Result
	require.Panics(t, func() {
		res, _ = h(ctx, msg)
	}, fmt.Sprintf("%v", res))
}
