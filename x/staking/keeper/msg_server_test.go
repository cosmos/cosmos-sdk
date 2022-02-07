package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestCreateValidatorWithLessThanMinCommission(t *testing.T) {
	PKS := simapp.CreateTestPubKeys(2)
	valConsPk1 := PKS[0]
	valConsPk2 := PKS[1]

	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addrs := simapp.AddTestAddrs(app, ctx, 3, sdk.NewInt(1234))

	// set min commission rate to non-zero
	params := app.StakingKeeper.GetParams(ctx)
	params.MinCommissionRate = sdk.NewDecWithPrec(1, 2)
	app.StakingKeeper.SetParams(ctx, params)

	// create validator with 0% commission
	msg1, err := stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(addrs[0]),
		valConsPk1,
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
		stakingtypes.Description{},
		stakingtypes.NewCommissionRates(sdk.NewDec(0), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0)),
		sdk.OneInt())
	require.NoError(t, err)
	msg2, err := stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(addrs[1]),
		valConsPk2,
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
		stakingtypes.Description{},
		stakingtypes.NewCommissionRates(sdk.NewDec(0), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0)),
		sdk.OneInt())
	require.NoError(t, err)

	sh := staking.NewHandler(app.StakingKeeper)
	_, err = sh(ctx, msg1)
	require.NoError(t, err)
	_, err = sh(ctx.WithBlockHeight(712001), msg2)
	require.Error(t, err)
}
