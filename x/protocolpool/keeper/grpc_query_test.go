package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/x/protocolpool/types"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestUnclaimedBudget() {
	startTime := suite.environment.HeaderService.GetHeaderInfo(suite.ctx).Time.Add(-70 * time.Second)
	period := time.Duration(60) * time.Second
	zeroCoin := sdk.NewCoin("foo", math.ZeroInt())
	nextClaimFrom := startTime.Add(period)
	recipientStrAddr, err := codectestutil.CodecOptions{}.GetAddressCodec().BytesToString(recipientAddr)
	suite.Require().NoError(err)
	testCases := []struct {
		name           string
		preRun         func()
		req            *types.QueryUnclaimedBudgetRequest
		expErr         bool
		expErrMsg      string
		unclaimedFunds *sdk.Coin
		resp           *types.QueryUnclaimedBudgetResponse
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
				Address: recipientStrAddr,
			},
			expErr:    true,
			expErrMsg: "no budget proposal found for address",
		},
		{
			name: "valid case",
			preRun: func() {
				// Prepare a valid budget proposal
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
			req: &types.QueryUnclaimedBudgetRequest{
				Address: recipientStrAddr,
			},
			expErr:         false,
			unclaimedFunds: &fooCoin,
			resp: &types.QueryUnclaimedBudgetResponse{
				TotalBudget:     &fooCoin,
				ClaimedAmount:   &zeroCoin,
				UnclaimedAmount: &fooCoin,
				NextClaimFrom:   &startTime,
				Period:          &period,
				TranchesLeft:    2,
			},
		},
		{
			name: "valid case with claim",
			preRun: func() {
				// Prepare a valid budget proposal
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

			req: &types.QueryUnclaimedBudgetRequest{
				Address: recipientStrAddr,
			},
			expErr:         false,
			unclaimedFunds: &fooCoin2,
			resp: &types.QueryUnclaimedBudgetResponse{
				TotalBudget:     &fooCoin,
				ClaimedAmount:   &fooCoin2,
				UnclaimedAmount: &fooCoin2,
				NextClaimFrom:   &nextClaimFrom,
				Period:          &period,
				TranchesLeft:    1,
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			if tc.preRun != nil {
				tc.preRun()
			}
			resp, err := suite.queryServer.UnclaimedBudget(suite.ctx, tc.req)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.resp, resp)
			}
		})
	}
}
