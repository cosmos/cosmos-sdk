package keeper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	addrAcc1, addrAcc2 = Addrs[0], Addrs[1]
	addrVal1, addrVal2 = sdk.ValAddress(Addrs[0]), sdk.ValAddress(Addrs[1])
	pk1, pk2           = PKs[0], PKs[1]
)

func TestNewQuerier(t *testing.T) {
	cdc := codec.New()
	ctx, _, keeper, _ := CreateTestInput(t, false, 1000)
	// Create Validators
	amts := []sdk.Int{sdk.NewInt(9), sdk.NewInt(8)}
	var validators [2]types.Validator
	for i, amt := range amts {
		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{})
		validators[i], _ = validators[i].AddTokensFromDel(amt)
		keeper.SetValidator(ctx, validators[i])
		keeper.SetValidatorByPowerIndex(ctx, validators[i])
	}

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	querier := NewQuerier(keeper)

	bz, err := querier(ctx, []string{"other"}, query)
	require.NotNil(t, err)
	require.Nil(t, bz)

	_, err = querier(ctx, []string{"pool"}, query)
	require.Nil(t, err)

	_, err = querier(ctx, []string{"parameters"}, query)
	require.Nil(t, err)

	queryValParams := types.NewQueryValidatorParams(addrVal1)
	bz, errRes := cdc.MarshalJSON(queryValParams)
	require.Nil(t, errRes)

	query.Path = "/custom/staking/validator"
	query.Data = bz

	_, err = querier(ctx, []string{"validator"}, query)
	require.Nil(t, err)

	_, err = querier(ctx, []string{"validatorDelegations"}, query)
	require.Nil(t, err)

	_, err = querier(ctx, []string{"validatorUnbondingDelegations"}, query)
	require.Nil(t, err)

	queryDelParams := types.NewQueryDelegatorParams(addrAcc2)
	bz, errRes = cdc.MarshalJSON(queryDelParams)
	require.Nil(t, errRes)

	query.Path = "/custom/staking/validator"
	query.Data = bz

	_, err = querier(ctx, []string{"delegatorDelegations"}, query)
	require.Nil(t, err)

	_, err = querier(ctx, []string{"delegatorUnbondingDelegations"}, query)
	require.Nil(t, err)

	_, err = querier(ctx, []string{"delegatorValidators"}, query)
	require.Nil(t, err)

	bz, errRes = cdc.MarshalJSON(types.NewQueryRedelegationParams(nil, nil, nil))
	require.Nil(t, errRes)
	query.Data = bz

	_, err = querier(ctx, []string{"redelegations"}, query)
	require.Nil(t, err)
}

func TestQueryParametersPool(t *testing.T) {
	cdc := codec.New()
	ctx, _, keeper, _ := CreateTestInput(t, false, 1000)
	bondDenom := keeper.BondDenom(ctx)

	res, err := queryParameters(ctx, keeper)
	require.Nil(t, err)

	var params types.Params
	errRes := cdc.UnmarshalJSON(res, &params)
	require.Nil(t, errRes)
	require.Equal(t, keeper.GetParams(ctx), params)

	res, err = queryPool(ctx, keeper)
	require.Nil(t, err)

	var pool types.Pool
	bondedPool := keeper.GetBondedPool(ctx)
	notBondedPool := keeper.GetNotBondedPool(ctx)
	errRes = cdc.UnmarshalJSON(res, &pool)
	require.Nil(t, errRes)
	require.Equal(t, bondedPool.GetCoins().AmountOf(bondDenom), pool.BondedTokens)
	require.Equal(t, notBondedPool.GetCoins().AmountOf(bondDenom), pool.NotBondedTokens)
}

func TestQueryValidators(t *testing.T) {
	cdc := codec.New()
	ctx, _, keeper, _ := CreateTestInput(t, false, 10000)
	params := keeper.GetParams(ctx)

	// Create Validators
	amts := []sdk.Int{sdk.NewInt(9), sdk.NewInt(8), sdk.NewInt(7)}
	status := []sdk.BondStatus{sdk.Bonded, sdk.Unbonded, sdk.Unbonding}
	var validators [3]types.Validator
	for i, amt := range amts {
		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{})
		validators[i], _ = validators[i].AddTokensFromDel(amt)
		validators[i] = validators[i].UpdateStatus(status[i])
	}

	keeper.SetValidator(ctx, validators[0])
	keeper.SetValidator(ctx, validators[1])
	keeper.SetValidator(ctx, validators[2])

	// Query Validators
	queriedValidators := keeper.GetValidators(ctx, params.MaxValidators)

	for i, s := range status {
		queryValsParams := types.NewQueryValidatorsParams(1, int(params.MaxValidators), s.String())
		bz, err := cdc.MarshalJSON(queryValsParams)
		require.Nil(t, err)

		req := abci.RequestQuery{
			Path: fmt.Sprintf("/custom/%s/%s", types.QuerierRoute, types.QueryValidators),
			Data: bz,
		}

		res, err := queryValidators(ctx, req, keeper)
		require.Nil(t, err)

		var validatorsResp []types.Validator
		err = cdc.UnmarshalJSON(res, &validatorsResp)
		require.Nil(t, err)

		require.Equal(t, 1, len(validatorsResp))
		require.ElementsMatch(t, validators[i].OperatorAddress, validatorsResp[0].OperatorAddress)

	}

	// Query each validator
	queryParams := types.NewQueryValidatorParams(addrVal1)
	bz, err := cdc.MarshalJSON(queryParams)
	require.Nil(t, err)

	query := abci.RequestQuery{
		Path: "/custom/staking/validator",
		Data: bz,
	}
	res, err := queryValidator(ctx, query, keeper)
	require.Nil(t, err)

	var validator types.Validator
	err = cdc.UnmarshalJSON(res, &validator)
	require.Nil(t, err)

	require.Equal(t, queriedValidators[0], validator)
}

func TestQueryDelegation(t *testing.T) {
	cdc := codec.New()
	ctx, _, keeper, _ := CreateTestInput(t, false, 10000)
	params := keeper.GetParams(ctx)

	// Create Validators and Delegation
	val1 := types.NewValidator(addrVal1, pk1, types.Description{})
	keeper.SetValidator(ctx, val1)
	keeper.SetValidatorByPowerIndex(ctx, val1)

	val2 := types.NewValidator(addrVal2, pk2, types.Description{})
	keeper.SetValidator(ctx, val2)
	keeper.SetValidatorByPowerIndex(ctx, val2)

	delTokens := sdk.TokensFromConsensusPower(20)
	keeper.Delegate(ctx, addrAcc2, delTokens, sdk.Unbonded, val1, true)

	// apply TM updates
	keeper.ApplyAndReturnValidatorSetUpdates(ctx)

	// Query Delegator bonded validators
	queryParams := types.NewQueryDelegatorParams(addrAcc2)
	bz, errRes := cdc.MarshalJSON(queryParams)
	require.Nil(t, errRes)

	query := abci.RequestQuery{
		Path: "/custom/staking/delegatorValidators",
		Data: bz,
	}

	delValidators := keeper.GetDelegatorValidators(ctx, addrAcc2, params.MaxValidators)

	res, err := queryDelegatorValidators(ctx, query, keeper)
	require.Nil(t, err)

	var validatorsResp []types.Validator
	errRes = cdc.UnmarshalJSON(res, &validatorsResp)
	require.Nil(t, errRes)

	require.Equal(t, len(delValidators), len(validatorsResp))
	require.ElementsMatch(t, delValidators, validatorsResp)

	// error unknown request
	query.Data = bz[:len(bz)-1]

	_, err = queryDelegatorValidators(ctx, query, keeper)
	require.NotNil(t, err)

	// Query bonded validator
	queryBondParams := types.NewQueryBondsParams(addrAcc2, addrVal1)
	bz, errRes = cdc.MarshalJSON(queryBondParams)
	require.Nil(t, errRes)

	query = abci.RequestQuery{
		Path: "/custom/staking/delegatorValidator",
		Data: bz,
	}

	res, err = queryDelegatorValidator(ctx, query, keeper)
	require.Nil(t, err)

	var validator types.Validator
	errRes = cdc.UnmarshalJSON(res, &validator)
	require.Nil(t, errRes)

	require.Equal(t, delValidators[0], validator)

	// error unknown request
	query.Data = bz[:len(bz)-1]

	_, err = queryDelegatorValidator(ctx, query, keeper)
	require.NotNil(t, err)

	// Query delegation

	query = abci.RequestQuery{
		Path: "/custom/staking/delegation",
		Data: bz,
	}

	delegation, found := keeper.GetDelegation(ctx, addrAcc2, addrVal1)
	require.True(t, found)

	res, err = queryDelegation(ctx, query, keeper)
	require.Nil(t, err)

	var delegationRes types.DelegationResponse
	errRes = cdc.UnmarshalJSON(res, &delegationRes)
	require.Nil(t, errRes)

	require.Equal(t, delegation.ValidatorAddress, delegationRes.ValidatorAddress)
	require.Equal(t, delegation.DelegatorAddress, delegationRes.DelegatorAddress)
	require.Equal(t, delegation.Shares.TruncateInt(), delegationRes.Balance)

	// Query Delegator Delegations
	query = abci.RequestQuery{
		Path: "/custom/staking/delegatorDelegations",
		Data: bz,
	}

	res, err = queryDelegatorDelegations(ctx, query, keeper)
	require.Nil(t, err)

	var delegatorDelegations types.DelegationResponses
	errRes = cdc.UnmarshalJSON(res, &delegatorDelegations)
	require.Nil(t, errRes)
	require.Len(t, delegatorDelegations, 1)
	require.Equal(t, delegation.ValidatorAddress, delegatorDelegations[0].ValidatorAddress)
	require.Equal(t, delegation.DelegatorAddress, delegatorDelegations[0].DelegatorAddress)
	require.Equal(t, delegation.Shares.TruncateInt(), delegatorDelegations[0].Balance)

	// error unknown request
	query.Data = bz[:len(bz)-1]

	_, err = queryDelegation(ctx, query, keeper)
	require.NotNil(t, err)

	// Query validator delegations

	bz, errRes = cdc.MarshalJSON(types.NewQueryValidatorParams(addrVal1))
	require.Nil(t, errRes)

	query = abci.RequestQuery{
		Path: "custom/staking/validatorDelegations",
		Data: bz,
	}

	res, err = queryValidatorDelegations(ctx, query, keeper)
	require.Nil(t, err)

	var delegationsRes types.DelegationResponses
	errRes = cdc.UnmarshalJSON(res, &delegationsRes)
	require.Nil(t, errRes)
	require.Len(t, delegatorDelegations, 1)
	require.Equal(t, delegation.ValidatorAddress, delegationsRes[0].ValidatorAddress)
	require.Equal(t, delegation.DelegatorAddress, delegationsRes[0].DelegatorAddress)
	require.Equal(t, delegation.Shares.TruncateInt(), delegationsRes[0].Balance)

	// Query unbonging delegation
	unbondingTokens := sdk.TokensFromConsensusPower(10)
	_, err = keeper.Undelegate(ctx, addrAcc2, val1.OperatorAddress, unbondingTokens.ToDec())
	require.Nil(t, err)

	queryBondParams = types.NewQueryBondsParams(addrAcc2, addrVal1)
	bz, errRes = cdc.MarshalJSON(queryBondParams)
	require.Nil(t, errRes)

	query = abci.RequestQuery{
		Path: "/custom/staking/unbondingDelegation",
		Data: bz,
	}

	unbond, found := keeper.GetUnbondingDelegation(ctx, addrAcc2, addrVal1)
	require.True(t, found)

	res, err = queryUnbondingDelegation(ctx, query, keeper)
	require.Nil(t, err)

	var unbondRes types.UnbondingDelegation
	errRes = cdc.UnmarshalJSON(res, &unbondRes)
	require.Nil(t, errRes)

	require.Equal(t, unbond, unbondRes)

	// error unknown request
	query.Data = bz[:len(bz)-1]

	_, err = queryUnbondingDelegation(ctx, query, keeper)
	require.NotNil(t, err)

	// Query Delegator Delegations

	query = abci.RequestQuery{
		Path: "/custom/staking/delegatorUnbondingDelegations",
		Data: bz,
	}

	res, err = queryDelegatorUnbondingDelegations(ctx, query, keeper)
	require.Nil(t, err)

	var delegatorUbds []types.UnbondingDelegation
	errRes = cdc.UnmarshalJSON(res, &delegatorUbds)
	require.Nil(t, errRes)
	require.Equal(t, unbond, delegatorUbds[0])

	// error unknown request
	query.Data = bz[:len(bz)-1]

	_, err = queryDelegatorUnbondingDelegations(ctx, query, keeper)
	require.NotNil(t, err)

	// Query redelegation
	redelegationTokens := sdk.TokensFromConsensusPower(10)
	_, err = keeper.BeginRedelegation(ctx, addrAcc2, val1.OperatorAddress,
		val2.OperatorAddress, redelegationTokens.ToDec())
	require.Nil(t, err)
	redel, found := keeper.GetRedelegation(ctx, addrAcc2, val1.OperatorAddress, val2.OperatorAddress)
	require.True(t, found)

	bz, errRes = cdc.MarshalJSON(types.NewQueryRedelegationParams(addrAcc2, val1.OperatorAddress, val2.OperatorAddress))
	require.Nil(t, errRes)

	query = abci.RequestQuery{
		Path: "/custom/staking/redelegations",
		Data: bz,
	}

	res, err = queryRedelegations(ctx, query, keeper)
	require.Nil(t, err)

	var redelRes types.RedelegationResponses
	errRes = cdc.UnmarshalJSON(res, &redelRes)
	require.Nil(t, errRes)
	require.Len(t, redelRes, 1)
	require.Equal(t, redel.DelegatorAddress, redelRes[0].DelegatorAddress)
	require.Equal(t, redel.ValidatorSrcAddress, redelRes[0].ValidatorSrcAddress)
	require.Equal(t, redel.ValidatorDstAddress, redelRes[0].ValidatorDstAddress)
	require.Len(t, redel.Entries, len(redelRes[0].Entries))
}

func TestQueryRedelegations(t *testing.T) {
	cdc := codec.New()
	ctx, _, keeper, _ := CreateTestInput(t, false, 10000)

	// Create Validators and Delegation
	val1 := types.NewValidator(addrVal1, pk1, types.Description{})
	val2 := types.NewValidator(addrVal2, pk2, types.Description{})
	keeper.SetValidator(ctx, val1)
	keeper.SetValidator(ctx, val2)

	delAmount := sdk.TokensFromConsensusPower(100)
	keeper.Delegate(ctx, addrAcc2, delAmount, sdk.Unbonded, val1, true)
	_ = keeper.ApplyAndReturnValidatorSetUpdates(ctx)

	rdAmount := sdk.TokensFromConsensusPower(20)
	keeper.BeginRedelegation(ctx, addrAcc2, val1.GetOperator(), val2.GetOperator(), rdAmount.ToDec())
	keeper.ApplyAndReturnValidatorSetUpdates(ctx)

	redel, found := keeper.GetRedelegation(ctx, addrAcc2, val1.OperatorAddress, val2.OperatorAddress)
	require.True(t, found)

	// delegator redelegations
	queryDelegatorParams := types.NewQueryDelegatorParams(addrAcc2)
	bz, errRes := cdc.MarshalJSON(queryDelegatorParams)
	require.Nil(t, errRes)

	query := abci.RequestQuery{
		Path: "/custom/staking/redelegations",
		Data: bz,
	}

	res, err := queryRedelegations(ctx, query, keeper)
	require.Nil(t, err)

	var redelRes types.RedelegationResponses
	errRes = cdc.UnmarshalJSON(res, &redelRes)
	require.Nil(t, errRes)
	require.Len(t, redelRes, 1)
	require.Equal(t, redel.DelegatorAddress, redelRes[0].DelegatorAddress)
	require.Equal(t, redel.ValidatorSrcAddress, redelRes[0].ValidatorSrcAddress)
	require.Equal(t, redel.ValidatorDstAddress, redelRes[0].ValidatorDstAddress)
	require.Len(t, redel.Entries, len(redelRes[0].Entries))

	// validator redelegations
	queryValidatorParams := types.NewQueryValidatorParams(val1.GetOperator())
	bz, errRes = cdc.MarshalJSON(queryValidatorParams)
	require.Nil(t, errRes)

	query = abci.RequestQuery{
		Path: "/custom/staking/redelegations",
		Data: bz,
	}

	res, err = queryRedelegations(ctx, query, keeper)
	require.Nil(t, err)

	errRes = cdc.UnmarshalJSON(res, &redelRes)
	require.Nil(t, errRes)
	require.Len(t, redelRes, 1)
	require.Equal(t, redel.DelegatorAddress, redelRes[0].DelegatorAddress)
	require.Equal(t, redel.ValidatorSrcAddress, redelRes[0].ValidatorSrcAddress)
	require.Equal(t, redel.ValidatorDstAddress, redelRes[0].ValidatorDstAddress)
	require.Len(t, redel.Entries, len(redelRes[0].Entries))
}

func TestQueryUnbondingDelegation(t *testing.T) {
	cdc := codec.New()
	ctx, _, keeper, _ := CreateTestInput(t, false, 10000)

	// Create Validators and Delegation
	val1 := types.NewValidator(addrVal1, pk1, types.Description{})
	keeper.SetValidator(ctx, val1)

	// delegate
	delAmount := sdk.TokensFromConsensusPower(100)
	_, err := keeper.Delegate(ctx, addrAcc1, delAmount, sdk.Unbonded, val1, true)
	require.NoError(t, err)
	_ = keeper.ApplyAndReturnValidatorSetUpdates(ctx)

	// undelegate
	undelAmount := sdk.TokensFromConsensusPower(20)
	_, err = keeper.Undelegate(ctx, addrAcc1, val1.GetOperator(), undelAmount.ToDec())
	require.NoError(t, err)
	keeper.ApplyAndReturnValidatorSetUpdates(ctx)

	_, found := keeper.GetUnbondingDelegation(ctx, addrAcc1, val1.OperatorAddress)
	require.True(t, found)

	//
	// found: query unbonding delegation by delegator and validator
	//
	queryValidatorParams := types.NewQueryBondsParams(addrAcc1, val1.GetOperator())
	bz, errRes := cdc.MarshalJSON(queryValidatorParams)
	require.Nil(t, errRes)
	query := abci.RequestQuery{
		Path: "/custom/staking/unbondingDelegation",
		Data: bz,
	}
	res, err := queryUnbondingDelegation(ctx, query, keeper)
	require.Nil(t, err)
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
	require.Nil(t, errRes)
	query = abci.RequestQuery{
		Path: "/custom/staking/unbondingDelegation",
		Data: bz,
	}
	res, err = queryUnbondingDelegation(ctx, query, keeper)
	require.NotNil(t, err)

	//
	// found: query unbonding delegation by delegator and validator
	//
	queryDelegatorParams := types.NewQueryDelegatorParams(addrAcc1)
	bz, errRes = cdc.MarshalJSON(queryDelegatorParams)
	require.Nil(t, errRes)
	query = abci.RequestQuery{
		Path: "/custom/staking/delegatorUnbondingDelegations",
		Data: bz,
	}
	res, err = queryDelegatorUnbondingDelegations(ctx, query, keeper)
	require.Nil(t, err)
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
	require.Nil(t, errRes)
	query = abci.RequestQuery{
		Path: "/custom/staking/delegatorUnbondingDelegations",
		Data: bz,
	}
	res, err = queryDelegatorUnbondingDelegations(ctx, query, keeper)
	require.Nil(t, err)
	require.NotNil(t, res)
	require.NoError(t, cdc.UnmarshalJSON(res, &ubDels))
	require.Equal(t, 0, len(ubDels))
}
