package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
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
					ConnectionID: ibctesting.InvalidID,
				}
			},
			false,
		},
		{
			"success",
			func() {
				clientA, clientB := suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
				connA := suite.chainA.GetFirstTestConnection(clientA, clientB)
				connB := suite.chainB.GetFirstTestConnection(clientB, clientA)

				counterparty := types.NewCounterparty(clientB, connB.ID, suite.chainB.GetPrefix())
				expConnection = types.NewConnectionEnd(types.INIT, connA.ID, clientA, counterparty, types.GetCompatibleVersions())
				suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainA.GetContext(), connA.ID, expConnection)

				req = &types.QueryConnectionRequest{
					ConnectionID: connA.ID,
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
				clientA, clientB, connA0, connB0 := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
				connA1, connB1, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
				suite.Require().NoError(err)

				clientA1, clientB1, connA2, connB2 := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)

				counterparty1 := types.NewCounterparty(clientB, connB0.ID, suite.chainB.GetPrefix())
				counterparty2 := types.NewCounterparty(clientB, connB1.ID, suite.chainB.GetPrefix())
				counterparty3 := types.NewCounterparty(clientB1, connB2.ID, suite.chainB.GetPrefix())

				conn1 := types.NewConnectionEnd(types.OPEN, connA0.ID, clientA, counterparty1, types.GetCompatibleVersions())
				conn2 := types.NewConnectionEnd(types.INIT, connA1.ID, clientA, counterparty2, types.GetCompatibleVersions())
				conn3 := types.NewConnectionEnd(types.OPEN, connA2.ID, clientA1, counterparty3, types.GetCompatibleVersions())

				expConnections = []*types.ConnectionEnd{&conn1, &conn2, &conn3}

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
					ClientID: ibctesting.InvalidID,
				}
			},
			false,
		},
		{
			"success",
			func() {
				clientA, clientB, connA0, _ := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
				connA1, _ := suite.coordinator.CreateConnection(suite.chainA, suite.chainB, clientA, clientB)
				expPaths = []string{connA0.ID, connA1.ID}
				suite.chainA.App.IBCKeeper.ConnectionKeeper.SetClientConnectionPaths(suite.chainA.GetContext(), clientA, expPaths)

				req = &types.QueryClientConnectionsRequest{
					ClientID: clientA,
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
