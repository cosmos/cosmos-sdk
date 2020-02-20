package keeper

import (
	"testing"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

var (
	addrAcc1, addrAcc2 = Addrs[0], Addrs[1]
	addrVal1, addrVal2 = sdk.ValAddress(Addrs[0]), sdk.ValAddress(Addrs[1])
	pk1, pk2           = PKs[0], PKs[1]
)

func TestQueryUnbondingDelegation(t *testing.T) {
	cdc := codec.New()
	ctx, _, _, keeper, _ := CreateTestInput(t, false, 10000)

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
	ctx, _, _, keeper, _ := CreateTestInput(t, false, 10000)

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
