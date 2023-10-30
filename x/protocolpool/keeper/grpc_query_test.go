package keeper_test

import (
	"cosmossdk.io/x/protocolpool/keeper"
	"cosmossdk.io/x/protocolpool/types"
)

func (suite *KeeperTestSuite) TestUnclaimedBudget() {
	queryServer := keeper.NewQuerier(suite.poolKeeper)
	testCases := []struct {
		name      string
		preRun    func()
		req       *types.QueryUnclaimedBudgetRequest
		expErr    bool
		expErrMsg string
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
					StartTime:        suite.ctx.BlockTime().Unix() - 70,
					Tranches:         2,
					Period:           60,
				}
				suite.poolKeeper.BudgetProposal.Set(suite.ctx, recipientAddr, budget)
			},
			req: &types.QueryUnclaimedBudgetRequest{
				Address: recipientAddr.String(),
			},
			expErr: false,
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
				suite.Require().Equal(resp.UnclaimedAmount, &fooCoin)
			}
		})
	}
}
