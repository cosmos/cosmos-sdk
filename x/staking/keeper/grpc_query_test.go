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

	// Query bonded validator
	valResp, err := queryClient.DelegatorValidator(gocontext.Background(), &types.QueryBondParamsRequest{})
	require.Error(t, err)
	require.Nil(t, valResp)

	req := &types.QueryBondParamsRequest{DelegatorAddr: addrAcc2, ValidatorAddr: addrVal1}
	valResp, err = queryClient.DelegatorValidator(gocontext.Background(), req)
	require.NoError(t, err)

	require.Equal(t, delValidators[0], valResp.Validator)

	// Query delegation
	delegationResp, err := queryClient.DelegationQ(gocontext.Background(), &types.QueryBondParamsRequest{})
	require.Error(t, err)
	require.Nil(t, delegationResp)

	req = &types.QueryBondParamsRequest{DelegatorAddr: addrAcc2, ValidatorAddr: addrVal1, Req: &query.PageRequest{}}
	delegation, found := app.StakingKeeper.GetDelegation(ctx, addrAcc2, addrVal1)
	require.True(t, found)

	delegationResp, err = queryClient.DelegationQ(gocontext.Background(), req)
	require.NoError(t, err)

	require.Equal(t, delegation.ValidatorAddress, delegationResp.DelegationResponse.Delegation.ValidatorAddress)
	require.Equal(t, delegation.DelegatorAddress, delegationResp.DelegationResponse.Delegation.DelegatorAddress)
	require.Equal(t, sdk.NewCoin(sdk.DefaultBondDenom, delegation.Shares.TruncateInt()), delegationResp.DelegationResponse.Balance)

	// Query Delegator Delegations
	delegatorDel, err := queryClient.DelegatorDelegations(gocontext.Background(), &types.QueryDelegatorParamsRequest{})
	require.Error(t, err)
	require.Nil(t, delegatorDel)

	delReq := &types.QueryDelegatorParamsRequest{DelegatorAddr: addrAcc2, Req: &query.PageRequest{}}
	delegatorDelegations, err := queryClient.DelegatorDelegations(gocontext.Background(), delReq)
	require.NoError(t, err)
	require.NotNil(t, delegatorDelegations)

	require.Len(t, delegatorDelegations.DelegationResponses, 1)
	require.Equal(t, delegation.ValidatorAddress, delegatorDelegations.DelegationResponses[0].Delegation.ValidatorAddress)
	require.Equal(t, delegation.DelegatorAddress, delegatorDelegations.DelegationResponses[0].Delegation.DelegatorAddress)
	require.Equal(t, sdk.NewCoin(sdk.DefaultBondDenom, delegation.Shares.TruncateInt()), delegatorDelegations.DelegationResponses[0].Balance)

	// Query validator delegations
	validatorDelegations, err := queryClient.ValidatorDelegations(gocontext.Background(), &types.QueryValidatorRequest{})
	require.Error(t, err)
	require.Nil(t, validatorDelegations)

	valReq := &types.QueryValidatorRequest{ValidatorAddr: addrVal1, Req: &query.PageRequest{}}
	delegationsRes, err := queryClient.ValidatorDelegations(gocontext.Background(), valReq)

	require.NoError(t, err)
	require.Len(t, delegatorDelegations.DelegationResponses, 1)
	require.Equal(t, delegation.ValidatorAddress, delegationsRes.DelegationResponses[0].Delegation.ValidatorAddress)
	require.Equal(t, delegation.DelegatorAddress, delegationsRes.DelegationResponses[0].Delegation.DelegatorAddress)
	require.Equal(t, sdk.NewCoin(sdk.DefaultBondDenom, delegation.Shares.TruncateInt()), delegationsRes.DelegationResponses[0].Balance)

	// Query unbonding delegation
	unbondingTokens := sdk.TokensFromConsensusPower(10)
	_, err = app.StakingKeeper.Undelegate(ctx, addrAcc2, val1.OperatorAddress, unbondingTokens.ToDec())
	require.NoError(t, err)

	unbondRes, err := queryClient.UnbondingDelegation(gocontext.Background(), &types.QueryBondParamsRequest{})
	require.Error(t, err)
	require.Nil(t, unbondRes)

	unbondReq := &types.QueryBondParamsRequest{DelegatorAddr: addrAcc2, ValidatorAddr: addrVal1, Req: &query.PageRequest{}}
	unbondRes, err = queryClient.UnbondingDelegation(gocontext.Background(), unbondReq)

	unbond, found := app.StakingKeeper.GetUnbondingDelegation(ctx, addrAcc2, addrVal1)
	require.True(t, found)
	require.NotNil(t, unbondRes)
	require.Equal(t, unbond, unbondRes.Unbond)

	// Query Delegator Unbonding Delegations
	delegatorUbds, err := queryClient.DelegatorUnbondingDelegations(gocontext.Background(), &types.QueryDelegatorParamsRequest{})
	require.Error(t, err)
	require.Nil(t, delegatorUbds)

	unbReq := &types.QueryDelegatorParamsRequest{DelegatorAddr: addrAcc2, Req: &query.PageRequest{}}
	delegatorUbds, err = queryClient.DelegatorUnbondingDelegations(gocontext.Background(), unbReq)
	require.NoError(t, err)
	require.Equal(t, unbond, delegatorUbds.UnbondingResponses[0])

	// Query Redelegations
	redelResp, err := queryClient.Redelegations(gocontext.Background(), &types.QueryRedelegationsRequest{})
	require.Error(t, err)
	require.Nil(t, redelResp)

	redelegationTokens := sdk.TokensFromConsensusPower(10)
	_, err = app.StakingKeeper.BeginRedelegation(ctx, addrAcc2, val1.OperatorAddress,
		val2.OperatorAddress, redelegationTokens.ToDec())
	require.NoError(t, err)
	redel, found := app.StakingKeeper.GetRedelegation(ctx, addrAcc2, val1.OperatorAddress, val2.OperatorAddress)
	require.True(t, found)

	redelReq := &types.QueryRedelegationsRequest{
		DelegatorAddr: addrAcc2, SrcValidatorAddr: val1.OperatorAddress, DstValidatorAddr: val2.OperatorAddress,
		Req: &query.PageRequest{}}
	redelResp, err = queryClient.Redelegations(gocontext.Background(), redelReq)

	require.NoError(t, err)
	require.Len(t, redelResp.RedelegationResponses, 1)
	require.Equal(t, redel.DelegatorAddress, redelResp.RedelegationResponses[0].Redelegation.DelegatorAddress)
	require.Equal(t, redel.ValidatorSrcAddress, redelResp.RedelegationResponses[0].Redelegation.ValidatorSrcAddress)
	require.Equal(t, redel.ValidatorDstAddress, redelResp.RedelegationResponses[0].Redelegation.ValidatorDstAddress)
	require.Len(t, redel.Entries, len(redelResp.RedelegationResponses[0].Entries))
}
