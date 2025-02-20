package keeper_test

import (
	gocontext "context"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/suite"

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

type KeeperTestSuite struct {
	suite.Suite

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

func (suite *KeeperTestSuite) SetupTest() {
	app, err := simtestutil.Setup(testutil.AppConfig,
		&suite.interfaceRegistry,
		&suite.bankKeeper,
		&suite.distrKeeper,
		&suite.stakingKeeper,
	)
	suite.NoError(err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, suite.interfaceRegistry)
	types.RegisterQueryServer(queryHelper, keeper.NewQuerier(suite.distrKeeper))
	queryClient := types.NewQueryClient(queryHelper)

	suite.ctx = ctx
	suite.queryClient = queryClient

	suite.addrs = simtestutil.AddTestAddrs(suite.bankKeeper, suite.stakingKeeper, ctx, 2, sdk.NewInt(1000000000))
	suite.valAddrs = simtestutil.ConvertAddrsToValAddrs(suite.addrs)
	suite.msgServer = keeper.NewMsgServerImpl(suite.distrKeeper)
}

func (suite *KeeperTestSuite) TestGRPCParams() {
	ctx, queryClient := suite.ctx, suite.queryClient

	var (
		params    types.Params
		req       *types.QueryParamsRequest
		expParams types.Params
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty params request",
			func() {
				req = &types.QueryParamsRequest{}
				expParams = types.DefaultParams()
			},
			true,
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

				suite.NoError(suite.distrKeeper.SetParams(ctx, params))
				req = &types.QueryParamsRequest{}
				expParams = params
			},
			true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.msg), func() {
			testCase.malleate()

			paramsRes, err := queryClient.Params(gocontext.Background(), req)

			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(paramsRes)
				suite.Require().Equal(expParams, paramsRes.Params)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCValidatorOutstandingRewards() {
	ctx, queryClient, valAddrs := suite.ctx, suite.queryClient, suite.valAddrs

	valCommission := sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", math.LegacyNewDec(5000)),
		sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(300)),
	}

	// set outstanding rewards
	suite.distrKeeper.SetValidatorOutstandingRewards(ctx, valAddrs[0], types.ValidatorOutstandingRewards{Rewards: valCommission})
	rewards := suite.distrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[0])

	var req *types.QueryValidatorOutstandingRewardsRequest

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = &types.QueryValidatorOutstandingRewardsRequest{}
			},
			false,
		}, {
			"valid request",
			func() {
				req = &types.QueryValidatorOutstandingRewardsRequest{ValidatorAddress: valAddrs[0].String()}
			},
			true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.msg), func() {
			testCase.malleate()

			validatorOutstandingRewards, err := queryClient.ValidatorOutstandingRewards(gocontext.Background(), req)

			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(rewards, validatorOutstandingRewards.Rewards)
				suite.Require().Equal(valCommission, validatorOutstandingRewards.Rewards.Rewards)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(validatorOutstandingRewards)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCValidatorCommission() {
	ctx, queryClient, valAddrs := suite.ctx, suite.queryClient, suite.valAddrs

	commission := sdk.DecCoins{{Denom: "token1", Amount: math.LegacyNewDec(4)}, {Denom: "token2", Amount: math.LegacyNewDec(2)}}
	suite.distrKeeper.SetValidatorAccumulatedCommission(ctx, valAddrs[0], types.ValidatorAccumulatedCommission{Commission: commission})

	var req *types.QueryValidatorCommissionRequest

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = &types.QueryValidatorCommissionRequest{}
			},
			false,
		},
		{
			"valid request",
			func() {
				req = &types.QueryValidatorCommissionRequest{ValidatorAddress: valAddrs[0].String()}
			},
			true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.msg), func() {
			testCase.malleate()

			commissionRes, err := queryClient.ValidatorCommission(gocontext.Background(), req)

			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(commissionRes)
				suite.Require().Equal(commissionRes.Commission.Commission, commission)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(commissionRes)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCValidatorSlashes() {
	ctx, queryClient, valAddrs := suite.ctx, suite.queryClient, suite.valAddrs

	slashes := []types.ValidatorSlashEvent{
		types.NewValidatorSlashEvent(3, sdk.NewDecWithPrec(5, 1)),
		types.NewValidatorSlashEvent(5, sdk.NewDecWithPrec(5, 1)),
		types.NewValidatorSlashEvent(7, sdk.NewDecWithPrec(5, 1)),
		types.NewValidatorSlashEvent(9, sdk.NewDecWithPrec(5, 1)),
	}

	for i, slash := range slashes {
		suite.distrKeeper.SetValidatorSlashEvent(ctx, valAddrs[0], uint64(i+2), 0, slash)
	}

	var (
		req    *types.QueryValidatorSlashesRequest
		expRes *types.QueryValidatorSlashesResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = &types.QueryValidatorSlashesRequest{}
				expRes = &types.QueryValidatorSlashesResponse{}
			},
			false,
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
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.msg), func() {
			testCase.malleate()

			slashesRes, err := queryClient.ValidatorSlashes(gocontext.Background(), req)

			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes.GetSlashes(), slashesRes.GetSlashes())
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(slashesRes)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCDelegationRewards() {
	ctx, addrs, valAddrs := suite.ctx, suite.addrs, suite.valAddrs

	tstaking := stakingtestutil.NewHelper(suite.T(), ctx, suite.stakingKeeper)
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), math.LegacyNewDec(0))
	tstaking.CreateValidator(valAddrs[0], valConsPk0, sdk.NewInt(100), true)

	staking.EndBlocker(ctx, suite.stakingKeeper)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, suite.interfaceRegistry)
	types.RegisterQueryServer(queryHelper, keeper.NewQuerier(suite.distrKeeper))
	queryClient := types.NewQueryClient(queryHelper)

	val := suite.stakingKeeper.Validator(ctx, valAddrs[0])

	initial := int64(10)
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial)}}
	suite.distrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// test command delegation rewards grpc
	var (
		req    *types.QueryDelegationRewardsRequest
		expRes *types.QueryDelegationRewardsResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = &types.QueryDelegationRewardsRequest{}
			},
			false,
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
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.msg), func() {
			testCase.malleate()

			rewards, err := queryClient.DelegationRewards(gocontext.Background(), req)

			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes, rewards)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(rewards)
			}
		})
	}

	// test command delegator total rewards grpc
	var (
		totalRewardsReq    *types.QueryDelegationTotalRewardsRequest
		expTotalRewardsRes *types.QueryDelegationTotalRewardsResponse
	)

	testCases = []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				totalRewardsReq = &types.QueryDelegationTotalRewardsRequest{}
			},
			false,
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
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.msg), func() {
			testCase.malleate()

			totalRewardsRes, err := queryClient.DelegationTotalRewards(gocontext.Background(), totalRewardsReq)

			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(totalRewardsRes, expTotalRewardsRes)
			} else {

				suite.Require().Error(err)
				suite.Require().Nil(totalRewardsRes)
			}
		})
	}

	// test command validator delegators grpc
	var (
		delegatorValidatorsReq    *types.QueryDelegatorValidatorsRequest
		expDelegatorValidatorsRes *types.QueryDelegatorValidatorsResponse
	)

	testCases = []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				delegatorValidatorsReq = &types.QueryDelegatorValidatorsRequest{}
			},
			false,
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
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.msg), func() {
			testCase.malleate()

			validators, err := queryClient.DelegatorValidators(gocontext.Background(), delegatorValidatorsReq)

			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expDelegatorValidatorsRes, validators)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(validators)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCDelegatorWithdrawAddress() {
	ctx, queryClient, addrs := suite.ctx, suite.queryClient, suite.addrs

	err := suite.distrKeeper.SetWithdrawAddr(ctx, addrs[0], addrs[1])
	suite.Require().Nil(err)

	var req *types.QueryDelegatorWithdrawAddressRequest

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = &types.QueryDelegatorWithdrawAddressRequest{}
			},
			false,
		},
		{
			"valid request",
			func() {
				req = &types.QueryDelegatorWithdrawAddressRequest{DelegatorAddress: addrs[0].String()}
			},
			true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.msg), func() {
			testCase.malleate()

			withdrawAddress, err := queryClient.DelegatorWithdrawAddress(gocontext.Background(), req)

			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(withdrawAddress.WithdrawAddress, addrs[1].String())
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(withdrawAddress)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCCommunityPool() {
	ctx, queryClient, addrs := suite.ctx, suite.queryClient, suite.addrs
	// reset fee pool
	suite.distrKeeper.SetFeePool(ctx, types.InitialFeePool())

	var (
		req     *types.QueryCommunityPoolRequest
		expPool *types.QueryCommunityPoolResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"valid request empty community pool",
			func() {
				req = &types.QueryCommunityPoolRequest{}
				expPool = &types.QueryCommunityPoolResponse{}
			},
			true,
		},
		{
			"valid request",
			func() {
				amount := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
				suite.Require().NoError(banktestutil.FundAccount(suite.bankKeeper, ctx, addrs[0], amount))

				err := suite.distrKeeper.FundCommunityPool(ctx, amount, addrs[0])
				suite.Require().Nil(err)
				req = &types.QueryCommunityPoolRequest{}

				expPool = &types.QueryCommunityPoolResponse{Pool: sdk.NewDecCoinsFromCoins(amount...)}
			},
			true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.msg), func() {
			testCase.malleate()

			pool, err := queryClient.CommunityPool(gocontext.Background(), req)

			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expPool, pool)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(pool)
			}
		})
	}
}

func TestDistributionTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
