package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestNewQuerier(t *testing.T) {
	cdc, app, ctx := createTestInput()

	addrs := simapp.AddTestAddrs(app, ctx, 500, sdk.NewInt(10000))
	_, addrAcc2 := addrs[0], addrs[1]
	addrVal1, _ := sdk.ValAddress(addrs[0]), sdk.ValAddress(addrs[1])

	// Create Validators
	amts := []sdk.Int{sdk.NewInt(9), sdk.NewInt(8)}
	var validators [2]types.Validator
	for i, amt := range amts {
		validators[i] = types.NewValidator(sdk.ValAddress(addrs[i]), PKs[i], types.Description{})
		validators[i], _ = validators[i].AddTokensFromDel(amt)
		app.StakingKeeper.SetValidator(ctx, validators[i])
		app.StakingKeeper.SetValidatorByPowerIndex(ctx, validators[i])
	}

	header := abci.Header{
		ChainID: "HelloChain",
		Height:  5,
	}
	hi := types.NewHistoricalInfo(header, validators[:])
	app.StakingKeeper.SetHistoricalInfo(ctx, 5, hi)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	querier := keeper.NewQuerier(app.StakingKeeper)

	bz, err := querier(ctx, []string{"other"}, query)
	require.Error(t, err)
	require.Nil(t, bz)

	_, err = querier(ctx, []string{"pool"}, query)
	require.NoError(t, err)

	_, err = querier(ctx, []string{"parameters"}, query)
	require.NoError(t, err)

	queryValParams := types.NewQueryValidatorParams(addrVal1, 0, 0)
	bz, errRes := cdc.MarshalJSON(queryValParams)
	require.NoError(t, errRes)

	query.Path = "/custom/staking/validator"
	query.Data = bz

	_, err = querier(ctx, []string{"validator"}, query)
	require.NoError(t, err)

	_, err = querier(ctx, []string{"validatorDelegations"}, query)
	require.NoError(t, err)

	_, err = querier(ctx, []string{"validatorUnbondingDelegations"}, query)
	require.NoError(t, err)

	queryDelParams := types.NewQueryDelegatorParams(addrAcc2)
	bz, errRes = cdc.MarshalJSON(queryDelParams)
	require.NoError(t, errRes)

	query.Path = "/custom/staking/validator"
	query.Data = bz

	_, err = querier(ctx, []string{"delegatorDelegations"}, query)
	require.NoError(t, err)

	_, err = querier(ctx, []string{"delegatorUnbondingDelegations"}, query)
	require.NoError(t, err)

	_, err = querier(ctx, []string{"delegatorValidators"}, query)
	require.NoError(t, err)

	bz, errRes = cdc.MarshalJSON(types.NewQueryRedelegationParams(nil, nil, nil))
	require.NoError(t, errRes)
	query.Data = bz

	_, err = querier(ctx, []string{"redelegations"}, query)
	require.NoError(t, err)

	queryHisParams := types.NewQueryHistoricalInfoParams(5)
	bz, errRes = cdc.MarshalJSON(queryHisParams)
	require.NoError(t, errRes)

	query.Path = "/custom/staking/historicalInfo"
	query.Data = bz

	_, err = querier(ctx, []string{"historicalInfo"}, query)
	require.NoError(t, err)
}

func TestQueryParametersPool(t *testing.T) {
	cdc, app, ctx := createTestInput()
	querier := keeper.NewQuerier(app.StakingKeeper)

	bondDenom := sdk.DefaultBondDenom

	res, err := querier(ctx, []string{types.QueryParameters}, abci.RequestQuery{})
	require.NoError(t, err)

	var params types.Params
	errRes := cdc.UnmarshalJSON(res, &params)
	require.NoError(t, errRes)
	require.Equal(t, app.StakingKeeper.GetParams(ctx), params)

	res, err = querier(ctx, []string{types.QueryPool}, abci.RequestQuery{})
	require.NoError(t, err)

	var pool types.Pool
	bondedPool := app.StakingKeeper.GetBondedPool(ctx)
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
	require.NoError(t, cdc.UnmarshalJSON(res, &pool))
	require.Equal(t, app.BankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), bondDenom).Amount, pool.NotBondedTokens)
	require.Equal(t, app.BankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount, pool.BondedTokens)
}

func TestQueryValidators(t *testing.T) {
	cdc, app, ctx := createTestInput()
	params := app.StakingKeeper.GetParams(ctx)
	querier := keeper.NewQuerier(app.StakingKeeper)

	addrs := simapp.AddTestAddrs(app, ctx, 500, sdk.TokensFromConsensusPower(10000))

	// Create Validators
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

	// Query Validators
	queriedValidators := app.StakingKeeper.GetValidators(ctx, params.MaxValidators)
	require.Len(t, queriedValidators, 3)

	for i, s := range status {
		queryValsParams := types.NewQueryValidatorsParams(1, int(params.MaxValidators), s.String())
		bz, err := cdc.MarshalJSON(queryValsParams)
		require.NoError(t, err)

		req := abci.RequestQuery{
			Path: fmt.Sprintf("/custom/%s/%s", types.QuerierRoute, types.QueryValidators),
			Data: bz,
		}

		res, err := querier(ctx, []string{types.QueryValidators}, req)
		require.NoError(t, err)

		var validatorsResp []types.Validator
		err = cdc.UnmarshalJSON(res, &validatorsResp)
		require.NoError(t, err)

		require.Equal(t, 1, len(validatorsResp))
		require.ElementsMatch(t, validators[i].OperatorAddress, validatorsResp[0].OperatorAddress)
	}

	// Query each validator
	for _, validator := range validators {
		queryParams := types.NewQueryValidatorParams(validator.OperatorAddress, 0, 0)
		bz, err := cdc.MarshalJSON(queryParams)
		require.NoError(t, err)

		query := abci.RequestQuery{
			Path: "/custom/staking/validator",
			Data: bz,
		}
		res, err := querier(ctx, []string{types.QueryValidator}, query)
		require.NoError(t, err)

		var queriedValidator types.Validator
		err = cdc.UnmarshalJSON(res, &queriedValidator)
		require.NoError(t, err)

		require.Equal(t, validator, queriedValidator)
	}
}

func TestQueryDelegation(t *testing.T) {
	cdc, app, ctx := createTestInput()
	params := app.StakingKeeper.GetParams(ctx)
	querier := keeper.NewQuerier(app.StakingKeeper)

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

	// Query Delegator bonded validators
	queryParams := types.NewQueryDelegatorParams(addrAcc2)
	bz, errRes := cdc.MarshalJSON(queryParams)
	require.NoError(t, errRes)

	query := abci.RequestQuery{
		Path: "/custom/staking/delegatorValidators",
		Data: bz,
	}

	delValidators := app.StakingKeeper.GetDelegatorValidators(ctx, addrAcc2, params.MaxValidators)

	res, err := querier(ctx, []string{types.QueryDelegatorValidators}, query)
	require.NoError(t, err)

	var validatorsResp []types.Validator
	errRes = cdc.UnmarshalJSON(res, &validatorsResp)
	require.NoError(t, errRes)

	require.Equal(t, len(delValidators), len(validatorsResp))
	require.ElementsMatch(t, delValidators, validatorsResp)

	// error unknown request
	query.Data = bz[:len(bz)-1]

	_, err = querier(ctx, []string{types.QueryDelegatorValidators}, query)
	require.Error(t, err)

	// Query bonded validator
	queryBondParams := types.NewQueryBondsParams(addrAcc2, addrVal1)
	bz, errRes = cdc.MarshalJSON(queryBondParams)
	require.NoError(t, errRes)

	query = abci.RequestQuery{
		Path: "/custom/staking/delegatorValidator",
		Data: bz,
	}

	res, err = querier(ctx, []string{types.QueryDelegatorValidator}, query)
	require.NoError(t, err)

	var validator types.Validator
	errRes = cdc.UnmarshalJSON(res, &validator)
	require.NoError(t, errRes)

	require.Equal(t, delValidators[0], validator)

	// error unknown request
	query.Data = bz[:len(bz)-1]

	_, err = querier(ctx, []string{types.QueryDelegatorValidator}, query)
	require.Error(t, err)

	// Query delegation

	query = abci.RequestQuery{
		Path: "/custom/staking/delegation",
		Data: bz,
	}

	delegation, found := app.StakingKeeper.GetDelegation(ctx, addrAcc2, addrVal1)
	require.True(t, found)

	res, err = querier(ctx, []string{types.QueryDelegation}, query)
	require.NoError(t, err)

	var delegationRes types.DelegationResponse
	errRes = cdc.UnmarshalJSON(res, &delegationRes)
	require.NoError(t, errRes)

	require.Equal(t, delegation.ValidatorAddress, delegationRes.ValidatorAddress)
	require.Equal(t, delegation.DelegatorAddress, delegationRes.DelegatorAddress)
	require.Equal(t, sdk.NewCoin(sdk.DefaultBondDenom, delegation.Shares.TruncateInt()), delegationRes.Balance)

	// Query Delegator Delegations
	query = abci.RequestQuery{
		Path: "/custom/staking/delegatorDelegations",
		Data: bz,
	}

	res, err = querier(ctx, []string{types.QueryDelegatorDelegations}, query)
	require.NoError(t, err)

	var delegatorDelegations types.DelegationResponses
	errRes = cdc.UnmarshalJSON(res, &delegatorDelegations)
	require.NoError(t, errRes)
	require.Len(t, delegatorDelegations, 1)
	require.Equal(t, delegation.ValidatorAddress, delegatorDelegations[0].ValidatorAddress)
	require.Equal(t, delegation.DelegatorAddress, delegatorDelegations[0].DelegatorAddress)
	require.Equal(t, sdk.NewCoin(sdk.DefaultBondDenom, delegation.Shares.TruncateInt()), delegatorDelegations[0].Balance)

	// error unknown request
	query.Data = bz[:len(bz)-1]

	_, err = querier(ctx, []string{types.QueryDelegation}, query)
	require.Error(t, err)

	// Query validator delegations
	bz, errRes = cdc.MarshalJSON(types.NewQueryValidatorParams(addrVal1, 1, 100))
	require.NoError(t, errRes)

	query = abci.RequestQuery{
		Path: "custom/staking/validatorDelegations",
		Data: bz,
	}

	res, err = querier(ctx, []string{types.QueryValidatorDelegations}, query)
	require.NoError(t, err)

	var delegationsRes types.DelegationResponses
	errRes = cdc.UnmarshalJSON(res, &delegationsRes)
	require.NoError(t, errRes)
	require.Len(t, delegatorDelegations, 1)
	require.Equal(t, delegation.ValidatorAddress, delegationsRes[0].ValidatorAddress)
	require.Equal(t, delegation.DelegatorAddress, delegationsRes[0].DelegatorAddress)
	require.Equal(t, sdk.NewCoin(sdk.DefaultBondDenom, delegation.Shares.TruncateInt()), delegationsRes[0].Balance)

	// Query unbonding delegation
	unbondingTokens := sdk.TokensFromConsensusPower(10)
	_, err = app.StakingKeeper.Undelegate(ctx, addrAcc2, val1.OperatorAddress, unbondingTokens.ToDec())
	require.NoError(t, err)

	queryBondParams = types.NewQueryBondsParams(addrAcc2, addrVal1)
	bz, errRes = cdc.MarshalJSON(queryBondParams)
	require.NoError(t, errRes)

	query = abci.RequestQuery{
		Path: "/custom/staking/unbondingDelegation",
		Data: bz,
	}

	unbond, found := app.StakingKeeper.GetUnbondingDelegation(ctx, addrAcc2, addrVal1)
	require.True(t, found)

	res, err = querier(ctx, []string{types.QueryUnbondingDelegation}, query)
	require.NoError(t, err)

	var unbondRes types.UnbondingDelegation
	errRes = cdc.UnmarshalJSON(res, &unbondRes)
	require.NoError(t, errRes)

	require.Equal(t, unbond, unbondRes)

	// error unknown request
	query.Data = bz[:len(bz)-1]

	_, err = querier(ctx, []string{types.QueryUnbondingDelegation}, query)
	require.Error(t, err)

	// Query Delegator Delegations

	query = abci.RequestQuery{
		Path: "/custom/staking/delegatorUnbondingDelegations",
		Data: bz,
	}

	res, err = querier(ctx, []string{types.QueryDelegatorUnbondingDelegations}, query)
	require.NoError(t, err)

	var delegatorUbds []types.UnbondingDelegation
	errRes = cdc.UnmarshalJSON(res, &delegatorUbds)
	require.NoError(t, errRes)
	require.Equal(t, unbond, delegatorUbds[0])

	// error unknown request
	query.Data = bz[:len(bz)-1]

	_, err = querier(ctx, []string{types.QueryDelegatorUnbondingDelegations}, query)
	require.Error(t, err)

	// Query redelegation
	redelegationTokens := sdk.TokensFromConsensusPower(10)
	_, err = app.StakingKeeper.BeginRedelegation(ctx, addrAcc2, val1.OperatorAddress,
		val2.OperatorAddress, redelegationTokens.ToDec())
	require.NoError(t, err)
	redel, found := app.StakingKeeper.GetRedelegation(ctx, addrAcc2, val1.OperatorAddress, val2.OperatorAddress)
	require.True(t, found)

	bz, errRes = cdc.MarshalJSON(types.NewQueryRedelegationParams(addrAcc2, val1.OperatorAddress, val2.OperatorAddress))
	require.NoError(t, errRes)

	query = abci.RequestQuery{
		Path: "/custom/staking/redelegations",
		Data: bz,
	}

	res, err = querier(ctx, []string{types.QueryRedelegations}, query)
	require.NoError(t, err)

	var redelRes types.RedelegationResponses
	errRes = cdc.UnmarshalJSON(res, &redelRes)
	require.NoError(t, errRes)
	require.Len(t, redelRes, 1)
	require.Equal(t, redel.DelegatorAddress, redelRes[0].DelegatorAddress)
	require.Equal(t, redel.ValidatorSrcAddress, redelRes[0].ValidatorSrcAddress)
	require.Equal(t, redel.ValidatorDstAddress, redelRes[0].ValidatorDstAddress)
	require.Len(t, redel.Entries, len(redelRes[0].Entries))
}

func TestQueryValidatorDelegations_Pagination(t *testing.T) {
	cases := []struct {
		page            int
		limit           int
		expectedResults int
	}{
		{
			page:            1,
			limit:           75,
			expectedResults: 75,
		},
		{
			page:            2,
			limit:           75,
			expectedResults: 25,
		},
		{
			page:            1,
			limit:           100,
			expectedResults: 100,
		},
	}

	cdc, app, ctx := createTestInput()
	querier := keeper.NewQuerier(app.StakingKeeper)

	addrs := simapp.AddTestAddrs(app, ctx, 100, sdk.TokensFromConsensusPower(10000))
	pubKeys := simapp.CreateTestPubKeys(1)

	valAddress := sdk.ValAddress(addrs[0])

	val1 := types.NewValidator(valAddress, pubKeys[0], types.Description{})
	app.StakingKeeper.SetValidator(ctx, val1)
	app.StakingKeeper.SetValidatorByPowerIndex(ctx, val1)

	// Create Validators and Delegation
	for _, addr := range addrs {
		validator, found := app.StakingKeeper.GetValidator(ctx, valAddress)
		if !found {
			t.Error("expected validator not found")
		}

		delTokens := sdk.TokensFromConsensusPower(20)
		_, err := app.StakingKeeper.Delegate(ctx, addr, delTokens, sdk.Unbonded, validator, true)
		require.NoError(t, err)
	}

	// apply TM updates
	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)

	for _, c := range cases {
		// Query Delegator bonded validators
		queryParams := types.NewQueryDelegatorParams(addrs[0])
		bz, errRes := cdc.MarshalJSON(queryParams)
		require.NoError(t, errRes)

		// Query valAddress delegations
		bz, errRes = cdc.MarshalJSON(types.NewQueryValidatorParams(valAddress, c.page, c.limit))
		require.NoError(t, errRes)

		query := abci.RequestQuery{
			Path: "custom/staking/validatorDelegations",
			Data: bz,
		}

		res, err := querier(ctx, []string{types.QueryValidatorDelegations}, query)
		require.NoError(t, err)

		var delegationsRes types.DelegationResponses
		errRes = cdc.UnmarshalJSON(res, &delegationsRes)
		require.NoError(t, errRes)
		require.Len(t, delegationsRes, c.expectedResults)
	}

	// Undelegate
	for _, addr := range addrs {
		delTokens := sdk.TokensFromConsensusPower(20)
		_, err := app.StakingKeeper.Undelegate(ctx, addr, val1.GetOperator(), delTokens.ToDec())
		require.NoError(t, err)
	}

	// apply TM updates
	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)

	for _, c := range cases {
		// Query Unbonding delegations with pagination.
		queryParams := types.NewQueryDelegatorParams(addrs[0])
		bz, errRes := cdc.MarshalJSON(queryParams)
		require.NoError(t, errRes)

		bz, errRes = cdc.MarshalJSON(types.NewQueryValidatorParams(valAddress, c.page, c.limit))
		require.NoError(t, errRes)
		query := abci.RequestQuery{
			Data: bz,
		}

		unbondingDelegations := types.UnbondingDelegations{}
		res, err := querier(ctx, []string{types.QueryValidatorUnbondingDelegations}, query)
		require.NoError(t, err)

		errRes = cdc.UnmarshalJSON(res, &unbondingDelegations)
		require.NoError(t, errRes)
		require.Len(t, unbondingDelegations, c.expectedResults)
	}
}

func TestQueryRedelegations(t *testing.T) {
	cdc, app, ctx := createTestInput()
	querier := keeper.NewQuerier(app.StakingKeeper)

	addrs := simapp.AddTestAddrs(app, ctx, 2, sdk.TokensFromConsensusPower(10000))
	addrAcc1, addrAcc2 := addrs[0], addrs[1]
	addrVal1, addrVal2 := sdk.ValAddress(addrAcc1), sdk.ValAddress(addrAcc2)

	// Create Validators and Delegation
	val1 := types.NewValidator(addrVal1, PKs[0], types.Description{})
	val2 := types.NewValidator(addrVal2, PKs[1], types.Description{})
	app.StakingKeeper.SetValidator(ctx, val1)
	app.StakingKeeper.SetValidator(ctx, val2)

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

	// delegator redelegations
	queryDelegatorParams := types.NewQueryDelegatorParams(addrAcc2)
	bz, errRes := cdc.MarshalJSON(queryDelegatorParams)
	require.NoError(t, errRes)

	query := abci.RequestQuery{
		Path: "/custom/staking/redelegations",
		Data: bz,
	}

	res, err := querier(ctx, []string{types.QueryRedelegations}, query)
	require.NoError(t, err)

	var redelRes types.RedelegationResponses
	errRes = cdc.UnmarshalJSON(res, &redelRes)
	require.NoError(t, errRes)
	require.Len(t, redelRes, 1)
	require.Equal(t, redel.DelegatorAddress, redelRes[0].DelegatorAddress)
	require.Equal(t, redel.ValidatorSrcAddress, redelRes[0].ValidatorSrcAddress)
	require.Equal(t, redel.ValidatorDstAddress, redelRes[0].ValidatorDstAddress)
	require.Len(t, redel.Entries, len(redelRes[0].Entries))

	// validator redelegations
	queryValidatorParams := types.NewQueryValidatorParams(val1.GetOperator(), 0, 0)
	bz, errRes = cdc.MarshalJSON(queryValidatorParams)
	require.NoError(t, errRes)

	query = abci.RequestQuery{
		Path: "/custom/staking/redelegations",
		Data: bz,
	}

	res, err = querier(ctx, []string{types.QueryRedelegations}, query)
	require.NoError(t, err)

	errRes = cdc.UnmarshalJSON(res, &redelRes)
	require.NoError(t, errRes)
	require.Len(t, redelRes, 1)
	require.Equal(t, redel.DelegatorAddress, redelRes[0].DelegatorAddress)
	require.Equal(t, redel.ValidatorSrcAddress, redelRes[0].ValidatorSrcAddress)
	require.Equal(t, redel.ValidatorDstAddress, redelRes[0].ValidatorDstAddress)
	require.Len(t, redel.Entries, len(redelRes[0].Entries))
}

func TestQueryUnbondingDelegation(t *testing.T) {
	cdc, app, ctx := createTestInput()
	querier := keeper.NewQuerier(app.StakingKeeper)

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

	_, found := app.StakingKeeper.GetUnbondingDelegation(ctx, addrAcc1, val1.OperatorAddress)
	require.True(t, found)

	//
	// found: query unbonding delegation by delegator and validator
	//
	queryValidatorParams := types.NewQueryBondsParams(addrAcc1, val1.GetOperator())
	bz, errRes := cdc.MarshalJSON(queryValidatorParams)
	require.NoError(t, errRes)
	query := abci.RequestQuery{
		Path: "/custom/staking/unbondingDelegation",
		Data: bz,
	}
	res, err := querier(ctx, []string{types.QueryUnbondingDelegation}, query)
	require.NoError(t, err)
	require.NotNil(t, res)
	var ubDel types.UnbondingDelegation
	require.NoError(t, cdc.UnmarshalJSON(res, &ubDel))
	require.Equal(t, addrAcc1, ubDel.DelegatorAddress)
	require.Equal(t, val1.OperatorAddress, ubDel.ValidatorAddress)
	require.Equal(t, 1, len(ubDel.Entries))

	//
	// not found: query unbonding delegation by delegator and validator
	//
	queryValidatorParams = types.NewQueryBondsParams(addrAcc2, val1.GetOperator())
	bz, errRes = cdc.MarshalJSON(queryValidatorParams)
	require.NoError(t, errRes)
	query = abci.RequestQuery{
		Path: "/custom/staking/unbondingDelegation",
		Data: bz,
	}
	_, err = querier(ctx, []string{types.QueryUnbondingDelegation}, query)
	require.Error(t, err)

	//
	// found: query unbonding delegation by delegator and validator
	//
	queryDelegatorParams := types.NewQueryDelegatorParams(addrAcc1)
	bz, errRes = cdc.MarshalJSON(queryDelegatorParams)
	require.NoError(t, errRes)
	query = abci.RequestQuery{
		Path: "/custom/staking/delegatorUnbondingDelegations",
		Data: bz,
	}
	res, err = querier(ctx, []string{types.QueryDelegatorUnbondingDelegations}, query)
	require.NoError(t, err)
	require.NotNil(t, res)
	var ubDels []types.UnbondingDelegation
	require.NoError(t, cdc.UnmarshalJSON(res, &ubDels))
	require.Equal(t, 1, len(ubDels))
	require.Equal(t, addrAcc1, ubDels[0].DelegatorAddress)
	require.Equal(t, val1.OperatorAddress, ubDels[0].ValidatorAddress)

	//
	// not found: query unbonding delegation by delegator and validator
	//
	queryDelegatorParams = types.NewQueryDelegatorParams(addrAcc2)
	bz, errRes = cdc.MarshalJSON(queryDelegatorParams)
	require.NoError(t, errRes)
	query = abci.RequestQuery{
		Path: "/custom/staking/delegatorUnbondingDelegations",
		Data: bz,
	}
	res, err = querier(ctx, []string{types.QueryDelegatorUnbondingDelegations}, query)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NoError(t, cdc.UnmarshalJSON(res, &ubDels))
	require.Equal(t, 0, len(ubDels))
}

func TestQueryHistoricalInfo(t *testing.T) {
	cdc, app, ctx := createTestInput()
	querier := keeper.NewQuerier(app.StakingKeeper)

	addrs := simapp.AddTestAddrs(app, ctx, 2, sdk.TokensFromConsensusPower(10000))
	addrAcc1, addrAcc2 := addrs[0], addrs[1]
	addrVal1, addrVal2 := sdk.ValAddress(addrAcc1), sdk.ValAddress(addrAcc2)

	// Create Validators and Delegation
	val1 := types.NewValidator(addrVal1, PKs[0], types.Description{})
	val2 := types.NewValidator(addrVal2, PKs[1], types.Description{})
	vals := []types.Validator{val1, val2}
	app.StakingKeeper.SetValidator(ctx, val1)
	app.StakingKeeper.SetValidator(ctx, val2)

	header := abci.Header{
		ChainID: "HelloChain",
		Height:  5,
	}
	hi := types.NewHistoricalInfo(header, vals)
	app.StakingKeeper.SetHistoricalInfo(ctx, 5, hi)

	queryHistoricalParams := types.NewQueryHistoricalInfoParams(4)
	bz, errRes := cdc.MarshalJSON(queryHistoricalParams)
	require.NoError(t, errRes)
	query := abci.RequestQuery{
		Path: "/custom/staking/historicalInfo",
		Data: bz,
	}
	res, err := querier(ctx, []string{types.QueryHistoricalInfo}, query)
	require.Error(t, err, "Invalid query passed")
	require.Nil(t, res, "Invalid query returned non-nil result")

	queryHistoricalParams = types.NewQueryHistoricalInfoParams(5)
	bz, errRes = cdc.MarshalJSON(queryHistoricalParams)
	require.NoError(t, errRes)
	query.Data = bz
	res, err = querier(ctx, []string{types.QueryHistoricalInfo}, query)
	require.NoError(t, err, "Valid query passed")
	require.NotNil(t, res, "Valid query returned nil result")

	var recv types.HistoricalInfo
	require.NoError(t, cdc.UnmarshalJSON(res, &recv))
	require.Equal(t, hi, recv, "HistoricalInfo query returned wrong result")
}
