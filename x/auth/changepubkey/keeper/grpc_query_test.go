package keeper_test

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/changepubkey/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/changepubkey/types"
)

func (suite *KeeperTestSuite) TestGRPCQueryAccountPubKeyHistory() {
	var (
		req *types.QueryPubKeyHistoryRequest
	)
	_, pub1, addr := testdata.KeyTestPubAddr()
	_, pub2, _ := testdata.KeyTestPubAddr()

	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		posttests func(res *types.QueryPubKeyHistoryResponse)
	}{
		{
			"empty request",
			func() {
				req = &types.QueryPubKeyHistoryRequest{}
			},
			false,
			func(res *types.QueryPubKeyHistoryResponse) {},
		},
		{
			"invalid request",
			func() {
				req = &types.QueryPubKeyHistoryRequest{Address: ""}
			},
			false,
			func(res *types.QueryPubKeyHistoryResponse) {},
		},
		{
			"account not found",
			func() {
				req = &types.QueryPubKeyHistoryRequest{Address: addr.String()}
			},
			false,
			func(res *types.QueryPubKeyHistoryResponse) {},
		},
		{
			"success",
			func() {
				suite.app.AccountKeeper.SetAccount(suite.ctx,
					suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr))
				req = &types.QueryPubKeyHistoryRequest{Address: addr.String()}
			},
			true,
			func(res *types.QueryPubKeyHistoryResponse) {
				suite.Require().Equal(len(res.History), 1)
			},
		},
		{
			"success with more history",
			func() {
				suite.app.AccountKeeper.SetAccount(suite.ctx, suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr))

				msgServer := keeper.NewMsgServerImpl(suite.app.AccountKeeper, suite.app.AccountHistoryKeeper)
				_, err := msgServer.ChangePubKey(sdk.WrapSDKContext(suite.ctx), types.NewMsgChangePubKey(addr, pub1))
				suite.Require().NoError(err)
				suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Minute))

				_, err = msgServer.ChangePubKey(sdk.WrapSDKContext(suite.ctx), types.NewMsgChangePubKey(addr, pub2))
				suite.Require().NoError(err)
				suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Minute))

				req = &types.QueryPubKeyHistoryRequest{Address: addr.String()}
			},
			true,
			func(res *types.QueryPubKeyHistoryResponse) {
				suite.Require().Equal(len(res.History), 3)
				suite.Require().Equal(types.DecodePubKey(suite.app.AppCodec(), res.History[0].PubKey), nil)
				suite.Require().Equal(types.DecodePubKey(suite.app.AppCodec(), res.History[1].PubKey), pub1)
				suite.Require().Equal(types.DecodePubKey(suite.app.AppCodec(), res.History[2].PubKey), pub2)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)

			res, err := suite.queryClient.PubKeyHistory(ctx, req)

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

func (suite *KeeperTestSuite) TestGRPCQueryAccountPubKeyHistoricalEntry() {
	var (
		req *types.QueryPubKeyHistoricalEntryRequest
	)
	_, pub1, addr := testdata.KeyTestPubAddr()
	_, pub2, _ := testdata.KeyTestPubAddr()

	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		posttests func(res *types.QueryPubKeyHistoricalEntryResponse)
	}{
		{
			"empty request",
			func() {
				req = &types.QueryPubKeyHistoricalEntryRequest{}
			},
			false,
			func(res *types.QueryPubKeyHistoricalEntryResponse) {},
		},
		{
			"invalid request",
			func() {
				req = &types.QueryPubKeyHistoricalEntryRequest{Address: ""}
			},
			false,
			func(res *types.QueryPubKeyHistoricalEntryResponse) {},
		},
		{
			"account not found",
			func() {
				req = &types.QueryPubKeyHistoricalEntryRequest{Address: addr.String()}
			},
			true,
			func(res *types.QueryPubKeyHistoricalEntryResponse) {
				suite.Require().Equal(types.DecodePubKey(suite.app.AppCodec(), res.Entry.PubKey), nil)
			},
		},
		{
			"account found empty but pubkey",
			func() {
				suite.app.AccountKeeper.SetAccount(suite.ctx,
					suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr))
				req = &types.QueryPubKeyHistoricalEntryRequest{Address: addr.String()}
			},
			true,
			func(res *types.QueryPubKeyHistoricalEntryResponse) {
				suite.Require().Equal(types.DecodePubKey(suite.app.AppCodec(), res.Entry.PubKey), nil)
			},
		},
		{
			"query with more history",
			func() {
				suite.app.AccountKeeper.SetAccount(suite.ctx, suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr))
				suite.ChangeAccountPubKeys(addr, pub1, pub2)
				req = &types.QueryPubKeyHistoricalEntryRequest{Address: addr.String(), Time: suite.ctx.BlockTime().Truncate(100 * time.Minute)}
			},
			true,
			func(res *types.QueryPubKeyHistoricalEntryResponse) {
				suite.Require().Equal(types.DecodePubKey(suite.app.AppCodec(), res.Entry.PubKey), nil)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)

			res, err := suite.queryClient.PubKeyHistoricalEntry(ctx, req)

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

func (suite *KeeperTestSuite) TestGRPCQueryAccountLastPubKeyHistoricalEntry() {
	var (
		req *types.QueryLastPubKeyHistoricalEntryRequest
	)
	_, pub1, addr := testdata.KeyTestPubAddr()
	_, pub2, _ := testdata.KeyTestPubAddr()

	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		posttests func(res *types.QueryLastPubKeyHistoricalEntryResponse)
	}{
		{
			"empty request",
			func() {
				req = &types.QueryLastPubKeyHistoricalEntryRequest{}
			},
			false,
			func(res *types.QueryLastPubKeyHistoricalEntryResponse) {},
		},
		{
			"invalid request",
			func() {
				req = &types.QueryLastPubKeyHistoricalEntryRequest{Address: ""}
			},
			false,
			func(res *types.QueryLastPubKeyHistoricalEntryResponse) {},
		},
		{
			"account not found",
			func() {
				req = &types.QueryLastPubKeyHistoricalEntryRequest{Address: addr.String()}
			},
			true,
			func(res *types.QueryLastPubKeyHistoricalEntryResponse) {
				suite.Require().Equal(types.DecodePubKey(suite.app.AppCodec(), res.Entry.PubKey), nil)
			},
		},
		{
			"account found empty but pubkey",
			func() {
				suite.app.AccountKeeper.SetAccount(suite.ctx,
					suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr))
				req = &types.QueryLastPubKeyHistoricalEntryRequest{Address: addr.String()}
			},
			true,
			func(res *types.QueryLastPubKeyHistoricalEntryResponse) {
				suite.Require().Equal(types.DecodePubKey(suite.app.AppCodec(), res.Entry.PubKey), nil)
			},
		},
		{
			"query with more history",
			func() {
				suite.app.AccountKeeper.SetAccount(suite.ctx, suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr))
				suite.ChangeAccountPubKeys(addr, pub1, pub2)
				req = &types.QueryLastPubKeyHistoricalEntryRequest{Address: addr.String()}
			},
			true,
			func(res *types.QueryLastPubKeyHistoricalEntryResponse) {
				suite.Require().Equal(types.DecodePubKey(suite.app.AppCodec(), res.Entry.PubKey), pub1)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)

			res, err := suite.queryClient.LastPubKeyHistoricalEntry(ctx, req)

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

func (suite *KeeperTestSuite) TestGRPCQueryAccountCurrentPubKeyHistoricalEntry() {
	var (
		req *types.QueryCurrentPubKeyEntryRequest
	)
	_, pub1, addr := testdata.KeyTestPubAddr()
	_, pub2, _ := testdata.KeyTestPubAddr()

	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		posttests func(res *types.QueryCurrentPubKeyEntryResponse)
	}{
		{
			"empty request",
			func() {
				req = &types.QueryCurrentPubKeyEntryRequest{}
			},
			false,
			func(res *types.QueryCurrentPubKeyEntryResponse) {},
		},
		{
			"invalid request",
			func() {
				req = &types.QueryCurrentPubKeyEntryRequest{Address: ""}
			},
			false,
			func(res *types.QueryCurrentPubKeyEntryResponse) {},
		},
		{
			"account not found",
			func() {
				req = &types.QueryCurrentPubKeyEntryRequest{Address: addr.String()}
			},
			true,
			func(res *types.QueryCurrentPubKeyEntryResponse) {
				suite.Require().Equal(types.DecodePubKey(suite.app.AppCodec(), res.Entry.PubKey), nil)
			},
		},
		{
			"account found empty but pubkey",
			func() {
				suite.app.AccountKeeper.SetAccount(suite.ctx,
					suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr))
				req = &types.QueryCurrentPubKeyEntryRequest{Address: addr.String()}
			},
			true,
			func(res *types.QueryCurrentPubKeyEntryResponse) {
				suite.Require().Equal(types.DecodePubKey(suite.app.AppCodec(), res.Entry.PubKey), nil)
			},
		},
		{
			"query with more history",
			func() {
				suite.app.AccountKeeper.SetAccount(suite.ctx, suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr))
				suite.ChangeAccountPubKeys(addr, pub1, pub2)
				req = &types.QueryCurrentPubKeyEntryRequest{Address: addr.String()}
			},
			true,
			func(res *types.QueryCurrentPubKeyEntryResponse) {
				suite.Require().Equal(types.DecodePubKey(suite.app.AppCodec(), res.Entry.PubKey), pub2)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)

			res, err := suite.queryClient.CurrentPubKeyEntry(ctx, req)

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
