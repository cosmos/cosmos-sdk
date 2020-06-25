package keeper_test

import (
	gocontext "context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestGRPCQueryValidators(t *testing.T) {
	_, app, ctx := createTestInput()

	addrs := simapp.AddTestAddrs(app, ctx, 500, sdk.TokensFromConsensusPower(10000))

	amts := []sdk.Int{sdk.NewInt(9), sdk.NewInt(8), sdk.NewInt(7)}
	status := []sdk.BondStatus{sdk.Bonded, sdk.Unbonded, sdk.Unbonding}
	var validators [3]types.Validator
	for i, amt := range amts {
		validators[i] = types.NewValidator(sdk.ValAddress(addrs[i]), PKs[i], types.Description{})
		validators[i], _ = validators[i].AddTokensFromDel(amt)
		validators[i] = validators[i].UpdateStatus(status[i])
	}

	app.StakingKeeper.SetValidator(ctx, validators[0])
	app.StakingKeeper.SetValidator(ctx, validators[1])
	app.StakingKeeper.SetValidator(ctx, validators[2])

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, app.StakingKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	_, err := queryClient.Validators(gocontext.Background(), &types.QueryValidatorsRequest{})
	require.Error(t, err)

	for i, s := range status {
		req := &types.QueryValidatorsRequest{Status: s.String(), Req: &query.PageRequest{Limit: 5}}
		res, err := queryClient.Validators(gocontext.Background(), req)
		require.NoError(t, err)

		require.Equal(t, 1, len(res.Validators))
		require.Equal(t, validators[i].OperatorAddress, res.Validators[0].OperatorAddress)
	}

	// check pagination
	pageReq := &query.PageRequest{Limit: 1}

	res, err := queryClient.Validators(gocontext.Background(), &types.QueryValidatorsRequest{Status: status[0].String(), Req: pageReq})
	require.NoError(t, err)
	require.NotNil(t, res)
	// require.Equal(t, 1, len(res.Validators))
	require.NotNil(t, res.Res.NextKey)

	valRes, err := queryClient.ValidatorQ(gocontext.Background(), &types.QueryValidatorRequest{})
	require.Error(t, err)
	require.Nil(t, valRes)

	for _, validator := range validators {
		valRes, err = queryClient.ValidatorQ(gocontext.Background(), &types.QueryValidatorRequest{ValidatorAddr: validator.OperatorAddress})
		require.NoError(t, err)
		require.Equal(t, validator, valRes.Validator)
	}
}

func TestGRPCQueryDelegation(t *testing.T) {
	_, app, ctx := createTestInput()
	params := app.StakingKeeper.GetParams(ctx)

	addrs := simapp.AddTestAddrs(app, ctx, 2, sdk.TokensFromConsensusPower(10000))
	addrAcc1, addrAcc2 := addrs[0], addrs[1]
	addrVal1, addrVal2 := sdk.ValAddress(addrAcc1), sdk.ValAddress(addrAcc2)

	pubKeys := simapp.CreateTestPubKeys(2)
	pk1, pk2 := pubKeys[0], pubKeys[1]

	// Create Validators and Delegation
	val1 := types.NewValidator(addrVal1, pk1, types.Description{})
	app.StakingKeeper.SetValidator(ctx, val1)
	app.StakingKeeper.SetValidatorByPowerIndex(ctx, val1)

	val2 := types.NewValidator(addrVal2, pk2, types.Description{})
	app.StakingKeeper.SetValidator(ctx, val2)
	app.StakingKeeper.SetValidatorByPowerIndex(ctx, val2)

	delTokens := sdk.TokensFromConsensusPower(20)
	_, err := app.StakingKeeper.Delegate(ctx, addrAcc2, delTokens, sdk.Unbonded, val1, true)
	require.NoError(t, err)

	// apply TM updates
	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, app.StakingKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	res, err := queryClient.DelegatorValidators(gocontext.Background(), &types.QueryDelegatorParamsRequest{})
	require.Error(t, err)
	require.Nil(t, res)

	delegatorParamsReq := &types.QueryDelegatorParamsRequest{
		DelegatorAddr: addrAcc2,
		Req:           &query.PageRequest{},
	}

	delValidators := app.StakingKeeper.GetDelegatorValidators(ctx, addrAcc2, params.MaxValidators)

	res, err = queryClient.DelegatorValidators(gocontext.Background(), delegatorParamsReq)
	require.NoError(t, err)

	require.Equal(t, len(delValidators), len(res.Validators))
	require.ElementsMatch(t, delValidators, res.Validators)
}
