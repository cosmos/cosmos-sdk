package keeper

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/poa/internal/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	addrVal1 = sdk.ValAddress(Addrs[0])
)

func TestNewQuerier(t *testing.T) {
	cdc := codec.New()
	ctx, _, keeper, _ := CreateTestInput(t, false)

	var validators [2]types.Validator
	for i := range validators {
		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], stakingtypes.Description{})
		keeper.SetValidator(ctx, validators[i])
		keeper.SetValidatorByPowerIndex(ctx, validators[i])
	}

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	querier := NewQuerier(keeper)

	_, err := querier(ctx, []string{"parameters"}, query)
	require.Nil(t, err)

	queryValParams := types.NewQueryValidatorParams(addrVal1)
	bz, errRes := cdc.MarshalJSON(queryValParams)
	require.Nil(t, errRes)

	query.Path = fmt.Sprintf("/custom/%s/%s", types.QuerierRoute, types.QueryValidator)
	query.Data = bz

	_, err = querier(ctx, []string{"validator"}, query)
	require.Nil(t, err)

	queryValsParams := types.NewQueryValidatorsParams(0, 0, "active")
	bz1, errRes1 := cdc.MarshalJSON(queryValsParams)
	require.Nil(t, errRes1)

	query.Path = fmt.Sprintf("/custom/%s/%s", types.QuerierRoute, types.QueryValidator)
	query.Data = bz1

	_, err = querier(ctx, []string{"validators"}, query)
	require.Nil(t, err)
}

func TestQueryValidators(t *testing.T) {
	cdc := codec.New()
	ctx, _, keeper, _ := CreateTestInput(t, false)
	params := keeper.GetParams(ctx)

	// Create Validators
	status := []sdk.BondStatus{sdk.Bonded, sdk.Unbonded, sdk.Unbonding}
	var validators [3]types.Validator
	for i := range validators {
		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], stakingtypes.Description{})
		validators[i] = validators[i].UpdateStatus(status[i])
		keeper.SetValidator(ctx, validators[i])
	}

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

		var validatorsResp types.Validators
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
		Path: fmt.Sprintf("/custom/%s/%s", types.QuerierRoute, types.QueryValidator),
		Data: bz,
	}
	res, err := queryValidator(ctx, query, keeper)
	require.Nil(t, err)

	var validator types.Validator
	err = cdc.UnmarshalJSON(res, &validator)
	require.Nil(t, err)

	require.Equal(t, queriedValidators[0], validator)
}
