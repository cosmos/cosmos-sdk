package querier

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keep "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	addrAcc1, addrAcc2 = keep.Addrs[0], keep.Addrs[1]
	addrVal1, addrVal2 = sdk.ValAddress(keep.Addrs[0]), sdk.ValAddress(keep.Addrs[1])
	pk1, pk2           = keep.PKs[0], keep.PKs[1]
)

func TestNewQuerier(t *testing.T) {
	cdc := codec.New()
	ctx, _, keeper := keep.CreateTestInput(t, false, 1000)
	pool := keeper.GetPool(ctx)
	// Create Validators
	amts := []sdk.Int{sdk.NewInt(9), sdk.NewInt(8)}
	var validators [2]types.Validator
	for i, amt := range amts {
		validators[i] = types.NewValidator(sdk.ValAddress(keep.Addrs[i]), keep.PKs[i], types.Description{})
		validators[i], pool, _ = validators[i].AddTokensFromDel(pool, amt)
		keeper.SetValidator(ctx, validators[i])
		keeper.SetValidatorByPowerIndex(ctx, validators[i])
	}
	keeper.SetPool(ctx, pool)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	querier := NewQuerier(keeper, cdc)

	bz, err := querier(ctx, []string{"other"}, query)
	require.NotNil(t, err)
	require.Nil(t, bz)

	_, err = querier(ctx, []string{"validators"}, query)
	require.Nil(t, err)

	_, err = querier(ctx, []string{"pool"}, query)
	require.Nil(t, err)

	_, err = querier(ctx, []string{"parameters"}, query)
	require.Nil(t, err)

	queryValParams := NewQueryValidatorParams(addrVal1)
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

	queryDelParams := NewQueryDelegatorParams(addrAcc2)
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

	bz, errRes = cdc.MarshalJSON(NewQueryRedelegationParams(nil, nil, nil))
	require.Nil(t, errRes)
	query.Data = bz

	_, err = querier(ctx, []string{"redelegations"}, query)
	require.Nil(t, err)
}

func TestQueryParametersPool(t *testing.T) {
	cdc := codec.New()
	ctx, _, keeper := keep.CreateTestInput(t, false, 1000)

	res, err := queryParameters(ctx, cdc, keeper)
	require.Nil(t, err)

	var params types.Params
	errRes := cdc.UnmarshalJSON(res, &params)
	require.Nil(t, errRes)
	require.Equal(t, keeper.GetParams(ctx), params)

	res, err = queryPool(ctx, cdc, keeper)
	require.Nil(t, err)

	var pool types.Pool
	errRes = cdc.UnmarshalJSON(res, &pool)
	require.Nil(t, errRes)
	require.Equal(t, keeper.GetPool(ctx), pool)
}

func TestQueryValidators(t *testing.T) {
	cdc := codec.New()
	ctx, _, keeper := keep.CreateTestInput(t, false, 10000)
	pool := keeper.GetPool(ctx)
	params := keeper.GetParams(ctx)

	// Create Validators
	amts := []sdk.Int{sdk.NewInt(9), sdk.NewInt(8)}
	var validators [2]types.Validator
	for i, amt := range amts {
		validators[i] = types.NewValidator(sdk.ValAddress(keep.Addrs[i]), keep.PKs[i], types.Description{})
		validators[i], pool, _ = validators[i].AddTokensFromDel(pool, amt)
	}
	keeper.SetPool(ctx, pool)
	keeper.SetValidator(ctx, validators[0])
	keeper.SetValidator(ctx, validators[1])

	// Query Validators
	queriedValidators := keeper.GetValidators(ctx, params.MaxValidators)

	res, err := queryValidators(ctx, cdc, keeper)
	require.Nil(t, err)

	var validatorsResp []types.Validator
	errRes := cdc.UnmarshalJSON(res, &validatorsResp)
	require.Nil(t, errRes)

	require.Equal(t, len(queriedValidators), len(validatorsResp))
	require.ElementsMatch(t, queriedValidators, validatorsResp)

	// Query each validator
	queryParams := NewQueryValidatorParams(addrVal1)
	bz, errRes := cdc.MarshalJSON(queryParams)
	require.Nil(t, errRes)

	query := abci.RequestQuery{
		Path: "/custom/staking/validator",
		Data: bz,
	}
	res, err = queryValidator(ctx, cdc, query, keeper)
	require.Nil(t, err)

	var validator types.Validator
	errRes = cdc.UnmarshalJSON(res, &validator)
	require.Nil(t, errRes)

	require.Equal(t, queriedValidators[0], validator)
}

func TestQueryDelegation(t *testing.T) {
	cdc := codec.New()
	ctx, _, keeper := keep.CreateTestInput(t, false, 10000)
	params := keeper.GetParams(ctx)

	// Create Validators and Delegation
	val1 := types.NewValidator(addrVal1, pk1, types.Description{})
	keeper.SetValidator(ctx, val1)
	keeper.SetValidatorByPowerIndex(ctx, val1)

	val2 := types.NewValidator(addrVal2, pk2, types.Description{})
	keeper.SetValidator(ctx, val2)
	keeper.SetValidatorByPowerIndex(ctx, val2)

	delTokens := sdk.TokensFromTendermintPower(20)
	keeper.Delegate(ctx, addrAcc2, delTokens, val1, true)

	// apply TM updates
	keeper.ApplyAndReturnValidatorSetUpdates(ctx)

	// Query Delegator bonded validators
	queryParams := NewQueryDelegatorParams(addrAcc2)
	bz, errRes := cdc.MarshalJSON(queryParams)
	require.Nil(t, errRes)

	query := abci.RequestQuery{
		Path: "/custom/staking/delegatorValidators",
		Data: bz,
	}

	delValidators := keeper.GetDelegatorValidators(ctx, addrAcc2, params.MaxValidators)

	res, err := queryDelegatorValidators(ctx, cdc, query, keeper)
	require.Nil(t, err)

	var validatorsResp []types.Validator
	errRes = cdc.UnmarshalJSON(res, &validatorsResp)
	require.Nil(t, errRes)

	require.Equal(t, len(delValidators), len(validatorsResp))
	require.ElementsMatch(t, delValidators, validatorsResp)

	// error unknown request
	query.Data = bz[:len(bz)-1]

	_, err = queryDelegatorValidators(ctx, cdc, query, keeper)
	require.NotNil(t, err)

	// Query bonded validator
	queryBondParams := NewQueryBondsParams(addrAcc2, addrVal1)
	bz, errRes = cdc.MarshalJSON(queryBondParams)
	require.Nil(t, errRes)

	query = abci.RequestQuery{
		Path: "/custom/staking/delegatorValidator",
		Data: bz,
	}

	res, err = queryDelegatorValidator(ctx, cdc, query, keeper)
	require.Nil(t, err)

	var validator types.Validator
	errRes = cdc.UnmarshalJSON(res, &validator)
	require.Nil(t, errRes)

	require.Equal(t, delValidators[0], validator)

	// error unknown request
	query.Data = bz[:len(bz)-1]

	_, err = queryDelegatorValidator(ctx, cdc, query, keeper)
	require.NotNil(t, err)

	// Query delegation

	query = abci.RequestQuery{
		Path: "/custom/staking/delegation",
		Data: bz,
	}

	delegation, found := keeper.GetDelegation(ctx, addrAcc2, addrVal1)
	require.True(t, found)

	res, err = queryDelegation(ctx, cdc, query, keeper)
	require.Nil(t, err)

	var delegationRes types.Delegation
	errRes = cdc.UnmarshalJSON(res, &delegationRes)
	require.Nil(t, errRes)

	require.Equal(t, delegation, delegationRes)

	// Query Delegator Delegations

	query = abci.RequestQuery{
		Path: "/custom/staking/delegatorDelegations",
		Data: bz,
	}

	res, err = queryDelegatorDelegations(ctx, cdc, query, keeper)
	require.Nil(t, err)

	var delegatorDelegations []types.Delegation
	errRes = cdc.UnmarshalJSON(res, &delegatorDelegations)
	require.Nil(t, errRes)
	require.Len(t, delegatorDelegations, 1)
	require.Equal(t, delegation, delegatorDelegations[0])

	// error unknown request
	query.Data = bz[:len(bz)-1]

	_, err = queryDelegation(ctx, cdc, query, keeper)
	require.NotNil(t, err)

	// Query validator delegations

	bz, errRes = cdc.MarshalJSON(NewQueryValidatorParams(addrVal1))
	require.Nil(t, errRes)

	query = abci.RequestQuery{
		Path: "custom/staking/validatorDelegations",
		Data: bz,
	}

	res, err = queryValidatorDelegations(ctx, cdc, query, keeper)
	require.Nil(t, err)

	var delegationsRes []types.Delegation
	errRes = cdc.UnmarshalJSON(res, &delegationsRes)
	require.Nil(t, errRes)

	require.Equal(t, delegationsRes[0], delegation)

	// Query unbonging delegation
	unbondingTokens := sdk.TokensFromTendermintPower(10)
	_, err = keeper.Undelegate(ctx, addrAcc2, val1.OperatorAddr, sdk.NewDecFromInt(unbondingTokens))
	require.Nil(t, err)

	queryBondParams = NewQueryBondsParams(addrAcc2, addrVal1)
	bz, errRes = cdc.MarshalJSON(queryBondParams)
	require.Nil(t, errRes)

	query = abci.RequestQuery{
		Path: "/custom/staking/unbondingDelegation",
		Data: bz,
	}

	unbond, found := keeper.GetUnbondingDelegation(ctx, addrAcc2, addrVal1)
	require.True(t, found)

	res, err = queryUnbondingDelegation(ctx, cdc, query, keeper)
	require.Nil(t, err)

	var unbondRes types.UnbondingDelegation
	errRes = cdc.UnmarshalJSON(res, &unbondRes)
	require.Nil(t, errRes)

	require.Equal(t, unbond, unbondRes)

	// error unknown request
	query.Data = bz[:len(bz)-1]

	_, err = queryUnbondingDelegation(ctx, cdc, query, keeper)
	require.NotNil(t, err)

	// Query Delegator Delegations

	query = abci.RequestQuery{
		Path: "/custom/staking/delegatorUnbondingDelegations",
		Data: bz,
	}

	res, err = queryDelegatorUnbondingDelegations(ctx, cdc, query, keeper)
	require.Nil(t, err)

	var delegatorUbds []types.UnbondingDelegation
	errRes = cdc.UnmarshalJSON(res, &delegatorUbds)
	require.Nil(t, errRes)
	require.Equal(t, unbond, delegatorUbds[0])

	// error unknown request
	query.Data = bz[:len(bz)-1]

	_, err = queryDelegatorUnbondingDelegations(ctx, cdc, query, keeper)
	require.NotNil(t, err)

	// Query redelegation
	redelegationTokens := sdk.TokensFromTendermintPower(10)
	_, err = keeper.BeginRedelegation(ctx, addrAcc2, val1.OperatorAddr,
		val2.OperatorAddr, sdk.NewDecFromInt(redelegationTokens))
	require.Nil(t, err)
	redel, found := keeper.GetRedelegation(ctx, addrAcc2, val1.OperatorAddr, val2.OperatorAddr)
	require.True(t, found)

	bz, errRes = cdc.MarshalJSON(NewQueryRedelegationParams(addrAcc2, val1.OperatorAddr, val2.OperatorAddr))
	require.Nil(t, errRes)

	query = abci.RequestQuery{
		Path: "/custom/staking/redelegations",
		Data: bz,
	}

	res, err = queryRedelegations(ctx, cdc, query, keeper)
	require.Nil(t, err)

	var redelRes []types.Redelegation
	errRes = cdc.UnmarshalJSON(res, &redelRes)
	require.Nil(t, errRes)

	require.Equal(t, redel, redelRes[0])
}

func TestQueryRedelegations(t *testing.T) {
	cdc := codec.New()
	ctx, _, keeper := keep.CreateTestInput(t, false, 10000)

	// Create Validators and Delegation
	val1 := types.NewValidator(addrVal1, pk1, types.Description{})
	val2 := types.NewValidator(addrVal2, pk2, types.Description{})
	keeper.SetValidator(ctx, val1)
	keeper.SetValidator(ctx, val2)

	delAmount := sdk.TokensFromTendermintPower(100)
	keeper.Delegate(ctx, addrAcc2, delAmount, val1, true)
	_ = keeper.ApplyAndReturnValidatorSetUpdates(ctx)

	rdAmount := sdk.TokensFromTendermintPower(20)
	keeper.BeginRedelegation(ctx, addrAcc2, val1.GetOperator(), val2.GetOperator(), sdk.NewDecFromInt(rdAmount))
	keeper.ApplyAndReturnValidatorSetUpdates(ctx)

	redelegation, found := keeper.GetRedelegation(ctx, addrAcc2, val1.OperatorAddr, val2.OperatorAddr)
	require.True(t, found)

	// delegator redelegations
	queryDelegatorParams := NewQueryDelegatorParams(addrAcc2)
	bz, errRes := cdc.MarshalJSON(queryDelegatorParams)
	require.Nil(t, errRes)

	query := abci.RequestQuery{
		Path: "/custom/staking/redelegations",
		Data: bz,
	}

	res, err := queryRedelegations(ctx, cdc, query, keeper)
	require.Nil(t, err)

	var redsRes []types.Redelegation
	errRes = cdc.UnmarshalJSON(res, &redsRes)
	require.Nil(t, errRes)

	require.Equal(t, redelegation, redsRes[0])

	// validator redelegations
	queryValidatorParams := NewQueryValidatorParams(val1.GetOperator())
	bz, errRes = cdc.MarshalJSON(queryValidatorParams)
	require.Nil(t, errRes)

	query = abci.RequestQuery{
		Path: "/custom/staking/redelegations",
		Data: bz,
	}

	res, err = queryRedelegations(ctx, cdc, query, keeper)
	require.Nil(t, err)

	errRes = cdc.UnmarshalJSON(res, &redsRes)
	require.Nil(t, errRes)

	require.Equal(t, redelegation, redsRes[0])
}
