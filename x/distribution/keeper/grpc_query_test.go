package keeper_test

import (
	gocontext "context"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestGRPCParams(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.DistrKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	params := types.Params{
		CommunityTax:        sdk.NewDecWithPrec(3, 1),
		BaseProposerReward:  sdk.NewDecWithPrec(2, 1),
		BonusProposerReward: sdk.NewDecWithPrec(1, 1),
		WithdrawAddrEnabled: true,
	}

	app.DistrKeeper.SetParams(ctx, params)

	paramsRes, err := queryClient.Params(gocontext.Background(), &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, paramsRes.Params, params)
}

func TestGRPCValidatorOutstandingRewards(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.DistrKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	valCommission := sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", sdk.NewDec(5000)),
		sdk.NewDecCoinFromDec("stake", sdk.NewDec(300)),
	}

	addr := simapp.AddTestAddrs(app, ctx, 1, sdk.NewInt(1000000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addr)

	pageReq := &query.PageRequest{Limit: 1}

	req := types.NewQueryValidatorOutstandingRewardsRequest(nil, pageReq)

	validatorOutstandingRewards, err := queryClient.ValidatorOutstandingRewards(gocontext.Background(), req)
	require.Error(t, err)
	require.Nil(t, validatorOutstandingRewards)

	// set outstanding rewards
	app.DistrKeeper.SetValidatorOutstandingRewards(ctx, valAddrs[0], types.ValidatorOutstandingRewards{Rewards: valCommission})
	rewards := app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[0])
	req = types.NewQueryValidatorOutstandingRewardsRequest(valAddrs[0], pageReq)

	validatorOutstandingRewards, err = queryClient.ValidatorOutstandingRewards(gocontext.Background(), req)
	require.NoError(t, err)
	require.Equal(t, rewards, validatorOutstandingRewards.Rewards)
	require.Equal(t, valCommission, validatorOutstandingRewards.Rewards.Rewards)
}

func TestGRPCValidatorCommission(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.DistrKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	addr := simapp.AddTestAddrs(app, ctx, 1, sdk.NewInt(1000000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addr)
	valOpAddr1 := valAddrs[0]

	commission := sdk.DecCoins{{Denom: "token1", Amount: sdk.NewDec(4)}, {Denom: "token2", Amount: sdk.NewDec(2)}}
	app.DistrKeeper.SetValidatorAccumulatedCommission(ctx, valOpAddr1, types.ValidatorAccumulatedCommission{Commission: commission})

	req := &types.QueryValidatorCommissionRequest{ValidatorAddress: valOpAddr1}

	commissionRes, err := queryClient.ValidatorCommission(gocontext.Background(), req)
	require.NoError(t, err)
	require.Equal(t, commissionRes.Commission.Commission, commission)
}

func TestGRPCValidatorSlashes(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.DistrKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	addr := simapp.AddTestAddrs(app, ctx, 1, sdk.NewInt(1000000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addr)
	valOpAddr1 := valAddrs[0]

	slashOne := types.NewValidatorSlashEvent(3, sdk.NewDecWithPrec(5, 1))
	slashTwo := types.NewValidatorSlashEvent(7, sdk.NewDecWithPrec(6, 1))

	app.DistrKeeper.SetValidatorSlashEvent(ctx, valOpAddr1, 3, 0, slashOne)
	app.DistrKeeper.SetValidatorSlashEvent(ctx, valOpAddr1, 7, 0, slashTwo)

	pageReq := &query.PageRequest{
		Limit: 1,
	}

	req := &types.QueryValidatorSlashesRequest{
		ValidatorAddress: valOpAddr1,
		StartingHeight:   1,
		EndingHeight:     10,
		Req:              pageReq,
	}

	slashes, err := queryClient.ValidatorSlashes(gocontext.Background(), req)

	require.NoError(t, err)
	require.Len(t, slashes.Slashes, 1)
	require.Equal(t, slashOne, slashes.Slashes[0])
	require.NotNil(t, slashes.Res.NextKey)

	pageReq.Key = slashes.Res.NextKey
	req.Req = pageReq

	slashes, err = queryClient.ValidatorSlashes(gocontext.Background(), req)
	require.NoError(t, err)
	require.Len(t, slashes.Slashes, 1)
	require.Equal(t, slashTwo, slashes.Slashes[0])
	require.Empty(t, slashes.Res)

	pageReq = &query.PageRequest{Limit: 2}
	req.Req = pageReq

	slashes, err = queryClient.ValidatorSlashes(gocontext.Background(), req)
	require.NoError(t, err)
	require.Len(t, slashes.Slashes, 2)
	require.Equal(t, slashOne, slashes.Slashes[0])
	require.Equal(t, slashTwo, slashes.Slashes[1])
	require.Empty(t, slashes.Res)
}

func TestGRPCDelegationRewards(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.DistrKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	addr := simapp.AddTestAddrs(app, ctx, 1, sdk.NewInt(1000000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addr)
	valOpAddr1 := valAddrs[0]

	sh := staking.NewHandler(app.StakingKeeper)
	comm := stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := stakingtypes.NewMsgCreateValidator(
		valOpAddr1, valConsPk1, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), stakingtypes.Description{}, comm, sdk.OneInt(),
	)

	res, err := sh(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	staking.EndBlocker(ctx, app.StakingKeeper)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	val := app.StakingKeeper.Validator(ctx, valOpAddr1)

	initial := int64(10)
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial)}}
	app.DistrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	req := &types.QueryDelegationRewardsRequest{
		DelegatorAddress: sdk.AccAddress(valOpAddr1),
		ValidatorAddress: valOpAddr1,
	}

	_, err = queryClient.DelegationRewards(gocontext.Background(), req)
	require.NoError(t, err)
	// TODO debug delegation rewards
	// require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial / 2)}}, rewards.Rewards)

	// totalRewardsReq := &types.QueryDelegationTotalRewardsRequest{
	// 	DelegatorAddress: sdk.AccAddress(valOpAddr1),
	// }

	// totalRewards, err := queryClient.DelegationTotalRewards(gocontext.Background(), totalRewardsReq)
	// require.NoError(t, err)
	// expectedDelReward := types.NewDelegationDelegatorReward(valOpAddr1,
	// 	sdk.DecCoins{sdk.NewInt64DecCoin("stake", 5)})
	// wantDelRewards := types.NewQueryDelegatorTotalRewardsResponse(
	// 	[]types.DelegationDelegatorReward{expectedDelReward}, expectedDelReward.Reward)
	// require.Equal(t, wantDelRewards, totalRewards)

	validators, err := queryClient.DelegatorValidators(gocontext.Background(),
		&types.QueryDelegatorValidatorsRequest{DelegatorAddress: sdk.AccAddress(valOpAddr1)})
	require.NoError(t, err)
	require.Len(t, validators.Validators, 1)
	require.Equal(t, validators.Validators[0], valOpAddr1)
}

func TestGRPCDelegatorWithdrawAddress(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	addr := simapp.AddTestAddrs(app, ctx, 2, sdk.NewInt(1000000000))

	err := app.DistrKeeper.SetWithdrawAddr(ctx, addr[0], addr[1])
	require.Nil(t, err)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.DistrKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	// queryClient.ValidatorCommission()

	_, err = queryClient.DelegatorWithdrawAddress(gocontext.Background(), &types.QueryDelegatorWithdrawAddressRequest{})
	require.Error(t, err)

	withdrawAddress, err := queryClient.DelegatorWithdrawAddress(gocontext.Background(),
		&types.QueryDelegatorWithdrawAddressRequest{DelegatorAddress: addr[0]})
	require.NoError(t, err)
	require.Equal(t, withdrawAddress.WithdrawAddress, addr[1])
}

func TestGRPCCommunityPool(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.DistrKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	pool, err := queryClient.CommunityPool(gocontext.Background(), &types.QueryCommunityPoolRequest{})
	require.NoError(t, err)
	require.Empty(t, pool)
}
