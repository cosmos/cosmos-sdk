package keeper_test

import (
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/header"
	"cosmossdk.io/math"
	"cosmossdk.io/x/protocolpool/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	recipientAddr = sdk.AccAddress([]byte("to1__________________"))

	fooCoin  = sdk.NewInt64Coin("foo", 100)
	fooCoin2 = sdk.NewInt64Coin("foo", 50)
)

func (suite *KeeperTestSuite) TestMsgSubmitBudgetProposal() {
	invalidCoin := sdk.NewInt64Coin("foo", 0)
	startTime := suite.ctx.BlockTime().Add(10 * time.Second)
	invalidStartTime := suite.ctx.BlockTime().Add(-15 * time.Second)
	period := time.Duration(60) * time.Second
	zeroPeriod := time.Duration(0) * time.Second
	testCases := map[string]struct {
		input     *types.MsgSubmitBudgetProposal
		expErr    bool
		expErrMsg string
	}{
		"empty recipient address": {
			input: &types.MsgSubmitBudgetProposal{
				Authority:        suite.poolKeeper.GetAuthority(),
				RecipientAddress: "",
				TotalBudget:      &fooCoin,
				StartTime:        &startTime,
				Tranches:         2,
				Period:           &period,
			},
			expErr:    true,
			expErrMsg: "empty address string is not allowed",
		},
		"empty authority": {
			input: &types.MsgSubmitBudgetProposal{
				Authority:        "",
				RecipientAddress: recipientAddr.String(),
				TotalBudget:      &fooCoin,
				StartTime:        &startTime,
				Tranches:         2,
				Period:           &period,
			},
			expErr:    true,
			expErrMsg: "empty address string is not allowed",
		},
		"invalid authority": {
			input: &types.MsgSubmitBudgetProposal{
				Authority:        "invalid_authority",
				RecipientAddress: recipientAddr.String(),
				TotalBudget:      &fooCoin,
				StartTime:        &startTime,
				Tranches:         2,
				Period:           &period,
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		"invalid budget": {
			input: &types.MsgSubmitBudgetProposal{
				Authority:        suite.poolKeeper.GetAuthority(),
				RecipientAddress: recipientAddr.String(),
				TotalBudget:      &invalidCoin,
				StartTime:        &startTime,
				Tranches:         2,
				Period:           &period,
			},
			expErr:    true,
			expErrMsg: "total budget cannot be zero",
		},
		"invalid start time": {
			input: &types.MsgSubmitBudgetProposal{
				Authority:        suite.poolKeeper.GetAuthority(),
				RecipientAddress: recipientAddr.String(),
				TotalBudget:      &fooCoin,
				StartTime:        &invalidStartTime,
				Tranches:         2,
				Period:           &period,
			},
			expErr:    true,
			expErrMsg: "start time cannot be less than the current block time",
		},
		"invalid tranches": {
			input: &types.MsgSubmitBudgetProposal{
				Authority:        suite.poolKeeper.GetAuthority(),
				RecipientAddress: recipientAddr.String(),
				TotalBudget:      &fooCoin,
				StartTime:        &startTime,
				Tranches:         0,
				Period:           &period,
			},
			expErr:    true,
			expErrMsg: "tranches must be greater than zero",
		},
		"invalid period": {
			input: &types.MsgSubmitBudgetProposal{
				Authority:        suite.poolKeeper.GetAuthority(),
				RecipientAddress: recipientAddr.String(),
				TotalBudget:      &fooCoin,
				StartTime:        &startTime,
				Tranches:         2,
				Period:           &zeroPeriod,
			},
			expErr:    true,
			expErrMsg: "period length should be greater than zero",
		},
		"all good": {
			input: &types.MsgSubmitBudgetProposal{
				Authority:        suite.poolKeeper.GetAuthority(),
				RecipientAddress: recipientAddr.String(),
				TotalBudget:      &fooCoin,
				StartTime:        &startTime,
				Tranches:         2,
				Period:           &period,
			},
			expErr: false,
		},
	}

	for name, tc := range testCases {
		suite.Run(name, func() {
			suite.SetupTest()

			_, err := suite.msgServer.SubmitBudgetProposal(suite.ctx, tc.input)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestMsgClaimBudget() {
	startTime := suite.ctx.BlockTime().Add(-70 * time.Second)
	period := time.Duration(60) * time.Second

	testCases := map[string]struct {
		preRun           func()
		recipientAddress sdk.AccAddress
		expErr           bool
		expErrMsg        string
		claimableFunds   sdk.Coin
	}{
		"empty recipient addr": {
			recipientAddress: sdk.AccAddress(""),
			expErr:           true,
			expErrMsg:        "invalid recipient address: empty address string is not allowed",
		},
		"no budget found": {
			recipientAddress: sdk.AccAddress([]byte("acc1__________")),
			expErr:           true,
			expErrMsg:        "no budget found for recipient",
		},
		"claiming before start time": {
			preRun: func() {
				startTime := suite.ctx.BlockTime().Add(3600 * time.Second)
				// Prepare the budget proposal with a future start time
				budget := types.Budget{
					RecipientAddress: recipientAddr.String(),
					TotalBudget:      &fooCoin,
					StartTime:        &startTime,
					Tranches:         2,
					Period:           &period,
				}
				err := suite.poolKeeper.BudgetProposal.Set(suite.ctx, recipientAddr, budget)
				suite.Require().NoError(err)
			},
			recipientAddress: recipientAddr,
			expErr:           true,
			expErrMsg:        "distribution has not started yet",
		},
		"budget period has not passed": {
			preRun: func() {
				startTime := suite.ctx.BlockTime().Add(-50 * time.Second)
				// Prepare the budget proposal with start time and a short period
				budget := types.Budget{
					RecipientAddress: recipientAddr.String(),
					TotalBudget:      &fooCoin,
					StartTime:        &startTime,
					Tranches:         1,
					Period:           &period,
				}
				err := suite.poolKeeper.BudgetProposal.Set(suite.ctx, recipientAddr, budget)
				suite.Require().NoError(err)
			},
			recipientAddress: recipientAddr,
			expErr:           true,
			expErrMsg:        "budget period has not passed yet",
		},
		"valid claim": {
			preRun: func() {
				// Prepare the budget proposal with valid start time and period
				budget := types.Budget{
					RecipientAddress: recipientAddr.String(),
					TotalBudget:      &fooCoin,
					StartTime:        &startTime,
					Tranches:         2,
					Period:           &period,
				}
				err := suite.poolKeeper.BudgetProposal.Set(suite.ctx, recipientAddr, budget)
				suite.Require().NoError(err)
			},
			recipientAddress: recipientAddr,
			expErr:           false,
			claimableFunds:   sdk.NewInt64Coin("foo", 50),
		},
		"double claim attempt with budget period not passed": {
			preRun: func() {
				// Prepare the budget proposal with valid start time and period
				budget := types.Budget{
					RecipientAddress: recipientAddr.String(),
					TotalBudget:      &fooCoin,
					StartTime:        &startTime,
					Tranches:         2,
					Period:           &period,
				}
				err := suite.poolKeeper.BudgetProposal.Set(suite.ctx, recipientAddr, budget)
				suite.Require().NoError(err)

				// Claim the funds once
				msg := &types.MsgClaimBudget{
					RecipientAddress: recipientAddr.String(),
				}
				suite.mockSendCoinsFromModuleToAccount(recipientAddr)
				_, err = suite.msgServer.ClaimBudget(suite.ctx, msg)
				suite.Require().NoError(err)
			},
			recipientAddress: recipientAddr,
			expErr:           true,
			expErrMsg:        "budget period has not passed yet",
		},
		"valid double claim attempt": {
			preRun: func() {
				oneMonthInSeconds := int64(30 * 24 * 60 * 60) // Approximate number of seconds in 1 month
				startTimeBeforeMonth := suite.ctx.BlockTime().Add(time.Duration(-oneMonthInSeconds) * time.Second)
				oneMonthPeriod := time.Duration(oneMonthInSeconds) * time.Second
				// Prepare the budget proposal with valid start time and period of 1 month (in seconds)
				budget := types.Budget{
					RecipientAddress: recipientAddr.String(),
					TotalBudget:      &fooCoin,
					StartTime:        &startTimeBeforeMonth,
					Tranches:         2,
					Period:           &oneMonthPeriod,
				}
				err := suite.poolKeeper.BudgetProposal.Set(suite.ctx, recipientAddr, budget)
				suite.Require().NoError(err)

				// Claim the funds once
				msg := &types.MsgClaimBudget{
					RecipientAddress: recipientAddr.String(),
				}
				suite.mockSendCoinsFromModuleToAccount(recipientAddr)
				_, err = suite.msgServer.ClaimBudget(suite.ctx, msg)
				suite.Require().NoError(err)

				// Create a new context with an updated block time to simulate a delay
				newBlockTime := suite.ctx.BlockTime().Add(time.Duration(oneMonthInSeconds) * time.Second)
				suite.ctx = suite.ctx.WithHeaderInfo(header.Info{
					Time: newBlockTime,
				})
			},
			recipientAddress: recipientAddr,
			expErr:           false,
			claimableFunds:   sdk.NewInt64Coin("foo", 50),
		},
		"budget ended for recipient": {
			preRun: func() {
				// Prepare the budget proposal with valid start time and period
				budget := types.Budget{
					RecipientAddress: recipientAddr.String(),
					TotalBudget:      &fooCoin,
					StartTime:        &startTime,
					Tranches:         2,
					Period:           &period,
				}
				err := suite.poolKeeper.BudgetProposal.Set(suite.ctx, recipientAddr, budget)
				suite.Require().NoError(err)

				// Claim the funds once
				msg := &types.MsgClaimBudget{
					RecipientAddress: recipientAddr.String(),
				}
				suite.mockSendCoinsFromModuleToAccount(recipientAddr)
				_, err = suite.msgServer.ClaimBudget(suite.ctx, msg)
				suite.Require().NoError(err)

				// Create a new context with an updated block time to simulate a delay
				newBlockTime := suite.ctx.BlockTime().Add(60 * time.Second)
				suite.ctx = suite.ctx.WithHeaderInfo(header.Info{
					Time: newBlockTime,
				})

				// Claim the funds twice
				msg = &types.MsgClaimBudget{
					RecipientAddress: recipientAddr.String(),
				}
				suite.mockSendCoinsFromModuleToAccount(recipientAddr)
				_, err = suite.msgServer.ClaimBudget(suite.ctx, msg)
				suite.Require().NoError(err)
			},
			recipientAddress: recipientAddr,
			expErr:           true,
			expErrMsg:        "budget ended for recipient",
		},
	}

	for name, tc := range testCases {
		suite.Run(name, func() {
			suite.SetupTest()
			if tc.preRun != nil {
				tc.preRun()
			}
			msg := &types.MsgClaimBudget{
				RecipientAddress: tc.recipientAddress.String(),
			}
			suite.mockSendCoinsFromModuleToAccount(tc.recipientAddress)
			resp, err := suite.msgServer.ClaimBudget(suite.ctx, msg)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.claimableFunds, resp.Amount)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestWithdrawContinuousFund() {
	recipient := sdk.AccAddress([]byte("recipientAddr1__________________"))
	recipient2 := sdk.AccAddress([]byte("recipientAddr2___________________"))
	testCases := map[string]struct {
		preRun           func()
		recipientAddress sdk.AccAddress
		expErr           bool
		expErrMsg        string
		withdrawnAmount  sdk.Coin
	}{
		"empty recipient": {
			recipientAddress: sdk.AccAddress([]byte("")),
			expErr:           true,
			expErrMsg:        "invalid recipient address",
		},
		"recipient with no continuous fund": {
			recipientAddress: recipient,
			expErr:           true,
			expErrMsg:        "no continuous fund found for recipient",
		},
		"funds percentage > 100": {
			preRun: func() {
				// Set fund 1
				percentage, err := math.LegacyNewDecFromStr("0.2")
				suite.Require().NoError(err)
				oneMonthInSeconds := int64(30 * 24 * 60 * 60) // Approximate number of seconds in 1 month
				expiry := suite.ctx.BlockTime().Add(time.Duration(oneMonthInSeconds) * time.Second)
				cf := types.ContinuousFund{
					Recipient:  recipient.String(),
					Percentage: percentage,
					Expiry:     &expiry,
				}
				// Set continuous fund
				err = suite.poolKeeper.ContinuousFund.Set(suite.ctx, recipient, cf)
				suite.Require().NoError(err)
				// Set recipient fund percentage and recipient fund distribution
				intPercentage := percentage.MulInt64(100)
				err = suite.poolKeeper.RecipientFundPercentage.Set(suite.ctx, recipient, intPercentage.TruncateInt())
				suite.Require().NoError(err)
				err = suite.poolKeeper.RecipientFundDistribution.Set(suite.ctx, recipient, math.ZeroInt())
				suite.Require().NoError(err)

				// Set fund 2
				percentage, err = math.LegacyNewDecFromStr("0.9")
				suite.Require().NoError(err)
				cf = types.ContinuousFund{
					Recipient:  recipient2.String(),
					Percentage: percentage,
					Expiry:     &expiry,
				}
				// Set continuous fund
				err = suite.poolKeeper.ContinuousFund.Set(suite.ctx, recipient2, cf)
				suite.Require().NoError(err)
				// Set recipient fund percentage and recipient fund distribution
				intPercentage = percentage.MulInt64(100)
				err = suite.poolKeeper.RecipientFundPercentage.Set(suite.ctx, recipient2, intPercentage.TruncateInt())
				suite.Require().NoError(err)
				err = suite.poolKeeper.RecipientFundDistribution.Set(suite.ctx, recipient2, math.ZeroInt())
				suite.Require().NoError(err)

				// Set ToDistribute
				toDistribute := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100000)))
				err = suite.poolKeeper.SetToDistribute(suite.ctx, toDistribute)
				suite.Require().NoError(err)
			},
			recipientAddress: recipient,
			expErr:           true,
			expErrMsg:        "error while iterating all the continuous funds: total funds percentage cannot exceed 100",
		},
		"expired case": {
			preRun: func() {
				percentage, err := math.LegacyNewDecFromStr("0.2")
				suite.Require().NoError(err)
				expiry := suite.ctx.BlockTime().Add(time.Duration(-1) * time.Second)
				cf := types.ContinuousFund{
					Recipient:  recipient.String(),
					Percentage: percentage,
					Expiry:     &expiry,
				}
				// Set continuous fund
				err = suite.poolKeeper.ContinuousFund.Set(suite.ctx, recipient, cf)
				suite.Require().NoError(err)
			},
			recipientAddress: recipient,
			expErr:           true,
			expErrMsg:        "cannot withdraw continuous funds\ncontinuous fund expired for recipient",
		},
		"valid case with ToDistribute amount zero": {
			preRun: func() {
				percentage, err := math.LegacyNewDecFromStr("0.2")
				suite.Require().NoError(err)
				oneMonthInSeconds := int64(30 * 24 * 60 * 60) // Approximate number of seconds in 1 month
				expiry := suite.ctx.BlockTime().Add(time.Duration(oneMonthInSeconds) * time.Second)
				cf := types.ContinuousFund{
					Recipient:  recipient.String(),
					Percentage: percentage,
					Expiry:     &expiry,
				}
				// Set continuous fund
				err = suite.poolKeeper.ContinuousFund.Set(suite.ctx, recipient, cf)
				suite.Require().NoError(err)
				// Set recipient fund percentage and recipient fund distribution
				intPercentage := percentage.MulInt64(100)
				err = suite.poolKeeper.RecipientFundPercentage.Set(suite.ctx, recipient, intPercentage.TruncateInt())
				suite.Require().NoError(err)
				err = suite.poolKeeper.RecipientFundDistribution.Set(suite.ctx, recipient, math.ZeroInt())
				suite.Require().NoError(err)
				err = suite.poolKeeper.ToDistribute.Set(suite.ctx, math.ZeroInt())
				suite.Require().NoError(err)
			},
			recipientAddress: recipient,
			expErr:           false,
			withdrawnAmount:  sdk.Coin{},
		},
		"valid case": {
			preRun: func() {
				percentage, err := math.LegacyNewDecFromStr("0.2")
				suite.Require().NoError(err)
				oneMonthInSeconds := int64(30 * 24 * 60 * 60) // Approximate number of seconds in 1 month
				expiry := suite.ctx.BlockTime().Add(time.Duration(oneMonthInSeconds) * time.Second)
				cf := types.ContinuousFund{
					Recipient:  recipient.String(),
					Percentage: percentage,
					Expiry:     &expiry,
				}
				// Set continuous fund
				err = suite.poolKeeper.ContinuousFund.Set(suite.ctx, recipient, cf)
				suite.Require().NoError(err)
				// Set recipient fund percentage and recipient fund distribution
				intPercentage := percentage.MulInt64(100)
				err = suite.poolKeeper.RecipientFundPercentage.Set(suite.ctx, recipient, intPercentage.TruncateInt())
				suite.Require().NoError(err)
				err = suite.poolKeeper.RecipientFundDistribution.Set(suite.ctx, recipient, math.ZeroInt())
				suite.Require().NoError(err)
				toDistribute := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100000)))
				err = suite.poolKeeper.SetToDistribute(suite.ctx, toDistribute)
				suite.Require().NoError(err)
			},
			recipientAddress: recipient,
			expErr:           false,
			withdrawnAmount:  sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(20000)),
		},
	}

	for name, tc := range testCases {
		suite.Run(name, func() {
			suite.SetupTest()
			if tc.preRun != nil {
				tc.preRun()
			}
			msg := &types.MsgWithdrawContinuousFund{
				RecipientAddress: tc.recipientAddress.String(),
			}

			suite.mockWithdrawContinuousFund()

			resp, err := suite.msgServer.WithdrawContinuousFund(suite.ctx, msg)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.withdrawnAmount, resp.Amount)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestCreateContinuousFund() {
	percentage, err := math.LegacyNewDecFromStr("0.2")
	suite.Require().NoError(err)
	negativePercentage, err := math.LegacyNewDecFromStr("-0.2")
	suite.Require().NoError(err)
	invalidExpirty := suite.ctx.BlockTime().Add(-15 * time.Second)
	oneMonthInSeconds := int64(30 * 24 * 60 * 60) // Approximate number of seconds in 1 month
	expiry := suite.ctx.BlockTime().Add(time.Duration(oneMonthInSeconds) * time.Second)
	testCases := map[string]struct {
		input     *types.MsgCreateContinuousFund
		expErr    bool
		expErrMsg string
	}{
		"empty recipient address": {
			input: &types.MsgCreateContinuousFund{
				Authority:  suite.poolKeeper.GetAuthority(),
				Recipient:  "",
				Percentage: percentage,
				Expiry:     &expiry,
			},
			expErr:    true,
			expErrMsg: "empty address string is not allowed",
		},
		"empty authority": {
			input: &types.MsgCreateContinuousFund{
				Authority:  "",
				Recipient:  recipientAddr.String(),
				Percentage: percentage,
				Expiry:     &expiry,
			},
			expErr:    true,
			expErrMsg: "empty address string is not allowed",
		},
		"invalid authority": {
			input: &types.MsgCreateContinuousFund{
				Authority:  "invalid_authority",
				Recipient:  recipientAddr.String(),
				Percentage: percentage,
				Expiry:     &expiry,
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		"zero percentage": {
			input: &types.MsgCreateContinuousFund{
				Authority:  suite.poolKeeper.GetAuthority(),
				Recipient:  recipientAddr.String(),
				Percentage: math.LegacyNewDec(0),
				Expiry:     &expiry,
			},
			expErr:    true,
			expErrMsg: "percentage cannot be zero or empty",
		},
		"negative percentage": {
			input: &types.MsgCreateContinuousFund{
				Authority:  suite.poolKeeper.GetAuthority(),
				Recipient:  recipientAddr.String(),
				Percentage: negativePercentage,
				Expiry:     &expiry,
			},
			expErr:    true,
			expErrMsg: "percentage cannot be negative",
		},
		"invalid percentage": {
			input: &types.MsgCreateContinuousFund{
				Authority:  suite.poolKeeper.GetAuthority(),
				Recipient:  recipientAddr.String(),
				Percentage: math.LegacyNewDec(1),
				Expiry:     &expiry,
			},
			expErr:    true,
			expErrMsg: "percentage cannot be greater than or equal to one",
		},
		"invalid expiry": {
			input: &types.MsgCreateContinuousFund{
				Authority:  suite.poolKeeper.GetAuthority(),
				Recipient:  recipientAddr.String(),
				Percentage: percentage,
				Expiry:     &invalidExpirty,
			},
			expErr:    true,
			expErrMsg: "expiry time cannot be less than the current block time",
		},
		"all good": {
			input: &types.MsgCreateContinuousFund{
				Authority:  suite.poolKeeper.GetAuthority(),
				Recipient:  recipientAddr.String(),
				Percentage: percentage,
				Expiry:     &expiry,
			},
			expErr: false,
		},
	}

	for name, tc := range testCases {
		suite.Run(name, func() {
			suite.SetupTest()

			_, err := suite.msgServer.CreateContinuousFund(suite.ctx, tc.input)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestCancelContinuousFund() {
	testCases := map[string]struct {
		preRun        func()
		recipientAddr sdk.AccAddress
		expErr        bool
		expErrMsg     string
		postRun       func()
	}{
		"empty recipient": {
			preRun: func() {
				percentage, err := math.LegacyNewDecFromStr("0.2")
				suite.Require().NoError(err)
				oneMonthInSeconds := int64(30 * 24 * 60 * 60) // Approximate number of seconds in 1 month
				expiry := suite.ctx.BlockTime().Add(time.Duration(oneMonthInSeconds) * time.Second)
				cf := types.ContinuousFund{
					Recipient:  "",
					Percentage: percentage,
					Expiry:     &expiry,
				}
				err = suite.poolKeeper.ContinuousFund.Set(suite.ctx, recipientAddr, cf)
				suite.Require().NoError(err)
			},
			expErr:    true,
			expErrMsg: "empty address string is not allowed",
		},
		"no recipient found": {
			recipientAddr: recipientAddr,
			expErr:        true,
			expErrMsg:     "no recipient found to cancel continuous fund",
		},
		"all good": {
			preRun: func() {
				percentage, err := math.LegacyNewDecFromStr("0.2")
				suite.Require().NoError(err)
				oneMonthInSeconds := int64(30 * 24 * 60 * 60) // Approximate number of seconds in 1 month
				expiry := suite.ctx.BlockTime().Add(time.Duration(oneMonthInSeconds) * time.Second)
				cf := types.ContinuousFund{
					Recipient:  recipientAddr.String(),
					Percentage: percentage,
					Expiry:     &expiry,
				}
				err = suite.poolKeeper.ContinuousFund.Set(suite.ctx, recipientAddr, cf)
				suite.Require().NoError(err)
			},
			recipientAddr: recipientAddr,
			expErr:        false,
			postRun: func() {
				_, err := suite.poolKeeper.ContinuousFund.Get(suite.ctx, recipientAddr)
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, collections.ErrNotFound)
			},
		},
	}

	for name, tc := range testCases {
		suite.Run(name, func() {
			suite.SetupTest()
			if tc.preRun != nil {
				tc.preRun()
			}
			msg := &types.MsgCancelContinuousFund{
				Authority:        suite.poolKeeper.GetAuthority(),
				RecipientAddress: tc.recipientAddr.String(),
			}
			_, err := suite.msgServer.CancelContinuousFund(suite.ctx, msg)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
			}
			if tc.postRun != nil {
				tc.postRun()
			}
		})
	}
}
