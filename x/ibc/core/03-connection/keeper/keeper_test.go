package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

type KeeperTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestSetAndGetConnection() {
	clientA, clientB := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
	connA := suite.chainA.GetFirstTestConnection(clientA, clientB)
	_, existed := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetConnection(suite.chainA.GetContext(), connA.ID)
	suite.Require().False(existed)

	suite.coordinator.CreateConnection(suite.chainA, suite.chainB, clientA, clientB)
	_, existed = suite.chainA.App.IBCKeeper.ConnectionKeeper.GetConnection(suite.chainA.GetContext(), connA.ID)
	suite.Require().True(existed)
}

func (suite *KeeperTestSuite) TestSetAndGetClientConnectionPaths() {
	clientA, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
	_, existed := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetClientConnectionPaths(suite.chainA.GetContext(), clientA)
	suite.False(existed)

	connections := []string{"connectionA", "connectionB"}
	suite.chainA.App.IBCKeeper.ConnectionKeeper.SetClientConnectionPaths(suite.chainA.GetContext(), clientA, connections)
	paths, existed := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetClientConnectionPaths(suite.chainA.GetContext(), clientA)
	suite.True(existed)
	suite.EqualValues(connections, paths)
}

// create 2 connections: A0 - B0, A1 - B1
func (suite KeeperTestSuite) TestGetAllConnections() {
	clientA, clientB, connA0, connB0 := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
	connA1, connB1 := suite.coordinator.CreateConnection(suite.chainA, suite.chainB, clientA, clientB)

	counterpartyB0 := types.NewCounterparty(clientB, connB0.ID, suite.chainB.GetPrefix()) // connection B0
	counterpartyB1 := types.NewCounterparty(clientB, connB1.ID, suite.chainB.GetPrefix()) // connection B1

	conn1 := types.NewConnectionEnd(types.OPEN, clientA, counterpartyB0, types.ExportedVersionsToProto(types.GetCompatibleVersions()), 0) // A0 - B0
	conn2 := types.NewConnectionEnd(types.OPEN, clientA, counterpartyB1, types.ExportedVersionsToProto(types.GetCompatibleVersions()), 0) // A1 - B1

	iconn1 := types.NewIdentifiedConnection(connA0.ID, conn1)
	iconn2 := types.NewIdentifiedConnection(connA1.ID, conn2)

	expConnections := []types.IdentifiedConnection{iconn1, iconn2}

	connections := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetAllConnections(suite.chainA.GetContext())
	suite.Require().Len(connections, len(expConnections))
	suite.Require().Equal(expConnections, connections)
}

// the test creates 2 clients clientA0 and clientA1. clientA0 has a single
// connection and clientA1 has 2 connections.
func (suite KeeperTestSuite) TestGetAllClientConnectionPaths() {
	clientA0, _, connA0, _ := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
	clientA1, clientB1, connA1, _ := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
	connA2, _ := suite.coordinator.CreateConnection(suite.chainA, suite.chainB, clientA1, clientB1)

	expPaths := []types.ConnectionPaths{
		types.NewConnectionPaths(clientA0, []string{connA0.ID}),
		types.NewConnectionPaths(clientA1, []string{connA1.ID, connA2.ID}),
	}

	connPaths := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetAllClientConnectionPaths(suite.chainA.GetContext())
	suite.Require().Len(connPaths, 2)
	suite.Require().Equal(expPaths, connPaths)
}

// TestGetTimestampAtHeight verifies if the clients on each chain return the
// correct timestamp for the other chain.
func (suite *KeeperTestSuite) TestGetTimestampAtHeight() {
	var connection types.ConnectionEnd

	cases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"verification success", func() {
			_, _, connA, _ := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			connection = suite.chainA.GetConnection(connA)
		}, true},
		{"consensus state not found", func() {
			// any non-nil value of connection is valid
			suite.Require().NotNil(connection)
		}, false},
	}

	for _, tc := range cases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			actualTimestamp, err := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetTimestampAtHeight(
				suite.chainA.GetContext(), connection, suite.chainB.LastHeader.GetHeight(),
			)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().EqualValues(uint64(suite.chainB.LastHeader.GetTime().UnixNano()), actualTimestamp)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
