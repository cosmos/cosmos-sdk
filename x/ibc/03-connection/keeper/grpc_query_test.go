package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

func (suite *KeeperTestSuite) TestQueryConnection() {
	var (
		req           *types.QueryConnectionRequest
		expConnection types.ConnectionEnd
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = nil
			},
			false,
		},
		{"invalid connectionID",
			func() {
				req = &types.QueryConnectionRequest{}
			},
			false,
		},
		{"connection not found",
			func() {
				req = &types.QueryConnectionRequest{
					ConnectionID: testConnectionIDB,
				}
			},
			false,
		},
		{
			"success",
			func() {
				counterparty := types.NewCounterparty(testClientIDA, testConnectionIDA, commitmenttypes.NewMerklePrefix(suite.chainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes()))
				expConnection = types.NewConnectionEnd(types.INIT, testConnectionIDB, testClientIDB, counterparty, types.GetCompatibleVersions())
				suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainA.GetContext(), testConnectionIDB, expConnection)

				req = &types.QueryConnectionRequest{
					ConnectionID: testConnectionIDB,
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

			res, err := suite.chainA.QueryServer.Connection(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(&expConnection, res.Connection)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryConnections() {
	var (
		req            *types.QueryConnectionsRequest
		expConnections = []*types.ConnectionEnd{}
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = nil
			},
			false,
		},
		{
			"empty pagination",
			func() {
				req = &types.QueryConnectionsRequest{}
			},
			true,
		},
		{
			"success",
			func() {
				counterparty1 := types.NewCounterparty(testClientIDA, testConnectionIDA, commitmenttypes.NewMerklePrefix(suite.oldchainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes()))
				counterparty2 := types.NewCounterparty(testClientIDB, testConnectionIDB, commitmenttypes.NewMerklePrefix(suite.oldchainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes()))
				counterparty3 := types.NewCounterparty(testClientID3, testConnectionID3, commitmenttypes.NewMerklePrefix(suite.oldchainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes()))

				conn1 := types.NewConnectionEnd(types.INIT, testConnectionIDA, testClientIDA, counterparty3, types.GetCompatibleVersions())
				conn2 := types.NewConnectionEnd(types.INIT, testConnectionIDB, testClientIDB, counterparty1, types.GetCompatibleVersions())
				conn3 := types.NewConnectionEnd(types.UNINITIALIZED, testConnectionID3, testClientID3, counterparty2, types.GetCompatibleVersions())

				expConnections = []*types.ConnectionEnd{&conn1, &conn2, &conn3}

				for i := range expConnections {
					suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainA.GetContext(), expConnections[i].ID, *expConnections[i])
				}

				req = &types.QueryConnectionsRequest{
					Req: &query.PageRequest{
						Limit:      3,
						CountTotal: true,
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

			res, err := suite.chainA.QueryServer.Connections(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(expConnections, res.Connections)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryClientConnections() {
	var (
		req      *types.QueryClientConnectionsRequest
		expPaths []string
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = nil
			},
			false,
		},
		{"invalid connectionID",
			func() {
				req = &types.QueryClientConnectionsRequest{}
			},
			false,
		},
		{"connection not found",
			func() {
				req = &types.QueryClientConnectionsRequest{
					ClientID: testClientIDA,
				}
			},
			false,
		},
		{
			"success",
			func() {
				expPaths = []string{testConnectionIDA, testConnectionIDB}
				suite.chainA.App.IBCKeeper.ConnectionKeeper.SetClientConnectionPaths(suite.chainA.GetContext(), testClientIDA, expPaths)

				req = &types.QueryClientConnectionsRequest{
					ClientID: testClientIDA,
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

			res, err := suite.chainA.QueryServer.ClientConnections(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(expPaths, res.ConnectionPaths)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
