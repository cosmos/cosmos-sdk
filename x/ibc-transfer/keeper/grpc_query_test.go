package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
)

func (suite *KeeperTestSuite) TestQueryConnection() {
	var (
		req      *types.QueryDenomTraceRequest
		expTrace types.DenomTrace
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"invalid hex hash",
			func() {
				req = &types.QueryDenomTraceRequest{
					Hash: "!@#!@#!",
				}
			},
			false,
		},
		{
			"not found denom trace",
			func() {
				expTrace.Trace = "transfer/channelToA/transfer/channelToB"
				expTrace.BaseDenom = "uatom"
				req = &types.QueryDenomTraceRequest{
					Hash: expTrace.Hash().String(),
				}
			},
			false,
		},
		{
			"success",
			func() {
				expTrace.Trace = "transfer/channelToA/transfer/channelToB"
				expTrace.BaseDenom = "uatom"
				suite.chainA.App.TransferKeeper.SetDenomTrace(suite.chainA.GetContext(), expTrace)

				req = &types.QueryDenomTraceRequest{
					Hash: expTrace.Hash().String(),
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.chainA.GetContext())

			res, err := suite.queryClient.DenomTrace(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(&expTrace, res.DenomTrace)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryConnections() {
	var (
		req       *types.QueryDenomTracesRequest
		expTraces = types.Traces(nil)
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty pagination",
			func() {
				req = &types.QueryDenomTracesRequest{}
			},
			true,
		},
		{
			"success",
			func() {
				expTraces = append(expTraces, types.DenomTrace{Trace: "", BaseDenom: "uatom"})
				expTraces = append(expTraces, types.DenomTrace{Trace: "transfer/channelToB", BaseDenom: "uatom"})
				expTraces = append(expTraces, types.DenomTrace{Trace: "transfer/channelToA/transfer/channelToB", BaseDenom: "uatom"})

				for _, trace := range expTraces {
					suite.chainA.App.TransferKeeper.SetDenomTrace(suite.chainA.GetContext(), trace)
				}

				req = &types.QueryDenomTracesRequest{
					Pagination: &query.PageRequest{
						Limit:      5,
						CountTotal: false,
					},
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.chainA.GetContext())

			res, err := suite.queryClient.DenomTraces(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				// FIXME: response is non-deterministic so we can't use Equal!
				suite.Require().ElementsMatch(expTraces, res.DenomTraces)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
