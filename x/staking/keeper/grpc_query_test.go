package keeper_test

import (
	gocontext "context"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestGRPCQueryValidators(t *testing.T) {
	_, app, ctx := createTestInput()

	addrs := simapp.AddTestAddrs(app, ctx, 500, sdk.TokensFromConsensusPower(10000))
	querier := keeper.Querier{app.StakingKeeper}

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
	types.RegisterQueryServer(queryHelper, querier)
	queryClient := types.NewQueryClient(queryHelper)

	vals, err := queryClient.Validators(gocontext.Background(), &types.QueryValidatorsRequest{})
	require.Error(t, err)
	require.Nil(t, vals)

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
	require.Equal(t, 1, len(res.Validators))
	require.NotNil(t, res.Res.NextKey)

	valRes, err := queryClient.Validator(gocontext.Background(), &types.QueryValidatorRequest{})
	require.Error(t, err)
	require.Nil(t, valRes)

	for _, validator := range validators {
		valRes, err = queryClient.Validator(gocontext.Background(), &types.QueryValidatorRequest{ValidatorAddr: validator.OperatorAddress})
		require.NoError(t, err)
		require.Equal(t, validator, valRes.Validator)
	}
}

func TestGRPCQueryDelegation(t *testing.T) {
	_, app, ctx := createTestInput()
	params := app.StakingKeeper.GetParams(ctx)
	querier := keeper.Querier{app.StakingKeeper}

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
	types.RegisterQueryServer(queryHelper, querier)
	queryClient := types.NewQueryClient(queryHelper)

	res, err := queryClient.DelegatorValidators(gocontext.Background(), &types.QueryDelegatorValidatorsRequest{})
	require.Error(t, err)
	require.Nil(t, res)

	delegatorParamsReq := &types.QueryDelegatorValidatorsRequest{DelegatorAddr: addrAcc2}

	delValidators := app.StakingKeeper.GetDelegatorValidators(ctx, addrAcc2, params.MaxValidators)

	res, err = queryClient.DelegatorValidators(gocontext.Background(), delegatorParamsReq)
	require.NoError(t, err)

	require.Equal(t, len(delValidators), len(res.Validators))
	require.ElementsMatch(t, delValidators, res.Validators)

	delegatorParamsReq = &types.QueryDelegatorValidatorsRequest{
		DelegatorAddr: addrAcc2,
		Req:           &query.PageRequest{Limit: 1},
	}

	res, err = queryClient.DelegatorValidators(gocontext.Background(), delegatorParamsReq)
	require.NoError(t, err)

	require.Equal(t, 1, len(res.Validators))
	require.ElementsMatch(t, delValidators, res.Validators)

	// Query bonded validator
	valResp, err := queryClient.DelegatorValidator(gocontext.Background(), &types.QueryDelegatorValidatorRequest{})
	require.Error(t, err)
	require.Nil(t, valResp)

	req := &types.QueryDelegatorValidatorRequest{DelegatorAddr: addrAcc2, ValidatorAddr: addrVal1}
	valResp, err = queryClient.DelegatorValidator(gocontext.Background(), req)
	require.NoError(t, err)

	require.Equal(t, delValidators[0], valResp.Validator)

	// Query delegation
	delegationResp, err := queryClient.Delegation(gocontext.Background(), &types.QueryDelegationRequest{})
	require.Error(t, err)
	require.Nil(t, delegationResp)

	delReq := &types.QueryDelegationRequest{DelegatorAddr: addrAcc2, ValidatorAddr: addrVal1, Req: &query.PageRequest{}}
	delegation, found := app.StakingKeeper.GetDelegation(ctx, addrAcc2, addrVal1)
	require.True(t, found)

	delegationResp, err = queryClient.Delegation(gocontext.Background(), delReq)
	require.NoError(t, err)

	require.Equal(t, delegation.ValidatorAddress, delegationResp.DelegationResponse.Delegation.ValidatorAddress)
	require.Equal(t, delegation.DelegatorAddress, delegationResp.DelegationResponse.Delegation.DelegatorAddress)
	require.Equal(t, sdk.NewCoin(sdk.DefaultBondDenom, delegation.Shares.TruncateInt()), delegationResp.DelegationResponse.Balance)

	// Query Delegator Delegations
	delegatorDel, err := queryClient.DelegatorDelegations(gocontext.Background(), &types.QueryDelegatorDelegationsRequest{})
	require.Error(t, err)
	require.Nil(t, delegatorDel)

	delDelReq := &types.QueryDelegatorDelegationsRequest{DelegatorAddr: addrAcc2, Req: &query.PageRequest{}}
	delegatorDelegations, err := queryClient.DelegatorDelegations(gocontext.Background(), delDelReq)
	require.NoError(t, err)
	require.NotNil(t, delegatorDelegations)

	require.Len(t, delegatorDelegations.DelegationResponses, 1)
	require.Equal(t, delegation.ValidatorAddress, delegatorDelegations.DelegationResponses[0].Delegation.ValidatorAddress)
	require.Equal(t, delegation.DelegatorAddress, delegatorDelegations.DelegationResponses[0].Delegation.DelegatorAddress)
	require.Equal(t, sdk.NewCoin(sdk.DefaultBondDenom, delegation.Shares.TruncateInt()), delegatorDelegations.DelegationResponses[0].Balance)

	// Query validator delegations
	validatorDelegations, err := queryClient.ValidatorDelegations(gocontext.Background(), &types.QueryValidatorDelegationsRequest{})
	require.Error(t, err)
	require.Nil(t, validatorDelegations)

	valReq := &types.QueryValidatorDelegationsRequest{ValidatorAddr: addrVal1, Req: &query.PageRequest{}}
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

	unbondRes, err := queryClient.UnbondingDelegation(gocontext.Background(), &types.QueryUnbondingDelegationRequest{})
	require.Error(t, err)
	require.Nil(t, unbondRes)

	unbondReq := &types.QueryUnbondingDelegationRequest{DelegatorAddr: addrAcc2, ValidatorAddr: addrVal1, Req: &query.PageRequest{}}
	unbondRes, err = queryClient.UnbondingDelegation(gocontext.Background(), unbondReq)

	unbond, found := app.StakingKeeper.GetUnbondingDelegation(ctx, addrAcc2, addrVal1)
	require.True(t, found)
	require.NotNil(t, unbondRes)
	require.Equal(t, unbond, unbondRes.Unbond)

	// Query Delegator Unbonding Delegations
	delegatorUbds, err := queryClient.DelegatorUnbondingDelegations(gocontext.Background(), &types.QueryDelegatorUnbondingDelegationsRequest{})
	require.Error(t, err)
	require.Nil(t, delegatorUbds)

	unbReq := &types.QueryDelegatorUnbondingDelegationsRequest{DelegatorAddr: addrAcc2, Req: &query.PageRequest{}}
	delegatorUbds, err = queryClient.DelegatorUnbondingDelegations(gocontext.Background(), unbReq)
	require.NoError(t, err)
	require.Equal(t, unbond, delegatorUbds.UnbondingResponses[0])

	// Query Redelegations
	redelResp, err := queryClient.Redelegations(gocontext.Background(), &types.QueryRedelegationsRequest{})
	require.NoError(t, err)

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

func TestGRPCQueryPoolParameters(t *testing.T) {
	_, app, ctx := createTestInput()
	bondDenom := sdk.DefaultBondDenom
	querier := keeper.Querier{app.StakingKeeper}

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, querier)
	queryClient := types.NewQueryClient(queryHelper)

	// Query pool
	res, err := queryClient.Pool(gocontext.Background(), &types.QueryPoolRequest{})
	require.NoError(t, err)
	bondedPool := app.StakingKeeper.GetBondedPool(ctx)
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
	require.Equal(t, app.BankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), bondDenom).Amount, res.Pool.NotBondedTokens)
	require.Equal(t, app.BankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount, res.Pool.BondedTokens)

	// Query Params
	resp, err := queryClient.Parameters(gocontext.Background(), &types.QueryParametersRequest{})
	require.NoError(t, err)
	require.Equal(t, app.StakingKeeper.GetParams(ctx), resp.Params)
}

func TestGRPCHistoricalInfo(t *testing.T) {
	_, app, ctx := createTestInput()
	querier := keeper.Querier{app.StakingKeeper}

	addrs := simapp.AddTestAddrs(app, ctx, 2, sdk.TokensFromConsensusPower(10000))
	addrAcc1, addrAcc2 := addrs[0], addrs[1]
	addrVal1, addrVal2 := sdk.ValAddress(addrAcc1), sdk.ValAddress(addrAcc2)

	// Create Validators and Delegation
	val1 := types.NewValidator(addrVal1, PKs[0], types.Description{})
	val2 := types.NewValidator(addrVal2, PKs[1], types.Description{})
	vals := []types.Validator{val1, val2}
	app.StakingKeeper.SetValidator(ctx, val1)
	app.StakingKeeper.SetValidator(ctx, val2)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, querier)
	queryClient := types.NewQueryClient(queryHelper)

	header := abci.Header{
		ChainID: "HelloChain",
		Height:  5,
	}
	hi := types.NewHistoricalInfo(header, vals)
	app.StakingKeeper.SetHistoricalInfo(ctx, 5, hi)

	hist, err := queryClient.HistoricalInfo(gocontext.Background(), &types.QueryHistoricalInfoRequest{})
	require.Error(t, err)
	require.Nil(t, hist)

	hist, err = queryClient.HistoricalInfo(gocontext.Background(), &types.QueryHistoricalInfoRequest{Height: 4})
	require.Error(t, err)
	require.Nil(t, hist)

	hist, err = queryClient.HistoricalInfo(gocontext.Background(), &types.QueryHistoricalInfoRequest{Height: 5})
	require.NoError(t, err)
	require.NotNil(t, hist)
	require.Equal(t, &hi, hist.Hist)
}

func TestGRPCQueryRedelegation(t *testing.T) {
	_, app, ctx := createTestInput()
	addrs := simapp.AddTestAddrs(app, ctx, 2, sdk.TokensFromConsensusPower(10000))
	addrAcc1, addrAcc2 := addrs[0], addrs[1]
	addrVal1, addrVal2 := sdk.ValAddress(addrAcc1), sdk.ValAddress(addrAcc2)
	querier := keeper.Querier{app.StakingKeeper}

	// Create Validators and Delegation
	val1 := types.NewValidator(addrVal1, PKs[0], types.Description{})
	val2 := types.NewValidator(addrVal2, PKs[1], types.Description{})
	app.StakingKeeper.SetValidator(ctx, val1)
	app.StakingKeeper.SetValidator(ctx, val2)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, querier)
	queryClient := types.NewQueryClient(queryHelper)

	delAmount := sdk.TokensFromConsensusPower(100)
	_, err := app.StakingKeeper.Delegate(ctx, addrAcc2, delAmount, sdk.Unbonded, val1, true)
	require.NoError(t, err)
	_ = app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)

	rdAmount := sdk.TokensFromConsensusPower(20)
	_, err = app.StakingKeeper.BeginRedelegation(ctx, addrAcc2, val1.GetOperator(), val2.GetOperator(), rdAmount.ToDec())
	require.NoError(t, err)
	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)

	redel, found := app.StakingKeeper.GetRedelegation(ctx, addrAcc2, val1.OperatorAddress, val2.OperatorAddress)
	require.True(t, found)
	t.Log(redel)

	// delegator redelegations
	redelResp, err := queryClient.Redelegations(gocontext.Background(), &types.QueryRedelegationsRequest{
		DelegatorAddr: addrAcc2, SrcValidatorAddr: val1.OperatorAddress, Req: &query.PageRequest{}})
	require.NoError(t, err)

	require.Len(t, redelResp.RedelegationResponses, 1)
	require.Equal(t, redel.DelegatorAddress, redelResp.RedelegationResponses[0].Redelegation.DelegatorAddress)
	require.Equal(t, redel.ValidatorSrcAddress, redelResp.RedelegationResponses[0].Redelegation.ValidatorSrcAddress)
	require.Equal(t, redel.ValidatorDstAddress, redelResp.RedelegationResponses[0].Redelegation.ValidatorDstAddress)
	require.Len(t, redel.Entries, len(redelResp.RedelegationResponses[0].Entries))

	redels := app.StakingKeeper.GetRedelegationsFromSrcValidator(ctx, val1.OperatorAddress)
	require.True(t, found)
	t.Log(redels)

	redelResp, err = queryClient.Redelegations(gocontext.Background(), &types.QueryRedelegationsRequest{
		SrcValidatorAddr: val1.GetOperator(), Req: &query.PageRequest{}})
	require.NoError(t, err)
	t.Log(val1.OperatorAddress)
	require.Len(t, redelResp.RedelegationResponses, 1)
	require.Equal(t, redel.DelegatorAddress, redelResp.RedelegationResponses[0].Redelegation.DelegatorAddress)
	require.Equal(t, redel.ValidatorSrcAddress, redelResp.RedelegationResponses[0].Redelegation.ValidatorSrcAddress)
	require.Equal(t, redel.ValidatorDstAddress, redelResp.RedelegationResponses[0].Redelegation.ValidatorDstAddress)
	require.Len(t, redel.Entries, len(redelResp.RedelegationResponses[0].Entries))
}

func TestGRPCQueryUnbondingDelegation(t *testing.T) {
	_, app, ctx := createTestInput()
	querier := keeper.Querier{app.StakingKeeper}

	addrs := simapp.AddTestAddrs(app, ctx, 2, sdk.TokensFromConsensusPower(10000))
	addrAcc1, addrAcc2 := addrs[0], addrs[1]
	addrVal1 := sdk.ValAddress(addrAcc1)

	// Create Validators and Delegation
	val1 := types.NewValidator(addrVal1, PKs[0], types.Description{})
	app.StakingKeeper.SetValidator(ctx, val1)

	// delegate
	delAmount := sdk.TokensFromConsensusPower(100)
	_, err := app.StakingKeeper.Delegate(ctx, addrAcc1, delAmount, sdk.Unbonded, val1, true)
	require.NoError(t, err)
	_ = app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)

	// undelegate
	undelAmount := sdk.TokensFromConsensusPower(20)
	_, err = app.StakingKeeper.Undelegate(ctx, addrAcc1, val1.GetOperator(), undelAmount.ToDec())
	require.NoError(t, err)
	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, querier)
	queryClient := types.NewQueryClient(queryHelper)

	unbDelsResp, err := queryClient.UnbondingDelegation(gocontext.Background(), &types.QueryUnbondingDelegationRequest{
		DelegatorAddr: addrAcc1, ValidatorAddr: val1.GetOperator()})

	require.Equal(t, addrAcc1, unbDelsResp.Unbond.DelegatorAddress)
	require.Equal(t, val1.OperatorAddress, unbDelsResp.Unbond.ValidatorAddress)
	require.Equal(t, 1, len(unbDelsResp.Unbond.Entries))

	unbDelsResp, err = queryClient.UnbondingDelegation(gocontext.Background(), &types.QueryUnbondingDelegationRequest{
		DelegatorAddr: addrAcc2, ValidatorAddr: val1.GetOperator()})
	require.Error(t, err)

	valUnbonds, err := queryClient.ValidatorUnbondingDelegations(gocontext.Background(), &types.QueryValidatorUnbondingDelegationsRequest{})
	require.Error(t, err)
	require.Nil(t, valUnbonds)

	valUnbonds, err = queryClient.ValidatorUnbondingDelegations(gocontext.Background(),
		&types.QueryValidatorUnbondingDelegationsRequest{ValidatorAddr: val1.GetOperator()})
	require.NoError(t, err)
	require.Equal(t, 1, len(valUnbonds.UnbondingResponses))
}
