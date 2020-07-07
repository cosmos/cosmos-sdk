package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
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
	clientA, clientB := suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
	connA := suite.chainA.GetFirstTestConnection(clientA, clientB)
	_, existed := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetConnection(suite.chainA.GetContext(), connA.ID)
	suite.Require().False(existed)

	suite.coordinator.CreateConnection(suite.chainA, suite.chainB, clientA, clientB)
	_, existed = suite.chainA.App.IBCKeeper.ConnectionKeeper.GetConnection(suite.chainA.GetContext(), connA.ID)
	suite.Require().True(existed)
}

func (suite *KeeperTestSuite) TestSetAndGetClientConnectionPaths() {
	clientA, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
	_, existed := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetClientConnectionPaths(suite.chainA.GetContext(), clientA)
	suite.False(existed)

	suite.chainA.App.IBCKeeper.ConnectionKeeper.SetClientConnectionPaths(suite.chainA.GetContext(), clientA, types.GetCompatibleVersions())
	paths, existed := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetClientConnectionPaths(suite.chainA.GetContext(), clientA)
	suite.True(existed)
	suite.EqualValues(types.GetCompatibleVersions(), paths)
}

// create 2 connections: A0 - B0, A1 - B1
func (suite KeeperTestSuite) TestGetAllConnections() {
	clientA, clientB, connA0, connB0 := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
	connA1, connB1 := suite.coordinator.CreateConnection(suite.chainA, suite.chainB, clientA, clientB)

	counterpartyB0 := types.NewCounterparty(clientB, connB0.ID, suite.chainB.GetPrefix()) // connection B0
	counterpartyB1 := types.NewCounterparty(clientB, connB1.ID, suite.chainB.GetPrefix()) // connection B1

	conn1 := types.NewConnectionEnd(types.OPEN, connA0.ID, clientA, counterpartyB0, types.GetCompatibleVersions()) // A0 - B0
	conn2 := types.NewConnectionEnd(types.OPEN, connA1.ID, clientA, counterpartyB1, types.GetCompatibleVersions()) // A1 - B1
	expConnections := []types.ConnectionEnd{conn1, conn2}

	connections := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetAllConnections(suite.chainA.GetContext())
	suite.Require().Len(connections, len(expConnections))
	suite.Require().Equal(expConnections, connections)
}

// the test creates 2 clients clientA0 and clientA1. clientA0 has a single
// connection and clientA1 has 2 connections.
func (suite KeeperTestSuite) TestGetAllClientConnectionPaths() {
	clientA0, _, connA0, _ := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
	clientA1, clientB1, connA1, _ := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
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
			_, _, connA, _ := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
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
				suite.chainA.GetContext(), connection, uint64(suite.chainB.LastHeader.Height),
			)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().EqualValues(uint64(suite.chainB.LastHeader.Time.UnixNano()), actualTimestamp)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
