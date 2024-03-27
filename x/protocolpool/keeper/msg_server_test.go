package keeper_test

import (
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/header"
	"cosmossdk.io/math"
	"cosmossdk.io/x/protocolpool/types"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	recipientAddr = sdk.AccAddress([]byte("to1__________________"))

	fooCoin  = sdk.NewInt64Coin("foo", 100)
	fooCoin2 = sdk.NewInt64Coin("foo", 50)
)

func (suite *KeeperTestSuite) TestMsgSubmitBudgetProposal() {
	invalidCoin := sdk.NewInt64Coin("foo", 0)
	startTime := suite.environment.HeaderService.GetHeaderInfo(suite.ctx).Time.Add(10 * time.Second)
	invalidStartTime := suite.environment.HeaderService.GetHeaderInfo(suite.ctx).Time.Add(-15 * time.Second)
	period := time.Duration(60) * time.Second
	zeroPeriod := time.Duration(0) * time.Second
	recipientStrAddr, err := codectestutil.CodecOptions{}.GetAddressCodec().BytesToString(recipientAddr)
	suite.Require().NoError(err)
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
				RecipientAddress: recipientStrAddr,
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
				RecipientAddress: recipientStrAddr,
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
				RecipientAddress: recipientStrAddr,
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
				RecipientAddress: recipientStrAddr,
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
				RecipientAddress: recipientStrAddr,
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
				RecipientAddress: recipientStrAddr,
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
				RecipientAddress: recipientStrAddr,
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
	startTime := suite.environment.HeaderService.GetHeaderInfo(suite.ctx).Time.Add(-70 * time.Second)
	period := time.Duration(60) * time.Second
	recipientStrAddr, err := codectestutil.CodecOptions{}.GetAddressCodec().BytesToString(recipientAddr)
	suite.Require().NoError(err)

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
				startTime := suite.environment.HeaderService.GetHeaderInfo(suite.ctx).Time.Add(3600 * time.Second)
				// Prepare the budget proposal with a future start time
				budget := types.Budget{
					RecipientAddress: recipientStrAddr,
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
				startTime := suite.environment.HeaderService.GetHeaderInfo(suite.ctx).Time.Add(-50 * time.Second)
				// Prepare the budget proposal with start time and a short period
				budget := types.Budget{
					RecipientAddress: recipientStrAddr,
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
					RecipientAddress: recipientStrAddr,
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
					RecipientAddress: recipientStrAddr,
					TotalBudget:      &fooCoin,
					StartTime:        &startTime,
					Tranches:         2,
					Period:           &period,
				}
				err := suite.poolKeeper.BudgetProposal.Set(suite.ctx, recipientAddr, budget)
				suite.Require().NoError(err)

				// Claim the funds once
				msg := &types.MsgClaimBudget{
					RecipientAddress: recipientStrAddr,
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
				startTimeBeforeMonth := suite.environment.HeaderService.GetHeaderInfo(suite.ctx).Time.Add(time.Duration(-oneMonthInSeconds) * time.Second)
				oneMonthPeriod := time.Duration(oneMonthInSeconds) * time.Second
				// Prepare the budget proposal with valid start time and period of 1 month (in seconds)
				budget := types.Budget{
					RecipientAddress: recipientStrAddr,
					TotalBudget:      &fooCoin,
					StartTime:        &startTimeBeforeMonth,
					Tranches:         2,
					Period:           &oneMonthPeriod,
				}
				err := suite.poolKeeper.BudgetProposal.Set(suite.ctx, recipientAddr, budget)
				suite.Require().NoError(err)

				// Claim the funds once
				msg := &types.MsgClaimBudget{
					RecipientAddress: recipientStrAddr,
				}
				suite.mockSendCoinsFromModuleToAccount(recipientAddr)
				_, err = suite.msgServer.ClaimBudget(suite.ctx, msg)
				suite.Require().NoError(err)

				// Create a new context with an updated block time to simulate a delay
				newBlockTime := suite.environment.HeaderService.GetHeaderInfo(suite.ctx).Time.Add(time.Duration(oneMonthInSeconds) * time.Second)
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
					RecipientAddress: recipientStrAddr,
					TotalBudget:      &fooCoin,
					StartTime:        &startTime,
					Tranches:         2,
					Period:           &period,
				}
				err := suite.poolKeeper.BudgetProposal.Set(suite.ctx, recipientAddr, budget)
				suite.Require().NoError(err)

				// Claim the funds once
				msg := &types.MsgClaimBudget{
					RecipientAddress: recipientStrAddr,
				}
				suite.mockSendCoinsFromModuleToAccount(recipientAddr)
				_, err = suite.msgServer.ClaimBudget(suite.ctx, msg)
				suite.Require().NoError(err)

				// Create a new context with an updated block time to simulate a delay
				newBlockTime := suite.environment.HeaderService.GetHeaderInfo(suite.ctx).Time.Add(60 * time.Second)
				suite.ctx = suite.ctx.WithHeaderInfo(header.Info{
					Time: newBlockTime,
				})

				// Claim the funds twice
				msg = &types.MsgClaimBudget{
					RecipientAddress: recipientStrAddr,
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
			addr, err := codectestutil.CodecOptions{}.GetAddressCodec().BytesToString(tc.recipientAddress)
			suite.Require().NoError(err)
			msg := &types.MsgClaimBudget{
				RecipientAddress: addr,
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
	addressCodec := codectestutil.CodecOptions{}.GetAddressCodec()
	recipient := sdk.AccAddress([]byte("recipientAddr1__________________"))
	recipientStrAddr, err := addressCodec.BytesToString(recipient)
	suite.Require().NoError(err)
	recipient2 := sdk.AccAddress([]byte("recipientAddr2___________________"))
	recipient2StrAddr, err := addressCodec.BytesToString(recipient2)
	suite.Require().NoError(err)
	recipient3 := sdk.AccAddress([]byte("recipientAddr3___________________"))
	recipient3StrAddr, err := addressCodec.BytesToString(recipient3)
	suite.Require().NoError(err)
	testCases := map[string]struct {
		preRun           func()
		recipientAddress []sdk.AccAddress
		expErr           bool
		expErrMsg        string
		withdrawnAmount  sdk.Coin
	}{
		"empty recipient": {
			recipientAddress: []sdk.AccAddress{sdk.AccAddress([]byte(""))},
			expErr:           true,
			expErrMsg:        "invalid recipient address",
		},
		"recipient with no continuous fund": {
			recipientAddress: []sdk.AccAddress{recipient},
			expErr:           true,
			expErrMsg:        "no continuous fund found for recipient",
		},
		"funds percentage > 100": {
			preRun: func() {
				// Set fund 1
				percentage, err := math.LegacyNewDecFromStr("0.2")
				suite.Require().NoError(err)
				oneMonthInSeconds := int64(30 * 24 * 60 * 60) // Approximate number of seconds in 1 month
				expiry := suite.environment.HeaderService.GetHeaderInfo(suite.ctx).Time.Add(time.Duration(oneMonthInSeconds) * time.Second)
				cf := types.ContinuousFund{
					Recipient:  recipientStrAddr,
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
					Recipient:  recipient2StrAddr,
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
				err = suite.poolKeeper.ToDistribute.Set(suite.ctx, math.NewInt(100000))
				suite.Require().NoError(err)
			},
			recipientAddress: []sdk.AccAddress{recipient},
			expErr:           true,
			expErrMsg:        "error while iterating all the continuous funds: total funds percentage cannot exceed 100",
		},
		"expired case": {
			preRun: func() {
				percentage, err := math.LegacyNewDecFromStr("0.2")
				suite.Require().NoError(err)
				expiry := suite.environment.HeaderService.GetHeaderInfo(suite.ctx).Time.Add(time.Duration(-1) * time.Second)
				cf := types.ContinuousFund{
					Recipient:  recipientStrAddr,
					Percentage: percentage,
					Expiry:     &expiry,
				}
				// Set continuous fund
				err = suite.poolKeeper.ContinuousFund.Set(suite.ctx, recipient, cf)
				suite.Require().NoError(err)
			},
			recipientAddress: []sdk.AccAddress{recipient},
			expErr:           true,
			expErrMsg:        "cannot withdraw continuous funds: continuous fund expired for recipient",
		},
		"valid case with ToDistribute amount zero": {
			preRun: func() {
				percentage, err := math.LegacyNewDecFromStr("0.2")
				suite.Require().NoError(err)
				oneMonthInSeconds := int64(30 * 24 * 60 * 60) // Approximate number of seconds in 1 month
				expiry := suite.environment.HeaderService.GetHeaderInfo(suite.ctx).Time.Add(time.Duration(oneMonthInSeconds) * time.Second)
				cf := types.ContinuousFund{
					Recipient:  recipientStrAddr,
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
			recipientAddress: []sdk.AccAddress{recipient},
			expErr:           false,
			withdrawnAmount:  sdk.NewCoin(sdk.DefaultBondDenom, math.ZeroInt()),
		},
		"valid case with empty expiry": {
			preRun: func() {
				percentage, err := math.LegacyNewDecFromStr("0.2")
				suite.Require().NoError(err)
				cf := types.ContinuousFund{
					Recipient:  recipientStrAddr,
					Percentage: percentage,
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
				suite.mockStreamFunds()
				err = suite.poolKeeper.SetToDistribute(suite.ctx, toDistribute, suite.poolKeeper.GetAuthority())
				suite.Require().NoError(err)
			},
			recipientAddress: []sdk.AccAddress{recipient},
			expErr:           false,
			withdrawnAmount:  sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(20000)),
		},
		"valid case": {
			preRun: func() {
				percentage, err := math.LegacyNewDecFromStr("0.2")
				suite.Require().NoError(err)
				oneMonthInSeconds := int64(30 * 24 * 60 * 60) // Approximate number of seconds in 1 month
				expiry := suite.environment.HeaderService.GetHeaderInfo(suite.ctx).Time.Add(time.Duration(oneMonthInSeconds) * time.Second)
				cf := types.ContinuousFund{
					Recipient:  recipientStrAddr,
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
				suite.mockStreamFunds()
				err = suite.poolKeeper.SetToDistribute(suite.ctx, toDistribute, suite.poolKeeper.GetAuthority())
				suite.Require().NoError(err)
			},
			recipientAddress: []sdk.AccAddress{recipient},
			expErr:           false,
			withdrawnAmount:  sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(20000)),
		},
		"valid case with multiple funds": {
			preRun: func() {
				// Set continuous fund 1
				percentage, err := math.LegacyNewDecFromStr("0.3")
				suite.Require().NoError(err)
				oneMonthInSeconds := int64(30 * 24 * 60 * 60) // Approximate number of seconds in 1 month
				expiry := suite.environment.HeaderService.GetHeaderInfo(suite.ctx).Time.Add(time.Duration(oneMonthInSeconds) * time.Second)
				cf := types.ContinuousFund{
					Recipient:  recipientStrAddr,
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

				// Set continuous fund 2
				percentage2, err := math.LegacyNewDecFromStr("0.2")
				suite.Require().NoError(err)
				cf = types.ContinuousFund{
					Recipient:  recipient2StrAddr,
					Percentage: percentage,
					Expiry:     &expiry,
				}
				// Set continuous fund
				err = suite.poolKeeper.ContinuousFund.Set(suite.ctx, recipient2, cf)
				suite.Require().NoError(err)
				// Set recipient fund percentage and recipient fund distribution
				intPercentage = percentage2.MulInt64(100)
				err = suite.poolKeeper.RecipientFundPercentage.Set(suite.ctx, recipient2, intPercentage.TruncateInt())
				suite.Require().NoError(err)
				err = suite.poolKeeper.RecipientFundDistribution.Set(suite.ctx, recipient2, math.ZeroInt())
				suite.Require().NoError(err)

				// Set continuous fund 3
				percentage3, err := math.LegacyNewDecFromStr("0.3")
				suite.Require().NoError(err)
				cf = types.ContinuousFund{
					Recipient:  recipient3StrAddr,
					Percentage: percentage2,
					Expiry:     &expiry,
				}
				// Set continuous fund
				err = suite.poolKeeper.ContinuousFund.Set(suite.ctx, recipient3, cf)
				suite.Require().NoError(err)
				// Set recipient fund percentage and recipient fund distribution
				intPercentage = percentage3.MulInt64(100)
				err = suite.poolKeeper.RecipientFundPercentage.Set(suite.ctx, recipient3, intPercentage.TruncateInt())
				suite.Require().NoError(err)
				err = suite.poolKeeper.RecipientFundDistribution.Set(suite.ctx, recipient3, math.ZeroInt())
				suite.Require().NoError(err)

				toDistribute := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100000)))
				suite.mockStreamFunds()
				err = suite.poolKeeper.SetToDistribute(suite.ctx, toDistribute, suite.poolKeeper.GetAuthority())
				suite.Require().NoError(err)
			},
			recipientAddress: []sdk.AccAddress{recipient, recipient2, recipient3},
			expErr:           false,
			withdrawnAmount:  sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(30000)),
		},
	}

	for name, tc := range testCases {
		suite.Run(name, func() {
			suite.SetupTest()
			if tc.preRun != nil {
				tc.preRun()
			}
			addr, err := addressCodec.BytesToString(tc.recipientAddress[0])
			suite.Require().NoError(err)
			msg := &types.MsgWithdrawContinuousFund{
				RecipientAddress: addr,
			}

			suite.mockWithdrawContinuousFund()

			resp, err := suite.msgServer.WithdrawContinuousFund(suite.ctx, msg)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.withdrawnAmount, resp.Amount)

				// this condition is valid only for request with multiple continuous funds
				if len(tc.recipientAddress) > 1 {
					toClaim, err := suite.poolKeeper.RecipientFundDistribution.Get(suite.ctx, tc.recipientAddress[1])
					suite.Require().NoError(err)
					suite.Require().Equal(toClaim, math.NewInt(20000))
					toClaim, err = suite.poolKeeper.RecipientFundDistribution.Get(suite.ctx, tc.recipientAddress[2])
					suite.Require().NoError(err)
					suite.Require().Equal(toClaim, math.NewInt(30000))
				}
			}
		})
	}
}

func (suite *KeeperTestSuite) TestCreateContinuousFund() {
	percentage, err := math.LegacyNewDecFromStr("0.2")
	suite.Require().NoError(err)
	negativePercentage, err := math.LegacyNewDecFromStr("-0.2")
	suite.Require().NoError(err)
	invalidExpirty := suite.environment.HeaderService.GetHeaderInfo(suite.ctx).Time.Add(-15 * time.Second)
	oneMonthInSeconds := int64(30 * 24 * 60 * 60) // Approximate number of seconds in 1 month
	expiry := suite.environment.HeaderService.GetHeaderInfo(suite.ctx).Time.Add(time.Duration(oneMonthInSeconds) * time.Second)
	recipientStrAddr, err := codectestutil.CodecOptions{}.GetAddressCodec().BytesToString(recipientAddr)
	suite.Require().NoError(err)
	testCases := map[string]struct {
		preRun    func()
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
				Recipient:  recipientStrAddr,
				Percentage: percentage,
				Expiry:     &expiry,
			},
			expErr:    true,
			expErrMsg: "empty address string is not allowed",
		},
		"invalid authority": {
			input: &types.MsgCreateContinuousFund{
				Authority:  "invalid_authority",
				Recipient:  recipientStrAddr,
				Percentage: percentage,
				Expiry:     &expiry,
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		"zero percentage": {
			input: &types.MsgCreateContinuousFund{
				Authority:  suite.poolKeeper.GetAuthority(),
				Recipient:  recipientStrAddr,
				Percentage: math.LegacyNewDec(0),
				Expiry:     &expiry,
			},
			expErr:    true,
			expErrMsg: "percentage cannot be zero or empty",
		},
		"negative percentage": {
			input: &types.MsgCreateContinuousFund{
				Authority:  suite.poolKeeper.GetAuthority(),
				Recipient:  recipientStrAddr,
				Percentage: negativePercentage,
				Expiry:     &expiry,
			},
			expErr:    true,
			expErrMsg: "percentage cannot be negative",
		},
		"invalid percentage": {
			input: &types.MsgCreateContinuousFund{
				Authority:  suite.poolKeeper.GetAuthority(),
				Recipient:  recipientStrAddr,
				Percentage: math.LegacyNewDec(1),
				Expiry:     &expiry,
			},
			expErr:    true,
			expErrMsg: "percentage cannot be greater than or equal to one",
		},
		"invalid expiry": {
			input: &types.MsgCreateContinuousFund{
				Authority:  suite.poolKeeper.GetAuthority(),
				Recipient:  recipientStrAddr,
				Percentage: percentage,
				Expiry:     &invalidExpirty,
			},
			expErr:    true,
			expErrMsg: "expiry time cannot be less than the current block time",
		},
		"all good": {
			input: &types.MsgCreateContinuousFund{
				Authority:  suite.poolKeeper.GetAuthority(),
				Recipient:  recipientStrAddr,
				Percentage: percentage,
				Expiry:     &expiry,
			},
			expErr: false,
		},
		"total funds percentage > 100": {
			preRun: func() {
				percentage, err := math.LegacyNewDecFromStr("0.9")
				suite.Require().NoError(err)
				recipient2 := sdk.AccAddress([]byte("recipientAddr2___________________"))
				recipient2StrAddr, err := codectestutil.CodecOptions{}.GetAddressCodec().BytesToString(recipient2)
				suite.Require().NoError(err)
				cf := types.ContinuousFund{
					Recipient:  recipient2StrAddr,
					Percentage: percentage,
					Expiry:     &time.Time{},
				}
				err = suite.poolKeeper.ContinuousFund.Set(suite.ctx, recipient2, cf)
				suite.Require().NoError(err)
				intPercentage := percentage.MulInt64(100)
				err = suite.poolKeeper.RecipientFundPercentage.Set(suite.ctx, recipient2, intPercentage.TruncateInt())
				suite.Require().NoError(err)
			},
			input: &types.MsgCreateContinuousFund{
				Authority:  suite.poolKeeper.GetAuthority(),
				Recipient:  recipientStrAddr,
				Percentage: percentage,
				Expiry:     &expiry,
			},
			expErr:    true,
			expErrMsg: "cannot set continuous fund proposal\ntotal funds percentage exceeds 100\ncurrent total percentage: 90",
		},
	}

	for name, tc := range testCases {
		suite.Run(name, func() {
			suite.SetupTest()
			if tc.preRun != nil {
				tc.preRun()
			}

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

// TestCancelContinuousFund tests the cancellation of a continuous fund.
// It verifies various scenarios such as canceling a fund with an empty recipient,
// canceling a fund with no recipient found, canceling a fund with unclaimed funds for the recipient,
// and canceling a fund with no errors.
func (suite *KeeperTestSuite) TestCancelContinuousFund() {
	recipientStrAddr, err := codectestutil.CodecOptions{}.GetAddressCodec().BytesToString(recipientAddr)
	suite.Require().NoError(err)
	recipient2 := sdk.AccAddress([]byte("recipientAddr2___________________"))
	recipient2StrAddr, err := codectestutil.CodecOptions{}.GetAddressCodec().BytesToString(recipient2)
	suite.Require().NoError(err)
	testCases := map[string]struct {
		preRun         func()
		recipientAddr  sdk.AccAddress
		expErr         bool
		expErrMsg      string
		postRun        func()
		withdrawnFunds sdk.Coin
	}{
		"empty recipient": {
			preRun: func() {
				percentage, err := math.LegacyNewDecFromStr("0.2")
				suite.Require().NoError(err)
				oneMonthInSeconds := int64(30 * 24 * 60 * 60) // Approximate number of seconds in 1 month
				expiry := suite.environment.HeaderService.GetHeaderInfo(suite.ctx).Time.Add(time.Duration(oneMonthInSeconds) * time.Second)
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
		"all good with unclaimed funds for recipient": {
			preRun: func() {
				// Set fund 1
				percentage, err := math.LegacyNewDecFromStr("0.2")
				suite.Require().NoError(err)
				oneMonthInSeconds := int64(30 * 24 * 60 * 60) // Approximate number of seconds in 1 month
				expiry := suite.environment.HeaderService.GetHeaderInfo(suite.ctx).Time.Add(time.Duration(oneMonthInSeconds) * time.Second)
				cf := types.ContinuousFund{
					Recipient:  recipientStrAddr,
					Percentage: percentage,
					Expiry:     &expiry,
				}
				// Set continuous fund
				err = suite.poolKeeper.ContinuousFund.Set(suite.ctx, recipientAddr, cf)
				suite.Require().NoError(err)
				// Set recipient fund percentage and recipient fund distribution
				intPercentage := percentage.MulInt64(100)
				err = suite.poolKeeper.RecipientFundPercentage.Set(suite.ctx, recipientAddr, intPercentage.TruncateInt())
				suite.Require().NoError(err)
				err = suite.poolKeeper.RecipientFundDistribution.Set(suite.ctx, recipientAddr, math.ZeroInt())
				suite.Require().NoError(err)

				// Set fund 2
				percentage, err = math.LegacyNewDecFromStr("0.3")
				suite.Require().NoError(err)
				cf = types.ContinuousFund{
					Recipient:  recipient2StrAddr,
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
				suite.mockStreamFunds()
				err = suite.poolKeeper.SetToDistribute(suite.ctx, toDistribute, suite.poolKeeper.GetAuthority())
				suite.Require().NoError(err)

				// withdraw funds for fund request 2
				suite.mockWithdrawContinuousFund()
				msg := &types.MsgWithdrawContinuousFund{RecipientAddress: recipient2StrAddr}
				_, err = suite.msgServer.WithdrawContinuousFund(suite.ctx, msg)
				suite.Require().NoError(err)
			},
			recipientAddr: recipientAddr,
			expErr:        false,
			postRun: func() {
				_, err := suite.poolKeeper.ContinuousFund.Get(suite.ctx, recipientAddr)
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, collections.ErrNotFound)
			},
			withdrawnFunds: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(20000)),
		},
		"all good": {
			preRun: func() {
				percentage, err := math.LegacyNewDecFromStr("0.2")
				suite.Require().NoError(err)
				oneMonthInSeconds := int64(30 * 24 * 60 * 60) // Approximate number of seconds in 1 month
				expiry := suite.environment.HeaderService.GetHeaderInfo(suite.ctx).Time.Add(time.Duration(oneMonthInSeconds) * time.Second)
				cf := types.ContinuousFund{
					Recipient:  recipientStrAddr,
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
			addr, err := codectestutil.CodecOptions{}.GetAddressCodec().BytesToString(tc.recipientAddr)
			suite.Require().NoError(err)
			msg := &types.MsgCancelContinuousFund{
				Authority:        suite.poolKeeper.GetAuthority(),
				RecipientAddress: addr,
			}
			resp, err := suite.msgServer.CancelContinuousFund(suite.ctx, msg)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(resp.WithdrawnAllocatedFund, tc.withdrawnFunds)
			}
			if tc.postRun != nil {
				tc.postRun()
			}
		})
	}
}
