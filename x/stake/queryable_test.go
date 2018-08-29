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

func newTestAddrQuery(accountAddr sdk.AccAddress) QueryAddressParams {
	return QueryAddressParams{
		AccountAddr: accountAddr,
	}
}

func newTestBondQuery(delegatorAddr, validatorAddr sdk.AccAddress) QueryBondsParams {
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

	addr1, addr2 := keep.Addrs[0], keep.Addrs[1]
	pk1, pk2 := keep.PKs[0], keep.PKs[1]

	// Create Validators
	msg1 := types.NewMsgCreateValidator(addr1, pk1, sdk.NewCoin("steak", sdk.NewInt(1000)), Description{})
	handleMsgCreateValidator(ctx, msg1, keeper)
	msg2 := types.NewMsgCreateValidator(addr2, pk2, sdk.NewCoin("steak", sdk.NewInt(100)), Description{})
	handleMsgCreateValidator(ctx, msg2, keeper)

	// Query Validators
	validators := keeper.GetValidators(ctx)
	res, err := queryValidators(ctx, []string{""}, keeper)
	assert.Nil(t, err)

	var validatorsResp []types.Validator
	errRes := keeper.Codec().UnmarshalJSON(res, &validatorsResp)
	assert.Nil(t, errRes)

	assert.Equal(t, len(validators), len(validatorsResp))
	assert.ElementsMatch(t, validators, validatorsResp)

	// Query each validator
	queryParams := newTestAddrQuery(addr1)
	bz, errRes := keeper.Codec().MarshalJSON(queryParams)
	assert.Nil(t, errRes)

	query := abci.RequestQuery{
		Path: "/custom/stake/delegation",
		Data: bz,
	}
	res, err = queryValidator(ctx, []string{query.Path}, query, keeper)
	assert.Nil(t, err)

	var validator types.Validator
	errRes = keeper.Codec().UnmarshalJSON(res, &validator)
	assert.Nil(t, errRes)

	assert.Equal(t, validators[0], validator)
}

func TestQueryDelegation(t *testing.T) {
	ctx, _, keeper := keep.CreateTestInput(t, false, 10000)

	addr1, addr2 := keep.Addrs[0], keep.Addrs[1]
	pk1, _ := keep.PKs[0], keep.PKs[1]

	// Create Validators and Delegation
	msg1 := types.NewMsgCreateValidator(addr1, pk1, sdk.NewCoin("steak", sdk.NewInt(1000)), Description{})
	handleMsgCreateValidator(ctx, msg1, keeper)
	msg2 := types.NewMsgDelegate(addr2, addr1, sdk.NewCoin("steak", sdk.NewInt(20)))
	handleMsgDelegate(ctx, msg2, keeper)

	// Query Delegator bonded validators
	queryParams := newTestAddrQuery(addr2)
	bz, errRes := keeper.Codec().MarshalJSON(queryParams)
	assert.Nil(t, errRes)

	query := abci.RequestQuery{
		Path: "/custom/stake/delegatorValidators",
		Data: bz,
	}

	delValidators := keeper.GetDelegatorValidators(ctx, addr2)
	res, err := queryDelegatorValidators(ctx, []string{query.Path}, query, keeper)
	assert.Nil(t, err)

	var validatorsResp []types.Validator
	errRes = keeper.Codec().UnmarshalJSON(res, &validatorsResp)
	assert.Nil(t, errRes)

	assert.Equal(t, len(delValidators), len(validatorsResp))
	assert.ElementsMatch(t, delValidators, validatorsResp)

	// Query bonded validator
	queryBondParams := newTestBondQuery(addr2, addr1)
	bz, errRes = keeper.Codec().MarshalJSON(queryBondParams)
	assert.Nil(t, errRes)

	query = abci.RequestQuery{
		Path: "/custom/stake/delegatorValidator",
		Data: bz,
	}

	res, err = queryDelegatorValidator(ctx, []string{query.Path}, query, keeper)
	assert.Nil(t, err)

	var validator types.Validator
	errRes = keeper.Codec().UnmarshalJSON(res, &validator)
	assert.Nil(t, errRes)

	assert.Equal(t, delValidators[0], validator)

	// Query delegation

	query = abci.RequestQuery{
		Path: "/custom/stake/delegation",
		Data: bz,
	}

	delegation, found := keeper.GetDelegation(ctx, addr2, addr1)
	assert.True(t, found)

	res, err = queryDelegation(ctx, []string{query.Path}, query, keeper)
	assert.Nil(t, err)

	var delegationRes types.Delegation
	errRes = keeper.Codec().UnmarshalJSON(res, &delegationRes)
	assert.Nil(t, errRes)

	assert.Equal(t, delegation, delegationRes)

	// Query unbonging delegation

	msg3 := types.NewMsgBeginUnbonding(addr2, addr1, sdk.NewDec(10))
	handleMsgBeginUnbonding(ctx, msg3, keeper)

	query = abci.RequestQuery{
		Path: "/custom/stake/unbonding-delegation",
		Data: bz,
	}

	unbond, found := keeper.GetUnbondingDelegation(ctx, addr2, addr1)
	assert.True(t, found)

	res, err = queryUnbondingDelegation(ctx, []string{query.Path}, query, keeper)
	assert.Nil(t, err)

	var unbondRes types.UnbondingDelegation
	errRes = keeper.Codec().UnmarshalJSON(res, &unbondRes)
	assert.Nil(t, errRes)

	assert.Equal(t, unbond, unbondRes)

}
