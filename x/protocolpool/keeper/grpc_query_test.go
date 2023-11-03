package keeper_test

import (
	"time"

	"cosmossdk.io/x/protocolpool/keeper"
	"cosmossdk.io/x/protocolpool/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestUnclaimedBudget() {
	queryServer := keeper.NewQuerier(suite.poolKeeper)
	startTime := suite.ctx.BlockTime().Add(-70 * time.Second)
	period := time.Duration(60) * time.Second
	testCases := []struct {
		name           string
		preRun         func()
		req            *types.QueryUnclaimedBudgetRequest
		expErr         bool
		expErrMsg      string
		unclaimedFunds *sdk.Coin
	}{
		{
			name: "empty recipient address",
			req: &types.QueryUnclaimedBudgetRequest{
				Address: "",
			},
			expErr:    true,
			expErrMsg: "empty address string is not allowed",
		},
		{
			name: "no budget proposal found",
			req: &types.QueryUnclaimedBudgetRequest{
				Address: recipientAddr.String(),
			},
			expErr:    true,
			expErrMsg: "no budget proposal found for address",
		},
		{
			name: "valid case",
			preRun: func() {
				// Prepare a valid budget proposal
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
			req: &types.QueryUnclaimedBudgetRequest{
				Address: recipientAddr.String(),
			},
			expErr:         false,
			unclaimedFunds: &fooCoin,
		},
		{
			name: "valid case with claim",
			preRun: func() {
				// Prepare a valid budget proposal
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

			req: &types.QueryUnclaimedBudgetRequest{
				Address: recipientAddr.String(),
			},
			expErr:         false,
			unclaimedFunds: &fooCoin2,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			if tc.preRun != nil {
				tc.preRun()
			}
			resp, err := queryServer.UnclaimedBudget(suite.ctx, tc.req)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.unclaimedFunds, resp.UnclaimedAmount)
			}
		})
	}
}
