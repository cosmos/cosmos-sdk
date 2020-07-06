package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

const (
	storeKey = host.StoreKey

	trustingPeriod time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod      time.Duration = time.Hour * 24 * 7 * 3
	maxClockDrift  time.Duration = time.Second * 10

	nextTimestamp = 10 // increment used for the next header's timestamp

	testPort1 = "firstport"
	testPort2 = "secondport"

	testChannel1 = "firstchannel"
	testChannel2 = "secondchannel"
)

var (
	timestamp = time.Now() // starting timestamp for the client test chain
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

// create 2 connections: A0 - B, A1 - B
func (suite KeeperTestSuite) TestGetAllConnections() {
	clientA, clientB := suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
	connA0 := suite.chainA.GetFirstTestConnection(clientA, clientB)
	connB := suite.chainB.GetFirstTestConnection(clientB, clientA)

	connA1ID := "testconnection"

	counterpartyB := types.NewCounterparty(clientB, connB.ID, suite.chainB.GetPrefix()) // connection B

	conn1 := types.NewConnectionEnd(types.INIT, connA0.ID, clientA, counterpartyB, types.GetCompatibleVersions()) // A0 - B
	conn2 := types.NewConnectionEnd(types.INIT, connA1ID, clientA, counterpartyB, types.GetCompatibleVersions())  // A1 - B

	expConnections := []types.ConnectionEnd{conn1, conn2}

	for i := range expConnections {
		suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainA.GetContext(), expConnections[i].ID, expConnections[i])
	}

	connections := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetAllConnections(suite.chainA.GetContext())
	suite.Require().Len(connections, len(expConnections))
	suite.Require().Equal(expConnections, connections)
}

/*
func (suite KeeperTestSuite) TestGetAllClientConnectionPaths() {
	clients := []clientexported.ClientState{
		ibctmtypes.NewClientState(testClientIDA, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, ibctmtypes.Header{}, commitmenttypes.GetSDKSpecs()),
		ibctmtypes.NewClientState(testClientIDB, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, ibctmtypes.Header{}, commitmenttypes.GetSDKSpecs()),
		ibctmtypes.NewClientState(testClientID3, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, ibctmtypes.Header{}, commitmenttypes.GetSDKSpecs()),
	}

	for i := range clients {
		suite.oldchainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.oldchainA.GetContext(), clients[i])
	}

	expPaths := []types.ConnectionPaths{
		types.NewConnectionPaths(testClientIDA, []string{host.ConnectionPath(testConnectionIDA)}),
		types.NewConnectionPaths(testClientIDB, []string{host.ConnectionPath(testConnectionIDB), host.ConnectionPath(testConnectionID3)}),
	}

	for i := range expPaths {
		suite.oldchainA.App.IBCKeeper.ConnectionKeeper.SetClientConnectionPaths(suite.oldchainA.GetContext(), expPaths[i].ClientID, expPaths[i].Paths)
	}

	connPaths := suite.oldchainA.App.IBCKeeper.ConnectionKeeper.GetAllClientConnectionPaths(suite.oldchainA.GetContext())
	suite.Require().Len(connPaths, 2)
	suite.Require().Equal(connPaths, expPaths)
}

// TestGetTimestampAtHeight verifies if the clients on each chain return the correct timestamp
// for the other chain.
func (suite *KeeperTestSuite) TestGetTimestampAtHeight() {
	cases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"verification success", func() {
			suite.oldchainA.CreateClient(suite.oldchainB)
		}, true},
		{"client state not found", func() {}, false},
	}

	for i, tc := range cases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			// create and store a connection to chainB on chainA
			connection := suite.oldchainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.OPEN)

			actualTimestamp, err := suite.oldchainA.App.IBCKeeper.ConnectionKeeper.GetTimestampAtHeight(
				suite.oldchainA.GetContext(), connection, uint64(suite.oldchainB.Header.Height),
			)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().EqualValues(uint64(suite.oldchainB.Header.Time.UnixNano()), actualTimestamp)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
*/
