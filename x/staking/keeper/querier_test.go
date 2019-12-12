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

	header := abci.Header{
		ChainID: "HelloChain",
		Height:  5,
	}
	hi := types.NewHistoricalInfo(header, validators[:])
	keeper.SetHistoricalInfo(ctx, 5, hi)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	querier := NewQuerier(keeper)

	bz, err := querier(ctx, []string{"other"}, query)
	require.Error(t, err)
	require.Nil(t, bz)

	_, err = querier(ctx, []string{"pool"}, query)
	require.NoError(t, err)

	_, err = querier(ctx, []string{"parameters"}, query)
	require.NoError(t, err)

	queryValParams := types.NewQueryValidatorParams(addrVal1)
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
	cdc := codec.New()
	ctx, _, keeper, _ := CreateTestInput(t, false, 1000)
	bondDenom := sdk.DefaultBondDenom

	res, err := queryParameters(ctx, keeper)
	require.NoError(t, err)

	var params types.Params
	errRes := cdc.UnmarshalJSON(res, &params)
	require.NoError(t, errRes)
	require.Equal(t, keeper.GetParams(ctx), params)

	res, err = queryPool(ctx, keeper)
	require.NoError(t, err)

	var pool types.Pool
	bondedPool := keeper.GetBondedPool(ctx)
	notBondedPool := keeper.GetNotBondedPool(ctx)
	errRes = cdc.UnmarshalJSON(res, &pool)
	require.NoError(t, errRes)
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
		require.NoError(t, err)

		req := abci.RequestQuery{
			Path: fmt.Sprintf("/custom/%s/%s", types.QuerierRoute, types.QueryValidators),
			Data: bz,
		}

		res, err := queryValidators(ctx, req, keeper)
		require.NoError(t, err)

		var validatorsResp []types.Validator
		err = cdc.UnmarshalJSON(res, &validatorsResp)
		require.NoError(t, err)

		require.Equal(t, 1, len(validatorsResp))
		require.ElementsMatch(t, validators[i].OperatorAddress, validatorsResp[0].OperatorAddress)

	}

	// Query each validator
	queryParams := types.NewQueryValidatorParams(addrVal1)
	bz, err := cdc.MarshalJSON(queryParams)
	require.NoError(t, err)

	query := abci.RequestQuery{
		Path: "/custom/staking/validator",
		Data: bz,
	}
	res, err := queryValidator(ctx, query, keeper)
	require.NoError(t, err)

	var validator types.Validator
	err = cdc.UnmarshalJSON(res, &validator)
	require.NoError(t, err)

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
	require.NoError(t, errRes)

	query := abci.RequestQuery{
		Path: "/custom/staking/delegatorValidators",
		Data: bz,
	}

	delValidators := keeper.GetDelegatorValidators(ctx, addrAcc2, params.MaxValidators)

	res, err := queryDelegatorValidators(ctx, query, keeper)
	require.NoError(t, err)

	var validatorsResp []types.Validator
	errRes = cdc.UnmarshalJSON(res, &validatorsResp)
	require.NoError(t, errRes)

	require.Equal(t, len(delValidators), len(validatorsResp))
	require.ElementsMatch(t, delValidators, validatorsResp)

	// error unknown request
	query.Data = bz[:len(bz)-1]

	_, err = queryDelegatorValidators(ctx, query, keeper)
	require.Error(t, err)

	// Query bonded validator
	queryBondParams := types.NewQueryBondsParams(addrAcc2, addrVal1)
	bz, errRes = cdc.MarshalJSON(queryBondParams)
	require.NoError(t, errRes)

	query = abci.RequestQuery{
		Path: "/custom/staking/delegatorValidator",
		Data: bz,
	}

	res, err = queryDelegatorValidator(ctx, query, keeper)
	require.NoError(t, err)

	var validator types.Validator
	errRes = cdc.UnmarshalJSON(res, &validator)
	require.NoError(t, errRes)

	require.Equal(t, delValidators[0], validator)

	// error unknown request
	query.Data = bz[:len(bz)-1]

	_, err = queryDelegatorValidator(ctx, query, keeper)
	require.Error(t, err)

	// Query delegation

	query = abci.RequestQuery{
		Path: "/custom/staking/delegation",
		Data: bz,
	}

	delegation, found := keeper.GetDelegation(ctx, addrAcc2, addrVal1)
	require.True(t, found)

	res, err = queryDelegation(ctx, query, keeper)
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

	res, err = queryDelegatorDelegations(ctx, query, keeper)
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

	_, err = queryDelegation(ctx, query, keeper)
	require.Error(t, err)

	// Query validator delegations

	bz, errRes = cdc.MarshalJSON(types.NewQueryValidatorParams(addrVal1))
	require.NoError(t, errRes)

	query = abci.RequestQuery{
		Path: "custom/staking/validatorDelegations",
		Data: bz,
	}

	res, err = queryValidatorDelegations(ctx, query, keeper)
	require.NoError(t, err)

	var delegationsRes types.DelegationResponses
	errRes = cdc.UnmarshalJSON(res, &delegationsRes)
	require.NoError(t, errRes)
	require.Len(t, delegatorDelegations, 1)
	require.Equal(t, delegation.ValidatorAddress, delegationsRes[0].ValidatorAddress)
	require.Equal(t, delegation.DelegatorAddress, delegationsRes[0].DelegatorAddress)
	require.Equal(t, sdk.NewCoin(sdk.DefaultBondDenom, delegation.Shares.TruncateInt()), delegationsRes[0].Balance)

	// Query unbonging delegation
	unbondingTokens := sdk.TokensFromConsensusPower(10)
	_, err = keeper.Undelegate(ctx, addrAcc2, val1.OperatorAddress, unbondingTokens.ToDec())
	require.NoError(t, err)

	queryBondParams = types.NewQueryBondsParams(addrAcc2, addrVal1)
	bz, errRes = cdc.MarshalJSON(queryBondParams)
	require.NoError(t, errRes)

	query = abci.RequestQuery{
		Path: "/custom/staking/unbondingDelegation",
		Data: bz,
	}

	unbond, found := keeper.GetUnbondingDelegation(ctx, addrAcc2, addrVal1)
	require.True(t, found)

	res, err = queryUnbondingDelegation(ctx, query, keeper)
	require.NoError(t, err)

	var unbondRes types.UnbondingDelegation
	errRes = cdc.UnmarshalJSON(res, &unbondRes)
	require.NoError(t, errRes)

	require.Equal(t, unbond, unbondRes)

	// error unknown request
	query.Data = bz[:len(bz)-1]

	_, err = queryUnbondingDelegation(ctx, query, keeper)
	require.Error(t, err)

	// Query Delegator Delegations

	query = abci.RequestQuery{
		Path: "/custom/staking/delegatorUnbondingDelegations",
		Data: bz,
	}

	res, err = queryDelegatorUnbondingDelegations(ctx, query, keeper)
	require.NoError(t, err)

	var delegatorUbds []types.UnbondingDelegation
	errRes = cdc.UnmarshalJSON(res, &delegatorUbds)
	require.NoError(t, errRes)
	require.Equal(t, unbond, delegatorUbds[0])

	// error unknown request
	query.Data = bz[:len(bz)-1]

	_, err = queryDelegatorUnbondingDelegations(ctx, query, keeper)
	require.Error(t, err)

	// Query redelegation
	redelegationTokens := sdk.TokensFromConsensusPower(10)
	_, err = keeper.BeginRedelegation(ctx, addrAcc2, val1.OperatorAddress,
		val2.OperatorAddress, redelegationTokens.ToDec())
	require.NoError(t, err)
	redel, found := keeper.GetRedelegation(ctx, addrAcc2, val1.OperatorAddress, val2.OperatorAddress)
	require.True(t, found)

	bz, errRes = cdc.MarshalJSON(types.NewQueryRedelegationParams(addrAcc2, val1.OperatorAddress, val2.OperatorAddress))
	require.NoError(t, errRes)

	query = abci.RequestQuery{
		Path: "/custom/staking/redelegations",
		Data: bz,
	}

	res, err = queryRedelegations(ctx, query, keeper)
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
	require.NoError(t, errRes)

	query := abci.RequestQuery{
		Path: "/custom/staking/redelegations",
		Data: bz,
	}

	res, err := queryRedelegations(ctx, query, keeper)
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
	queryValidatorParams := types.NewQueryValidatorParams(val1.GetOperator())
	bz, errRes = cdc.MarshalJSON(queryValidatorParams)
	require.NoError(t, errRes)

	query = abci.RequestQuery{
		Path: "/custom/staking/redelegations",
		Data: bz,
	}

	res, err = queryRedelegations(ctx, query, keeper)
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
	require.NoError(t, errRes)
	query := abci.RequestQuery{
		Path: "/custom/staking/unbondingDelegation",
		Data: bz,
	}
	res, err := queryUnbondingDelegation(ctx, query, keeper)
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
	_, err = queryUnbondingDelegation(ctx, query, keeper)
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
	res, err = queryDelegatorUnbondingDelegations(ctx, query, keeper)
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
	res, err = queryDelegatorUnbondingDelegations(ctx, query, keeper)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NoError(t, cdc.UnmarshalJSON(res, &ubDels))
	require.Equal(t, 0, len(ubDels))
}

func TestQueryHistoricalInfo(t *testing.T) {
	cdc := codec.New()
	ctx, _, keeper, _ := CreateTestInput(t, false, 10000)

	// Create Validators and Delegation
	val1 := types.NewValidator(addrVal1, pk1, types.Description{})
	val2 := types.NewValidator(addrVal2, pk2, types.Description{})
	vals := []types.Validator{val1, val2}
	keeper.SetValidator(ctx, val1)
	keeper.SetValidator(ctx, val2)

	header := abci.Header{
		ChainID: "HelloChain",
		Height:  5,
	}
	hi := types.NewHistoricalInfo(header, vals)
	keeper.SetHistoricalInfo(ctx, 5, hi)

	queryHistoricalParams := types.NewQueryHistoricalInfoParams(4)
	bz, errRes := cdc.MarshalJSON(queryHistoricalParams)
	require.NoError(t, errRes)
	query := abci.RequestQuery{
		Path: "/custom/staking/historicalInfo",
		Data: bz,
	}
	res, err := queryHistoricalInfo(ctx, query, keeper)
	require.Error(t, err, "Invalid query passed")
	require.Nil(t, res, "Invalid query returned non-nil result")

	queryHistoricalParams = types.NewQueryHistoricalInfoParams(5)
	bz, errRes = cdc.MarshalJSON(queryHistoricalParams)
	require.NoError(t, errRes)
	query.Data = bz
	res, err = queryHistoricalInfo(ctx, query, keeper)
	require.NoError(t, err, "Valid query passed")
	require.NotNil(t, res, "Valid query returned nil result")

	var recv types.HistoricalInfo
	require.NoError(t, cdc.UnmarshalJSON(res, &recv))
	require.Equal(t, hi, recv, "HistoricalInfo query returned wrong result")
}
