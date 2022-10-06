package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

var (
	fromAddr   = sdk.AccAddress([]byte("from1________________"))
	to1Addr    = sdk.AccAddress([]byte("to1__________________"))
	to2Addr    = sdk.AccAddress([]byte("to2__________________"))
	to3Addr    = sdk.AccAddress([]byte("to3__________________"))
	fooCoin    = sdk.NewInt64Coin("foo", 100)
	periodCoin = sdk.NewInt64Coin("foo", 20)
)

func (suite *KeeperTestSuite) TestCreateVestingAccount() {
	testCases := map[string]struct {
		preRun    func(*types.MsgCreateVestingAccount)
		input     *types.MsgCreateVestingAccount
		expErr    bool
		expErrMsg string
	}{
		"create for existing account": {
			preRun: func(msg *types.MsgCreateVestingAccount) {
				suite.mockNewMsgCreateVestingAccount(msg, true)
			},
			input: types.NewMsgCreateVestingAccount(
				fromAddr,
				to1Addr,
				sdk.Coins{fooCoin},
				suite.endTime.Unix(),
				true,
			),
			expErr:    true,
			expErrMsg: "already exists",
		},
		"create a valid delayed vesting account": {
			preRun: func(msg *types.MsgCreateVestingAccount) {
				suite.mockNewMsgCreateVestingAccount(msg, false)
			},
			input: types.NewMsgCreateVestingAccount(
				fromAddr,
				to2Addr,
				sdk.Coins{fooCoin},
				suite.endTime.Unix(),
				true,
			),
			expErr:    false,
			expErrMsg: "",
		},
		"create a valid continuous vesting account": {
			preRun: func(msg *types.MsgCreateVestingAccount) {
				suite.mockNewMsgCreateVestingAccount(msg, false)
			},
			input: types.NewMsgCreateVestingAccount(
				fromAddr,
				to3Addr,
				sdk.Coins{fooCoin},
				suite.endTime.Unix(),
				false,
			),
			expErr:    false,
			expErrMsg: "",
		},
	}

	for name, tc := range testCases {
		suite.Run(name, func() {
			tc.preRun(tc.input)
			_, err := suite.msgServer.CreateVestingAccount(suite.ctx, tc.input)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestCreatePermanentLockedAccount() {
	testCases := map[string]struct {
		preRun    func(*types.MsgCreatePermanentLockedAccount)
		input     *types.MsgCreatePermanentLockedAccount
		expErr    bool
		expErrMsg string
	}{
		"create for existing account": {
			preRun: func(msg *types.MsgCreatePermanentLockedAccount) {
				suite.mockCreatePermanentLockedAccount(msg, true)
			},
			input: types.NewMsgCreatePermanentLockedAccount(
				fromAddr,
				to1Addr,
				sdk.Coins{fooCoin},
			),
			expErr:    true,
			expErrMsg: "already exists",
		},
		"create a valid permanent locked account": {
			preRun: func(msg *types.MsgCreatePermanentLockedAccount) {
				suite.mockCreatePermanentLockedAccount(msg, false)
			},
			input: types.NewMsgCreatePermanentLockedAccount(
				fromAddr,
				to2Addr,
				sdk.Coins{fooCoin},
			),
			expErr:    false,
			expErrMsg: "",
		},
	}

	for name, tc := range testCases {
		suite.Run(name, func() {
			tc.preRun(tc.input)
			_, err := suite.msgServer.CreatePermanentLockedAccount(suite.ctx, tc.input)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestCreatePeriodicVestingAccount() {
	testCases := map[string]struct {
		preRun    func(*types.MsgCreatePeriodicVestingAccount)
		input     *types.MsgCreatePeriodicVestingAccount
		expErr    bool
		expErrMsg string
	}{
		"create for existing account": {
			preRun: func(msg *types.MsgCreatePeriodicVestingAccount) {
				suite.mockCreatePeriodicVestingAccount(msg, true)
			},
			input: types.NewMsgCreatePeriodicVestingAccount(
				fromAddr,
				to1Addr,
				time.Now().Unix(),
				[]types.Period{
					{
						Length: 10,
						Amount: sdk.NewCoins(periodCoin),
					},
				},
			),
			expErr:    true,
			expErrMsg: "already exists",
		},
		"create a valid periodic vesting account": {
			preRun: func(msg *types.MsgCreatePeriodicVestingAccount) {
				suite.mockCreatePeriodicVestingAccount(msg, false)
			},
			input: types.NewMsgCreatePeriodicVestingAccount(
				fromAddr,
				to2Addr,
				time.Now().Unix(),
				[]types.Period{
					{
						Length: 10,
						Amount: sdk.NewCoins(periodCoin),
					},
					{
						Length: 20,
						Amount: sdk.NewCoins(fooCoin),
					},
				},
			),
			expErr:    false,
			expErrMsg: "",
		},
	}

	for name, tc := range testCases {
		suite.Run(name, func() {
			tc.preRun(tc.input)
			_, err := suite.msgServer.CreatePeriodicVestingAccount(suite.ctx, tc.input)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) mockNewMsgCreateVestingAccount(msg *types.MsgCreateVestingAccount, accountExisted bool) {
	from := sdk.MustAccAddressFromBech32(msg.FromAddress)
	to := sdk.MustAccAddressFromBech32(msg.ToAddress)

	suite.bankKeeper.EXPECT().IsSendEnabledCoins(suite.ctx, msg.Amount).Return(nil)
	suite.bankKeeper.EXPECT().BlockedAddr(to).Return(false)
	baseAccount := authtypes.NewBaseAccountWithAddress(to)
	if accountExisted {
		suite.accountKeeper.EXPECT().GetAccount(suite.ctx, to).Return(baseAccount)
		return
	}
	suite.accountKeeper.EXPECT().GetAccount(suite.ctx, to).Return(nil)
	suite.accountKeeper.EXPECT().NewAccount(suite.ctx, baseAccount).Return(baseAccount)

	baseVestingAccount := types.NewBaseVestingAccount(baseAccount, msg.Amount.Sort(), msg.EndTime)

	var vestingAccount authtypes.AccountI
	if msg.Delayed {
		vestingAccount = types.NewDelayedVestingAccountRaw(baseVestingAccount)
	} else {
		vestingAccount = types.NewContinuousVestingAccountRaw(baseVestingAccount, suite.ctx.BlockTime().Unix())
	}

	suite.accountKeeper.EXPECT().SetAccount(suite.ctx, vestingAccount)
	suite.bankKeeper.EXPECT().SendCoins(suite.ctx, from, to, msg.Amount).Return(nil)
}

func (suite *KeeperTestSuite) mockCreatePermanentLockedAccount(msg *types.MsgCreatePermanentLockedAccount, accountExisted bool) {
	from := sdk.MustAccAddressFromBech32(msg.FromAddress)
	to := sdk.MustAccAddressFromBech32(msg.ToAddress)

	suite.bankKeeper.EXPECT().IsSendEnabledCoins(suite.ctx, msg.Amount).Return(nil)
	suite.bankKeeper.EXPECT().BlockedAddr(to).Return(false)
	baseAccount := authtypes.NewBaseAccountWithAddress(to)
	if accountExisted {
		suite.accountKeeper.EXPECT().GetAccount(suite.ctx, to).Return(baseAccount)
		return
	}
	suite.accountKeeper.EXPECT().GetAccount(suite.ctx, to).Return(nil)
	suite.accountKeeper.EXPECT().NewAccount(suite.ctx, baseAccount).Return(baseAccount)

	vestingAccount := types.NewPermanentLockedAccount(baseAccount, msg.Amount)

	suite.accountKeeper.EXPECT().SetAccount(suite.ctx, vestingAccount)
	suite.bankKeeper.EXPECT().SendCoins(suite.ctx, from, to, msg.Amount).Return(nil)
}

func (suite *KeeperTestSuite) mockCreatePeriodicVestingAccount(msg *types.MsgCreatePeriodicVestingAccount, accountExisted bool) {
	from := sdk.MustAccAddressFromBech32(msg.FromAddress)
	to := sdk.MustAccAddressFromBech32(msg.ToAddress)

	baseAccount := authtypes.NewBaseAccountWithAddress(to)
	if accountExisted {
		suite.accountKeeper.EXPECT().GetAccount(suite.ctx, to).Return(baseAccount)
		return
	}
	suite.accountKeeper.EXPECT().GetAccount(suite.ctx, to).Return(nil)
	suite.accountKeeper.EXPECT().NewAccount(suite.ctx, baseAccount).Return(baseAccount)

	var totalCoins sdk.Coins
	for _, period := range msg.VestingPeriods {
		totalCoins = totalCoins.Add(period.Amount...)
	}

	vestingAccount := types.NewPeriodicVestingAccount(baseAccount, totalCoins.Sort(), msg.StartTime, msg.VestingPeriods)

	suite.accountKeeper.EXPECT().SetAccount(suite.ctx, vestingAccount)
	suite.bankKeeper.EXPECT().SendCoins(suite.ctx, from, to, totalCoins).Return(nil)
}
