package keeper_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const custom = "custom"

func getQueriedParams(t *testing.T, ctx sdk.Context, cdc *codec.LegacyAmino, querier sdk.Querier) types.Params {
	var params types.Params

	bz, err := querier(ctx, []string{types.QueryParams}, abci.RequestQuery{})
	require.Nil(t, err)
	require.Nil(t, cdc.UnmarshalJSON(bz, &params))

	return params
}

func getQueriedValidatorOutstandingRewards(t *testing.T, ctx sdk.Context, cdc *codec.LegacyAmino, querier sdk.Querier, validatorAddr sdk.ValAddress) sdk.DecCoins {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryValidatorOutstandingRewards}, "/"),
		Data: cdc.MustMarshalJSON(types.NewQueryValidatorOutstandingRewardsParams(validatorAddr)),
	}

	bz, err := querier(ctx, []string{types.QueryValidatorOutstandingRewards}, query)
	require.Nil(t, err)
	outstandingRewards := types.ValidatorOutstandingRewards{}
	require.Nil(t, cdc.UnmarshalJSON(bz, &outstandingRewards))

	return outstandingRewards.GetRewards()
}

func getQueriedValidatorCommission(t *testing.T, ctx sdk.Context, cdc *codec.LegacyAmino, querier sdk.Querier, validatorAddr sdk.ValAddress) sdk.DecCoins {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryValidatorCommission}, "/"),
		Data: cdc.MustMarshalJSON(types.NewQueryValidatorCommissionParams(validatorAddr)),
	}

	bz, err := querier(ctx, []string{types.QueryValidatorCommission}, query)
	require.Nil(t, err)
	validatorCommission := types.ValidatorAccumulatedCommission{}
	require.Nil(t, cdc.UnmarshalJSON(bz, &validatorCommission))

	return validatorCommission.GetCommission()
}

func getQueriedValidatorSlashes(t *testing.T, ctx sdk.Context, cdc *codec.LegacyAmino, querier sdk.Querier, validatorAddr sdk.ValAddress, startHeight uint64, endHeight uint64) (slashes []types.ValidatorSlashEvent) {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryValidatorSlashes}, "/"),
		Data: cdc.MustMarshalJSON(types.NewQueryValidatorSlashesParams(validatorAddr, startHeight, endHeight)),
	}

	bz, err := querier(ctx, []string{types.QueryValidatorSlashes}, query)
	require.Nil(t, err)
	require.Nil(t, cdc.UnmarshalJSON(bz, &slashes))

	return
}

func getQueriedDelegationRewards(t *testing.T, ctx sdk.Context, cdc *codec.LegacyAmino, querier sdk.Querier, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) (rewards sdk.DecCoins) {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryDelegationRewards}, "/"),
		Data: cdc.MustMarshalJSON(types.NewQueryDelegationRewardsParams(delegatorAddr, validatorAddr)),
	}

	bz, err := querier(ctx, []string{types.QueryDelegationRewards}, query)
	require.Nil(t, err)
	require.Nil(t, cdc.UnmarshalJSON(bz, &rewards))

	return
}

func getQueriedDelegatorTotalRewards(t *testing.T, ctx sdk.Context, cdc *codec.LegacyAmino, querier sdk.Querier, delegatorAddr sdk.AccAddress) (response types.QueryDelegatorTotalRewardsResponse) {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryDelegatorTotalRewards}, "/"),
		Data: cdc.MustMarshalJSON(types.NewQueryDelegatorParams(delegatorAddr)),
	}

	bz, err := querier(ctx, []string{types.QueryDelegatorTotalRewards}, query)
	require.Nil(t, err)
	require.Nil(t, cdc.UnmarshalJSON(bz, &response))

	return
}

func getQueriedCommunityPool(t *testing.T, ctx sdk.Context, cdc *codec.LegacyAmino, querier sdk.Querier) (ptr []byte) {
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
	cdc := codec.NewLegacyAmino()
	types.RegisterLegacyAminoCodec(cdc)
	banktypes.RegisterLegacyAminoCodec(cdc)

	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addr := simapp.AddTestAddrs(app, ctx, 1, sdk.NewInt(1000000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addr)
	valOpAddr1 := valAddrs[0]

	querier := keeper.NewQuerier(app.DistrKeeper, cdc)

	// test param queries
	params := types.Params{
		CommunityTax:        sdk.NewDecWithPrec(3, 1),
		BaseProposerReward:  sdk.NewDecWithPrec(2, 1),
		BonusProposerReward: sdk.NewDecWithPrec(1, 1),
		WithdrawAddrEnabled: true,
	}

	app.DistrKeeper.SetParams(ctx, params)

	paramsRes := getQueriedParams(t, ctx, cdc, querier)
	require.Equal(t, params.CommunityTax, paramsRes.CommunityTax)
	require.Equal(t, params.BaseProposerReward, paramsRes.BaseProposerReward)
	require.Equal(t, params.BonusProposerReward, paramsRes.BonusProposerReward)
	require.Equal(t, params.WithdrawAddrEnabled, paramsRes.WithdrawAddrEnabled)

	// test outstanding rewards query
	outstandingRewards := sdk.DecCoins{{Denom: "mytoken", Amount: sdk.NewDec(3)}, {Denom: "myothertoken", Amount: sdk.NewDecWithPrec(3, 7)}}
	app.DistrKeeper.SetValidatorOutstandingRewards(ctx, valOpAddr1, types.ValidatorOutstandingRewards{Rewards: outstandingRewards})
	retOutstandingRewards := getQueriedValidatorOutstandingRewards(t, ctx, cdc, querier, valOpAddr1)
	require.Equal(t, outstandingRewards, retOutstandingRewards)

	// test validator commission query
	commission := sdk.DecCoins{{Denom: "token1", Amount: sdk.NewDec(4)}, {Denom: "token2", Amount: sdk.NewDec(2)}}
	app.DistrKeeper.SetValidatorAccumulatedCommission(ctx, valOpAddr1, types.ValidatorAccumulatedCommission{Commission: commission})
	retCommission := getQueriedValidatorCommission(t, ctx, cdc, querier, valOpAddr1)
	require.Equal(t, commission, retCommission)

	// test delegator's total rewards query
	delRewards := getQueriedDelegatorTotalRewards(t, ctx, cdc, querier, sdk.AccAddress(valOpAddr1))
	require.Equal(t, types.QueryDelegatorTotalRewardsResponse{}, delRewards)

	// test validator slashes query with height range
	slashOne := types.NewValidatorSlashEvent(3, sdk.NewDecWithPrec(5, 1))
	slashTwo := types.NewValidatorSlashEvent(7, sdk.NewDecWithPrec(6, 1))
	app.DistrKeeper.SetValidatorSlashEvent(ctx, valOpAddr1, 3, 0, slashOne)
	app.DistrKeeper.SetValidatorSlashEvent(ctx, valOpAddr1, 7, 0, slashTwo)
	slashes := getQueriedValidatorSlashes(t, ctx, cdc, querier, valOpAddr1, 0, 2)
	require.Equal(t, 0, len(slashes))
	slashes = getQueriedValidatorSlashes(t, ctx, cdc, querier, valOpAddr1, 0, 5)
	require.Equal(t, []types.ValidatorSlashEvent{slashOne}, slashes)
	slashes = getQueriedValidatorSlashes(t, ctx, cdc, querier, valOpAddr1, 0, 10)
	require.Equal(t, []types.ValidatorSlashEvent{slashOne, slashTwo}, slashes)

	// test delegation rewards query
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	tstaking.CreateValidator(valOpAddr1, valConsPk1, sdk.NewInt(100), true)

	staking.EndBlocker(ctx, app.StakingKeeper)

	val := app.StakingKeeper.Validator(ctx, valOpAddr1)
	rewards := getQueriedDelegationRewards(t, ctx, cdc, querier, sdk.AccAddress(valOpAddr1), valOpAddr1)
	require.True(t, rewards.IsZero())
	initial := int64(10)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial)}}
	app.DistrKeeper.AllocateTokensToValidator(ctx, val, tokens)
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
