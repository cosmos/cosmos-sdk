package crisis_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
)

var (
	testModuleName        = "dummy"
	dummyRouteWhichPasses = crisis.NewInvarRoute(testModuleName, "which-passes", func(_ sdk.Context) (string, bool) { return "", false })
	dummyRouteWhichFails  = crisis.NewInvarRoute(testModuleName, "which-fails", func(_ sdk.Context) (string, bool) { return "whoops", true })
	addrs                 = distr.TestAddrs
)

func CreateTestInput(t *testing.T) (sdk.Context, crisis.Keeper, auth.AccountKeeper, distr.Keeper) {

	communityTax := sdk.NewDecWithPrec(2, 2)
	ctx, accKeeper, _, distrKeeper, _, paramsKeeper, supplyKeeper :=
		distr.CreateTestInputAdvanced(t, false, 10, communityTax)

	paramSpace := paramsKeeper.Subspace(crisis.DefaultParamspace)
	crisisKeeper := crisis.NewKeeper(paramSpace, 1, supplyKeeper, auth.FeeCollectorName)
	constantFee := sdk.NewInt64Coin("stake", 10000000)
	crisisKeeper.SetConstantFee(ctx, constantFee)

	crisisKeeper.RegisterRoute(testModuleName, dummyRouteWhichPasses.Route, dummyRouteWhichPasses.Invar)
	crisisKeeper.RegisterRoute(testModuleName, dummyRouteWhichFails.Route, dummyRouteWhichFails.Invar)

	// set the community pool to pay back the constant fee
	feePool := distr.InitialFeePool()
	feePool.CommunityPool = sdk.NewDecCoins(sdk.NewCoins(constantFee))
	distrKeeper.SetFeePool(ctx, feePool)

	return ctx, crisisKeeper, accKeeper, distrKeeper
}

//____________________________________________________________________________

func TestHandleMsgVerifyInvariantWithNotEnoughSenderCoins(t *testing.T) {
	ctx, crisisKeeper, accKeeper, _ := CreateTestInput(t)
	sender := addrs[0]
	coin := accKeeper.GetAccount(ctx, sender).GetCoins()[0]
	excessCoins := sdk.NewCoin(coin.Denom, coin.Amount.AddRaw(1))
	crisisKeeper.SetConstantFee(ctx, excessCoins)

	h := crisis.NewHandler(crisisKeeper)
	msg := crisis.NewMsgVerifyInvariant(sender, testModuleName, dummyRouteWhichPasses.Route)
	require.False(t, h(ctx, msg).IsOK())
}

func TestHandleMsgVerifyInvariantWithBadInvariant(t *testing.T) {
	ctx, crisisKeeper, _, _ := CreateTestInput(t)
	sender := addrs[0]

	h := crisis.NewHandler(crisisKeeper)
	msg := crisis.NewMsgVerifyInvariant(sender, testModuleName, "route-that-doesnt-exist")
	res := h(ctx, msg)
	require.False(t, res.IsOK())
}

func TestHandleMsgVerifyInvariantWithInvariantBroken(t *testing.T) {
	ctx, crisisKeeper, _, _ := CreateTestInput(t)
	sender := addrs[0]

	h := crisis.NewHandler(crisisKeeper)
	msg := crisis.NewMsgVerifyInvariant(sender, testModuleName, dummyRouteWhichFails.Route)
	var res sdk.Result
	require.Panics(t, func() {
		res = h(ctx, msg)
	}, fmt.Sprintf("%v", res))
}

func TestHandleMsgVerifyInvariantWithInvariantBrokenAndNotEnoughPoolCoins(t *testing.T) {
	ctx, crisisKeeper, _, distrKeeper := CreateTestInput(t)
	sender := addrs[0]

	// set the community pool to empty
	feePool := distrKeeper.GetFeePool(ctx)
	feePool.CommunityPool = sdk.DecCoins{}
	distrKeeper.SetFeePool(ctx, feePool)

	h := crisis.NewHandler(crisisKeeper)
	msg := crisis.NewMsgVerifyInvariant(sender, testModuleName, dummyRouteWhichFails.Route)
	var res sdk.Result
	require.Panics(t, func() {
		res = h(ctx, msg)
	}, fmt.Sprintf("%v", res))
}

func TestHandleMsgVerifyInvariantWithInvariantNotBroken(t *testing.T) {
	ctx, crisisKeeper, _, _ := CreateTestInput(t)
	sender := addrs[0]

	h := crisis.NewHandler(crisisKeeper)
	msg := crisis.NewMsgVerifyInvariant(sender, testModuleName, dummyRouteWhichPasses.Route)
	require.True(t, h(ctx, msg).IsOK())
}

func TestInvalidMsg(t *testing.T) {
	k := crisis.Keeper{}
	h := crisis.NewHandler(k)

	res := h(sdk.NewContext(nil, abci.Header{}, false, nil), sdk.NewTestMsg())
	require.False(t, res.IsOK())
	require.True(t, strings.Contains(res.Log, "unrecognized crisis message type"))
}
