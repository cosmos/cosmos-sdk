package querier

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keep "github.com/cosmos/cosmos-sdk/x/stake/keeper"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	addrAcc1, addrAcc2 = keep.Addrs[0], keep.Addrs[1]
	addrVal1, addrVal2 = sdk.ValAddress(keep.Addrs[0]), sdk.ValAddress(keep.Addrs[1])
	pk1, pk2           = keep.PKs[0], keep.PKs[1]
)

func newTestDelegatorQuery(delegatorAddr sdk.AccAddress) QueryDelegatorParams {
	return QueryDelegatorParams{
		DelegatorAddr: delegatorAddr,
	}
}

func newTestValidatorQuery(validatorAddr sdk.ValAddress) QueryValidatorParams {
	return QueryValidatorParams{
		ValidatorAddr: validatorAddr,
	}
}

func newTestBondQuery(delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) QueryBondsParams {
	return QueryBondsParams{
		DelegatorAddr: delegatorAddr,
		ValidatorAddr: validatorAddr,
	}
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
	validators[0] = keeper.UpdateValidator(ctx, validators[0])
	validators[1] = keeper.UpdateValidator(ctx, validators[1])

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
	queryParams := newTestValidatorQuery(addrVal1)
	bz, errRes := cdc.MarshalJSON(queryParams)
	require.Nil(t, errRes)

	query := abci.RequestQuery{
		Path: "/custom/stake/validator",
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

	keeper.Delegate(ctx, addrAcc2, sdk.NewCoin("steak", sdk.NewInt(20)), val1, true)

	// Query Delegator bonded validators
	queryParams := newTestDelegatorQuery(addrAcc2)
	bz, errRes := cdc.MarshalJSON(queryParams)
	require.Nil(t, errRes)

	query := abci.RequestQuery{
		Path: "/custom/stake/delegatorValidators",
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

	// Query bonded validator
	queryBondParams := newTestBondQuery(addrAcc2, addrVal1)
	bz, errRes = cdc.MarshalJSON(queryBondParams)
	require.Nil(t, errRes)

	query = abci.RequestQuery{
		Path: "/custom/stake/delegatorValidator",
		Data: bz,
	}

	res, err = queryDelegatorValidator(ctx, cdc, query, keeper)
	require.Nil(t, err)

	var validator types.Validator
	errRes = cdc.UnmarshalJSON(res, &validator)
	require.Nil(t, errRes)

	require.Equal(t, delValidators[0], validator)

	// Query delegation

	query = abci.RequestQuery{
		Path: "/custom/stake/delegation",
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

	// Query unbonging delegation
	keeper.BeginUnbonding(ctx, addrAcc2, val1.OperatorAddr, sdk.NewDec(10))

	query = abci.RequestQuery{
		Path: "/custom/stake/unbondingDelegation",
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

	// Query Delegator Summary

	query = abci.RequestQuery{
		Path: "/custom/stake/delegator",
		Data: bz,
	}

	res, err = queryDelegator(ctx, cdc, query, keeper)
	require.Nil(t, err)

	var summary types.DelegationSummary
	errRes = cdc.UnmarshalJSON(res, &summary)
	require.Nil(t, errRes)

	require.Equal(t, unbond, summary.UnbondingDelegations[0])
}
