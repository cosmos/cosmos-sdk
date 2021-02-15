package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
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
					ConnectionId: ibctesting.InvalidID,
				}
			},
			false,
		},
		{
			"success",
			func() {
				clientA, clientB := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
				connA := suite.chainA.GetFirstTestConnection(clientA, clientB)
				connB := suite.chainB.GetFirstTestConnection(clientB, clientA)

				counterparty := types.NewCounterparty(clientB, connB.ID, suite.chainB.GetPrefix())
				expConnection = types.NewConnectionEnd(types.INIT, clientA, counterparty, types.ExportedVersionsToProto(types.GetCompatibleVersions()), 500)
				suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainA.GetContext(), connA.ID, expConnection)

				req = &types.QueryConnectionRequest{
					ConnectionId: connA.ID,
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
		expConnections = []*types.IdentifiedConnection{}
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
				clientA, clientB, connA0, connB0 := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
				clientA1, clientB1, connA1, connB1 := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
				connA2, _, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
				suite.Require().NoError(err)

				counterparty1 := types.NewCounterparty(clientB, connB0.ID, suite.chainB.GetPrefix())
				counterparty2 := types.NewCounterparty(clientB1, connB1.ID, suite.chainB.GetPrefix())
				// counterparty connection id is blank after open init
				counterparty3 := types.NewCounterparty(clientB, "", suite.chainB.GetPrefix())

				conn1 := types.NewConnectionEnd(types.OPEN, clientA, counterparty1, types.ExportedVersionsToProto(types.GetCompatibleVersions()), 0)
				conn2 := types.NewConnectionEnd(types.OPEN, clientA1, counterparty2, types.ExportedVersionsToProto(types.GetCompatibleVersions()), 0)
				conn3 := types.NewConnectionEnd(types.INIT, clientA, counterparty3, types.ExportedVersionsToProto(types.GetCompatibleVersions()), 0)

				iconn1 := types.NewIdentifiedConnection(connA0.ID, conn1)
				iconn2 := types.NewIdentifiedConnection(connA1.ID, conn2)
				iconn3 := types.NewIdentifiedConnection(connA2.ID, conn3)

				expConnections = []*types.IdentifiedConnection{&iconn1, &iconn2, &iconn3}

				req = &types.QueryConnectionsRequest{
					Pagination: &query.PageRequest{
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
					ClientId: ibctesting.InvalidID,
				}
			},
			false,
		},
		{
			"success",
			func() {
				clientA, clientB, connA0, _ := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
				connA1, _ := suite.coordinator.CreateConnection(suite.chainA, suite.chainB, clientA, clientB)
				expPaths = []string{connA0.ID, connA1.ID}
				suite.chainA.App.IBCKeeper.ConnectionKeeper.SetClientConnectionPaths(suite.chainA.GetContext(), clientA, expPaths)

				req = &types.QueryClientConnectionsRequest{
					ClientId: clientA,
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

func (suite *KeeperTestSuite) TestQueryConnectionClientState() {
	var (
		req                      *types.QueryConnectionClientStateRequest
		expIdentifiedClientState clienttypes.IdentifiedClientState
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
			"invalid connection ID",
			func() {
				req = &types.QueryConnectionClientStateRequest{
					ConnectionId: "",
				}
			},
			false,
		},
		{
			"connection not found",
			func() {
				req = &types.QueryConnectionClientStateRequest{
					ConnectionId: "test-connection-id",
				}
			},
			false,
		},
		{
			"client state not found",
			func() {
				_, _, connA, _, _, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, channeltypes.UNORDERED)

				// set connection to empty so clientID is empty
				suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainA.GetContext(), connA.ID, types.ConnectionEnd{})

				req = &types.QueryConnectionClientStateRequest{
					ConnectionId: connA.ID,
				}
			}, false,
		},
		{
			"success",
			func() {
				clientA, _, connA, _ := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)

				expClientState := suite.chainA.GetClientState(clientA)
				expIdentifiedClientState = clienttypes.NewIdentifiedClientState(clientA, expClientState)

				req = &types.QueryConnectionClientStateRequest{
					ConnectionId: connA.ID,
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

			res, err := suite.chainA.QueryServer.ConnectionClientState(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(&expIdentifiedClientState, res.IdentifiedClientState)

				// ensure UnpackInterfaces is defined
				cachedValue := res.IdentifiedClientState.ClientState.GetCachedValue()
				suite.Require().NotNil(cachedValue)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryConnectionConsensusState() {
	var (
		req               *types.QueryConnectionConsensusStateRequest
		expConsensusState exported.ConsensusState
		expClientID       string
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
			"invalid connection ID",
			func() {
				req = &types.QueryConnectionConsensusStateRequest{
					ConnectionId:   "",
					RevisionNumber: 0,
					RevisionHeight: 1,
				}
			},
			false,
		},
		{
			"connection not found",
			func() {
				req = &types.QueryConnectionConsensusStateRequest{
					ConnectionId:   "test-connection-id",
					RevisionNumber: 0,
					RevisionHeight: 1,
				}
			},
			false,
		},
		{
			"consensus state not found",
			func() {
				_, _, connA, _, _, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, channeltypes.UNORDERED)

				req = &types.QueryConnectionConsensusStateRequest{
					ConnectionId:   connA.ID,
					RevisionNumber: 0,
					RevisionHeight: uint64(suite.chainA.GetContext().BlockHeight()), // use current height
				}
			}, false,
		},
		{
			"success",
			func() {
				clientA, _, connA, _ := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)

				clientState := suite.chainA.GetClientState(clientA)
				expConsensusState, _ = suite.chainA.GetConsensusState(clientA, clientState.GetLatestHeight())
				suite.Require().NotNil(expConsensusState)
				expClientID = clientA

				req = &types.QueryConnectionConsensusStateRequest{
					ConnectionId:   connA.ID,
					RevisionNumber: clientState.GetLatestHeight().GetRevisionNumber(),
					RevisionHeight: clientState.GetLatestHeight().GetRevisionHeight(),
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

			res, err := suite.chainA.QueryServer.ConnectionConsensusState(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				consensusState, err := clienttypes.UnpackConsensusState(res.ConsensusState)
				suite.Require().NoError(err)
				suite.Require().Equal(expConsensusState, consensusState)
				suite.Require().Equal(expClientID, res.ClientId)

				// ensure UnpackInterfaces is defined
				cachedValue := res.ConsensusState.GetCachedValue()
				suite.Require().NotNil(cachedValue)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
