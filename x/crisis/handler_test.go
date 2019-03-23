package crisis

import (
	"errors"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/stretchr/testify/require"
)

var (
	dummyRouteWhichPasses = NewInvarRoute("dummy-pass", func(_ sdk.Context) error { return nil })
	dummyRouteWhichFails  = NewInvarRoute("dummy-fail", func(_ sdk.Context) error { return errors.New("whoops") })
	addrs                 = distr.TestAddrs
)

func CreateTestInput(t *testing.T) (sdk.Context, Keeper, auth.AccountKeeper) {

	communityTax := sdk.NewDecWithPrec(2, 2)
	ctx, accKeeper, bankKeeper, distrKeeper, _, feeCollectionKeeper, paramsKeeper :=
		distr.CreateTestInputAdvanced(t, false, 10, communityTax)

	paramSpace := paramsKeeper.Subspace(DefaultParamspace)
	crisisKeeper := NewKeeper(paramSpace, distrKeeper, bankKeeper, feeCollectionKeeper)
	crisisKeeper.SetConstantFee(ctx, sdk.NewInt64Coin("stake", 1000))
	crisisKeeper.RegisterRoute(dummyRouteWhichPasses.Route, dummyRouteWhichPasses.Invar)
	crisisKeeper.RegisterRoute(dummyRouteWhichFails.Route, dummyRouteWhichFails.Invar)

	return ctx, crisisKeeper, accKeeper
}

//____________________________________________________________________________

func TestHandleMsgVerifyInvariantWithNotEnoughCoins(t *testing.T) {
	ctx, crisisKeeper, accKeeper := CreateTestInput(t)
	sender := addrs[0]
	coin := accKeeper.GetAccount(ctx, sender).GetCoins()[0]
	excessCoins := sdk.NewCoin(coin.Denom, coin.Amount.AddRaw(1))
	crisisKeeper.SetConstantFee(ctx, excessCoins)

	msg := NewMsgVerifyInvariance(sender, dummyRouteWhichPasses.Route)
	res := handleMsgVerifyInvariant(ctx, msg, crisisKeeper)
	require.False(t, res.IsOK())
}

func TestHandleMsgVerifyInvariantWithBadInvariant(t *testing.T) {
	ctx, crisisKeeper, accKeeper := CreateTestInput(t)
	sender := addrs[0]
	coin := accKeeper.GetAccount(ctx, sender).GetCoins()[0]
	crisisKeeper.SetConstantFee(ctx, coin)

	msg := NewMsgVerifyInvariance(sender, "route-that-doesnt-exist")
	res := handleMsgVerifyInvariant(ctx, msg, crisisKeeper)
	require.False(t, res.IsOK())
}

func TestHandleMsgVerifyInvariantWithInvariantNotBroken(t *testing.T) {
	ctx, crisisKeeper, accKeeper := CreateTestInput(t)
	sender := addrs[0]
	coin := accKeeper.GetAccount(ctx, sender).GetCoins()[0]
	crisisKeeper.SetConstantFee(ctx, coin)

	msg := NewMsgVerifyInvariance(sender, dummyRouteWhichPasses.Route)
	res := handleMsgVerifyInvariant(ctx, msg, crisisKeeper)
	require.True(t, res.IsOK())
}

func TestHandleMsgVerifyInvariantWithInvariantBroken(t *testing.T) {
	ctx, crisisKeeper, accKeeper := CreateTestInput(t)
	sender := addrs[0]
	coin := accKeeper.GetAccount(ctx, sender).GetCoins()[0]
	crisisKeeper.SetConstantFee(ctx, coin)

	msg := NewMsgVerifyInvariance(sender, dummyRouteWhichFails.Route)
	var res sdk.Result
	require.Panics(t, func() {
		res = handleMsgVerifyInvariant(ctx, msg, crisisKeeper)
	}, fmt.Sprintf("%v", res))
}
