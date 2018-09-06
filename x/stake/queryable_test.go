package stake

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
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
	cdc := wire.NewCodec()
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
	cdc := wire.NewCodec()
	ctx, _, keeper := keep.CreateTestInput(t, false, 10000)

	// Create Validators
	msg1 := types.NewMsgCreateValidator(addrVal1, pk1, sdk.NewCoin("steak", sdk.NewInt(1000)), Description{})
	handleMsgCreateValidator(ctx, msg1, keeper)
	msg2 := types.NewMsgCreateValidator(addrVal2, pk2, sdk.NewCoin("steak", sdk.NewInt(100)), Description{})
	handleMsgCreateValidator(ctx, msg2, keeper)

	// Query Validators
	var bechValidators []types.BechValidator
	validators := keeper.GetValidators(ctx)
	for _, val := range validators {
		bechVal, err := val.Bech32Validator()
		require.Nil(t, err)
		bechValidators = append(bechValidators, bechVal)
	}
	res, err := queryValidators(ctx, cdc, []string{""}, keeper)
	require.Nil(t, err)

	var validatorsResp []types.BechValidator
	errRes := cdc.UnmarshalJSON(res, &validatorsResp)
	require.Nil(t, errRes)

	require.Equal(t, len(bechValidators), len(validatorsResp))
	require.ElementsMatch(t, bechValidators, validatorsResp)

	// Query each validator
	queryParams := newTestValidatorQuery(addrVal1)
	bz, errRes := cdc.MarshalJSON(queryParams)
	require.Nil(t, errRes)

	query := abci.RequestQuery{
		Path: "/custom/stake/validator",
		Data: bz,
	}
	res, err = queryValidator(ctx, cdc, []string{query.Path}, query, keeper)
	require.Nil(t, err)

	var validator types.BechValidator
	errRes = cdc.UnmarshalJSON(res, &validator)
	require.Nil(t, errRes)

	require.Equal(t, validators[0], validator)
}

func TestQueryDelegation(t *testing.T) {
	cdc := wire.NewCodec()
	ctx, _, keeper := keep.CreateTestInput(t, false, 10000)

	// Create Validators and Delegation
	msg1 := types.NewMsgCreateValidator(addrVal1, pk1, sdk.NewCoin("steak", sdk.NewInt(1000)), Description{})
	handleMsgCreateValidator(ctx, msg1, keeper)
	msg2 := types.NewMsgDelegate(addrAcc2, addrVal1, sdk.NewCoin("steak", sdk.NewInt(20)))
	handleMsgDelegate(ctx, msg2, keeper)

	// Query Delegator bonded validators
	queryParams := newTestDelegatorQuery(addrAcc2)
	bz, errRes := cdc.MarshalJSON(queryParams)
	require.Nil(t, errRes)

	query := abci.RequestQuery{
		Path: "/custom/stake/delegatorValidators",
		Data: bz,
	}

	delValidators := keeper.GetDelegatorBechValidators(ctx, addrAcc2)
	res, err := queryDelegatorValidators(ctx, cdc, []string{query.Path}, query, keeper)
	require.Nil(t, err)

	var validatorsResp []types.BechValidator
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

	res, err = queryDelegatorValidator(ctx, cdc, []string{query.Path}, query, keeper)
	require.Nil(t, err)

	var validator types.BechValidator
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

	delegationREST := delegation.ToRest()

	res, err = queryDelegation(ctx, cdc, []string{query.Path}, query, keeper)
	require.Nil(t, err)

	var delegationRestRes types.DelegationREST
	errRes = cdc.UnmarshalJSON(res, &delegationRestRes)
	require.Nil(t, errRes)

	require.Equal(t, delegationREST, delegationRestRes)

	// Query unbonging delegation

	msg3 := types.NewMsgBeginUnbonding(addrAcc2, addrVal1, sdk.NewDec(10))
	handleMsgBeginUnbonding(ctx, msg3, keeper)

	query = abci.RequestQuery{
		Path: "/custom/stake/unbondingDelegation",
		Data: bz,
	}

	unbond, found := keeper.GetUnbondingDelegation(ctx, addrAcc2, addrVal1)
	require.True(t, found)

	res, err = queryUnbondingDelegation(ctx, cdc, []string{query.Path}, query, keeper)
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

	res, err = queryDelegator(ctx, cdc, []string{query.Path}, query, keeper)
	require.Nil(t, err)

	var summary types.DelegationSummary
	errRes = cdc.UnmarshalJSON(res, &summary)
	require.Nil(t, errRes)

	require.Equal(t, unbond, summary.UnbondingDelegations[0])
}
