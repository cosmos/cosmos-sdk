package keeper_test

import (
	"go.uber.org/mock/gomock"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

func (suite *KeeperTestSuite) TestParams() {
	expectedParams := types.DefaultGenesisState().Params
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

func (suite *KeeperTestSuite) TestContinuousFund() {
	testCases := []struct {
		name      string
		preRun    func()
		req       *types.QueryContinuousFundRequest
		expErr    bool
		expErrMsg string
		resp      *types.QueryContinuousFundResponse
	}{
		{
			name:      "nil request - error",
			req:       nil,
			expErr:    true,
			expErrMsg: "empty request",
		},
		{
			name: "invalid address",
			req: &types.QueryContinuousFundRequest{
				Recipient: "invalid", // not a valid Bech32 address
			},
			preRun: func() {
				// Return a real codec; its StringToBytes will fail for an invalid address.
				suite.authKeeper.EXPECT().AddressCodec().
					Return(address.NewBech32Codec("cosmos")).AnyTimes()
			},
			expErr:    true,
			expErrMsg: "invalid address:",
		},
		{
			name: "fund not found",
			req: &types.QueryContinuousFundRequest{
				Recipient: recipientAddr.String(), // valid format but not set in store
			},
			preRun: func() {
				suite.authKeeper.EXPECT().AddressCodec().
					Return(address.NewBech32Codec("cosmos")).AnyTimes()
			},
			expErr:    true,
			expErrMsg: "not found",
		},
		{
			name: "valid continuous fund",
			req: &types.QueryContinuousFundRequest{
				Recipient: recipientAddr.String(),
			},
			preRun: func() {
				// Use the real codec to convert the address.
				suite.authKeeper.EXPECT().AddressCodec().
					Return(address.NewBech32Codec("cosmos")).AnyTimes()
				// Insert a continuous fund directly into the pool keeper.
				fund := types.ContinuousFund{
					Recipient:  recipientAddr.String(),
					Percentage: math.LegacyMustNewDecFromStr("0.5"),
					Expiry:     nil,
				}
				err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipientAddr, fund)
				suite.Require().NoError(err)
			},
			expErr: false,
			resp: &types.QueryContinuousFundResponse{
				ContinuousFund: types.ContinuousFund{
					Recipient:  recipientAddr.String(),
					Percentage: math.LegacyMustNewDecFromStr("0.5"),
					Expiry:     nil,
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
				suite.Require().Equal(tc.resp.ContinuousFund.Recipient, resp.ContinuousFund.Recipient)
				suite.Require().Equal(tc.resp.ContinuousFund.Percentage, resp.ContinuousFund.Percentage)
				suite.Require().Equal(tc.resp.ContinuousFund.Expiry, resp.ContinuousFund.Expiry)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestContinuousFunds() {
	testCases := []struct {
		name      string
		preRun    func()
		req       *types.QueryContinuousFundsRequest
		expErr    bool
		expErrMsg string
		resp      *types.QueryContinuousFundsResponse
	}{
		{
			name:      "nil request - error",
			req:       nil,
			expErr:    true,
			expErrMsg: "empty request",
		},
		{
			name:   "no continuous funds",
			req:    &types.QueryContinuousFundsRequest{},
			preRun: nil,
			expErr: false,
			resp: &types.QueryContinuousFundsResponse{
				ContinuousFunds: []types.ContinuousFund{},
			},
		},
		{
			name: "valid continuous funds",
			req:  &types.QueryContinuousFundsRequest{},
			preRun: func() {
				// Insert two continuous funds directly into the keeper.
				fund1 := types.ContinuousFund{
					Recipient:  recipientAddr.String(),
					Percentage: math.LegacyMustNewDecFromStr("0.3"),
					Expiry:     nil,
				}
				err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipientAddr, fund1)
				suite.Require().NoError(err)

				fund2 := types.ContinuousFund{
					Recipient:  recipientAddr2.String(),
					Percentage: math.LegacyMustNewDecFromStr("0.7"),
					Expiry:     nil,
				}
				err = suite.poolKeeper.ContinuousFunds.Set(suite.ctx, recipientAddr2, fund2)
				suite.Require().NoError(err)
			},
			expErr: false,
			resp: &types.QueryContinuousFundsResponse{
				ContinuousFunds: []types.ContinuousFund{
					{
						Recipient:  recipientAddr.String(),
						Percentage: math.LegacyMustNewDecFromStr("0.3"),
						Expiry:     nil,
					},
					{
						Recipient:  recipientAddr2.String(),
						Percentage: math.LegacyMustNewDecFromStr("0.7"),
						Expiry:     nil,
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
				suite.Require().ElementsMatch(tc.resp.ContinuousFunds, resp.ContinuousFunds)
			}
		})
	}
}
