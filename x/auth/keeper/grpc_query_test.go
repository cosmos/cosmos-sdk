package keeper_test

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

func (suite *KeeperTestSuite) TestBech32Prefix() {
		suite.SetupTest() // reset
		req := &types.Bech32PrefixRequest{}
		res, err := suite.queryClient.Bech32Prefix(context.Background(), req)
		suite.Require().NoError(err)
		suite.Require().NotNil(res)
		suite.Require().Equal(sdk.Bech32MainPrefix, res.Bech32Prefix)
	})
}

func (suite *KeeperTestSuite) TestAddressBytesToString() {

	_, _, addr := testdata.KeyTestPubAddr()
	addrBytes := []byte(addr)

	testCases := []struct {
		msg       string
		req       *types.AddressBytesToStringRequest
		expPass   bool
		posttests func(res *types.AddressBytesToStringResponse)
	}{
		{
			"success",
			&types.AddressBytesToStringRequest{AddressBytes: addrBytes},
			true,
			func(res *types.AddressBytesToStringResponse) {
				text, err := suite.app.AccountKeeper.GetAddressCdC().BytesToString(addrBytes)
				suite.Require().NoError(err)
				suite.Require().NotNil(text)
				suite.Require().Equal(text, res.AddressString)
			},
		},
		{
			"request is empty",
			&types.AddressBytesToStringRequest{},
			false,
			func(res *types.AddressBytesToStringResponse) {},
		},
		{
			"empty account address in request",
			&types.AddressBytesToStringRequest{AddressBytes: []byte{}},
			false,
			func(res *types.AddressBytesToStringResponse) {},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			ctx := sdk.WrapSDKContext(suite.ctx)

			res, err := suite.queryClient.AddressBytesToString(ctx, tc.req)

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

func (suite *KeeperTestSuite) TestAddressStringToBytes() {
	_, _, addr := testdata.KeyTestPubAddr()
	addrBytes := []byte(addr)
	text, err := suite.app.AccountKeeper.GetAddressCdC().BytesToString(addrBytes)
	suite.Require().NoError(err)

	testCases := []struct {
		msg       string
		req       *types.AddressStringToBytesRequest
		expPass   bool
		posttests func(res *types.AddressStringToBytesResponse)
	}{
		{
			"success",
			&types.AddressStringToBytesRequest{AddressString: text},
			true,
			func(res *types.AddressStringToBytesResponse) {
				suite.Require().True(bytes.Equal(res.AddressBytes, addrBytes))
			},
		},
		{
			"request is empty",
			&types.AddressStringToBytesRequest{},
			false,
			func(res *types.AddressStringToBytesResponse) {},
		},
		{
			"AddressString field in request is empty",
			&types.AddressStringToBytesRequest{AddressString: ""},
			false,
			func(res *types.AddressStringToBytesResponse) {},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			ctx := sdk.WrapSDKContext(suite.ctx)

			res, err := suite.queryClient.AddressStringToBytes(ctx, tc.req)

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
