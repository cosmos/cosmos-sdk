package crisis

import (
	"errors"
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

	return ctx, crisisKeeper, accKeeper
}

//____________________________________________________________________________

func TestHandleMsgVerifyInvariantWithNotEnoughCoins(t *testing.T) {
	ctx, crisisKeeper, accKeeper := CreateTestInput(t)
	sender := addrs[0]
	coins := accKeeper.GetAccount(ctx, sender).GetCoins()
	excessCoins := sdk.NewCoin(coins[0].Denom, coins[0].Amount.AddRaw(1))
	crisisKeeper.SetConstantFee(ctx, excessCoins)

	msg := NewMsgVerifyInvariance(sender, "dummy_route")
	res := handleMsgVerifyInvariant(ctx, msg, crisisKeeper)
	require.False(t, res.IsOK())
}

func TestHandleMsgVerifyInvariantWithBadInvariant(t *testing.T) {
	// TODO
}

func TestHandleMsgVerifyInvariantWithInvariantNotBroken(t *testing.T) {
	// TODO
}

func TestHandleMsgVerifyInvariantWithInvariantBroken(t *testing.T) {
	// TODO
}
