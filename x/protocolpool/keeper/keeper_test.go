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
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	poolkeeper "github.com/cosmos/cosmos-sdk/x/protocolpool/keeper"
	pooltestutil "github.com/cosmos/cosmos-sdk/x/protocolpool/testutil"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

var (
	poolAcc      = authtypes.NewEmptyModuleAccount(types.ModuleName)
	poolDistrAcc = authtypes.NewEmptyModuleAccount(types.ProtocolPoolEscrowAccount)

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
	accountKeeper.EXPECT().GetModuleAddress(types.ProtocolPoolEscrowAccount).Return(poolDistrAcc.GetAddress()).AnyTimes()
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()
	suite.authKeeper = accountKeeper

	bankKeeper := pooltestutil.NewMockBankKeeper(ctrl)
	suite.bankKeeper = bankKeeper

	authority, err := accountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(govtypes.ModuleName))
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
		for i := range totalFunds {
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
	initialBalance := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000))
	initalBalanceCoins := sdk.NewCoins(initialBalance)

	tests := []struct {
		name               string
		params             types.Params
		initialPoolBalance sdk.Coin
		setup              func()
		expectedErr        string
		verify             func()
	}{
		{
			name:   "module account missing",
			params: types.Params{EnabledDistributionDenoms: []string{sdk.DefaultBondDenom}},
			setup: func() {
				suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.ProtocolPoolEscrowAccount).Return(nil)
			},
			expectedErr: "module account",
		},
		{
			name:               "zero funds in module account",
			initialPoolBalance: sdk.NewCoin(sdk.DefaultBondDenom, math.ZeroInt()),
			params:             types.Params{EnabledDistributionDenoms: []string{sdk.DefaultBondDenom}},
			setup: func() {
				suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.ProtocolPoolEscrowAccount).
					Return(poolDistrAcc).AnyTimes()
			},
			expectedErr: "",
		},
		{
			name:               "one valid continuous fund",
			params:             types.Params{EnabledDistributionDenoms: []string{sdk.DefaultBondDenom}},
			initialPoolBalance: initialBalance,
			setup: func() {
				suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.ProtocolPoolEscrowAccount).
					Return(poolDistrAcc).AnyTimes()

				fund := types.ContinuousFund{
					Recipient:  recipientAddr.String(),
					Percentage: math.LegacyMustNewDecFromStr("0.3"),
					Expiry:     nil,
				}
				err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipientAddr, fund)
				suite.Require().NoError(err)

				amountToStream := poolkeeper.PercentageCoinMul(math.LegacyMustNewDecFromStr("0.3"), initalBalanceCoins)
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(suite.ctx, types.ProtocolPoolEscrowAccount, recipientAddr, amountToStream).
					Return(nil).Times(1)

				remainingCoins := initalBalanceCoins.Sub(amountToStream...)
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToModule(suite.ctx, types.ProtocolPoolEscrowAccount, types.ModuleName, remainingCoins).
					Return(nil).Times(1)
			},
			expectedErr: "",
		},
		{
			name:               "one valid continuous fund with a blocked account",
			params:             types.Params{EnabledDistributionDenoms: []string{sdk.DefaultBondDenom}},
			initialPoolBalance: initialBalance,
			setup: func() {
				suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.ProtocolPoolEscrowAccount).
					Return(poolDistrAcc).AnyTimes()

				fund := types.ContinuousFund{
					Recipient:  recipientAddr.String(),
					Percentage: math.LegacyMustNewDecFromStr("0.3"),
					Expiry:     nil,
				}
				err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipientAddr, fund)
				suite.Require().NoError(err)

				amountToStream := poolkeeper.PercentageCoinMul(math.LegacyMustNewDecFromStr("0.3"), initalBalanceCoins)
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(suite.ctx, types.ProtocolPoolEscrowAccount, recipientAddr, amountToStream).
					Return(sdkerrors.ErrUnauthorized).Times(1)

				suite.bankKeeper.EXPECT().SendCoinsFromModuleToModule(suite.ctx, types.ProtocolPoolEscrowAccount, types.ModuleName, initalBalanceCoins).
					Return(nil).Times(1)
			},
			expectedErr: "",
			verify: func() {
				// check that the broken continuous fund is removed
				_, err := suite.poolKeeper.ContinuousFunds.Get(suite.ctx, recipientAddr)
				suite.Require().Error(err)
			},
		},
		{
			name:               "one expired continuous fund",
			params:             types.Params{EnabledDistributionDenoms: []string{sdk.DefaultBondDenom}},
			initialPoolBalance: initialBalance,
			setup: func() {
				suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.ProtocolPoolEscrowAccount).
					Return(poolDistrAcc).AnyTimes()

				expiredTime := suite.ctx.BlockTime().Add(-time.Hour)
				fund := types.ContinuousFund{
					Recipient:  recipientAddr.String(),
					Percentage: math.LegacyMustNewDecFromStr("0.3"),
					Expiry:     &expiredTime,
				}
				err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipientAddr, fund)
				suite.Require().NoError(err)

				// And full amount to be sent to the community pool.
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToModule(suite.ctx, types.ProtocolPoolEscrowAccount, types.ModuleName, initalBalanceCoins).
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
			name:               "multiple valid continuous funds - one recipient is blocked",
			initialPoolBalance: initialBalance,
			params:             types.Params{EnabledDistributionDenoms: []string{sdk.DefaultBondDenom}},
			setup: func() {
				suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.ProtocolPoolEscrowAccount).
					Return(poolDistrAcc).AnyTimes()

				fund1 := types.ContinuousFund{
					Recipient:  recipientAddr.String(),
					Percentage: math.LegacyMustNewDecFromStr("0.3"),
					Expiry:     nil,
				}
				err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipientAddr, fund1)
				suite.Require().NoError(err)

				fund2 := types.ContinuousFund{
					Recipient:  recipientAddr2.String(),
					Percentage: math.LegacyMustNewDecFromStr("0.2"),
					Expiry:     nil,
				}
				accAddr, err := sdk.AccAddressFromBech32(fund2.Recipient)
				suite.Require().NoError(err)

				err = suite.poolKeeper.ContinuousFunds.Set(suite.ctx, accAddr, fund2)
				suite.Require().NoError(err)

				amountToStream1 := poolkeeper.PercentageCoinMul(math.LegacyMustNewDecFromStr("0.3"), initalBalanceCoins)
				amountToStream2 := poolkeeper.PercentageCoinMul(math.LegacyMustNewDecFromStr("0.2"), initalBalanceCoins)
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(suite.ctx, types.ProtocolPoolEscrowAccount, recipientAddr, amountToStream1).
					Return(sdkerrors.ErrUnauthorized).Times(1)
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(suite.ctx, types.ProtocolPoolEscrowAccount, recipientAddr2, amountToStream2).
					Return(nil).Times(1)

				remainingCoins := initalBalanceCoins.Sub(amountToStream2...)
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToModule(suite.ctx, types.ProtocolPoolEscrowAccount, types.ModuleName, remainingCoins).
					Return(nil).Times(1)
			},
			expectedErr: "",
			verify: func() {
				// check that the broken continuous fund is removed
				_, err := suite.poolKeeper.ContinuousFunds.Get(suite.ctx, recipientAddr)
				suite.Require().Error(err)
				// check that the valid continuous fund is in the store
				_, err = suite.poolKeeper.ContinuousFunds.Get(suite.ctx, recipientAddr2)
				suite.Require().NoError(err)
			},
		},
		{
			name:               "fund percentages sum over 1 (resulting in negative remainder)",
			initialPoolBalance: initialBalance,
			params:             types.Params{EnabledDistributionDenoms: []string{sdk.DefaultBondDenom}},
			setup: func() {
				suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.ProtocolPoolEscrowAccount).
					Return(poolDistrAcc).AnyTimes()

				// Two funds whose percentages sum to 1.3 (80% and 50%).
				fund1 := types.ContinuousFund{
					Recipient:  recipientAddr.String(),
					Percentage: math.LegacyMustNewDecFromStr("0.8"),
					Expiry:     nil,
				}
				err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipientAddr, fund1)
				suite.Require().NoError(err)

				fund2 := types.ContinuousFund{
					Recipient:  recipientAddr2.String(),
					Percentage: math.LegacyMustNewDecFromStr("0.5"),
					Expiry:     nil,
				}
				err = suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipientAddr2, fund2)
				suite.Require().NoError(err)

				amountToStream1 := poolkeeper.PercentageCoinMul(math.LegacyMustNewDecFromStr("0.8"), initalBalanceCoins) // 800 stake
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(suite.ctx, types.ProtocolPoolEscrowAccount, recipientAddr, amountToStream1).
					Return(nil).Times(1)
			},
			// we will fail on the second iteration of the loop
			expectedErr: "negative funds for distribution from ContinuousFunds: -300stake",
		},
		{
			name:   "bank send to account error",
			params: types.Params{EnabledDistributionDenoms: []string{sdk.DefaultBondDenom}},
			setup: func() {
				suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.ProtocolPoolEscrowAccount).
					Return(poolDistrAcc).AnyTimes()

				totalCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000)))
				suite.bankKeeper.EXPECT().GetBalance(suite.ctx, poolDistrAcc.GetAddress(), sdk.DefaultBondDenom).
					Return(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000))).Times(1)

				fund := types.ContinuousFund{
					Recipient:  recipientAddr.String(),
					Percentage: math.LegacyMustNewDecFromStr("0.3"),
					Expiry:     nil,
				}
				err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipientAddr, fund)
				suite.Require().NoError(err)

				amountToStream := poolkeeper.PercentageCoinMul(math.LegacyMustNewDecFromStr("0.3"), totalCoins)
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(suite.ctx, types.ProtocolPoolEscrowAccount, recipientAddr, amountToStream).
					Return(fmt.Errorf("send account error")).Times(1)
			},
			expectedErr: "failed to distribute fund",
		},
		{
			name:               "bank send to module error",
			initialPoolBalance: initialBalance,
			params:             types.Params{EnabledDistributionDenoms: []string{sdk.DefaultBondDenom}},
			setup: func() {
				suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.ProtocolPoolEscrowAccount).
					Return(poolDistrAcc).AnyTimes()

				fund := types.ContinuousFund{
					Recipient:  recipientAddr.String(),
					Percentage: math.LegacyMustNewDecFromStr("0.3"),
					Expiry:     nil,
				}
				err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipientAddr, fund)
				suite.Require().NoError(err)

				amountToStream := poolkeeper.PercentageCoinMul(math.LegacyMustNewDecFromStr("0.3"), initalBalanceCoins)
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(suite.ctx, types.ProtocolPoolEscrowAccount, recipientAddr, amountToStream).
					Return(nil).Times(1)

				remainingCoins := initalBalanceCoins.Sub(amountToStream...)
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToModule(suite.ctx, types.ProtocolPoolEscrowAccount, types.ModuleName, remainingCoins).
					Return(fmt.Errorf("send module error")).Times(1)
			},
			expectedErr: "failed to send coins to community pool",
		},
		{
			name:               "fund expiry equals block time (not expired)",
			params:             types.Params{EnabledDistributionDenoms: []string{sdk.DefaultBondDenom}},
			initialPoolBalance: initialBalance,
			setup: func() {
				suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, types.ProtocolPoolEscrowAccount).
					Return(poolDistrAcc).AnyTimes()

				expiryTime := suite.ctx.BlockTime() // exactly equal to block time.
				fund := types.ContinuousFund{
					Recipient:  recipientAddr.String(),
					Percentage: math.LegacyMustNewDecFromStr("0.3"),
					Expiry:     &expiryTime,
				}
				err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipientAddr, fund)
				suite.Require().NoError(err)

				amountToStream := poolkeeper.PercentageCoinMul(math.LegacyMustNewDecFromStr("0.3"), initalBalanceCoins)
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(suite.ctx, types.ProtocolPoolEscrowAccount, recipientAddr, amountToStream).
					Return(nil).Times(1)

				remainingCoins := initalBalanceCoins.Sub(amountToStream...)
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToModule(suite.ctx, types.ProtocolPoolEscrowAccount, types.ModuleName, remainingCoins).
					Return(nil).Times(1)
			},
			expectedErr: "",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			suite.Require().NoError(suite.poolKeeper.Params.Set(suite.ctx, tc.params))
			if !tc.initialPoolBalance.IsNil() {
				suite.bankKeeper.EXPECT().GetBalance(suite.ctx, poolDistrAcc.GetAddress(), initialBalance.Denom).
					Return(tc.initialPoolBalance).Times(1)
			}

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

// TestPercentageCoinMul tests the PercentageCoinMul function.
func TestPercentageCoinMul(t *testing.T) {
	tests := []struct {
		name       string
		percentage math.LegacyDec
		coins      sdk.Coins
		expected   sdk.Coins
	}{
		{
			name:       "zero percentage",
			percentage: math.LegacyMustNewDecFromStr("0.0"),
			coins:      sdk.NewCoins(sdk.NewCoin("atom", math.NewInt(100))),
			expected:   sdk.NewCoins(sdk.NewCoin("atom", math.NewInt(0))),
		},
		{
			name:       "100 percent",
			percentage: math.LegacyMustNewDecFromStr("1.0"),
			coins:      sdk.NewCoins(sdk.NewCoin("atom", math.NewInt(100))),
			expected:   sdk.NewCoins(sdk.NewCoin("atom", math.NewInt(100))),
		},
		{
			name:       "50 percent",
			percentage: math.LegacyMustNewDecFromStr("0.5"),
			coins:      sdk.NewCoins(sdk.NewCoin("atom", math.NewInt(100))),
			expected:   sdk.NewCoins(sdk.NewCoin("atom", math.NewInt(50))),
		},
		{
			name:       "fraction with truncation",
			percentage: math.LegacyMustNewDecFromStr("0.333333333333333333"), // Approx. 1/3.
			coins:      sdk.NewCoins(sdk.NewCoin("atom", math.NewInt(100))),
			// 100 * 1/3 = 33.333... which truncates to 33.
			expected: sdk.NewCoins(sdk.NewCoin("atom", math.NewInt(33))),
		},
		{
			name:       "multiple denominations",
			percentage: math.LegacyMustNewDecFromStr("0.5"),
			coins: sdk.NewCoins(
				sdk.NewCoin("atom", math.NewInt(100)),
				sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(200)),
			),
			expected: sdk.NewCoins(
				sdk.NewCoin("atom", math.NewInt(50)),
				sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100)),
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Call the function under test.
			result := poolkeeper.PercentageCoinMul(tc.percentage, tc.coins)

			// Compare the resulting coins with the expected coins.
			if !result.Equal(tc.expected) {
				t.Errorf("unexpected result for %s:\nexpected: %s\ngot:      %s", tc.name, tc.expected.String(), result.String())
			}
		})
	}
}
