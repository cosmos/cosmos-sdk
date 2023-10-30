package keeper_test

import (
	"cosmossdk.io/x/protocolpool/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	recipientAddr = sdk.AccAddress([]byte("to1__________________"))

	fooCoin = sdk.NewInt64Coin("foo", 100)
)

func (suite *KeeperTestSuite) TestMsgSubmitBudgetProposal() {
	invalidCoin := sdk.NewInt64Coin("foo", 0)
	testCases := map[string]struct {
		preRun    func()
		input     *types.MsgSubmitBudgetProposal
		expErr    bool
		expErrMsg string
	}{
		"empty recipient address": {
			input: &types.MsgSubmitBudgetProposal{
				Authority:        suite.poolKeeper.GetAuthority(),
				RecipientAddress: "",
				TotalBudget:      &fooCoin,
				StartTime:        suite.ctx.BlockTime().Unix(),
				Tranches:         2,
				Period:           60,
			},
			expErr:    true,
			expErrMsg: "empty address string is not allowed",
		},
		"empty authority": {
			input: &types.MsgSubmitBudgetProposal{
				Authority:        "",
				RecipientAddress: recipientAddr.String(),
				TotalBudget:      &fooCoin,
				StartTime:        suite.ctx.BlockTime().Unix(),
				Tranches:         2,
				Period:           60,
			},
			expErr:    true,
			expErrMsg: "empty address string is not allowed",
		},
		"invalid authority": {
			input: &types.MsgSubmitBudgetProposal{
				Authority:        "invalid_authority",
				RecipientAddress: recipientAddr.String(),
				TotalBudget:      &fooCoin,
				StartTime:        suite.ctx.BlockTime().Unix(),
				Tranches:         2,
				Period:           60,
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		"invalid budget": {
			input: &types.MsgSubmitBudgetProposal{
				Authority:        suite.poolKeeper.GetAuthority(),
				RecipientAddress: recipientAddr.String(),
				TotalBudget:      &invalidCoin,
				StartTime:        suite.ctx.BlockTime().Unix(),
				Tranches:         2,
				Period:           60,
			},
			expErr:    true,
			expErrMsg: "total budget cannot be zero",
		},
		"invalid start time": {
			input: &types.MsgSubmitBudgetProposal{
				Authority:        suite.poolKeeper.GetAuthority(),
				RecipientAddress: recipientAddr.String(),
				TotalBudget:      &fooCoin,
				StartTime:        -10,
				Tranches:         2,
				Period:           60,
			},
			expErr:    true,
			expErrMsg: "invalid start time",
		},
		"invalid tranches": {
			input: &types.MsgSubmitBudgetProposal{
				Authority:        suite.poolKeeper.GetAuthority(),
				RecipientAddress: recipientAddr.String(),
				TotalBudget:      &fooCoin,
				StartTime:        suite.ctx.BlockTime().Unix(),
				Tranches:         0,
				Period:           60,
			},
			expErr:    true,
			expErrMsg: "tranches must be a positive value",
		},
		"invalid period": {
			input: &types.MsgSubmitBudgetProposal{
				Authority:        suite.poolKeeper.GetAuthority(),
				RecipientAddress: recipientAddr.String(),
				TotalBudget:      &fooCoin,
				StartTime:        suite.ctx.BlockTime().Unix(),
				Tranches:         2,
				Period:           0,
			},
			expErr:    true,
			expErrMsg: "period length should be a positive value",
		},
		"all good": {
			input: &types.MsgSubmitBudgetProposal{
				Authority:        suite.poolKeeper.GetAuthority(),
				RecipientAddress: recipientAddr.String(),
				TotalBudget:      &fooCoin,
				StartTime:        suite.ctx.BlockTime().Unix(),
				Tranches:         2,
				Period:           60,
			},
			expErr: false,
		},
	}

	for name, tc := range testCases {
		suite.Run(name, func() {
			if tc.preRun != nil {
				tc.preRun()
			}
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
	testCases := map[string]struct {
		preRun           func()
		recipientAddress sdk.AccAddress
		expErr           bool
		expErrMsg        string
	}{
		"empty recipient addr": {
			recipientAddress: sdk.AccAddress(""),
			expErr:           true,
			expErrMsg:        "invalid recipient address: empty address string is not allowed",
		},
		"no budget found": {
			recipientAddress: sdk.AccAddress([]byte("acc1__________")),
			expErr:           true,
			expErrMsg:        "no claimable funds are present for recipient",
		},
		"claiming before start time": {
			preRun: func() {
				// Prepare the budget proposal with a future start time
				budget := types.Budget{
					RecipientAddress: recipientAddr.String(),
					TotalBudget:      &fooCoin,
					StartTime:        suite.ctx.BlockTime().Unix() + 3600,
					Tranches:         2,
					Period:           60,
				}
				suite.poolKeeper.BudgetProposal.Set(suite.ctx, recipientAddr, budget)
			},
			recipientAddress: recipientAddr,
			expErr:           true,
			expErrMsg:        "distribution has not started yet",
		},
		"budget period has not passed": {
			preRun: func() {
				// Prepare the budget proposal with start time and a short period
				budget := types.Budget{
					RecipientAddress: recipientAddr.String(),
					TotalBudget:      &fooCoin,
					StartTime:        suite.ctx.BlockTime().Unix() - 50,
					Tranches:         1,
					Period:           60,
				}
				suite.poolKeeper.BudgetProposal.Set(suite.ctx, recipientAddr, budget)
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
					StartTime:        suite.ctx.BlockTime().Unix() - 70,
					Tranches:         2,
					Period:           60,
				}
				suite.poolKeeper.BudgetProposal.Set(suite.ctx, recipientAddr, budget)
			},
			recipientAddress: recipientAddr,
			expErr:           false,
		},
	}

	for name, tc := range testCases {
		suite.Run(name, func() {
			if tc.preRun != nil {
				tc.preRun()
			}
			msg := &types.MsgClaimBudget{
				RecipientAddress: tc.recipientAddress.String(),
			}
			suite.mockSendCoinsFromModuleToAccount(tc.recipientAddress)
			_, err := suite.msgServer.ClaimBudget(suite.ctx, msg)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}
