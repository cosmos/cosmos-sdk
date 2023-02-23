package keeper_test

import (
	gocontext "context"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/testutil"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type fixture struct {
	ctx         sdk.Context
	queryClient types.QueryClient
	addrs       []sdk.AccAddress
	valAddrs    []sdk.ValAddress

	interfaceRegistry codectypes.InterfaceRegistry
	bankKeeper        bankkeeper.Keeper
	distrKeeper       keeper.Keeper
	stakingKeeper     *stakingkeeper.Keeper
	msgServer         types.MsgServer
}

func initFixture(t assert.TestingT) *fixture {
	f := &fixture{}

	app, err := simtestutil.Setup(
		testutil.AppConfig,
		&f.interfaceRegistry,
		&f.bankKeeper,
		&f.distrKeeper,
		&f.stakingKeeper,
	)
	assert.NilError(t, err)

	ctx := app.BaseApp.NewContext(false, cmtproto.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, f.interfaceRegistry)
	types.RegisterQueryServer(queryHelper, keeper.NewQuerier(f.distrKeeper))
	queryClient := types.NewQueryClient(queryHelper)

	f.ctx = ctx
	f.queryClient = queryClient

	f.addrs = simtestutil.AddTestAddrs(f.bankKeeper, f.stakingKeeper, ctx, 2, sdk.NewInt(1000000000))
	f.valAddrs = simtestutil.ConvertAddrsToValAddrs(f.addrs)
	f.msgServer = keeper.NewMsgServerImpl(f.distrKeeper)

	return f
}

func TestGRPCParams(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx, queryClient := f.ctx, f.queryClient

	var (
		params    types.Params
		req       *types.QueryParamsRequest
		expParams types.Params
	)

	testCases := []struct {
		msg      string
		malleate func()
	}{
		{
			"empty params request",
			func() {
				req = &types.QueryParamsRequest{}
				expParams = types.DefaultParams()
			},
		},
		{
			"valid request",
			func() {
				params = types.Params{
					CommunityTax:        sdk.NewDecWithPrec(3, 1),
					BaseProposerReward:  sdk.ZeroDec(),
					BonusProposerReward: sdk.ZeroDec(),
					WithdrawAddrEnabled: true,
				}

				assert.NilError(t, f.distrKeeper.SetParams(ctx, params))
				req = &types.QueryParamsRequest{}
				expParams = params
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Case %s", testCase.msg), func(t *testing.T) {
			testCase.malleate()

			paramsRes, err := queryClient.Params(gocontext.Background(), req)

			assert.NilError(t, err)
			assert.Assert(t, paramsRes != nil)
			assert.DeepEqual(t, expParams, paramsRes.Params)
		})
	}
}

func TestGRPCValidatorOutstandingRewards(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx, queryClient, valAddrs := f.ctx, f.queryClient, f.valAddrs

	valCommission := sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", math.LegacyNewDec(5000)),
		sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(300)),
	}

	// set outstanding rewards
	f.distrKeeper.SetValidatorOutstandingRewards(ctx, valAddrs[0], types.ValidatorOutstandingRewards{Rewards: valCommission})
	rewards := f.distrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[0])

	var req *types.QueryValidatorOutstandingRewardsRequest

	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		expErrMsg string
	}{
		{
			"empty request",
			func() {
				req = &types.QueryValidatorOutstandingRewardsRequest{}
			},
			false,
			"empty validator address",
		}, {
			"valid request",
			func() {
				req = &types.QueryValidatorOutstandingRewardsRequest{ValidatorAddress: valAddrs[0].String()}
			},
			true,
			"",
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Case %s", testCase.msg), func(t *testing.T) {
			testCase.malleate()

			validatorOutstandingRewards, err := queryClient.ValidatorOutstandingRewards(gocontext.Background(), req)

			if testCase.expPass {
				assert.NilError(t, err)
				assert.DeepEqual(t, rewards, validatorOutstandingRewards.Rewards)
				assert.DeepEqual(t, valCommission, validatorOutstandingRewards.Rewards.Rewards)
			} else {
				assert.ErrorContains(t, err, testCase.expErrMsg)
				assert.Assert(t, validatorOutstandingRewards == nil)
			}
		})
	}
}

func TestGRPCValidatorCommission(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx, queryClient, valAddrs := f.ctx, f.queryClient, f.valAddrs

	commission := sdk.DecCoins{{Denom: "token1", Amount: math.LegacyNewDec(4)}, {Denom: "token2", Amount: math.LegacyNewDec(2)}}
	f.distrKeeper.SetValidatorAccumulatedCommission(ctx, valAddrs[0], types.ValidatorAccumulatedCommission{Commission: commission})

	var req *types.QueryValidatorCommissionRequest

	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		expErrMsg string
	}{
		{
			"empty request",
			func() {
				req = &types.QueryValidatorCommissionRequest{}
			},
			false,
			"empty validator address",
		},
		{
			"valid request",
			func() {
				req = &types.QueryValidatorCommissionRequest{ValidatorAddress: valAddrs[0].String()}
			},
			true,
			"",
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Case %s", testCase.msg), func(t *testing.T) {
			testCase.malleate()

			commissionRes, err := queryClient.ValidatorCommission(gocontext.Background(), req)

			if testCase.expPass {
				assert.NilError(t, err)
				assert.Assert(t, commissionRes != nil)
				assert.DeepEqual(t, commissionRes.Commission.Commission, commission)
			} else {
				assert.ErrorContains(t, err, testCase.expErrMsg)
				assert.Assert(t, commissionRes == nil)
			}
		})
	}
}

func TestGRPCValidatorSlashes(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx, queryClient, valAddrs := f.ctx, f.queryClient, f.valAddrs

	slashes := []types.ValidatorSlashEvent{
		types.NewValidatorSlashEvent(3, sdk.NewDecWithPrec(5, 1)),
		types.NewValidatorSlashEvent(5, sdk.NewDecWithPrec(5, 1)),
		types.NewValidatorSlashEvent(7, sdk.NewDecWithPrec(5, 1)),
		types.NewValidatorSlashEvent(9, sdk.NewDecWithPrec(5, 1)),
	}

	for i, slash := range slashes {
		f.distrKeeper.SetValidatorSlashEvent(ctx, valAddrs[0], uint64(i+2), 0, slash)
	}

	var (
		req    *types.QueryValidatorSlashesRequest
		expRes *types.QueryValidatorSlashesResponse
	)

	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		expErrMsg string
	}{
		{
			"empty request",
			func() {
				req = &types.QueryValidatorSlashesRequest{}
				expRes = &types.QueryValidatorSlashesResponse{}
			},
			false,
			"empty validator address",
		},
		{
			"Ending height lesser than start height request",
			func() {
				req = &types.QueryValidatorSlashesRequest{
					ValidatorAddress: valAddrs[1].String(),
					StartingHeight:   10,
					EndingHeight:     1,
				}
				expRes = &types.QueryValidatorSlashesResponse{Pagination: &query.PageResponse{}}
			},
			false,
			"starting height greater than ending height",
		},
		{
			"no slash event validator request",
			func() {
				req = &types.QueryValidatorSlashesRequest{
					ValidatorAddress: valAddrs[1].String(),
					StartingHeight:   1,
					EndingHeight:     10,
				}
				expRes = &types.QueryValidatorSlashesResponse{Pagination: &query.PageResponse{}}
			},
			true,
			"",
		},
		{
			"request slashes with offset 2 and limit 2",
			func() {
				pageReq := &query.PageRequest{
					Offset: 2,
					Limit:  2,
				}

				req = &types.QueryValidatorSlashesRequest{
					ValidatorAddress: valAddrs[0].String(),
					StartingHeight:   1,
					EndingHeight:     10,
					Pagination:       pageReq,
				}

				expRes = &types.QueryValidatorSlashesResponse{
					Slashes: slashes[2:],
				}
			},
			true,
			"",
		},
		{
			"request slashes with page limit 3 and count total",
			func() {
				pageReq := &query.PageRequest{
					Limit:      3,
					CountTotal: true,
				}

				req = &types.QueryValidatorSlashesRequest{
					ValidatorAddress: valAddrs[0].String(),
					StartingHeight:   1,
					EndingHeight:     10,
					Pagination:       pageReq,
				}

				expRes = &types.QueryValidatorSlashesResponse{
					Slashes: slashes[:3],
				}
			},
			true,
			"",
		},
		{
			"request slashes with page limit 4 and count total",
			func() {
				pageReq := &query.PageRequest{
					Limit:      4,
					CountTotal: true,
				}

				req = &types.QueryValidatorSlashesRequest{
					ValidatorAddress: valAddrs[0].String(),
					StartingHeight:   1,
					EndingHeight:     10,
					Pagination:       pageReq,
				}

				expRes = &types.QueryValidatorSlashesResponse{
					Slashes: slashes[:4],
				}
			},
			true,
			"",
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Case %s", testCase.msg), func(t *testing.T) {
			testCase.malleate()

			slashesRes, err := queryClient.ValidatorSlashes(gocontext.Background(), req)

			if testCase.expPass {
				assert.NilError(t, err)
				assert.DeepEqual(t, expRes.GetSlashes(), slashesRes.GetSlashes())
			} else {
				assert.ErrorContains(t, err, testCase.expErrMsg)
				assert.Assert(t, slashesRes == nil)
			}
		})
	}
}

func TestGRPCDelegationRewards(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx, addrs, valAddrs := f.ctx, f.addrs, f.valAddrs

	tstaking := stakingtestutil.NewHelper(t, ctx, f.stakingKeeper)
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), math.LegacyNewDec(0))
	tstaking.CreateValidator(valAddrs[0], valConsPk0, sdk.NewInt(100), true)

	staking.EndBlocker(ctx, f.stakingKeeper)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, f.interfaceRegistry)
	types.RegisterQueryServer(queryHelper, keeper.NewQuerier(f.distrKeeper))
	queryClient := types.NewQueryClient(queryHelper)

	val := f.stakingKeeper.Validator(ctx, valAddrs[0])

	initial := int64(10)
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial)}}
	f.distrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// test command delegation rewards grpc
	var (
		req    *types.QueryDelegationRewardsRequest
		expRes *types.QueryDelegationRewardsResponse
	)

	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		expErrMsg string
	}{
		{
			"empty request",
			func() {
				req = &types.QueryDelegationRewardsRequest{}
			},
			false,
			"empty delegator address",
		},
		{
			"empty delegator request",
			func() {
				req = &types.QueryDelegationRewardsRequest{
					DelegatorAddress: "",
					ValidatorAddress: valAddrs[0].String(),
				}
			},
			false,
			"empty delegator address",
		},
		{
			"empty validator request",
			func() {
				req = &types.QueryDelegationRewardsRequest{
					DelegatorAddress: addrs[1].String(),
					ValidatorAddress: "",
				}
			},
			false,
			"empty validator address",
		},
		{
			"request with wrong delegator and validator",
			func() {
				req = &types.QueryDelegationRewardsRequest{
					DelegatorAddress: addrs[1].String(),
					ValidatorAddress: valAddrs[1].String(),
				}
			},
			false,
			"validator does not exist",
		},
		{
			"valid request",
			func() {
				req = &types.QueryDelegationRewardsRequest{
					DelegatorAddress: addrs[0].String(),
					ValidatorAddress: valAddrs[0].String(),
				}

				expRes = &types.QueryDelegationRewardsResponse{
					Rewards: sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial / 2)}},
				}
			},
			true,
			"",
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Case %s", testCase.msg), func(t *testing.T) {
			testCase.malleate()

			rewards, err := queryClient.DelegationRewards(gocontext.Background(), req)

			if testCase.expPass {
				assert.NilError(t, err)
				assert.DeepEqual(t, expRes, rewards)
			} else {
				assert.ErrorContains(t, err, testCase.expErrMsg)
				assert.Assert(t, rewards == nil)
			}
		})
	}

	// test command delegator total rewards grpc
	var (
		totalRewardsReq    *types.QueryDelegationTotalRewardsRequest
		expTotalRewardsRes *types.QueryDelegationTotalRewardsResponse
	)

	testCases = []struct {
		msg       string
		malleate  func()
		expPass   bool
		expErrMsg string
	}{
		{
			"empty request",
			func() {
				totalRewardsReq = &types.QueryDelegationTotalRewardsRequest{}
			},
			false,
			"empty delegator address",
		},
		{
			"valid total delegation rewards",
			func() {
				totalRewardsReq = &types.QueryDelegationTotalRewardsRequest{
					DelegatorAddress: addrs[0].String(),
				}

				expectedDelReward := types.NewDelegationDelegatorReward(valAddrs[0],
					sdk.DecCoins{sdk.NewInt64DecCoin("stake", 5)})

				expTotalRewardsRes = &types.QueryDelegationTotalRewardsResponse{
					Rewards: []types.DelegationDelegatorReward{expectedDelReward},
					Total:   expectedDelReward.Reward,
				}
			},
			true,
			"",
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Case %s", testCase.msg), func(t *testing.T) {
			testCase.malleate()

			totalRewardsRes, err := queryClient.DelegationTotalRewards(gocontext.Background(), totalRewardsReq)

			if testCase.expPass {
				assert.NilError(t, err)
				assert.DeepEqual(t, totalRewardsRes, expTotalRewardsRes)
			} else {
				assert.ErrorContains(t, err, testCase.expErrMsg)
				assert.Assert(t, totalRewardsRes == nil)
			}
		})
	}

	// test command validator delegators grpc
	var (
		delegatorValidatorsReq    *types.QueryDelegatorValidatorsRequest
		expDelegatorValidatorsRes *types.QueryDelegatorValidatorsResponse
	)

	testCases = []struct {
		msg       string
		malleate  func()
		expPass   bool
		expErrMsg string
	}{
		{
			"empty request",
			func() {
				delegatorValidatorsReq = &types.QueryDelegatorValidatorsRequest{}
			},
			false,
			"empty delegator address",
		},
		{
			"request no delegations address",
			func() {
				delegatorValidatorsReq = &types.QueryDelegatorValidatorsRequest{
					DelegatorAddress: addrs[1].String(),
				}

				expDelegatorValidatorsRes = &types.QueryDelegatorValidatorsResponse{}
			},
			true,
			"",
		},
		{
			"valid request",
			func() {
				delegatorValidatorsReq = &types.QueryDelegatorValidatorsRequest{
					DelegatorAddress: addrs[0].String(),
				}
				expDelegatorValidatorsRes = &types.QueryDelegatorValidatorsResponse{
					Validators: []string{valAddrs[0].String()},
				}
			},
			true,
			"",
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Case %s", testCase.msg), func(t *testing.T) {
			testCase.malleate()

			validators, err := queryClient.DelegatorValidators(gocontext.Background(), delegatorValidatorsReq)

			if testCase.expPass {
				assert.NilError(t, err)
				assert.DeepEqual(t, expDelegatorValidatorsRes, validators)
			} else {
				assert.ErrorContains(t, err, testCase.expErrMsg)
				assert.Assert(t, validators == nil)
			}
		})
	}
}

func TestGRPCDelegatorWithdrawAddress(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx, queryClient, addrs := f.ctx, f.queryClient, f.addrs

	err := f.distrKeeper.SetWithdrawAddr(ctx, addrs[0], addrs[1])
	assert.Assert(t, err == nil)

	var req *types.QueryDelegatorWithdrawAddressRequest

	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		expErrMsg string
	}{
		{
			"empty request",
			func() {
				req = &types.QueryDelegatorWithdrawAddressRequest{}
			},
			false,
			"empty delegator address",
		},
		{
			"valid request",
			func() {
				req = &types.QueryDelegatorWithdrawAddressRequest{DelegatorAddress: addrs[0].String()}
			},
			true,
			"",
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Case %s", testCase.msg), func(t *testing.T) {
			testCase.malleate()

			withdrawAddress, err := queryClient.DelegatorWithdrawAddress(gocontext.Background(), req)

			if testCase.expPass {
				assert.NilError(t, err)
				assert.Equal(t, withdrawAddress.WithdrawAddress, addrs[1].String())
			} else {
				assert.ErrorContains(t, err, testCase.expErrMsg)
				assert.Assert(t, withdrawAddress == nil)
			}
		})
	}
}

func TestGRPCCommunityPool(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx, queryClient, addrs := f.ctx, f.queryClient, f.addrs
	// reset fee pool
	f.distrKeeper.SetFeePool(ctx, types.InitialFeePool())

	var (
		req     *types.QueryCommunityPoolRequest
		expPool *types.QueryCommunityPoolResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
	}{
		{
			"valid request empty community pool",
			func() {
				req = &types.QueryCommunityPoolRequest{}
				expPool = &types.QueryCommunityPoolResponse{}
			},
		},
		{
			"valid request",
			func() {
				amount := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
				assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addrs[0], amount))

				err := f.distrKeeper.FundCommunityPool(ctx, amount, addrs[0])
				assert.Assert(t, err == nil)
				req = &types.QueryCommunityPoolRequest{}

				expPool = &types.QueryCommunityPoolResponse{Pool: sdk.NewDecCoinsFromCoins(amount...)}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Case %s", testCase.msg), func(t *testing.T) {
			testCase.malleate()

			pool, err := queryClient.CommunityPool(gocontext.Background(), req)

			assert.NilError(t, err)
			assert.DeepEqual(t, expPool, pool)
		})
	}
}
