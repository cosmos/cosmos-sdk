package keeper_test

import (
	"time"

	"go.uber.org/mock/gomock"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

func (suite *KeeperTestSuite) TestContinuousFunds() {
	t := time.Now()
	recipientStrAddr := recipientAddr.String()
	recipientStrAddr2 := recipientAddr2.String()
	testCases := []struct {
		name           string
		preRun         func()
		req            *types.QueryContinuousFundsRequest
		expErr         bool
		expErrMsg      string
		unclaimedFunds *sdk.Coin
		resp           *types.QueryContinuousFundsResponse
	}{
		{
			name:      "nil request",
			req:       nil,
			expErr:    true,
			expErrMsg: "empty request",
		},
		{
			name:      "empty recipient address",
			req:       &types.QueryContinuousFundsRequest{},
			expErr:    true,
			expErrMsg: "empty address string is not allowed",
		},
		{
			name:      "no budget proposal found",
			req:       &types.QueryContinuousFundsRequest{},
			expErr:    true,
			expErrMsg: "no budget proposal found for address",
		},
		{
			name: "valid case - single",
			preRun: func() {
				// Prepare a valid budget proposal
				fund := types.ContinuousFund{
					Recipient:  recipientStrAddr,
					Percentage: math.LegacyMustNewDecFromStr("0.1"),
				}
				err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipientAddr, fund)
				suite.Require().NoError(err)
			},
			req:            &types.QueryContinuousFundsRequest{},
			expErr:         false,
			unclaimedFunds: &fooCoin,
			resp: &types.QueryContinuousFundsResponse{
				ContinuousFunds: []types.ContinuousFund{
					{
						Recipient:  recipientStrAddr,
						Percentage: math.LegacyMustNewDecFromStr("0.1"),
					},
				},
			},
		},
		{
			name: "valid case - multiple",
			preRun: func() {
				// Prepare a valid budget proposal
				fund1 := types.ContinuousFund{
					Recipient:  recipientStrAddr,
					Percentage: math.LegacyMustNewDecFromStr("0.1"),
					Expiry:     &t,
				}
				err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipientAddr, fund1)
				suite.Require().NoError(err)

				// Prepare a valid budget proposal
				fund2 := types.ContinuousFund{
					Recipient:  recipientStrAddr2,
					Percentage: math.LegacyMustNewDecFromStr("0.2"),
					Expiry:     &t,
				}
				err = suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipientAddr2, fund2)
				suite.Require().NoError(err)
			},
			req:            &types.QueryContinuousFundsRequest{},
			expErr:         false,
			unclaimedFunds: &fooCoin,
			resp: &types.QueryContinuousFundsResponse{
				ContinuousFunds: []types.ContinuousFund{
					{
						Recipient:  recipientStrAddr,
						Percentage: math.LegacyMustNewDecFromStr("0.1"),
						Expiry:     &t,
					},
					{
						Recipient:  recipientStrAddr2,
						Percentage: math.LegacyMustNewDecFromStr("0.2"),
						Expiry:     &t,
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			if tc.preRun != nil {
				tc.preRun()
			}
			resp, err := suite.queryServer.ContinuousFunds(suite.ctx, tc.req)
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

func (suite *KeeperTestSuite) TestContinuousFund() {
	t := time.Now()
	recipientStrAddr := recipientAddr.String()
	testCases := []struct {
		name           string
		preRun         func()
		req            *types.QueryContinuousFundRequest
		expErr         bool
		expErrMsg      string
		unclaimedFunds *sdk.Coin
		resp           *types.QueryContinuousFundResponse
	}{
		{
			name: "empty recipient address",
			req: &types.QueryContinuousFundRequest{
				Recipient: "",
			},
			expErr:    true,
			expErrMsg: "empty address string is not allowed",
		},
		{
			name: "no continuous fund found",
			req: &types.QueryContinuousFundRequest{
				Recipient: recipientStrAddr,
			},
			expErr:    true,
			expErrMsg: "rpc error: code = NotFound desc = not found",
		},
		{
			name: "valid case - no expiry",
			preRun: func() {
				// Prepare a valid budget proposal
				fund := types.ContinuousFund{
					Recipient:  recipientStrAddr,
					Percentage: math.LegacyMustNewDecFromStr("0.1"),
				}
				err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipientAddr, fund)
				suite.Require().NoError(err)
			},
			req: &types.QueryContinuousFundRequest{
				Recipient: recipientStrAddr,
			},
			expErr:         false,
			unclaimedFunds: &fooCoin,
			resp: &types.QueryContinuousFundResponse{
				ContinuousFund: types.ContinuousFund{
					Recipient:  recipientStrAddr,
					Percentage: math.LegacyMustNewDecFromStr("0.1"),
				},
			},
		},
		{
			name: "valid case",
			preRun: func() {
				// Prepare a valid budget proposal
				fund := types.ContinuousFund{
					Recipient:  recipientStrAddr,
					Percentage: math.LegacyMustNewDecFromStr("0.1"),
					Expiry:     &t,
				}
				err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipientAddr, fund)
				suite.Require().NoError(err)
			},
			req: &types.QueryContinuousFundRequest{
				Recipient: recipientStrAddr,
			},
			expErr:         false,
			unclaimedFunds: &fooCoin,
			resp: &types.QueryContinuousFundResponse{
				ContinuousFund: types.ContinuousFund{
					Recipient:  recipientStrAddr,
					Percentage: math.LegacyMustNewDecFromStr("0.1"),
					Expiry:     &t,
				},
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			if tc.preRun != nil {
				tc.preRun()
			}
			resp, err := suite.queryServer.ContinuousFund(suite.ctx, tc.req)
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

func (suite *KeeperTestSuite) TestParams() {
	expectedParams := *types.DefaultGenesisState().Params
	err := suite.poolKeeper.Params.Set(suite.ctx, expectedParams)
	suite.Require().NoError(err)

	testCases := []struct {
		name      string
		preRun    func()
		req       *types.QueryParamsRequest
		expErr    bool
		expErrMsg string
		resp      *types.QueryParamsResponse
	}{
		{
			name:      "nil request	- error",
			req:       nil,
			expErr:    true,
			expErrMsg: "rpc error: code = InvalidArgument desc = empty request",
		},
		{
			name:      "valid",
			req:       &types.QueryParamsRequest{},
			expErr:    false,
			expErrMsg: "empty address string is not allowed",
			resp: &types.QueryParamsResponse{
				Params: expectedParams,
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			if tc.preRun != nil {
				tc.preRun()
			}
			resp, err := suite.queryServer.Params(suite.ctx, tc.req)
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

func (suite *KeeperTestSuite) TestCommunityPool() {
	testCases := []struct {
		name      string
		preRun    func()
		req       *types.QueryCommunityPoolRequest
		expErr    bool
		expErrMsg string
		resp      *types.QueryCommunityPoolResponse
	}{
		{
			name:      "nil request	- error",
			req:       nil,
			expErr:    true,
			expErrMsg: "rpc error: code = InvalidArgument desc = empty request",
		},
		{
			name: "valid",
			preRun: func() {
				suite.authKeeper.EXPECT().GetModuleAccount(gomock.Any(), types.ModuleName).Return(poolAcc).Times(1)
				suite.bankKeeper.EXPECT().GetAllBalances(gomock.Any(), poolAcc.GetAddress()).Return(sdk.NewCoins(fooCoin)).Times(1)
			},
			req: &types.QueryCommunityPoolRequest{},
			resp: &types.QueryCommunityPoolResponse{
				Pool: sdk.NewCoins(fooCoin),
			},
			expErr:    false,
			expErrMsg: "",
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			if tc.preRun != nil {
				tc.preRun()
			}
			resp, err := suite.queryServer.CommunityPool(suite.ctx, tc.req)
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
