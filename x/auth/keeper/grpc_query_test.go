package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func (suite *KeeperTestSuite) TestGRPCQueryAccount() {
	var (
		req *types.QueryAccountRequest
	)
	_, _, addr := testdata.KeyTestPubAddr()

	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		posttests func(res *types.QueryAccountResponse)
	}{
		{
			"empty request",
			func() {
				req = &types.QueryAccountRequest{}
			},
			false,
			func(res *types.QueryAccountResponse) {},
		},
		{
			"invalid request",
			func() {
				req = &types.QueryAccountRequest{Address: ""}
			},
			false,
			func(res *types.QueryAccountResponse) {},
		},
		{
			"invalid request with empty byte array",
			func() {
				req = &types.QueryAccountRequest{Address: ""}
			},
			false,
			func(res *types.QueryAccountResponse) {},
		},
		{
			"account not found",
			func() {
				req = &types.QueryAccountRequest{Address: addr.String()}
			},
			false,
			func(res *types.QueryAccountResponse) {},
		},
		{
			"success",
			func() {
				suite.app.AccountKeeper.SetAccount(suite.ctx,
					suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr))
				req = &types.QueryAccountRequest{Address: addr.String()}
			},
			true,
			func(res *types.QueryAccountResponse) {
				var newAccount types.AccountI
				err := suite.app.InterfaceRegistry().UnpackAny(res.Account, &newAccount)
				suite.Require().NoError(err)
				suite.Require().NotNil(newAccount)
				suite.Require().True(addr.Equals(newAccount.GetAddress()))
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)

			res, err := suite.queryClient.Account(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(res)
			}

			tc.posttests(res)
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCQueryParameters() {
	var (
		req       *types.QueryParamsRequest
		expParams types.Params
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"success",
			func() {
				req = &types.QueryParamsRequest{}
				expParams = suite.app.AccountKeeper.GetParams(suite.ctx)
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)

			res, err := suite.queryClient.Params(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(expParams, res.Params)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(res)
			}
		})
	}
}
