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

func (suite *KeeperTestSuite) TestCreateContinuousFund() {
	cap := sdk.NewInt64Coin("uatom", 100000)
	percentage, err := math.LegacyNewDecFromStr("0.2")
	suite.Require().NoError(err)
	negativePercentage, err := math.LegacyNewDecFromStr("-0.2")
	suite.Require().NoError(err)
	invalidCap := sdk.NewInt64Coin("foo", 0)
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
				Title:       "New title",
				Description: "New description",
				Authority:   suite.poolKeeper.GetAuthority(),
				Recipient:   "",
				Metadata:    "AQ==",
				Percentage:  percentage,
				Cap:         &cap,
				Expiry:      &expiry,
			},
			expErr:    true,
			expErrMsg: "empty address string is not allowed",
		},
		"empty authority": {
			input: &types.MsgCreateContinuousFund{
				Title:       "New title",
				Description: "New description",
				Authority:   "",
				Recipient:   recipientAddr.String(),
				Metadata:    "AQ==",
				Percentage:  percentage,
				Cap:         &cap,
				Expiry:      &expiry,
			},
			expErr:    true,
			expErrMsg: "empty address string is not allowed",
		},
		"invalid authority": {
			input: &types.MsgCreateContinuousFund{
				Title:       "New title",
				Description: "New description",
				Authority:   "invalid_authority",
				Recipient:   recipientAddr.String(),
				Metadata:    "AQ==",
				Percentage:  percentage,
				Cap:         &cap,
				Expiry:      &expiry,
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		"empty title": {
			input: &types.MsgCreateContinuousFund{
				Title:       "",
				Description: "New description",
				Authority:   suite.poolKeeper.GetAuthority(),
				Recipient:   recipientAddr.String(),
				Metadata:    "AQ==",
				Percentage:  percentage,
				Cap:         &cap,
				Expiry:      &expiry,
			},
			expErr:    true,
			expErrMsg: "title cannot be empty",
		},
		"empty description": {
			input: &types.MsgCreateContinuousFund{
				Title:       "New title",
				Description: "",
				Authority:   suite.poolKeeper.GetAuthority(),
				Recipient:   recipientAddr.String(),
				Metadata:    "AQ==",
				Percentage:  percentage,
				Cap:         &cap,
				Expiry:      &expiry,
			},
			expErr:    true,
			expErrMsg: "description cannot be empty",
		},
		"zero percentage": {
			input: &types.MsgCreateContinuousFund{
				Title:       "New title",
				Description: "New description",
				Authority:   suite.poolKeeper.GetAuthority(),
				Recipient:   recipientAddr.String(),
				Metadata:    "AQ==",
				Percentage:  math.LegacyNewDec(0),
				Cap:         &cap,
				Expiry:      &expiry,
			},
			expErr:    true,
			expErrMsg: "percentage cannot be zero or empty",
		},
		"negative percentage": {
			input: &types.MsgCreateContinuousFund{
				Title:       "New title",
				Description: "New description",
				Authority:   suite.poolKeeper.GetAuthority(),
				Recipient:   recipientAddr.String(),
				Metadata:    "AQ==",
				Percentage:  negativePercentage,
				Cap:         &cap,
				Expiry:      &expiry,
			},
			expErr:    true,
			expErrMsg: "percentage cannot be negative",
		},
		"invalid percentage": {
			input: &types.MsgCreateContinuousFund{
				Title:       "New title",
				Description: "New description",
				Authority:   suite.poolKeeper.GetAuthority(),
				Recipient:   recipientAddr.String(),
				Metadata:    "AQ==",
				Percentage:  math.LegacyNewDec(1),
				Cap:         &cap,
				Expiry:      &expiry,
			},
			expErr:    true,
			expErrMsg: "percentage cannot be greater than or equal to one",
		},
		"invalid cap": {
			input: &types.MsgCreateContinuousFund{
				Title:       "New title",
				Description: "New description",
				Authority:   suite.poolKeeper.GetAuthority(),
				Recipient:   recipientAddr.String(),
				Metadata:    "AQ==",
				Percentage:  percentage,
				Cap:         &invalidCap,
				Expiry:      &expiry,
			},
			expErr:    true,
			expErrMsg: "invalid capital: amount cannot be zero",
		},
		"invalid expiry": {
			input: &types.MsgCreateContinuousFund{
				Title:       "New title",
				Description: "New description",
				Authority:   suite.poolKeeper.GetAuthority(),
				Recipient:   recipientAddr.String(),
				Metadata:    "AQ==",
				Percentage:  percentage,
				Cap:         &cap,
				Expiry:      &invalidExpirty,
			},
			expErr:    true,
			expErrMsg: "expiry time cannot be less than the current block time",
		},
		"all good": {
			input: &types.MsgCreateContinuousFund{
				Title:       "New title",
				Description: "New description",
				Authority:   suite.poolKeeper.GetAuthority(),
				Recipient:   recipientAddr.String(),
				Metadata:    "AQ==",
				Percentage:  percentage,
				Cap:         &cap,
				Expiry:      &expiry,
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
				cap := sdk.NewInt64Coin("uatom", 100000)
				oneMonthInSeconds := int64(30 * 24 * 60 * 60) // Approximate number of seconds in 1 month
				expiry := suite.ctx.BlockTime().Add(time.Duration(oneMonthInSeconds) * time.Second)
				cf := types.ContinuousFund{
					Title:       "New Title",
					Description: "New description",
					Recipient:   "",
					Metadata:    "AQ==",
					Percentage:  percentage,
					Cap:         &cap,
					Expiry:      &expiry,
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
				cap := sdk.NewInt64Coin("uatom", 100000)
				oneMonthInSeconds := int64(30 * 24 * 60 * 60) // Approximate number of seconds in 1 month
				expiry := suite.ctx.BlockTime().Add(time.Duration(oneMonthInSeconds) * time.Second)
				cf := types.ContinuousFund{
					Title:       "New Title",
					Description: "New description",
					Recipient:   recipientAddr.String(),
					Metadata:    "AQ==",
					Percentage:  percentage,
					Cap:         &cap,
					Expiry:      &expiry,
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
