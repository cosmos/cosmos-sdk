package keeper

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

const custom = "custom"

func getQueriedParams(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier) types.Params {
	var params types.Params

	bz, err := querier(ctx, []string{types.QueryParams}, abci.RequestQuery{})
	require.Nil(t, err)
	require.Nil(t, cdc.UnmarshalJSON(bz, &params))

	return params
}

func getQueriedValidatorOutstandingRewards(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, validatorAddr sdk.ValAddress) (outstandingRewards sdk.DecCoins) {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryValidatorOutstandingRewards}, "/"),
		Data: cdc.MustMarshalJSON(types.NewQueryValidatorOutstandingRewardsParams(validatorAddr)),
	}

	bz, err := querier(ctx, []string{types.QueryValidatorOutstandingRewards}, query)
	require.Nil(t, err)
	require.Nil(t, cdc.UnmarshalJSON(bz, &outstandingRewards))

	return
}

func getQueriedValidatorCommission(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, validatorAddr sdk.ValAddress) (validatorCommission sdk.DecCoins) {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryValidatorCommission}, "/"),
		Data: cdc.MustMarshalJSON(types.NewQueryValidatorCommissionParams(validatorAddr)),
	}

	bz, err := querier(ctx, []string{types.QueryValidatorCommission}, query)
	require.Nil(t, err)
	require.Nil(t, cdc.UnmarshalJSON(bz, &validatorCommission))

	return
}

func getQueriedValidatorSlashes(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, validatorAddr sdk.ValAddress, startHeight uint64, endHeight uint64) (slashes []types.ValidatorSlashEvent) {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryValidatorSlashes}, "/"),
		Data: cdc.MustMarshalJSON(types.NewQueryValidatorSlashesParams(validatorAddr, startHeight, endHeight)),
	}

	bz, err := querier(ctx, []string{types.QueryValidatorSlashes}, query)
	require.Nil(t, err)
	require.Nil(t, cdc.UnmarshalJSON(bz, &slashes))

	return
}

func getQueriedDelegationRewards(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) (rewards sdk.DecCoins) {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryDelegationRewards}, "/"),
		Data: cdc.MustMarshalJSON(types.NewQueryDelegationRewardsParams(delegatorAddr, validatorAddr)),
	}

	bz, err := querier(ctx, []string{types.QueryDelegationRewards}, query)
	require.Nil(t, err)
	require.Nil(t, cdc.UnmarshalJSON(bz, &rewards))

	return
}

func getQueriedDelegatorTotalRewards(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, delegatorAddr sdk.AccAddress) (response types.QueryDelegatorTotalRewardsResponse) {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryDelegatorTotalRewards}, "/"),
		Data: cdc.MustMarshalJSON(types.NewQueryDelegatorParams(delegatorAddr)),
	}

	bz, err := querier(ctx, []string{types.QueryDelegatorTotalRewards}, query)
	require.Nil(t, err)
	require.Nil(t, cdc.UnmarshalJSON(bz, &response))

	return
}

func getQueriedCommunityPool(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier) (ptr []byte) {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryCommunityPool}, ""),
		Data: []byte{},
	}

	cp, err := querier(ctx, []string{types.QueryCommunityPool}, query)
	require.Nil(t, err)
	require.Nil(t, cdc.UnmarshalJSON(cp, &ptr))

	return
}

func TestQueries(t *testing.T) {
	cdc := codec.New()
	types.RegisterCodec(cdc)
	supply.RegisterCodec(cdc)
	ctx, _, keeper, sk, _ := CreateTestInputDefault(t, false, 100)
	querier := NewQuerier(keeper)

	// test param queries
	params := types.Params{
		CommunityTax:        sdk.NewDecWithPrec(3, 1),
		BaseProposerReward:  sdk.NewDecWithPrec(2, 1),
		BonusProposerReward: sdk.NewDecWithPrec(1, 1),
		WithdrawAddrEnabled: true,
	}

	keeper.SetParams(ctx, params)

	paramsRes := getQueriedParams(t, ctx, cdc, querier)
	require.Equal(t, params.CommunityTax, paramsRes.CommunityTax)
	require.Equal(t, params.BaseProposerReward, paramsRes.BaseProposerReward)
	require.Equal(t, params.BonusProposerReward, paramsRes.BonusProposerReward)
	require.Equal(t, params.WithdrawAddrEnabled, paramsRes.WithdrawAddrEnabled)

	// test outstanding rewards query
	outstandingRewards := sdk.DecCoins{{Denom: "mytoken", Amount: sdk.NewDec(3)}, {Denom: "myothertoken", Amount: sdk.NewDecWithPrec(3, 7)}}
	keeper.SetValidatorOutstandingRewards(ctx, valOpAddr1, outstandingRewards)
	retOutstandingRewards := getQueriedValidatorOutstandingRewards(t, ctx, cdc, querier, valOpAddr1)
	require.Equal(t, outstandingRewards, retOutstandingRewards)

	// test validator commission query
	commission := sdk.DecCoins{{Denom: "token1", Amount: sdk.NewDec(4)}, {Denom: "token2", Amount: sdk.NewDec(2)}}
	keeper.SetValidatorAccumulatedCommission(ctx, valOpAddr1, commission)
	retCommission := getQueriedValidatorCommission(t, ctx, cdc, querier, valOpAddr1)
	require.Equal(t, commission, retCommission)

	// test delegator's total rewards query
	delRewards := getQueriedDelegatorTotalRewards(t, ctx, cdc, querier, sdk.AccAddress(valOpAddr1))
	require.Equal(t, types.QueryDelegatorTotalRewardsResponse{}, delRewards)

	// test validator slashes query with height range
	slashOne := types.NewValidatorSlashEvent(3, sdk.NewDecWithPrec(5, 1))
	slashTwo := types.NewValidatorSlashEvent(7, sdk.NewDecWithPrec(6, 1))
	keeper.SetValidatorSlashEvent(ctx, valOpAddr1, 3, 0, slashOne)
	keeper.SetValidatorSlashEvent(ctx, valOpAddr1, 7, 0, slashTwo)
	slashes := getQueriedValidatorSlashes(t, ctx, cdc, querier, valOpAddr1, 0, 2)
	require.Equal(t, 0, len(slashes))
	slashes = getQueriedValidatorSlashes(t, ctx, cdc, querier, valOpAddr1, 0, 5)
	require.Equal(t, []types.ValidatorSlashEvent{slashOne}, slashes)
	slashes = getQueriedValidatorSlashes(t, ctx, cdc, querier, valOpAddr1, 0, 10)
	require.Equal(t, []types.ValidatorSlashEvent{slashOne, slashTwo}, slashes)

	// test delegation rewards query
	sh := staking.NewHandler(sk)
	comm := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(
		valOpAddr1, valConsPk1, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, comm, sdk.OneInt(),
	)

	res, err := sh(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	staking.EndBlocker(ctx, sk)

	val := sk.Validator(ctx, valOpAddr1)
	rewards := getQueriedDelegationRewards(t, ctx, cdc, querier, sdk.AccAddress(valOpAddr1), valOpAddr1)
	require.True(t, rewards.IsZero())
	initial := int64(10)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial)}}
	keeper.AllocateTokensToValidator(ctx, val, tokens)
	rewards = getQueriedDelegationRewards(t, ctx, cdc, querier, sdk.AccAddress(valOpAddr1), valOpAddr1)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial / 2)}}, rewards)

	// test delegator's total rewards query
	delRewards = getQueriedDelegatorTotalRewards(t, ctx, cdc, querier, sdk.AccAddress(valOpAddr1))
	expectedDelReward := types.NewDelegationDelegatorReward(valOpAddr1,
		sdk.DecCoins{sdk.NewInt64DecCoin("stake", 5)})
	wantDelRewards := types.NewQueryDelegatorTotalRewardsResponse(
		[]types.DelegationDelegatorReward{expectedDelReward}, expectedDelReward.Reward)
	require.Equal(t, wantDelRewards, delRewards)

	// currently community pool hold nothing so we should return null
	communityPool := getQueriedCommunityPool(t, ctx, cdc, querier)
	require.Nil(t, communityPool)
}
