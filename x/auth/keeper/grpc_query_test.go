package keeper_test

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
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

func (suite *KeeperTestSuite) TestGRPCQueryModuleAccounts() {
	var req *types.QueryModuleAccountsRequest

	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		posttests func(res *types.QueryModuleAccountsResponse)
	}{
		{
			"success",
			func() {
				req = &types.QueryModuleAccountsRequest{}
			},
			true,
			func(res *types.QueryModuleAccountsResponse) {
				mintModuleExists := false
				for _, acc := range res.Accounts {
					var account sdk.AccountI
					err := suite.encCfg.InterfaceRegistry.UnpackAny(acc, &account)
					suite.Require().NoError(err)

					moduleAccount, ok := account.(sdk.ModuleAccountI)

					suite.Require().True(ok)
					if moduleAccount.GetName() == "mint" {
						mintModuleExists = true
					}
				}
				suite.Require().True(mintModuleExists)
			},
		},
		{
			"invalid module name",
			func() {
				req = &types.QueryModuleAccountsRequest{}
			},
			true,
			func(res *types.QueryModuleAccountsResponse) {
				mintModuleExists := false
				for _, acc := range res.Accounts {
					var account sdk.AccountI
					err := suite.encCfg.InterfaceRegistry.UnpackAny(acc, &account)
					suite.Require().NoError(err)

					moduleAccount, ok := account.(sdk.ModuleAccountI)

					suite.Require().True(ok)
					if moduleAccount.GetName() == "falseCase" {
						mintModuleExists = true
					}
				}
				suite.Require().False(mintModuleExists)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			res, err := suite.queryClient.ModuleAccounts(suite.ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				// Make sure output is sorted alphabetically.
				var moduleNames []string
				for _, any := range res.Accounts {
					var account sdk.AccountI
					err := suite.encCfg.InterfaceRegistry.UnpackAny(any, &account)
					suite.Require().NoError(err)
					moduleAccount, ok := account.(sdk.ModuleAccountI)
					suite.Require().True(ok)
					moduleNames = append(moduleNames, moduleAccount.GetName())
				}
				suite.Require().True(sort.StringsAreSorted(moduleNames))
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(res)
			}

			tc.posttests(res)
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
				var account sdk.AccountI
				err := suite.encCfg.InterfaceRegistry.UnpackAny(res.Account, &account)
				suite.Require().NoError(err)

				moduleAccount, ok := account.(sdk.ModuleAccountI)
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
			res, err := suite.queryClient.ModuleAccountByName(suite.ctx, req)
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

func (suite *KeeperTestSuite) TestBech32Prefix() {
	suite.SetupTest() // reset
	req := &types.Bech32PrefixRequest{}
	res, err := suite.queryClient.Bech32Prefix(context.Background(), req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(sdk.Bech32MainPrefix, res.Bech32Prefix)
}

func (suite *KeeperTestSuite) TestAddressBytesToString() {
	testCases := []struct {
		msg     string
		req     *types.AddressBytesToStringRequest
		expPass bool
	}{
		{
			"success",
			&types.AddressBytesToStringRequest{AddressBytes: addrBytes},
			true,
		},
		{
			"request is empty",
			&types.AddressBytesToStringRequest{},
			false,
		},
		{
			"empty account address in request",
			&types.AddressBytesToStringRequest{AddressBytes: []byte{}},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			res, err := suite.queryClient.AddressBytesToString(context.Background(), tc.req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(res.AddressString, addrStr)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestAddressStringToBytes() {
	testCases := []struct {
		msg     string
		req     *types.AddressStringToBytesRequest
		expPass bool
	}{
		{
			"success",
			&types.AddressStringToBytesRequest{AddressString: addrStr},
			true,
		},
		{
			"request is empty",
			&types.AddressStringToBytesRequest{},
			false,
		},
		{
			"AddressString field in request is empty",
			&types.AddressStringToBytesRequest{AddressString: ""},
			false,
		},
		{
			"address prefix is incorrect",
			&types.AddressStringToBytesRequest{AddressString: "regen13c3d4wq2t22dl0dstraf8jc3f902e3fsy9n3wv"},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			res, err := suite.queryClient.AddressStringToBytes(context.Background(), tc.req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().True(bytes.Equal(res.AddressBytes, addrBytes))
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryAccountInfo() {
	_, pk, addr := testdata.KeyTestPubAddr()
	acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr)
	suite.Require().NoError(acc.SetPubKey(pk))
	suite.Require().NoError(acc.SetSequence(10))
	suite.accountKeeper.SetAccount(suite.ctx, acc)

	res, err := suite.queryClient.AccountInfo(context.Background(), &types.QueryAccountInfoRequest{
		Address: addr.String(),
	})

	suite.Require().NoError(err)
	suite.Require().NotNil(res.Info)
	suite.Require().Equal(addr.String(), res.Info.Address)
	suite.Require().Equal(acc.GetAccountNumber(), res.Info.AccountNumber)
	suite.Require().Equal(acc.GetSequence(), res.Info.Sequence)
	suite.Require().Equal("/"+proto.MessageName(pk), res.Info.PubKey.TypeUrl)
	pkBz, err := proto.Marshal(pk)
	suite.Require().NoError(err)
	suite.Require().Equal(pkBz, res.Info.PubKey.Value)
}

func (suite *KeeperTestSuite) TestQueryAccountInfoWithoutPubKey() {
	acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr)
	suite.accountKeeper.SetAccount(suite.ctx, acc)

	res, err := suite.queryClient.AccountInfo(context.Background(), &types.QueryAccountInfoRequest{
		Address: addr.String(),
	})

	suite.Require().NoError(err)
	suite.Require().NotNil(res.Info)
	suite.Require().Equal(addr.String(), res.Info.Address)
	suite.Require().Nil(res.Info.PubKey)
}

func (suite *KeeperTestSuite) TestGRPCQueryRekeyedAccounts() {
	_, pk0, accAddr := testdata.KeyTestPubAddr()
	_, pk1, k1Addr := testdata.KeyTestPubAddr()
	_, pk2, k2Addr := testdata.KeyTestPubAddr()

	acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, accAddr)
	suite.Require().NoError(acc.SetPubKey(pk0))
	suite.accountKeeper.SetAccount(suite.ctx, acc)

	query := func(a sdk.AccAddress) []string {
		res, err := suite.queryClient.RekeyedAccounts(context.Background(), &types.QueryRekeyedAccountsRequest{Address: a.String()})
		suite.Require().NoError(err)
		return res.Addresses
	}

	// before any rotation, no key points at a rekeyed account
	suite.Require().Empty(query(accAddr))
	suite.Require().Empty(query(k1Addr))

	// rotate A to k1
	suite.Require().NoError(suite.accountKeeper.ApplyRekey(suite.ctx, suite.accountKeeper.GetAccount(suite.ctx, accAddr), pk1))
	suite.Require().Equal([]string{accAddr.String()}, query(k1Addr))
	suite.Require().Empty(query(accAddr))

	// rotate A to k2: the k1 entry goes away
	suite.Require().NoError(suite.accountKeeper.ApplyRekey(suite.ctx, suite.accountKeeper.GetAccount(suite.ctx, accAddr), pk2))
	suite.Require().Empty(query(k1Addr))
	suite.Require().Equal([]string{accAddr.String()}, query(k2Addr))

	// rotate A back to its natural key: no index entry remains
	suite.Require().NoError(suite.accountKeeper.ApplyRekey(suite.ctx, suite.accountKeeper.GetAccount(suite.ctx, accAddr), pk0))
	suite.Require().Empty(query(k2Addr))
	suite.Require().Empty(query(accAddr))
}

func (suite *KeeperTestSuite) TestGRPCQueryRekeyedAccountsMultiple() {
	_, pkA, addrA := testdata.KeyTestPubAddr()
	_, pkB, addrB := testdata.KeyTestPubAddr()
	_, pk1, k1Addr := testdata.KeyTestPubAddr()

	for _, kv := range []struct {
		addr sdk.AccAddress
		pk   cryptotypes.PubKey
	}{{addrA, pkA}, {addrB, pkB}} {
		acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, kv.addr)
		suite.Require().NoError(acc.SetPubKey(kv.pk))
		suite.accountKeeper.SetAccount(suite.ctx, acc)
		suite.Require().NoError(suite.accountKeeper.ApplyRekey(suite.ctx, suite.accountKeeper.GetAccount(suite.ctx, kv.addr), pk1))
	}

	res, err := suite.queryClient.RekeyedAccounts(context.Background(), &types.QueryRekeyedAccountsRequest{Address: k1Addr.String()})
	suite.Require().NoError(err)
	expected := []string{addrA.String(), addrB.String()}
	if bytes.Compare(addrA, addrB) > 0 {
		expected = []string{addrB.String(), addrA.String()}
	}
	suite.Require().Equal(expected, res.Addresses)
}

func (suite *KeeperTestSuite) TestGRPCQueryRekeyedAccountsPagination() {
	_, pk1, k1Addr := testdata.KeyTestPubAddr()

	var expected []string
	for range 3 {
		_, pk, addr := testdata.KeyTestPubAddr()
		acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr)
		suite.Require().NoError(acc.SetPubKey(pk))
		suite.accountKeeper.SetAccount(suite.ctx, acc)
		suite.Require().NoError(suite.accountKeeper.ApplyRekey(suite.ctx, suite.accountKeeper.GetAccount(suite.ctx, addr), pk1))
		expected = append(expected, addr.String())
	}
	sort.Slice(expected, func(i, j int) bool {
		a, _ := sdk.AccAddressFromBech32(expected[i])
		b, _ := sdk.AccAddressFromBech32(expected[j])
		return bytes.Compare(a, b) < 0
	})

	res, err := suite.queryClient.RekeyedAccounts(context.Background(), &types.QueryRekeyedAccountsRequest{
		Address:    k1Addr.String(),
		Pagination: &query.PageRequest{Limit: 2, CountTotal: true},
	})
	suite.Require().NoError(err)
	suite.Require().Equal(expected[:2], res.Addresses)
	suite.Require().NotNil(res.Pagination)
	suite.Require().NotEmpty(res.Pagination.NextKey)
	suite.Require().Equal(uint64(3), res.Pagination.Total)

	res, err = suite.queryClient.RekeyedAccounts(context.Background(), &types.QueryRekeyedAccountsRequest{
		Address:    k1Addr.String(),
		Pagination: &query.PageRequest{Key: res.Pagination.NextKey, Limit: 2},
	})
	suite.Require().NoError(err)
	suite.Require().Equal(expected[2:], res.Addresses)
	suite.Require().Empty(res.Pagination.NextKey)
}

func (suite *KeeperTestSuite) TestGRPCQueryPubKeyHistoryPagination() {
	_, pk0, accAddr := testdata.KeyTestPubAddr()
	acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, accAddr)
	suite.Require().NoError(acc.SetPubKey(pk0))
	suite.accountKeeper.SetAccount(suite.ctx, acc)

	for h := int64(1); h <= 3; h++ {
		_, pk, _ := testdata.KeyTestPubAddr()
		ctx := suite.ctx.WithBlockHeight(h)
		suite.Require().NoError(suite.accountKeeper.ApplyRekey(ctx, suite.accountKeeper.GetAccount(ctx, accAddr), pk))
	}

	res, err := suite.queryClient.PubKeyHistory(context.Background(), &types.QueryPubKeyHistoryRequest{
		Address:    accAddr.String(),
		Pagination: &query.PageRequest{Limit: 2, CountTotal: true},
	})
	suite.Require().NoError(err)
	suite.Require().Len(res.Entries, 2)
	suite.Require().Equal(int64(1), res.Entries[0].ReplacedAtHeight)
	suite.Require().Equal(int64(2), res.Entries[1].ReplacedAtHeight)
	suite.Require().NotNil(res.Pagination)
	suite.Require().NotEmpty(res.Pagination.NextKey)
	suite.Require().Equal(uint64(3), res.Pagination.Total)

	res, err = suite.queryClient.PubKeyHistory(context.Background(), &types.QueryPubKeyHistoryRequest{
		Address:    accAddr.String(),
		Pagination: &query.PageRequest{Key: res.Pagination.NextKey, Limit: 2},
	})
	suite.Require().NoError(err)
	suite.Require().Len(res.Entries, 1)
	suite.Require().Equal(int64(3), res.Entries[0].ReplacedAtHeight)
	suite.Require().Empty(res.Pagination.NextKey)
}

func (suite *KeeperTestSuite) TestGRPCQueryPubKeyHistory() {
	_, pk0, accAddr := testdata.KeyTestPubAddr()
	_, pk1, k1Addr := testdata.KeyTestPubAddr()
	_, pk2, k2Addr := testdata.KeyTestPubAddr()

	acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, accAddr)
	suite.Require().NoError(acc.SetPubKey(pk0))
	suite.accountKeeper.SetAccount(suite.ctx, acc)

	query := func(a sdk.AccAddress) []types.PubKeyHistoryEntry {
		res, err := suite.queryClient.PubKeyHistory(context.Background(), &types.QueryPubKeyHistoryRequest{Address: a.String()})
		suite.Require().NoError(err)
		return res.Entries
	}

	suite.Require().Empty(query(accAddr))

	ctx := suite.ctx.WithBlockHeight(7)
	suite.Require().NoError(suite.accountKeeper.ApplyRekey(ctx, suite.accountKeeper.GetAccount(ctx, accAddr), pk1))

	entries := query(accAddr)
	suite.Require().Len(entries, 1)
	suite.Require().Equal(int64(7), entries[0].ReplacedAtHeight)
	suite.Require().Equal(k1Addr.String(), entries[0].NewKeyAddress)
	suite.Require().Equal("/"+proto.MessageName(pk0), entries[0].PubKey.TypeUrl)
	pkBz, err := proto.Marshal(pk0)
	suite.Require().NoError(err)
	suite.Require().Equal(pkBz, entries[0].PubKey.Value)

	// the key's natural address has no history of its own
	suite.Require().Empty(query(k1Addr))

	ctx = suite.ctx.WithBlockHeight(9)
	suite.Require().NoError(suite.accountKeeper.ApplyRekey(ctx, suite.accountKeeper.GetAccount(ctx, accAddr), pk2))

	entries = query(accAddr)
	suite.Require().Len(entries, 2)
	suite.Require().Equal(int64(7), entries[0].ReplacedAtHeight)
	suite.Require().Equal(int64(9), entries[1].ReplacedAtHeight)
	suite.Require().Equal(k2Addr.String(), entries[1].NewKeyAddress)
	pkBz, err = proto.Marshal(pk1)
	suite.Require().NoError(err)
	suite.Require().Equal(pkBz, entries[1].PubKey.Value)
}

func (suite *KeeperTestSuite) TestGRPCQueryRekeyInvalidRequests() {
	for _, tc := range []struct {
		msg  string
		addr string
	}{
		{"empty address", ""},
		{"invalid bech32", "cosmos1invalid"},
		{"wrong prefix", "osmo13c3d4wq2t22dl0dstraf8jc3f902e3fsy9n3wv"},
	} {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			_, err := suite.queryClient.RekeyedAccounts(context.Background(), &types.QueryRekeyedAccountsRequest{Address: tc.addr})
			suite.Require().Equal(codes.InvalidArgument, status.Code(err), err)

			_, err = suite.queryClient.PubKeyHistory(context.Background(), &types.QueryPubKeyHistoryRequest{Address: tc.addr})
			suite.Require().Equal(codes.InvalidArgument, status.Code(err), err)
		})
	}

	qs := keeper.NewQueryServer(suite.accountKeeper)
	_, err := qs.RekeyedAccounts(suite.ctx, nil)
	suite.Require().Equal(codes.InvalidArgument, status.Code(err))
	_, err = qs.PubKeyHistory(suite.ctx, nil)
	suite.Require().Equal(codes.InvalidArgument, status.Code(err))
}
