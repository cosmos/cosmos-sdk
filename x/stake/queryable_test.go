package stake

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keep "github.com/cosmos/cosmos-sdk/x/stake/keeper"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
	"github.com/stretchr/testify/assert"
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
	ctx, _, keeper := keep.CreateTestInput(t, false, 1000)

	res, err := queryParameters(ctx, keeper)
	require.Nil(t, err)

	var params types.Params
	errRes := keeper.Codec().UnmarshalJSON(res, &params)
	assert.Nil(t, errRes)
	assert.Equal(t, keeper.GetParams(ctx), params)

	res, err = queryPool(ctx, keeper)
	require.Nil(t, err)

	var pool types.Pool
	errRes = keeper.Codec().UnmarshalJSON(res, &pool)
	assert.Nil(t, errRes)
	assert.Equal(t, keeper.GetPool(ctx), pool)
}

func TestQueryValidators(t *testing.T) {

	ctx, _, keeper := keep.CreateTestInput(t, false, 10000)

	// Create Validators
	msg1 := types.NewMsgCreateValidator(addrVal1, pk1, sdk.NewCoin("steak", sdk.NewInt(1000)), Description{})
	handleMsgCreateValidator(ctx, msg1, keeper)
	msg2 := types.NewMsgCreateValidator(addrVal2, pk2, sdk.NewCoin("steak", sdk.NewInt(100)), Description{})
	handleMsgCreateValidator(ctx, msg2, keeper)

	// Query Validators
	validators := keeper.GetBechValidators(ctx)
	res, err := queryValidators(ctx, []string{""}, keeper)
	assert.Nil(t, err)

	var validatorsResp []types.BechValidator
	errRes := keeper.Codec().UnmarshalJSON(res, &validatorsResp)
	assert.Nil(t, errRes)

	assert.Equal(t, len(validators), len(validatorsResp))
	assert.ElementsMatch(t, validators, validatorsResp)

	// Query each validator
	queryParams := newTestValidatorQuery(addrVal1)
	bz, errRes := keeper.Codec().MarshalJSON(queryParams)
	assert.Nil(t, errRes)

	query := abci.RequestQuery{
		Path: "/custom/stake/validator",
		Data: bz,
	}
	res, err = queryValidator(ctx, []string{query.Path}, query, keeper)
	assert.Nil(t, err)

	var validator types.BechValidator
	errRes = keeper.Codec().UnmarshalJSON(res, &validator)
	assert.Nil(t, errRes)

	assert.Equal(t, validators[0], validator)
}

func TestQueryDelegation(t *testing.T) {
	ctx, _, keeper := keep.CreateTestInput(t, false, 10000)

	// Create Validators and Delegation
	msg1 := types.NewMsgCreateValidator(addrVal1, pk1, sdk.NewCoin("steak", sdk.NewInt(1000)), Description{})
	handleMsgCreateValidator(ctx, msg1, keeper)
	msg2 := types.NewMsgDelegate(addrAcc2, addrVal1, sdk.NewCoin("steak", sdk.NewInt(20)))
	handleMsgDelegate(ctx, msg2, keeper)

	// Query Delegator bonded validators
	queryParams := newTestDelegatorQuery(addrAcc2)
	bz, errRes := keeper.Codec().MarshalJSON(queryParams)
	assert.Nil(t, errRes)

	query := abci.RequestQuery{
		Path: "/custom/stake/delegatorValidators",
		Data: bz,
	}

	delValidators := keeper.GetDelegatorBechValidators(ctx, addrAcc2)
	res, err := queryDelegatorValidators(ctx, []string{query.Path}, query, keeper)
	assert.Nil(t, err)

	var validatorsResp []types.BechValidator
	errRes = keeper.Codec().UnmarshalJSON(res, &validatorsResp)
	assert.Nil(t, errRes)

	assert.Equal(t, len(delValidators), len(validatorsResp))
	assert.ElementsMatch(t, delValidators, validatorsResp)

	// Query bonded validator
	queryBondParams := newTestBondQuery(addrAcc2, addrVal1)
	bz, errRes = keeper.Codec().MarshalJSON(queryBondParams)
	assert.Nil(t, errRes)

	query = abci.RequestQuery{
		Path: "/custom/stake/delegatorValidator",
		Data: bz,
	}

	res, err = queryDelegatorValidator(ctx, []string{query.Path}, query, keeper)
	assert.Nil(t, err)

	var validator types.BechValidator
	errRes = keeper.Codec().UnmarshalJSON(res, &validator)
	assert.Nil(t, errRes)

	assert.Equal(t, delValidators[0], validator)

	// Query delegation

	query = abci.RequestQuery{
		Path: "/custom/stake/delegation",
		Data: bz,
	}

	delegation, found := keeper.GetDelegation(ctx, addrAcc2, addrVal1)
	assert.True(t, found)

	delegationNoRat := types.NewDelegationWithoutDec(delegation)

	res, err = queryDelegation(ctx, []string{query.Path}, query, keeper)
	assert.Nil(t, err)

	var delegationRes types.DelegationWithoutDec
	errRes = keeper.Codec().UnmarshalJSON(res, &delegationRes)
	assert.Nil(t, errRes)

	assert.Equal(t, delegationNoRat, delegationRes)

	// Query unbonging delegation

	msg3 := types.NewMsgBeginUnbonding(addrAcc2, addrVal1, sdk.NewDec(10))
	handleMsgBeginUnbonding(ctx, msg3, keeper)

	query = abci.RequestQuery{
		Path: "/custom/stake/unbondingDelegation",
		Data: bz,
	}

	unbond, found := keeper.GetUnbondingDelegation(ctx, addrAcc2, addrVal1)
	assert.True(t, found)

	res, err = queryUnbondingDelegation(ctx, []string{query.Path}, query, keeper)
	assert.Nil(t, err)

	var unbondRes types.UnbondingDelegation
	errRes = keeper.Codec().UnmarshalJSON(res, &unbondRes)
	assert.Nil(t, errRes)

	assert.Equal(t, unbond, unbondRes)

	// Query Delegator Summary

	query = abci.RequestQuery{
		Path: "/custom/stake/delegator",
		Data: bz,
	}

	res, err = queryDelegator(ctx, []string{query.Path}, query, keeper)
	assert.Nil(t, err)

	var summary types.DelegationSummary
	errRes = keeper.Codec().UnmarshalJSON(res, &summary)
	assert.Nil(t, errRes)

	assert.Equal(t, unbond, summary.UnbondingDelegations[0])
}
