package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	poolkeeper "github.com/cosmos/cosmos-sdk/x/protocolpool/keeper"
	pooltestutil "github.com/cosmos/cosmos-sdk/x/protocolpool/testutil"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

var (
	poolAcc      = authtypes.NewEmptyModuleAccount(types.CommunityPoolAccount)
	streamAcc    = authtypes.NewEmptyModuleAccount(types.StreamAccount)
	poolDistrAcc = authtypes.NewEmptyModuleAccount(types.ProtocolPoolDistrAccount)

	recipientAddr = sdk.AccAddress("to1__________________")

	fooCoin  = sdk.NewInt64Coin("foo", 100)
	fooCoin2 = sdk.NewInt64Coin("foo", 50)
)

type KeeperTestSuite struct {
	suite.Suite

	ctx        sdk.Context
	poolKeeper poolkeeper.Keeper
	authKeeper *pooltestutil.MockAccountKeeper
	bankKeeper *pooltestutil.MockBankKeeper

	msgServer   types.MsgServer
	queryServer types.QueryServer
}

func (suite *KeeperTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockTime(time.Now())
	encCfg := moduletestutil.MakeTestEncodingConfig()

	// gomock initializations
	ctrl := gomock.NewController(suite.T())
	accountKeeper := pooltestutil.NewMockAccountKeeper(ctrl)
	accountKeeper.EXPECT().GetModuleAddress(types.CommunityPoolAccount).Return(poolAcc.GetAddress())
	accountKeeper.EXPECT().GetModuleAddress(types.ProtocolPoolDistrAccount).Return(poolDistrAcc.GetAddress())
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()
	accountKeeper.EXPECT().GetModuleAddress(types.StreamAccount).Return(streamAcc.GetAddress())
	suite.authKeeper = accountKeeper

	bankKeeper := pooltestutil.NewMockBankKeeper(ctrl)
	suite.bankKeeper = bankKeeper

	authority, err := accountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.GovModuleName))
	suite.Require().NoError(err)

	poolKeeper := poolkeeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		authority,
	)
	suite.ctx = ctx
	suite.poolKeeper = poolKeeper

	err = suite.poolKeeper.Params.Set(ctx, types.Params{
		EnabledDistributionDenoms: []string{sdk.DefaultBondDenom},
	})
	suite.Require().NoError(err)

	types.RegisterInterfaces(encCfg.InterfaceRegistry)
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, poolkeeper.Querier{Keeper: poolKeeper})
	suite.msgServer = poolkeeper.NewMsgServerImpl(poolKeeper)
	suite.queryServer = poolkeeper.NewQuerier(poolKeeper)
}

func (suite *KeeperTestSuite) mockSendCoinsFromModuleToAccount(accAddr sdk.AccAddress) {
	suite.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(suite.ctx, types.CommunityPoolAccount, accAddr, gomock.Any()).AnyTimes()
}

func (suite *KeeperTestSuite) mockWithdrawContinuousFund() {
	suite.authKeeper.EXPECT().GetModuleAccount(gomock.Any(), types.CommunityPoolAccount).Return(poolAcc).AnyTimes()
	distrBal := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100000))
	suite.bankKeeper.EXPECT().GetBalance(gomock.Any(), gomock.Any(), sdk.DefaultBondDenom).Return(distrBal).AnyTimes()
	suite.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
}

func (suite *KeeperTestSuite) mockStreamFunds(distributed math.Int) {
	suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.CommunityPoolAccount).Return(poolAcc).AnyTimes()
	suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.ProtocolPoolDistrAccount).Return(poolDistrAcc).AnyTimes()
	suite.authKeeper.EXPECT().GetModuleAddress(types.StreamAccount).Return(streamAcc.GetAddress()).AnyTimes()
	distrBal := sdk.NewCoin(sdk.DefaultBondDenom, distributed)
	suite.bankKeeper.EXPECT().GetBalance(suite.ctx, poolDistrAcc.GetAddress(), sdk.DefaultBondDenom).Return(distrBal).AnyTimes()
	suite.bankKeeper.EXPECT().SendCoinsFromModuleToModule(suite.ctx, poolDistrAcc.GetName(), streamAcc.GetName(), gomock.Any()).AnyTimes()
	suite.bankKeeper.EXPECT().SendCoinsFromModuleToModule(suite.ctx, poolDistrAcc.GetName(), poolAcc.GetName(), gomock.Any()).AnyTimes()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestIterateAndUpdateFundsDistribution() {
	// We'll create 2 continuous funds of 30% each, and the total pool is 1000000, meaning each fund should get 300000

	suite.SetupTest()
	suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.ProtocolPoolDistrAccount).Return(poolAcc).AnyTimes()
	distrBal := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000000))
	suite.bankKeeper.EXPECT().GetBalance(suite.ctx, poolAcc.GetAddress(), sdk.DefaultBondDenom).Return(distrBal).AnyTimes()
	suite.bankKeeper.EXPECT().SendCoinsFromModuleToModule(suite.ctx, poolDistrAcc.GetName(), streamAcc.GetName(), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(600000))))
	suite.bankKeeper.EXPECT().SendCoinsFromModuleToModule(suite.ctx, poolDistrAcc.GetName(), poolAcc.GetName(), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(400000))))

	_, err := suite.msgServer.CreateContinuousFund(suite.ctx, &types.MsgCreateContinuousFund{
		Authority:  suite.poolKeeper.GetAuthority(),
		Recipient:  "cosmos1qypq2q2l8z4wz2z2l8z4wz2z2l8z4wz2srklj6",
		Percentage: math.LegacyMustNewDecFromStr("0.3"),
	})
	suite.Require().NoError(err)

	_, err = suite.msgServer.CreateContinuousFund(suite.ctx, &types.MsgCreateContinuousFund{
		Authority:  suite.poolKeeper.GetAuthority(),
		Recipient:  "cosmos1tygms3xhhs3yv487phx3dw4a95jn7t7lpm470r",
		Percentage: math.LegacyMustNewDecFromStr("0.3"),
	})
	suite.Require().NoError(err)

	_ = suite.poolKeeper.SetToDistribute(suite.ctx)

	err = suite.poolKeeper.IterateAndUpdateFundsDistribution(suite.ctx)
	suite.Require().NoError(err)

	err = suite.poolKeeper.RecipientFundDistributions.Walk(suite.ctx, nil, func(key sdk.AccAddress, value types.DistributionAmount) (stop bool, err error) {
		strAddr, err := suite.authKeeper.AddressCodec().BytesToString(key)
		suite.Require().NoError(err)

		if strAddr == "cosmos1qypq2q2l8z4wz2z2l8z4wz2z2l8z4wz2srklj6" {
			suite.Require().Equal(value.Amount, sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(300000))))
		} else if strAddr == "cosmos1tygms3xhhs3yv487phx3dw4a95jn7t7lpm470r" {
			suite.Require().Equal(value.Amount, sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(300000))))
		}
		return false, nil
	})
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestGetCommunityPool() {
	suite.SetupTest()

	expectedBalance := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000000)))
	suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.CommunityPoolAccount).Return(poolAcc).Times(1)
	suite.bankKeeper.EXPECT().GetAllBalances(suite.ctx, poolAcc.GetAddress()).Return(expectedBalance).Times(1)

	balance, err := suite.poolKeeper.GetCommunityPool(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedBalance, balance)

	// Test error case when module account doesn't exist
	suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.CommunityPoolAccount).Return(nil).Times(1)
	_, err = suite.poolKeeper.GetCommunityPool(suite.ctx)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "module account protocolpool_community_pool does not exist")
}

func (suite *KeeperTestSuite) TestSetToDistribute() {
	suite.SetupTest()

	params, err := suite.poolKeeper.Params.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal([]string{sdk.DefaultBondDenom}, params.EnabledDistributionDenoms)

	// add another denom
	err = suite.poolKeeper.Params.Set(suite.ctx, types.Params{
		EnabledDistributionDenoms: []string{sdk.DefaultBondDenom, "foo"},
	})
	suite.Require().NoError(err)

	suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.ProtocolPoolDistrAccount).Return(poolDistrAcc).AnyTimes()
	distrBal := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000000))
	suite.bankKeeper.EXPECT().GetBalance(suite.ctx, poolDistrAcc.GetAddress(), sdk.DefaultBondDenom).Return(distrBal).Times(2)
	suite.bankKeeper.EXPECT().GetBalance(suite.ctx, poolDistrAcc.GetAddress(), "foo").Return(sdk.NewCoin("foo", math.NewInt(1234))).Times(2)

	// because there are no continuous funds, all are going to the community pool
	suite.bankKeeper.EXPECT().SendCoinsFromModuleToModule(suite.ctx, poolDistrAcc.GetName(), poolAcc.GetName(), sdk.NewCoins(distrBal, sdk.NewCoin("foo", math.NewInt(1234))))

	err = suite.poolKeeper.SetToDistribute(suite.ctx)
	suite.Require().NoError(err)

	// Verify that LastBalance was not set (zero balance)
	_, err = suite.poolKeeper.LastBalance.Get(suite.ctx)
	suite.Require().ErrorContains(err, "not found")

	// create new continuous fund and distribute again
	addrCdc := address.NewBech32Codec("cosmos")
	addrStr := "cosmos1qypq2q2l8z4wz2z2l8z4wz2z2l8z4wz2srklj6"
	addrBz, err := addrCdc.StringToBytes(addrStr)
	suite.Require().NoError(err)

	err = suite.poolKeeper.ContinuousFunds.Set(suite.ctx, addrBz, types.ContinuousFund{
		Recipient:  addrStr,
		Percentage: math.LegacyMustNewDecFromStr("0.3"),
		Expiry:     nil,
	})
	suite.Require().NoError(err)

	err = suite.poolKeeper.SetToDistribute(suite.ctx)
	suite.Require().NoError(err)

	// Verify that LastBalance was set correctly
	lastBalance, err := suite.poolKeeper.LastBalance.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(1000000)), sdk.NewCoin("foo", math.NewInt(1234))), lastBalance.Amount)

	// Verify that a distribution was set
	var distribution types.DistributionAmount
	err = suite.poolKeeper.Distributions.Walk(suite.ctx, nil, func(key time.Time, value types.DistributionAmount) (bool, error) {
		distribution = value
		return true, nil
	})
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(1000000)), sdk.NewCoin("foo", math.NewInt(1234))), distribution.Amount)

	// Test case when balance is zero
	zeroBal := sdk.NewCoin(sdk.DefaultBondDenom, math.ZeroInt())
	suite.bankKeeper.EXPECT().GetBalance(suite.ctx, poolDistrAcc.GetAddress(), sdk.DefaultBondDenom).Return(zeroBal)
	suite.bankKeeper.EXPECT().GetBalance(suite.ctx, poolDistrAcc.GetAddress(), "foo").Return(sdk.NewCoin("foo", math.ZeroInt()))

	err = suite.poolKeeper.SetToDistribute(suite.ctx)
	suite.Require().NoError(err)

	// Verify that no new distribution was set
	count := 0
	err = suite.poolKeeper.Distributions.Walk(suite.ctx, nil, func(key time.Time, value types.DistributionAmount) (bool, error) {
		count++
		return false, nil
	})
	suite.Require().NoError(err)
	suite.Require().Equal(1, count) // Only the previous distribution should exist
}
