package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/core/header"
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
	poolAcc      = authtypes.NewEmptyModuleAccount(types.ModuleName)
	poolDistrAcc = authtypes.NewEmptyModuleAccount(types.ProtocolPoolDistrAccount)

	recipientAddr  = sdk.AccAddress("to1__________________")
	recipientAddr2 = sdk.AccAddress("to2__________________")

	fooCoin = sdk.NewInt64Coin("foo", 100)
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
	ctx := testCtx.Ctx.WithHeaderInfo(header.Info{Time: time.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	// gomock initializations
	ctrl := gomock.NewController(suite.T())
	accountKeeper := pooltestutil.NewMockAccountKeeper(ctrl)
	accountKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return(poolAcc.GetAddress()).AnyTimes()
	accountKeeper.EXPECT().GetModuleAddress(types.ProtocolPoolDistrAccount).Return(poolDistrAcc.GetAddress()).AnyTimes()
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()
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

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestGetCommunityPool() {
	suite.SetupTest()

	expectedBalance := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000000)))
	suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.ModuleName).Return(poolAcc).Times(1)
	suite.bankKeeper.EXPECT().GetAllBalances(suite.ctx, poolAcc.GetAddress()).Return(expectedBalance).Times(1)

	balance, err := suite.poolKeeper.GetCommunityPool(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedBalance, balance)

	// Test error case when module account doesn't exist
	suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.ModuleName).Return(nil).Times(1)
	_, err = suite.poolKeeper.GetCommunityPool(suite.ctx)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "module account protocolpool does not exist")
}

func (suite *KeeperTestSuite) TestGetAllContinuousFunds() {
	suite.Run("empty store", func() {
		// Reset the context to start with a clean store.
		suite.SetupTest()

		funds, err := suite.poolKeeper.GetAllContinuousFunds(suite.ctx)
		suite.Require().NoError(err)
		suite.Require().Empty(funds, "expected no continuous funds in store")
	})

	suite.Run("one fund in store", func() {
		suite.SetupTest()

		fund := types.ContinuousFund{
			Recipient:  recipientAddr.String(),
			Percentage: math.LegacyMustNewDecFromStr("0.5"),
			Expiry:     nil,
		}

		err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipientAddr, fund)
		suite.Require().NoError(err)

		funds, err := suite.poolKeeper.GetAllContinuousFunds(suite.ctx)
		suite.Require().NoError(err)
		suite.Require().Len(funds, 1, "expected one continuous fund in store")
		suite.Require().Equal(fund.Recipient, funds[0].Recipient)
		suite.Require().Equal(fund.Percentage, funds[0].Percentage)
		suite.Require().Equal(fund.Expiry, funds[0].Expiry)
	})

	suite.Run("many funds in store", func() {
		suite.SetupTest()

		totalFunds := 10

		// Insert a number of funds.
		for i := 0; i < totalFunds; i++ {
			accAddr := sdk.AccAddress(fmt.Sprintf("ao%d__________________", i))

			fund := types.ContinuousFund{
				Recipient:  accAddr.String(),
				Percentage: math.LegacyMustNewDecFromStr("0.1"),
				Expiry:     nil,
			}
			err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, accAddr, fund)
			suite.Require().NoError(err)
		}

		funds, err := suite.poolKeeper.GetAllContinuousFunds(suite.ctx)
		suite.Require().NoError(err)
		suite.Require().Len(funds, totalFunds, "expected many continuous funds in store")

		// verify each inserted fund's percentage.
		for _, f := range funds {
			suite.Require().Equal(math.LegacyMustNewDecFromStr("0.1"), f.Percentage)
		}
	})
}

func (suite *KeeperTestSuite) TestDistributeFunds() {
	tests := []struct {
		name        string
		setup       func()
		expectedErr string
		verify      func()
	}{
		{
			name: "module account missing",
			setup: func() {
				suite.SetupTest()
				suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.ProtocolPoolDistrAccount).Return(nil)
			},
			expectedErr: "module account",
		},
		{
			name: "zero funds in module account",
			setup: func() {
				suite.SetupTest()
				suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.ProtocolPoolDistrAccount).
					Return(poolDistrAcc).AnyTimes()

				// Enable only the stake denom.
				params := types.Params{EnabledDistributionDenoms: []string{sdk.DefaultBondDenom}}
				suite.Require().NoError(suite.poolKeeper.Params.Set(suite.ctx, params))

				zeroCoin := sdk.NewCoin(sdk.DefaultBondDenom, math.ZeroInt())
				suite.bankKeeper.EXPECT().GetBalance(suite.ctx, poolDistrAcc.GetAddress(), sdk.DefaultBondDenom).
					Return(zeroCoin).Times(1)
			},
			expectedErr: "",
		},
		{
			name: "one valid continuous fund",
			setup: func() {
				suite.SetupTest()
				suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.ProtocolPoolDistrAccount).
					Return(poolDistrAcc).AnyTimes()

				params := types.Params{EnabledDistributionDenoms: []string{sdk.DefaultBondDenom}}
				suite.Require().NoError(suite.poolKeeper.Params.Set(suite.ctx, params))

				totalCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000)))
				suite.bankKeeper.EXPECT().GetBalance(suite.ctx, poolDistrAcc.GetAddress(), sdk.DefaultBondDenom).
					Return(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000))).Times(1)

				recipient := sdk.AccAddress("recipient1______________")
				fund := types.ContinuousFund{
					Recipient:  recipient.String(),
					Percentage: math.LegacyMustNewDecFromStr("0.3"),
					Expiry:     nil,
				}
				err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipient, fund)
				suite.Require().NoError(err)

				amountToStream := types.PercentageCoinMul(math.LegacyMustNewDecFromStr("0.3"), totalCoins)
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(suite.ctx, types.ProtocolPoolDistrAccount, recipient, amountToStream).
					Return(nil).Times(1)

				remainingCoins := totalCoins.Sub(amountToStream...)
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToModule(suite.ctx, types.ProtocolPoolDistrAccount, types.ModuleName, remainingCoins).
					Return(nil).Times(1)
			},
			expectedErr: "",
		},
		{
			name: "one expired continuous fund",
			setup: func() {
				suite.SetupTest()
				suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.ProtocolPoolDistrAccount).
					Return(poolDistrAcc).AnyTimes()

				params := types.Params{EnabledDistributionDenoms: []string{sdk.DefaultBondDenom}}
				suite.Require().NoError(suite.poolKeeper.Params.Set(suite.ctx, params))

				totalCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000)))
				suite.bankKeeper.EXPECT().GetBalance(suite.ctx, poolDistrAcc.GetAddress(), sdk.DefaultBondDenom).
					Return(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000))).Times(1)

				recipient := sdk.AccAddress("expiredrecipient_______")
				expiredTime := suite.ctx.BlockTime().Add(-time.Hour)
				fund := types.ContinuousFund{
					Recipient:  recipient.String(),
					Percentage: math.LegacyMustNewDecFromStr("0.3"),
					Expiry:     &expiredTime,
				}
				err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipient, fund)
				suite.Require().NoError(err)

				// And full amount to be sent to the community pool.
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToModule(suite.ctx, types.ProtocolPoolDistrAccount, types.ModuleName, totalCoins).
					Return(nil).Times(1)
			},
			expectedErr: "",
			verify: func() {
				// Verify that the expired fund was removed.
				funds, err := suite.poolKeeper.GetAllContinuousFunds(suite.ctx)
				suite.Require().NoError(err)
				suite.Require().Empty(funds, "expected expired continuous fund to be removed")
			},
		},
		{
			name: "multiple valid continuous funds",
			setup: func() {
				suite.SetupTest()
				suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.ProtocolPoolDistrAccount).
					Return(poolDistrAcc).AnyTimes()

				params := types.Params{EnabledDistributionDenoms: []string{sdk.DefaultBondDenom}}
				suite.Require().NoError(suite.poolKeeper.Params.Set(suite.ctx, params))

				totalCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000)))
				suite.bankKeeper.EXPECT().GetBalance(suite.ctx, poolDistrAcc.GetAddress(), sdk.DefaultBondDenom).
					Return(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000))).Times(1)

				recipient1 := sdk.AccAddress("recipient1______________")
				fund1 := types.ContinuousFund{
					Recipient:  recipient1.String(),
					Percentage: math.LegacyMustNewDecFromStr("0.3"),
					Expiry:     nil,
				}
				err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipient1, fund1)
				suite.Require().NoError(err)

				recipient2 := sdk.AccAddress("recipient2______________")
				fund2 := types.ContinuousFund{
					Recipient:  recipient2.String(),
					Percentage: math.LegacyMustNewDecFromStr("0.2"),
					Expiry:     nil,
				}
				err = suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipient2, fund2)
				suite.Require().NoError(err)

				amountToStream1 := types.PercentageCoinMul(math.LegacyMustNewDecFromStr("0.3"), totalCoins)
				amountToStream2 := types.PercentageCoinMul(math.LegacyMustNewDecFromStr("0.2"), totalCoins)
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(suite.ctx, types.ProtocolPoolDistrAccount, recipient1, amountToStream1).
					Return(nil).Times(1)
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(suite.ctx, types.ProtocolPoolDistrAccount, recipient2, amountToStream2).
					Return(nil).Times(1)

				remainingCoins := totalCoins.Sub(amountToStream1...).Sub(amountToStream2...)
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToModule(suite.ctx, types.ProtocolPoolDistrAccount, types.ModuleName, remainingCoins).
					Return(nil).Times(1)
			},
			expectedErr: "",
		},
		{
			name: "fund percentages sum over 1 (resulting in negative remainder)",
			setup: func() {
				suite.SetupTest()
				suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.ProtocolPoolDistrAccount).
					Return(poolDistrAcc).AnyTimes()

				params := types.Params{EnabledDistributionDenoms: []string{sdk.DefaultBondDenom}}
				suite.Require().NoError(suite.poolKeeper.Params.Set(suite.ctx, params))

				totalCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000)))
				suite.bankKeeper.EXPECT().GetBalance(suite.ctx, poolDistrAcc.GetAddress(), sdk.DefaultBondDenom).
					Return(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000))).Times(1)

				// Two funds whose percentages sum to 1.3 (80% and 50%).
				recipient1 := sdk.AccAddress("recipient1______________")
				fund1 := types.ContinuousFund{
					Recipient:  recipient1.String(),
					Percentage: math.LegacyMustNewDecFromStr("0.8"),
					Expiry:     nil,
				}
				err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipient1, fund1)
				suite.Require().NoError(err)

				recipient2 := sdk.AccAddress("recipient2______________")
				fund2 := types.ContinuousFund{
					Recipient:  recipient2.String(),
					Percentage: math.LegacyMustNewDecFromStr("0.5"),
					Expiry:     nil,
				}
				err = suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipient2, fund2)
				suite.Require().NoError(err)

				amountToStream1 := types.PercentageCoinMul(math.LegacyMustNewDecFromStr("0.8"), totalCoins) // 800 stake
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(suite.ctx, types.ProtocolPoolDistrAccount, recipient1, amountToStream1).
					Return(nil).Times(1)
			},
			// we will fail on the second iteration of the loop
			expectedErr: "negative funds for distribution from ContinuousFunds: -300stake",
		},
		{
			name: "bank send to account error",
			setup: func() {
				suite.SetupTest()
				suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.ProtocolPoolDistrAccount).
					Return(poolDistrAcc).AnyTimes()

				params := types.Params{EnabledDistributionDenoms: []string{sdk.DefaultBondDenom}}
				suite.Require().NoError(suite.poolKeeper.Params.Set(suite.ctx, params))

				totalCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000)))
				suite.bankKeeper.EXPECT().GetBalance(suite.ctx, poolDistrAcc.GetAddress(), sdk.DefaultBondDenom).
					Return(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000))).Times(1)

				recipient := sdk.AccAddress("recipient1______________")
				fund := types.ContinuousFund{
					Recipient:  recipient.String(),
					Percentage: math.LegacyMustNewDecFromStr("0.3"),
					Expiry:     nil,
				}
				err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipient, fund)
				suite.Require().NoError(err)

				amountToStream := types.PercentageCoinMul(math.LegacyMustNewDecFromStr("0.3"), totalCoins)
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(suite.ctx, types.ProtocolPoolDistrAccount, recipient, amountToStream).
					Return(fmt.Errorf("send account error")).Times(1)
			},
			expectedErr: "failed to distribute fund",
		},
		{
			name: "bank send to module error",
			setup: func() {
				suite.SetupTest()
				suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.ProtocolPoolDistrAccount).
					Return(poolDistrAcc).AnyTimes()

				params := types.Params{EnabledDistributionDenoms: []string{sdk.DefaultBondDenom}}
				suite.Require().NoError(suite.poolKeeper.Params.Set(suite.ctx, params))

				totalCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000)))
				suite.bankKeeper.EXPECT().GetBalance(suite.ctx, poolDistrAcc.GetAddress(), sdk.DefaultBondDenom).
					Return(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000))).Times(1)

				recipient := sdk.AccAddress("recipient1______________")
				fund := types.ContinuousFund{
					Recipient:  recipient.String(),
					Percentage: math.LegacyMustNewDecFromStr("0.3"),
					Expiry:     nil,
				}
				err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipient, fund)
				suite.Require().NoError(err)

				amountToStream := types.PercentageCoinMul(math.LegacyMustNewDecFromStr("0.3"), totalCoins)
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(suite.ctx, types.ProtocolPoolDistrAccount, recipient, amountToStream).
					Return(nil).Times(1)

				remainingCoins := totalCoins.Sub(amountToStream...)
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToModule(suite.ctx, types.ProtocolPoolDistrAccount, types.ModuleName, remainingCoins).
					Return(fmt.Errorf("send module error")).Times(1)
			},
			expectedErr: "failed to send coins to community pool",
		},
		{
			name: "fund expiry equals block time (not expired)",
			setup: func() {
				suite.SetupTest()
				suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.ProtocolPoolDistrAccount).
					Return(poolDistrAcc).AnyTimes()

				params := types.Params{EnabledDistributionDenoms: []string{sdk.DefaultBondDenom}}
				suite.Require().NoError(suite.poolKeeper.Params.Set(suite.ctx, params))

				totalCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000)))
				suite.bankKeeper.EXPECT().GetBalance(suite.ctx, poolDistrAcc.GetAddress(), sdk.DefaultBondDenom).
					Return(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000))).Times(1)

				recipient := sdk.AccAddress("recipient1______________")
				expiryTime := suite.ctx.BlockTime() // exactly equal to block time.
				fund := types.ContinuousFund{
					Recipient:  recipient.String(),
					Percentage: math.LegacyMustNewDecFromStr("0.3"),
					Expiry:     &expiryTime,
				}
				err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipient, fund)
				suite.Require().NoError(err)

				amountToStream := types.PercentageCoinMul(math.LegacyMustNewDecFromStr("0.3"), totalCoins)
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(suite.ctx, types.ProtocolPoolDistrAccount, recipient, amountToStream).
					Return(nil).Times(1)

				remainingCoins := totalCoins.Sub(amountToStream...)
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToModule(suite.ctx, types.ProtocolPoolDistrAccount, types.ModuleName, remainingCoins).
					Return(nil).Times(1)
			},
			expectedErr: "",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.setup()

			err := suite.poolKeeper.DistributeFunds(suite.ctx)
			switch {
			case tc.expectedErr != "":
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expectedErr)
			default:
				suite.Require().NoError(err)
			}
		})
	}
}
