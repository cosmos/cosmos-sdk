package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"pgregory.net/rapid"
)

type RapidTestSuite struct {
	suite.Suite

	ctx sdk.Context

	queryClient   types.QueryClient
	accountKeeper keeper.AccountKeeper
	msgServer     types.MsgServer
	encCfg        moduletestutil.TestEncodingConfig
}

func TestRapidTestSuite(t *testing.T) {
	suite.Run(t, new(RapidTestSuite))
}

func (suite *RapidTestSuite) SetupTest() {
	suite.encCfg = moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{})

	key := sdk.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, sdk.NewTransientStoreKey("transient_test"))
	suite.ctx = testCtx.Ctx.WithBlockHeader(tmproto.Header{})

	maccPerms := map[string][]string{
		"fee_collector":          nil,
		"mint":                   {"minter"},
		"bonded_tokens_pool":     {"burner", "staking"},
		"not_bonded_tokens_pool": {"burner", "staking"},
		multiPerm:                {"burner", "minter", "staking"},
		randomPerm:               {"random"},
	}

	suite.accountKeeper = keeper.NewAccountKeeper(
		suite.encCfg.Codec,
		key,
		types.ProtoBaseAccount,
		maccPerms,
		"cosmos",
		types.NewModuleAddress("gov").String(),
	)

	suite.msgServer = keeper.NewMsgServerImpl(suite.accountKeeper)
	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, suite.accountKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
}

func (suite *RapidTestSuite) TestGRPCQueryAccounts() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		pkBz := rapid.SliceOfN(rapid.Byte(), 20, 20).Draw(t, "hex")
		addr := sdk.AccAddress(pkBz)
		suite.accountKeeper.SetAccount(suite.ctx,
			suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr))

		for i := 0; i < 1000; i++ {
			acc, err := suite.queryClient.Account(suite.ctx, &types.QueryAccountRequest{Address: addr.String()})
			suite.Require().NoError(err)
			suite.Require().NotNil(acc)
			var account types.AccountI

			err = suite.encCfg.InterfaceRegistry.UnpackAny(acc.Account, &account)
			suite.Require().NoError(err)
			suite.Require().Equal(account.GetAddress(), addr)
		}
	})
}

// func (suite *KeeperTestSuite) TestGRPCQueryAccount() {
// 	var req *types.QueryAccountRequest
// 	_, _, addr := testdata.KeyTestPubAddr()

// 	testCases := []struct {
// 		msg       string
// 		malleate  func()
// 		expPass   bool
// 		posttests func(res *types.QueryAccountResponse)
// 	}{
// 		{
// 			"empty request",
// 			func() {
// 				req = &types.QueryAccountRequest{}
// 			},
// 			false,
// 			func(res *types.QueryAccountResponse) {},
// 		},
// 		{
// 			"invalid request",
// 			func() {
// 				req = &types.QueryAccountRequest{Address: ""}
// 			},
// 			false,
// 			func(res *types.QueryAccountResponse) {},
// 		},
// 		{
// 			"invalid request with empty byte array",
// 			func() {
// 				req = &types.QueryAccountRequest{Address: ""}
// 			},
// 			false,
// 			func(res *types.QueryAccountResponse) {},
// 		},
// 		{
// 			"account not found",
// 			func() {
// 				req = &types.QueryAccountRequest{Address: addr.String()}
// 			},
// 			false,
// 			func(res *types.QueryAccountResponse) {},
// 		},
// 		{
// 			"success",
// 			func() {
// 				suite.accountKeeper.SetAccount(suite.ctx,
// 					suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr))
// 				req = &types.QueryAccountRequest{Address: addr.String()}
// 			},
// 			true,
// 			func(res *types.QueryAccountResponse) {
// 				var newAccount types.AccountI
// 				err := suite.encCfg.InterfaceRegistry.UnpackAny(res.Account, &newAccount)
// 				suite.Require().NoError(err)
// 				suite.Require().NotNil(newAccount)
// 				suite.Require().True(addr.Equals(newAccount.GetAddress()))
// 			},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
// 			suite.SetupTest() // reset

// 			tc.malleate()
// 			ctx := sdk.WrapSDKContext(suite.ctx)

// 			res, err := suite.queryClient.Account(ctx, req)

// 			if tc.expPass {
// 				suite.Require().NoError(err)
// 				suite.Require().NotNil(res)
// 			} else {
// 				suite.Require().Error(err)
// 				suite.Require().Nil(res)
// 			}

// 			tc.posttests(res)
// 		})
// 	}
// }

// func (suite *KeeperTestSuite) TestGRPCQueryAccountAddressByID() {
// 	var req *types.QueryAccountAddressByIDRequest
// 	_, _, addr := testdata.KeyTestPubAddr()

// 	testCases := []struct {
// 		msg       string
// 		malleate  func()
// 		expPass   bool
// 		posttests func(res *types.QueryAccountAddressByIDResponse)
// 	}{
// 		{
// 			"invalid request",
// 			func() {
// 				req = &types.QueryAccountAddressByIDRequest{Id: -1}
// 			},
// 			false,
// 			func(res *types.QueryAccountAddressByIDResponse) {},
// 		},
// 		{
// 			"account address not found",
// 			func() {
// 				req = &types.QueryAccountAddressByIDRequest{Id: math.MaxInt64}
// 			},
// 			false,
// 			func(res *types.QueryAccountAddressByIDResponse) {},
// 		},
// 		{
// 			"valid request",
// 			func() {
// 				account := suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr)
// 				suite.accountKeeper.SetAccount(suite.ctx, account)
// 				req = &types.QueryAccountAddressByIDRequest{Id: int64(account.GetAccountNumber())}
// 			},
// 			true,
// 			func(res *types.QueryAccountAddressByIDResponse) {
// 				suite.Require().NotNil(res.AccountAddress)
// 			},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
// 			suite.SetupTest() // reset

// 			tc.malleate()
// 			ctx := sdk.WrapSDKContext(suite.ctx)

// 			res, err := suite.queryClient.AccountAddressByID(ctx, req)

// 			if tc.expPass {
// 				suite.Require().NoError(err)
// 				suite.Require().NotNil(res)
// 			} else {
// 				suite.Require().Error(err)
// 				suite.Require().Nil(res)
// 			}

// 			tc.posttests(res)
// 		})
// 	}
// }

// func (suite *KeeperTestSuite) TestGRPCQueryParameters() {
// 	var (
// 		req       *types.QueryParamsRequest
// 		expParams types.Params
// 	)

// 	testCases := []struct {
// 		msg      string
// 		malleate func()
// 		expPass  bool
// 	}{
// 		{
// 			"success",
// 			func() {
// 				req = &types.QueryParamsRequest{}
// 				expParams = suite.accountKeeper.GetParams(suite.ctx)
// 			},
// 			true,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
// 			suite.SetupTest() // reset

// 			tc.malleate()
// 			ctx := sdk.WrapSDKContext(suite.ctx)

// 			res, err := suite.queryClient.Params(ctx, req)

// 			if tc.expPass {
// 				suite.Require().NoError(err)
// 				suite.Require().NotNil(res)
// 				suite.Require().Equal(expParams, res.Params)
// 			} else {
// 				suite.Require().Error(err)
// 				suite.Require().Nil(res)
// 			}
// 		})
// 	}
// }

// func (suite *KeeperTestSuite) TestGRPCQueryModuleAccounts() {
// 	var req *types.QueryModuleAccountsRequest

// 	testCases := []struct {
// 		msg       string
// 		malleate  func()
// 		expPass   bool
// 		posttests func(res *types.QueryModuleAccountsResponse)
// 	}{
// 		{
// 			"success",
// 			func() {
// 				req = &types.QueryModuleAccountsRequest{}
// 			},
// 			true,
// 			func(res *types.QueryModuleAccountsResponse) {
// 				mintModuleExists := false
// 				for _, acc := range res.Accounts {
// 					var account types.AccountI
// 					err := suite.encCfg.InterfaceRegistry.UnpackAny(acc, &account)
// 					suite.Require().NoError(err)

// 					moduleAccount, ok := account.(types.ModuleAccountI)

// 					suite.Require().True(ok)
// 					if moduleAccount.GetName() == "mint" {
// 						mintModuleExists = true
// 					}
// 				}
// 				suite.Require().True(mintModuleExists)
// 			},
// 		},
// 		{
// 			"invalid module name",
// 			func() {
// 				req = &types.QueryModuleAccountsRequest{}
// 			},
// 			true,
// 			func(res *types.QueryModuleAccountsResponse) {
// 				mintModuleExists := false
// 				for _, acc := range res.Accounts {
// 					var account types.AccountI
// 					err := suite.encCfg.InterfaceRegistry.UnpackAny(acc, &account)
// 					suite.Require().NoError(err)

// 					moduleAccount, ok := account.(types.ModuleAccountI)

// 					suite.Require().True(ok)
// 					if moduleAccount.GetName() == "falseCase" {
// 						mintModuleExists = true
// 					}
// 				}
// 				suite.Require().False(mintModuleExists)
// 			},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
// 			suite.SetupTest() // reset

// 			tc.malleate()
// 			ctx := sdk.WrapSDKContext(suite.ctx)

// 			res, err := suite.queryClient.ModuleAccounts(ctx, req)

// 			if tc.expPass {
// 				suite.Require().NoError(err)
// 				suite.Require().NotNil(res)
// 				// Make sure output is sorted alphabetically.
// 				var moduleNames []string
// 				for _, any := range res.Accounts {
// 					var account types.AccountI
// 					err := suite.encCfg.InterfaceRegistry.UnpackAny(any, &account)
// 					suite.Require().NoError(err)
// 					moduleAccount, ok := account.(types.ModuleAccountI)
// 					suite.Require().True(ok)
// 					moduleNames = append(moduleNames, moduleAccount.GetName())
// 				}
// 				suite.Require().True(sort.StringsAreSorted(moduleNames))
// 			} else {
// 				suite.Require().Error(err)
// 				suite.Require().Nil(res)
// 			}

// 			tc.posttests(res)
// 		})
// 	}
// }

// func (suite *KeeperTestSuite) TestBech32Prefix() {
// 	suite.SetupTest() // reset
// 	req := &types.Bech32PrefixRequest{}
// 	res, err := suite.queryClient.Bech32Prefix(context.Background(), req)
// 	suite.Require().NoError(err)
// 	suite.Require().NotNil(res)
// 	suite.Require().Equal(sdk.Bech32MainPrefix, res.Bech32Prefix)
// }

// func (suite *KeeperTestSuite) TestAddressBytesToString() {
// 	testCases := []struct {
// 		msg     string
// 		req     *types.AddressBytesToStringRequest
// 		expPass bool
// 	}{
// 		{
// 			"success",
// 			&types.AddressBytesToStringRequest{AddressBytes: addrBytes},
// 			true,
// 		},
// 		{
// 			"request is empty",
// 			&types.AddressBytesToStringRequest{},
// 			false,
// 		},
// 		{
// 			"empty account address in request",
// 			&types.AddressBytesToStringRequest{AddressBytes: []byte{}},
// 			false,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
// 			suite.SetupTest() // reset

// 			res, err := suite.queryClient.AddressBytesToString(context.Background(), tc.req)

// 			if tc.expPass {
// 				suite.Require().NoError(err)
// 				suite.Require().NotNil(res)
// 				suite.Require().Equal(res.AddressString, addrStr)
// 			} else {
// 				suite.Require().Error(err)
// 				suite.Require().Nil(res)
// 			}
// 		})
// 	}
// }

// func (suite *KeeperTestSuite) TestAddressStringToBytes() {
// 	testCases := []struct {
// 		msg     string
// 		req     *types.AddressStringToBytesRequest
// 		expPass bool
// 	}{
// 		{
// 			"success",
// 			&types.AddressStringToBytesRequest{AddressString: addrStr},
// 			true,
// 		},
// 		{
// 			"request is empty",
// 			&types.AddressStringToBytesRequest{},
// 			false,
// 		},
// 		{
// 			"AddressString field in request is empty",
// 			&types.AddressStringToBytesRequest{AddressString: ""},
// 			false,
// 		},
// 		{
// 			"address prefix is incorrect",
// 			&types.AddressStringToBytesRequest{AddressString: "regen13c3d4wq2t22dl0dstraf8jc3f902e3fsy9n3wv"},
// 			false,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
// 			suite.SetupTest() // reset

// 			res, err := suite.queryClient.AddressStringToBytes(context.Background(), tc.req)

// 			if tc.expPass {
// 				suite.Require().NoError(err)
// 				suite.Require().NotNil(res)
// 				suite.Require().True(bytes.Equal(res.AddressBytes, addrBytes))
// 			} else {
// 				suite.Require().Error(err)
// 				suite.Require().Nil(res)
// 			}
// 		})
// 	}
// }
