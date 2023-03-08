package vesting_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

type HandlerTestSuite struct {
	suite.Suite

	handler sdk.Handler
	app     *simapp.SimApp
}

func (suite *HandlerTestSuite) SetupTest() {
	checkTx := false
	app := simapp.Setup(checkTx)

	suite.handler = vesting.NewHandler(app.AccountKeeper, app.BankKeeper, app.StakingKeeper)
	suite.app = app
}

func (suite *HandlerTestSuite) TestMsgCreateVestingAccount() {
	ctx := suite.app.BaseApp.NewContext(false, tmproto.Header{Height: suite.app.LastBlockHeight() + 1})

	balances := sdk.NewCoins(sdk.NewInt64Coin("test", 1000))
	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	addr3 := sdk.AccAddress([]byte("addr3_______________"))

	acc1 := suite.app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	suite.app.AccountKeeper.SetAccount(ctx, acc1)
	suite.Require().NoError(simapp.FundAccount(suite.app, ctx, addr1, balances))

	testCases := []struct {
		name      string
		msg       *types.MsgCreateVestingAccount
		expectErr bool
	}{
		{
			name:      "create delayed vesting account",
			msg:       types.NewMsgCreateVestingAccount(addr1, addr2, sdk.NewCoins(sdk.NewInt64Coin("test", 100)), ctx.BlockTime().Unix()+10000, true),
			expectErr: false,
		},
		{
			name:      "create continuous vesting account",
			msg:       types.NewMsgCreateVestingAccount(addr1, addr3, sdk.NewCoins(sdk.NewInt64Coin("test", 100)), ctx.BlockTime().Unix()+10000, false),
			expectErr: false,
		},
		{
			name:      "continuous vesting account already exists",
			msg:       types.NewMsgCreateVestingAccount(addr1, addr3, sdk.NewCoins(sdk.NewInt64Coin("test", 100)), ctx.BlockTime().Unix()+10000, false),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			res, err := suite.handler(ctx, tc.msg)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)

				toAddr, err := sdk.AccAddressFromBech32(tc.msg.ToAddress)
				suite.Require().NoError(err)
				accI := suite.app.AccountKeeper.GetAccount(ctx, toAddr)
				suite.Require().NotNil(accI)

				if tc.msg.Delayed {
					acc, ok := accI.(*types.DelayedVestingAccount)
					suite.Require().True(ok)
					suite.Require().Equal(tc.msg.Amount, acc.GetVestingCoins(ctx.BlockTime()))
				} else {
					acc, ok := accI.(*types.ContinuousVestingAccount)
					suite.Require().True(ok)
					suite.Require().Equal(tc.msg.Amount, acc.GetVestingCoins(ctx.BlockTime()))
				}
			}
		})
	}
}

func (suite *HandlerTestSuite) TestMsgCreatePeriodicVestingAccount() {
	ctx := suite.app.BaseApp.NewContext(false, tmproto.Header{Height: suite.app.LastBlockHeight() + 1})

	balances := sdk.NewCoins(sdk.NewInt64Coin("test", 1000))
	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr3 := sdk.AccAddress([]byte("addr3_______________"))

	acc1 := suite.app.AccountKeeper.NewAccountWithAddress(ctx, addr1)

	period := []types.Period{{Length: 5000, Amount: balances}}
	suite.app.AccountKeeper.SetAccount(ctx, acc1)
	suite.Require().NoError(simapp.FundAccount(suite.app.BankKeeper, ctx, addr1, balances))

	testCases := []struct {
		name      string
		msg       *types.MsgCreatePeriodicVestingAccount
		expectErr bool
	}{
		{
			name:      "create periodic vesting account",
			msg:       types.NewMsgCreatePeriodicVestingAccount(addr1, addr3, 0, period, false),
			expectErr: false,
		},
		{
			name: "bad from addr",
			msg: &types.MsgCreatePeriodicVestingAccount{
				FromAddress:    "foo",
				ToAddress:      addr3.String(),
				StartTime:      0,
				VestingPeriods: period,
			},
			expectErr: true,
		},
		{
			name: "bad to addr",
			msg: &types.MsgCreatePeriodicVestingAccount{
				FromAddress:    addr1.String(),
				ToAddress:      "foo",
				StartTime:      0,
				VestingPeriods: period,
			},
			expectErr: true,
		},
		{
			name:      "account exists",
			msg:       types.NewMsgCreatePeriodicVestingAccount(addr1, addr1, 0, period, false),
			expectErr: true,
		},
		{
			name:      "account exists - try merge",
			msg:       types.NewMsgCreatePeriodicVestingAccount(addr1, addr3, 0, period, false),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			res, err := suite.handler(ctx, tc.msg)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)

				toAddr, err := sdk.AccAddressFromBech32(tc.msg.ToAddress)

				suite.Require().NoError(err)
				fromAddr, err := sdk.AccAddressFromBech32(tc.msg.FromAddress)
				suite.Require().NoError(err)

				accI := suite.app.AccountKeeper.GetAccount(ctx, toAddr)
				suite.Require().NotNil(accI)
				suite.Require().IsType(&types.PeriodicVestingAccount{}, accI)
				balanceSource := suite.app.BankKeeper.GetBalance(ctx, fromAddr, "test")
				suite.Require().Equal(balanceSource, sdk.NewInt64Coin("test", 0))
				balanceDest := suite.app.BankKeeper.GetBalance(ctx, toAddr, "test")
				suite.Require().Equal(balanceDest, sdk.NewInt64Coin("test", 1000))

			}
		})
	}
}

func (suite *HandlerTestSuite) TestMsgCreatePeriodicVestingAccount_Merge() {
	tst := func(amt int64) sdk.Coin {
		return sdk.NewInt64Coin("test", amt)
	}
	ctx := suite.app.BaseApp.NewContext(false, tmproto.Header{Height: suite.app.LastBlockHeight() + 1})

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	addr3 := sdk.AccAddress([]byte("addr3_______________"))
	addr4 := sdk.AccAddress([]byte("addr4_______________"))

	// Create the funding account
	acc1 := suite.app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	suite.app.AccountKeeper.SetAccount(ctx, acc1)
	suite.Require().NoError(simapp.FundAccount(suite.app.BankKeeper, ctx, addr1, sdk.NewCoins(tst(1000))))

	// Create a normal account - cannot merge into it
	acc2 := suite.app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	suite.app.AccountKeeper.SetAccount(ctx, acc2)
	periods := []types.Period{
		{Length: 1000, Amount: sdk.NewCoins(tst(60))},
		{Length: 1000, Amount: sdk.NewCoins(tst(40))},
	}
	res, err := suite.handler(ctx, types.NewMsgCreatePeriodicVestingAccount(addr1, addr2, 0, periods, true))
	suite.Require().Nil(res, "want nil result when merging with non-periodic vesting account")
	suite.Require().Error(err, "want failure when merging with non-periodic vesting account")
	funderBalance := suite.app.BankKeeper.GetBalance(ctx, addr1, "test")
	suite.Require().Equal(funderBalance, tst(1000))

	// Create a PVA normally
	res, err = suite.handler(ctx, types.NewMsgCreatePeriodicVestingAccount(addr1, addr3, 0, periods, false))
	suite.Require().NotNil(res)
	suite.Require().NoError(err)
	acc3 := suite.app.AccountKeeper.GetAccount(ctx, addr3)
	suite.Require().NotNil(acc3)
	suite.Require().IsType(&types.PeriodicVestingAccount{}, acc3)
	funderBalance = suite.app.BankKeeper.GetBalance(ctx, addr1, "test")
	suite.Require().Equal(funderBalance, tst(900))
	balance := suite.app.BankKeeper.GetBalance(ctx, addr3, "test")
	suite.Require().Equal(balance, tst(100))

	// Add new funding to it
	res, err = suite.handler(ctx, types.NewMsgCreatePeriodicVestingAccount(addr1, addr3, 2000, periods, true))
	suite.Require().NotNil(res)
	suite.Require().NoError(err)
	acc3 = suite.app.AccountKeeper.GetAccount(ctx, addr3)
	suite.Require().NotNil(acc3)
	suite.Require().IsType(&types.PeriodicVestingAccount{}, acc3)
	funderBalance = suite.app.BankKeeper.GetBalance(ctx, addr1, "test")
	suite.Require().Equal(funderBalance, tst(800))
	balance = suite.app.BankKeeper.GetBalance(ctx, addr3, "test")
	suite.Require().Equal(balance, tst(200))
	pva := acc3.(*types.PeriodicVestingAccount)
	suite.Require().True(pva.GetVestingCoins(time.Unix(0, 0)).IsEqual(sdk.NewCoins(tst(200))))
	suite.Require().True(pva.GetVestingCoins(time.Unix(1005, 0)).IsEqual(sdk.NewCoins(tst(140))))
	suite.Require().True(pva.GetVestingCoins(time.Unix(2005, 0)).IsEqual(sdk.NewCoins(tst(100))))
	suite.Require().True(pva.GetVestingCoins(time.Unix(3005, 0)).IsEqual(sdk.NewCoins(tst(40))))
	suite.Require().True(pva.GetVestingCoins(time.Unix(4005, 0)).IsEqual(sdk.NewCoins(tst(0))))

	// Can create a new periodic vesting account using merge flag too
	acc4 := suite.app.AccountKeeper.GetAccount(ctx, addr4)
	suite.Require().Nil(acc4)
	res, err = suite.handler(ctx, types.NewMsgCreatePeriodicVestingAccount(addr1, addr4, 0, periods, true))
	suite.Require().NotNil(res)
	suite.Require().NoError(err)
	acc4 = suite.app.AccountKeeper.GetAccount(ctx, addr4)
	suite.Require().NotNil(acc4)
	suite.Require().IsType(&types.PeriodicVestingAccount{}, acc4)
	funderBalance = suite.app.BankKeeper.GetBalance(ctx, addr1, "test")
	suite.Require().Equal(funderBalance, tst(700))
	balance = suite.app.BankKeeper.GetBalance(ctx, addr4, "test")
	suite.Require().Equal(balance, tst(100))
}

func (suite *HandlerTestSuite) TestMsgCreateClawbackVestingAccount() {
	ctx := suite.app.BaseApp.NewContext(false, tmproto.Header{Height: suite.app.LastBlockHeight() + 1})

	balances := sdk.NewCoins(sdk.NewInt64Coin("test", 1000))
	quarter := sdk.NewCoins(sdk.NewInt64Coin("test", 250))
	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	addr3 := sdk.AccAddress([]byte("addr3_______________"))
	addr4 := sdk.AccAddress([]byte("addr4_______________"))

	acc1 := suite.app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	suite.app.AccountKeeper.SetAccount(ctx, acc1)

	lockupPeriods := []types.Period{{Length: 5000, Amount: balances}}
	vestingPeriods := []types.Period{
		{Length: 2000, Amount: quarter},
		{Length: 2000, Amount: quarter},
		{Length: 2000, Amount: quarter},
		{Length: 2000, Amount: quarter},
	}

	testCases := []struct {
		name               string
		msg                *types.MsgCreateClawbackVestingAccount
		expectErr          bool
		expectExtraBalance int64
	}{
		{
			name:      "create clawback vesting account",
			msg:       types.NewMsgCreateClawbackVestingAccount(addr1, addr2, 0, lockupPeriods, vestingPeriods, false),
			expectErr: false,
		},
		{
			name: "bad from addr",
			msg: &types.MsgCreateClawbackVestingAccount{
				FromAddress:    "foo",
				ToAddress:      addr2.String(),
				StartTime:      0,
				LockupPeriods:  lockupPeriods,
				VestingPeriods: vestingPeriods,
			},
			expectErr: true,
		},
		{
			name: "bad to addr",
			msg: &types.MsgCreateClawbackVestingAccount{
				FromAddress:    addr1.String(),
				ToAddress:      "foo",
				StartTime:      0,
				LockupPeriods:  lockupPeriods,
				VestingPeriods: vestingPeriods,
			},
			expectErr: true,
		},
		{
			name:      "default lockup",
			msg:       types.NewMsgCreateClawbackVestingAccount(addr1, addr3, 0, nil, vestingPeriods, false),
			expectErr: false,
		},
		{
			name:      "default vesting",
			msg:       types.NewMsgCreateClawbackVestingAccount(addr1, addr4, 0, lockupPeriods, nil, false),
			expectErr: false,
		},
		{
			name:      "different amounts",
			msg:       types.NewMsgCreateClawbackVestingAccount(addr1, addr2, 0, []types.Period{{Length: 5000, Amount: quarter}}, vestingPeriods, false),
			expectErr: true,
		},
		{
			name:      "account exists",
			msg:       types.NewMsgCreateClawbackVestingAccount(addr1, addr1, 0, lockupPeriods, vestingPeriods, false),
			expectErr: true,
		},
		{
			name:      "account exists no merge",
			msg:       types.NewMsgCreateClawbackVestingAccount(addr1, addr2, 0, lockupPeriods, vestingPeriods, false),
			expectErr: true,
		},
		{
			name:      "account exists merge not clawback",
			msg:       types.NewMsgCreateClawbackVestingAccount(addr1, addr1, 0, lockupPeriods, vestingPeriods, true),
			expectErr: true,
		},
		{
			name:               "account exists merge",
			msg:                types.NewMsgCreateClawbackVestingAccount(addr1, addr2, 0, lockupPeriods, vestingPeriods, true),
			expectErr:          false,
			expectExtraBalance: 1000,
		},
		{
			name:      "merge wrong funder",
			msg:       types.NewMsgCreateClawbackVestingAccount(addr2, addr2, 0, lockupPeriods, vestingPeriods, true),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			if !tc.expectErr {
				suite.Require().NoError(simapp.FundAccount(suite.app.BankKeeper, ctx, addr1, balances))
			}
			res, err := suite.handler(ctx, tc.msg)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)

				toAddr, err := sdk.AccAddressFromBech32(tc.msg.ToAddress)

				suite.Require().NoError(err)
				fromAddr, err := sdk.AccAddressFromBech32(tc.msg.FromAddress)
				suite.Require().NoError(err)

				accI := suite.app.AccountKeeper.GetAccount(ctx, toAddr)
				suite.Require().NotNil(accI)
				suite.Require().IsType(&types.ClawbackVestingAccount{}, accI)
				balanceSource := suite.app.BankKeeper.GetBalance(ctx, fromAddr, "test")
				suite.Require().Equal(balanceSource, sdk.NewInt64Coin("test", 0))
				balanceDest := suite.app.BankKeeper.GetBalance(ctx, toAddr, "test")
				suite.Require().Equal(balanceDest, sdk.NewInt64Coin("test", 1000+tc.expectExtraBalance))

			}
		})
	}
}

func (suite *HandlerTestSuite) TestMsgClawback() {
	ctx := suite.app.BaseApp.NewContext(false, tmproto.Header{Height: suite.app.LastBlockHeight() + 1})

	balances := sdk.NewCoins(sdk.NewInt64Coin("test", 1000))
	quarter := sdk.NewCoins(sdk.NewInt64Coin("test", 250))
	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	addr3 := sdk.AccAddress([]byte("addr3_______________"))
	addr4 := sdk.AccAddress([]byte("addr4_______________"))

	funder := suite.app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	suite.app.AccountKeeper.SetAccount(ctx, funder)
	suite.Require().NoError(simapp.FundAccount(suite.app.BankKeeper, ctx, addr1, balances))

	lockupPeriods := []types.Period{{Length: 5000, Amount: balances}}
	vestingPeriods := []types.Period{
		{Length: 2000, Amount: quarter},
		{Length: 2000, Amount: quarter},
		{Length: 2000, Amount: quarter},
		{Length: 2000, Amount: quarter},
	}

	createMsg := types.NewMsgCreateClawbackVestingAccount(addr1, addr2, 0, lockupPeriods, vestingPeriods, false)
	res, err := suite.handler(ctx, createMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	balanceDest := suite.app.BankKeeper.GetBalance(ctx, addr2, "test")
	suite.Require().Equal(balanceDest, sdk.NewInt64Coin("test", 1000))

	clawbackMsg := types.NewMsgClawback(addr1, addr2, addr3)
	clawRes, err := suite.handler(ctx, clawbackMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(clawRes)

	balanceDest = suite.app.BankKeeper.GetBalance(ctx, addr2, "test")
	suite.Require().Equal(balanceDest, sdk.NewInt64Coin("test", 0))
	balanceClaw := suite.app.BankKeeper.GetBalance(ctx, addr3, "test")
	suite.Require().Equal(balanceClaw, sdk.NewInt64Coin("test", 1000))

	// test bad messages

	// bad funder
	clawbackMsg = &types.MsgClawback{
		FunderAddress: "foo",
		Address:       addr2.String(),
		DestAddress:   addr3.String(),
	}
	_, err = suite.handler(ctx, clawbackMsg)
	suite.Require().Error(err)

	// bad addr
	clawbackMsg = &types.MsgClawback{
		FunderAddress: addr1.String(),
		Address:       "foo",
		DestAddress:   addr3.String(),
	}
	_, err = suite.handler(ctx, clawbackMsg)
	suite.Require().Error(err)

	// bad dest
	clawbackMsg = &types.MsgClawback{
		FunderAddress: addr1.String(),
		Address:       addr2.String(),
		DestAddress:   "foo",
	}
	_, err = suite.handler(ctx, clawbackMsg)
	suite.Require().Error(err)

	// no account
	clawbackMsg = types.NewMsgClawback(addr1, addr4, addr3)
	_, err = suite.handler(ctx, clawbackMsg)
	suite.Require().Error(err)

	// wrong account type
	clawbackMsg = types.NewMsgClawback(addr1, addr3, addr4)
	_, err = suite.handler(ctx, clawbackMsg)
	suite.Require().Error(err)

	// wrong funder
	clawbackMsg = types.NewMsgClawback(addr4, addr2, addr3)
	_, err = suite.handler(ctx, clawbackMsg)
	suite.Require().Error(err)
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
