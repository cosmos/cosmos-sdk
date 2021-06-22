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
	PKS := simapp.CreateTestPubKeys(1)
	valConsPk1 := PKS[0]

	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addrs := simapp.AddTestAddrs(app, ctx, 3, sdk.NewInt(1234))

	// set min commission rate to non-zero
	params := app.StakingKeeper.GetParams(ctx)
	params.MinCommissionRate = sdk.NewDecWithPrec(1, 2)
	app.StakingKeeper.SetParams(ctx, params)

	// create validator with 0% commission
	msg, err := stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(addrs[0]),
		valConsPk1,
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
		stakingtypes.Description{},
		stakingtypes.NewCommissionRates(sdk.NewDec(0), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0)),
		sdk.OneInt())
	require.NoError(t, err)

	sh := staking.NewHandler(app.StakingKeeper)
	_, err = sh(ctx, msg)
	require.Error(t, err)
}
