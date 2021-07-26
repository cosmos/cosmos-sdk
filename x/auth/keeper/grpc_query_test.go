package keeper_test

import (
	"fmt"
	"bytes"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func (suite *KeeperTestSuite) TestGRPCQueryAccounts() {
	var (
		req *types.QueryAccountsRequest
	)
	_, _, first := testdata.KeyTestPubAddr()
	_, _, second := testdata.KeyTestPubAddr()

	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		posttests func(res *types.QueryAccountsResponse)
	}{
		{
			"success",
			func() {
				suite.app.AccountKeeper.SetAccount(suite.ctx,
					suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, first))
				suite.app.AccountKeeper.SetAccount(suite.ctx,
					suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, second))
				req = &types.QueryAccountsRequest{}
			},
			true,
			func(res *types.QueryAccountsResponse) {
				for _, acc := range res.Accounts {
					var account types.AccountI
					err := suite.app.InterfaceRegistry().UnpackAny(acc, &account)
					suite.Require().NoError(err)

					suite.Require().True(
						first.Equals(account.GetAddress()) || second.Equals(account.GetAddress()))
				}
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)

			res, err := suite.queryClient.Accounts(ctx, req)

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

func (suite *KeeperTestSuite) TestBech32Prefix(){
	suite.Run("TestBech32Prefix", func() {
		suite.SetupTest() // reset
		ctx := sdk.WrapSDKContext(suite.ctx)
		req := &types.Bech32PrefixRequest{}
		res, err := suite.queryClient.Bech32Prefix(ctx, req)
		suite.Require().NoError(err)
		suite.Require().NotNil(res)
		suite.Require().Equal(res.Bech32Prefix, sdk.Bech32MainPrefix)
	})
}

func (suite *KeeperTestSuite) TestBech32FromAccAddr(){
	var req *types.Bech32FromAccAddrRequest
	_, _, addr := testdata.KeyTestPubAddr()

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		posttests func(res *types.Bech32FromAccAddrResponse)
	}{
		{
			"success",
			func() {
				addrBytes := []byte(addr)
				req = &types.Bech32FromAccAddrRequest{AccountAddr: addrBytes}
			},
			true,
			func(res *types.Bech32FromAccAddrResponse) {
				bech32, err := bech32.ConvertAndEncode(suite.app.AccountKeeper.GetBech32Prefix(), []byte(addr))
				suite.Require().NoError(err)
				suite.Require().NotNil(bech32)
				suite.Require().Equal(bech32, res.Bech32)
			},
		},
		{
			"req is nil",
			func() {
				req = &types.Bech32FromAccAddrRequest{}
			},
			false,
			func(res *types.Bech32FromAccAddrResponse) {},
		},
		{
			"empty account address in request",
			func() {
				emptyAddrBytes := []byte{}
				req = &types.Bech32FromAccAddrRequest{AccountAddr: emptyAddrBytes}
			},
			false,
			func(res *types.Bech32FromAccAddrResponse) {},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)

			res, err := suite.queryClient.Bech32FromAccAddr(ctx, req)

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

func (suite *KeeperTestSuite) TestAccAddrFromBech32() {
	var req  *types.AccAddrFromBech32Request
	_, _, addr := testdata.KeyTestPubAddr()

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		posttests func(res *types.AccAddrFromBech32Response)
	}{
		{
			"success",
			func() {
				bech32, _ := bech32.ConvertAndEncode(suite.app.AccountKeeper.GetBech32Prefix(), []byte(addr))
				req = &types.AccAddrFromBech32Request{Bech32: bech32}
			},
			true,
			func(res *types.AccAddrFromBech32Response) {
				suite.Require().True(bytes.Equal(res.AccountAddr,[]byte(addr)))
			},
		},
		{
			"req is nil",
			func() {
				req = &types.AccAddrFromBech32Request{}
			},
			false,
			func(res *types.AccAddrFromBech32Response) {},
		},
		{ 
			"Bech32 field in request is empty",
			func() {
				req = &types.AccAddrFromBech32Request{Bech32: ""}
			},
			false,
			func(res *types.AccAddrFromBech32Response) {},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)

			res, err := suite.queryClient.AccAddrFromBech32(ctx, req)

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





