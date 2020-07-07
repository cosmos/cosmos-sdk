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
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
)

type DistributionTestSuite struct {
	suite.Suite

	app         *simapp.SimApp
	ctx         sdk.Context
	queryClient types.QueryClient
}

func (suite *DistributionTestSuite) SetupTest() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.DistrKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.app = app
	suite.ctx = ctx
	suite.queryClient = queryClient
}

func (suite *DistributionTestSuite) TestGRPCParams() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient

	params := types.Params{
		CommunityTax:        sdk.NewDecWithPrec(3, 1),
		BaseProposerReward:  sdk.NewDecWithPrec(2, 1),
		BonusProposerReward: sdk.NewDecWithPrec(1, 1),
		WithdrawAddrEnabled: true,
	}

	app.DistrKeeper.SetParams(ctx, params)

	paramsRes, err := queryClient.Params(gocontext.Background(), &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(paramsRes.Params, params)
}

func (suite *DistributionTestSuite) TestGRPCValidatorOutstandingRewards() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient

	valCommission := sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", sdk.NewDec(5000)),
		sdk.NewDecCoinFromDec("stake", sdk.NewDec(300)),
	}

	addr := simapp.AddTestAddrs(app, ctx, 1, sdk.NewInt(1000000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addr)

	pageReq := &query.PageRequest{Limit: 1}

	req := types.NewQueryValidatorOutstandingRewardsRequest(nil, pageReq)

	validatorOutstandingRewards, err := queryClient.ValidatorOutstandingRewards(gocontext.Background(), req)
	suite.Require().Error(err)
	suite.Require().Nil(validatorOutstandingRewards)

	// set outstanding rewards
	app.DistrKeeper.SetValidatorOutstandingRewards(ctx, valAddrs[0], types.ValidatorOutstandingRewards{Rewards: valCommission})
	rewards := app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[0])
	req = types.NewQueryValidatorOutstandingRewardsRequest(valAddrs[0], pageReq)

	validatorOutstandingRewards, err = queryClient.ValidatorOutstandingRewards(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().Equal(rewards, validatorOutstandingRewards.Rewards)
	suite.Require().Equal(valCommission, validatorOutstandingRewards.Rewards.Rewards)
}

func (suite *DistributionTestSuite) TestGRPCValidatorCommission() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient

	addr := simapp.AddTestAddrs(app, ctx, 1, sdk.NewInt(1000000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addr)
	valOpAddr1 := valAddrs[0]

	commission := sdk.DecCoins{{Denom: "token1", Amount: sdk.NewDec(4)}, {Denom: "token2", Amount: sdk.NewDec(2)}}
	app.DistrKeeper.SetValidatorAccumulatedCommission(ctx, valOpAddr1, types.ValidatorAccumulatedCommission{Commission: commission})

	req := &types.QueryValidatorCommissionRequest{}
	commissionRes, err := queryClient.ValidatorCommission(gocontext.Background(), req)
	suite.Require().Error(err)
	suite.Require().Nil(commissionRes)

	req = &types.QueryValidatorCommissionRequest{ValidatorAddress: valOpAddr1}
	commissionRes, err = queryClient.ValidatorCommission(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().Equal(commissionRes.Commission.Commission, commission)
}

func (suite *DistributionTestSuite) TestGRPCValidatorSlashes() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient

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

	slashes, err := queryClient.ValidatorSlashes(gocontext.Background(), &types.QueryValidatorSlashesRequest{})
	suite.Require().Error(err)
	suite.Require().Nil(slashes)

	req := &types.QueryValidatorSlashesRequest{
		ValidatorAddress: valOpAddr1,
		StartingHeight:   1,
		EndingHeight:     10,
		Req:              pageReq,
	}

	slashes, err = queryClient.ValidatorSlashes(gocontext.Background(), req)

	suite.Require().NoError(err)
	suite.Require().Len(slashes.Slashes, 1)
	suite.Require().Equal(slashOne, slashes.Slashes[0])
	suite.Require().NotNil(slashes.Res.NextKey)

	pageReq.Key = slashes.Res.NextKey
	req.Req = pageReq

	slashes, err = queryClient.ValidatorSlashes(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().Len(slashes.Slashes, 1)
	suite.Require().Equal(slashTwo, slashes.Slashes[0])
	suite.Require().Empty(slashes.Res)

	pageReq = &query.PageRequest{Limit: 2}
	req.Req = pageReq

	slashes, err = queryClient.ValidatorSlashes(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().Len(slashes.Slashes, 2)
	suite.Require().Equal(slashOne, slashes.Slashes[0])
	suite.Require().Equal(slashTwo, slashes.Slashes[1])
	suite.Require().Empty(slashes.Res)
}

func (suite *DistributionTestSuite) TestGRPCDelegationRewards() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient

	addr := simapp.AddTestAddrs(app, ctx, 1, sdk.NewInt(1000000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addr)
	valOpAddr1 := valAddrs[0]

	sh := staking.NewHandler(app.StakingKeeper)
	comm := stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := stakingtypes.NewMsgCreateValidator(
		valOpAddr1, valConsPk1, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), stakingtypes.Description{}, comm, sdk.OneInt(),
	)

	res, err := sh(ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

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

	rewards, err := queryClient.DelegationRewards(gocontext.Background(), &types.QueryDelegationRewardsRequest{})
	suite.Require().Error(err)
	suite.Require().Nil(rewards)

	_, err = queryClient.DelegationRewards(gocontext.Background(), req)
	suite.Require().NoError(err)
	// TODO debug delegation rewards
	// suite.Require().Equal(sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial / 2)}}, rewards.Rewards)

	_, err = queryClient.DelegationTotalRewards(gocontext.Background(), &types.QueryDelegationTotalRewardsRequest{})
	suite.Require().Error(err)

	totalRewardsReq := &types.QueryDelegationTotalRewardsRequest{
		DelegatorAddress: sdk.AccAddress(valOpAddr1),
	}

	_, err = queryClient.DelegationTotalRewards(gocontext.Background(), totalRewardsReq)
	suite.Require().NoError(err)

	// TODO debug delegation rewards
	// expectedDelReward := types.NewDelegationDelegatorReward(valOpAddr1,
	// 	sdk.DecCoins{sdk.NewInt64DecCoin("stake", 5)})
	// wantDelRewards := types.NewQueryDelegatorTotalRewardsResponse(
	// 	[]types.DelegationDelegatorReward{expectedDelReward}, expectedDelReward.Reward)
	// suite.Require().Equal(wantDelRewards, totalRewards)

	validators, err := queryClient.DelegatorValidators(gocontext.Background(), &types.QueryDelegatorValidatorsRequest{})
	suite.Require().Error(err)
	suite.Require().Nil(validators, 1)

	validators, err = queryClient.DelegatorValidators(gocontext.Background(),
		&types.QueryDelegatorValidatorsRequest{DelegatorAddress: sdk.AccAddress(valOpAddr1)})
	suite.Require().NoError(err)
	suite.Require().Len(validators.Validators, 1)
	suite.Require().Equal(validators.Validators[0], valOpAddr1)
}

func (suite *DistributionTestSuite) TestGRPCDelegatorWithdrawAddress() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient

	addr := simapp.AddTestAddrs(app, ctx, 2, sdk.NewInt(1000000000))

	err := app.DistrKeeper.SetWithdrawAddr(ctx, addr[0], addr[1])
	suite.Require().Nil(err)

	withdrawAddress, err := queryClient.DelegatorWithdrawAddress(gocontext.Background(), &types.QueryDelegatorWithdrawAddressRequest{})
	suite.Require().Error(err)
	suite.Require().Nil(withdrawAddress)

	withdrawAddress, err = queryClient.DelegatorWithdrawAddress(gocontext.Background(),
		&types.QueryDelegatorWithdrawAddressRequest{DelegatorAddress: addr[0]})
	suite.Require().NoError(err)
	suite.Require().Equal(withdrawAddress.WithdrawAddress, addr[1])
}

func (suite *DistributionTestSuite) TestGRPCCommunityPool() {
	queryClient := suite.queryClient

	pool, err := queryClient.CommunityPool(gocontext.Background(), &types.QueryCommunityPoolRequest{})
	suite.Require().NoError(err)
	suite.Require().Empty(pool)
}

func TestDistributionTestSuite(t *testing.T) {
	suite.Run(t, new(DistributionTestSuite))
}
