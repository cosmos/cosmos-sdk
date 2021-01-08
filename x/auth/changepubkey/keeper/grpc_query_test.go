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
	_, _, addr := testdata.KeyTestPubAddr()

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
				_, pub1, _ := testdata.KeyTestPubAddr()
				_, err := msgServer.ChangePubKey(sdk.WrapSDKContext(suite.ctx), types.NewMsgChangePubKey(addr, pub1))
				suite.Require().NoError(err)
				suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Minute))

				_, pub2, _ := testdata.KeyTestPubAddr()
				_, err = msgServer.ChangePubKey(sdk.WrapSDKContext(suite.ctx), types.NewMsgChangePubKey(addr, pub2))
				suite.Require().NoError(err)
				suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Minute))

				req = &types.QueryPubKeyHistoryRequest{Address: addr.String()}
			},
			true,
			func(res *types.QueryPubKeyHistoryResponse) {
				suite.Require().Equal(len(res.History), 3)
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
