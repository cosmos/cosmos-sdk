package keeper_test

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/cosmos/gogoproto/proto"
	"github.com/golang/mock/gomock"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

const addrStr = "cosmos13c3d4wq2t22dl0dstraf8jc3f902e3fsy9n3wv"

var addrBytes = []byte{0x8e, 0x22, 0xda, 0xb8, 0xa, 0x5a, 0x94, 0xdf, 0xbd, 0xb0, 0x58, 0xfa, 0x93, 0xcb, 0x11, 0x49, 0x5e, 0xac, 0xc5, 0x30}

func (suite *KeeperTestSuite) TestGRPCQueryAccounts() {
	var req *types.QueryAccountsRequest
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
				suite.accountKeeper.SetAccount(suite.ctx,
					suite.accountKeeper.NewAccountWithAddress(suite.ctx, first))
				suite.accountKeeper.SetAccount(suite.ctx,
					suite.accountKeeper.NewAccountWithAddress(suite.ctx, second))
				req = &types.QueryAccountsRequest{}
			},
			true,
			func(res *types.QueryAccountsResponse) {
				addresses := make([]sdk.AccAddress, len(res.Accounts))
				for i, acc := range res.Accounts {
					var account sdk.AccountI
					err := suite.encCfg.InterfaceRegistry.UnpackAny(acc, &account)
					suite.Require().NoError(err)
					addresses[i] = account.GetAddress()
				}
				suite.Subset(addresses, []sdk.AccAddress{first, second})
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			res, err := suite.queryClient.Accounts(suite.ctx, req)

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
	var req *types.QueryAccountRequest
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
				suite.accountKeeper.SetAccount(suite.ctx,
					suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr))
				req = &types.QueryAccountRequest{Address: addr.String()}
			},
			true,
			func(res *types.QueryAccountResponse) {
				var newAccount sdk.AccountI
				err := suite.encCfg.InterfaceRegistry.UnpackAny(res.Account, &newAccount)
				suite.Require().NoError(err)
				suite.Require().NotNil(newAccount)
				suite.Require().True(addr.Equals(newAccount.GetAddress()))
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset
			suite.acctsModKeeper.EXPECT().IsAccountsModuleAccount(gomock.Any(), gomock.Any()).Return(false).AnyTimes()

			tc.malleate()
			res, err := suite.queryClient.Account(suite.ctx, req)

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

func (suite *KeeperTestSuite) TestGRPCQueryAccountAddressByID() {
	var req *types.QueryAccountAddressByIDRequest
	_, _, addr := testdata.KeyTestPubAddr()

	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		posttests func(res *types.QueryAccountAddressByIDResponse)
	}{
		{
			"invalid request",
			func() {
				req = &types.QueryAccountAddressByIDRequest{Id: -1}
			},
			false,
			func(res *types.QueryAccountAddressByIDResponse) {},
		},
		{
			"account address not found",
			func() {
				req = &types.QueryAccountAddressByIDRequest{Id: math.MaxInt64}
			},
			false,
			func(res *types.QueryAccountAddressByIDResponse) {},
		},
		{
			"valid account-id",
			func() {
				account := suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr)
				suite.accountKeeper.SetAccount(suite.ctx, account)
				req = &types.QueryAccountAddressByIDRequest{AccountId: account.GetAccountNumber()}
			},
			true,
			func(res *types.QueryAccountAddressByIDResponse) {
				suite.Require().NotNil(res.AccountAddress)
			},
		},
		{
			"invalid request",
			func() {
				account := suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr)
				suite.accountKeeper.SetAccount(suite.ctx, account)
				req = &types.QueryAccountAddressByIDRequest{Id: 1}
			},
			false,
			func(res *types.QueryAccountAddressByIDResponse) {},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			res, err := suite.queryClient.AccountAddressByID(suite.ctx, req)

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

func (suite *KeeperTestSuite) TestGRPCQueryParams() {
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
				expParams = suite.accountKeeper.GetParams(suite.ctx)
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			res, err := suite.queryClient.Params(suite.ctx, req)

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

func (suite *KeeperTestSuite) TestGRPCQueryModuleAccountByName() {
	var req *types.QueryModuleAccountByNameRequest

	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		posttests func(res *types.QueryModuleAccountByNameResponse)
	}{
		{
			"success",
			func() {
				req = &types.QueryModuleAccountByNameRequest{Name: "mint"}
			},
			true,
			func(res *types.QueryModuleAccountByNameResponse) {
				var account types.AccountI
				err := suite.app.InterfaceRegistry().UnpackAny(res.Account, &account)
				suite.Require().NoError(err)

				moduleAccount, ok := account.(types.ModuleAccountI)
				suite.Require().True(ok)
				suite.Require().Equal(moduleAccount.GetName(), "mint")
			},
		},
		{
			"invalid module name",
			func() {
				req = &types.QueryModuleAccountByNameRequest{Name: "gover"}
			},
			false,
			func(res *types.QueryModuleAccountByNameResponse) {
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset
			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)
			res, err := suite.queryClient.ModuleAccountByName(ctx, req)
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
