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
)

const custom = "custom"

func getQueriedParams(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier) (communityTax sdk.Dec, baseProposerReward sdk.Dec, bonusProposerReward sdk.Dec, withdrawAddrEnabled bool) {

	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, QueryParams, ParamCommunityTax}, "/"),
		Data: []byte{},
	}

	bz, err := querier(ctx, []string{QueryParams, ParamCommunityTax}, query)
	require.Nil(t, err)
	require.Nil(t, cdc.UnmarshalJSON(bz, &communityTax))

	query = abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, QueryParams, ParamBaseProposerReward}, "/"),
		Data: []byte{},
	}

	bz, err = querier(ctx, []string{QueryParams, ParamBaseProposerReward}, query)
	require.Nil(t, err)
	require.Nil(t, cdc.UnmarshalJSON(bz, &baseProposerReward))

	query = abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, QueryParams, ParamBonusProposerReward}, "/"),
		Data: []byte{},
	}

	bz, err = querier(ctx, []string{QueryParams, ParamBonusProposerReward}, query)
	require.Nil(t, err)
	require.Nil(t, cdc.UnmarshalJSON(bz, &bonusProposerReward))

	query = abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, QueryParams, ParamWithdrawAddrEnabled}, "/"),
		Data: []byte{},
	}

	bz, err = querier(ctx, []string{QueryParams, ParamWithdrawAddrEnabled}, query)
	require.Nil(t, err)
	require.Nil(t, cdc.UnmarshalJSON(bz, &withdrawAddrEnabled))

	return
}

func getQueriedValidatorOutstandingRewards(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, validatorAddr sdk.ValAddress) (outstandingRewards sdk.DecCoins) {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, QueryValidatorOutstandingRewards}, "/"),
		Data: cdc.MustMarshalJSON(NewQueryValidatorOutstandingRewardsParams(validatorAddr)),
	}

	bz, err := querier(ctx, []string{QueryValidatorOutstandingRewards}, query)
	require.Nil(t, err)
	require.Nil(t, cdc.UnmarshalJSON(bz, &outstandingRewards))

	return
}

func getQueriedValidatorCommission(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, validatorAddr sdk.ValAddress) (validatorCommission sdk.DecCoins) {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, QueryValidatorCommission}, "/"),
		Data: cdc.MustMarshalJSON(NewQueryValidatorCommissionParams(validatorAddr)),
	}

	bz, err := querier(ctx, []string{QueryValidatorCommission}, query)
	require.Nil(t, err)
	require.Nil(t, cdc.UnmarshalJSON(bz, &validatorCommission))

	return
}

func getQueriedValidatorSlashes(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, validatorAddr sdk.ValAddress, startHeight uint64, endHeight uint64) (slashes []types.ValidatorSlashEvent) {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, QueryValidatorSlashes}, "/"),
		Data: cdc.MustMarshalJSON(NewQueryValidatorSlashesParams(validatorAddr, startHeight, endHeight)),
	}

	bz, err := querier(ctx, []string{QueryValidatorSlashes}, query)
	require.Nil(t, err)
	require.Nil(t, cdc.UnmarshalJSON(bz, &slashes))

	return
}

func getQueriedDelegationRewards(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) (rewards sdk.DecCoins) {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, QueryDelegationRewards}, "/"),
		Data: cdc.MustMarshalJSON(NewQueryDelegationRewardsParams(delegatorAddr, validatorAddr)),
	}

	bz, err := querier(ctx, []string{QueryDelegationRewards}, query)
	require.Nil(t, err)
	require.Nil(t, cdc.UnmarshalJSON(bz, &rewards))

	return
}

func getQueriedDelegatorTotalRewards(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, delegatorAddr sdk.AccAddress) (response types.QueryDelegatorTotalRewardsResponse) {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, QueryDelegatorTotalRewards}, "/"),
		Data: cdc.MustMarshalJSON(NewQueryDelegatorParams(delegatorAddr)),
	}

	bz, err := querier(ctx, []string{QueryDelegatorTotalRewards}, query)
	require.Nil(t, err)
	require.Nil(t, cdc.UnmarshalJSON(bz, &response))

	return
}

func getQueriedCommunityPool(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier) (ptr []byte) {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, QueryCommunityPool}, ""),
		Data: []byte{},
	}

	cp, err := querier(ctx, []string{QueryCommunityPool}, query)
	require.Nil(t, err)
	require.Nil(t, cdc.UnmarshalJSON(cp, &ptr))

	return
}

func TestQueries(t *testing.T) {
	cdc := codec.New()
	types.RegisterCodec(cdc)
	ctx, _, keeper, sk, _ := CreateTestInputDefault(t, false, 100)
	querier := NewQuerier(keeper)

	// test param queries
	communityTax := sdk.NewDecWithPrec(3, 1)
	baseProposerReward := sdk.NewDecWithPrec(2, 1)
	bonusProposerReward := sdk.NewDecWithPrec(1, 1)
	withdrawAddrEnabled := true
	keeper.SetCommunityTax(ctx, communityTax)
	keeper.SetBaseProposerReward(ctx, baseProposerReward)
	keeper.SetBonusProposerReward(ctx, bonusProposerReward)
	keeper.SetWithdrawAddrEnabled(ctx, withdrawAddrEnabled)
	retCommunityTax, retBaseProposerReward, retBonusProposerReward, retWithdrawAddrEnabled := getQueriedParams(t, ctx, cdc, querier)
	require.Equal(t, communityTax, retCommunityTax)
	require.Equal(t, baseProposerReward, retBaseProposerReward)
	require.Equal(t, bonusProposerReward, retBonusProposerReward)
	require.Equal(t, withdrawAddrEnabled, retWithdrawAddrEnabled)

	// test outstanding rewards query
	outstandingRewards := sdk.DecCoins{{"mytoken", sdk.NewDec(3)}, {"myothertoken", sdk.NewDecWithPrec(3, 7)}}
	keeper.SetValidatorOutstandingRewards(ctx, valOpAddr1, outstandingRewards)
	retOutstandingRewards := getQueriedValidatorOutstandingRewards(t, ctx, cdc, querier, valOpAddr1)
	require.Equal(t, outstandingRewards, retOutstandingRewards)

	// test validator commission query
	commission := sdk.DecCoins{{"token1", sdk.NewDec(4)}, {"token2", sdk.NewDec(2)}}
	keeper.SetValidatorAccumulatedCommission(ctx, valOpAddr1, commission)
	retCommission := getQueriedValidatorCommission(t, ctx, cdc, querier, valOpAddr1)
	require.Equal(t, commission, retCommission)

	// test delegator's total rewards query
	delRewards := getQueriedDelegatorTotalRewards(t, ctx, cdc, querier, sdk.AccAddress(valOpAddr1))
	require.Equal(t, types.QueryDelegatorTotalRewardsResponse{}, delRewards)

	// test validator slashes query with height range
	slashOne := types.NewValidatorSlashEvent(3, sdk.NewDecWithPrec(5, 1))
	slashTwo := types.NewValidatorSlashEvent(7, sdk.NewDecWithPrec(6, 1))
	keeper.SetValidatorSlashEvent(ctx, valOpAddr1, 3, slashOne)
	keeper.SetValidatorSlashEvent(ctx, valOpAddr1, 7, slashTwo)
	slashes := getQueriedValidatorSlashes(t, ctx, cdc, querier, valOpAddr1, 0, 2)
	require.Equal(t, 0, len(slashes))
	slashes = getQueriedValidatorSlashes(t, ctx, cdc, querier, valOpAddr1, 0, 5)
	require.Equal(t, []types.ValidatorSlashEvent{slashOne}, slashes)
	slashes = getQueriedValidatorSlashes(t, ctx, cdc, querier, valOpAddr1, 0, 10)
	require.Equal(t, []types.ValidatorSlashEvent{slashOne, slashTwo}, slashes)

	// test delegation rewards query
	sh := staking.NewHandler(sk)
	comm := staking.NewCommissionMsg(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, comm, sdk.OneInt())
	require.True(t, sh(ctx, msg).IsOK())
	staking.EndBlocker(ctx, sk)
	val := sk.Validator(ctx, valOpAddr1)
	rewards := getQueriedDelegationRewards(t, ctx, cdc, querier, sdk.AccAddress(valOpAddr1), valOpAddr1)
	require.True(t, rewards.IsZero())
	initial := int64(10)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	tokens := sdk.DecCoins{{sdk.DefaultBondDenom, sdk.NewDec(initial)}}
	keeper.AllocateTokensToValidator(ctx, val, tokens)
	rewards = getQueriedDelegationRewards(t, ctx, cdc, querier, sdk.AccAddress(valOpAddr1), valOpAddr1)
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, sdk.NewDec(initial / 2)}}, rewards)

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
