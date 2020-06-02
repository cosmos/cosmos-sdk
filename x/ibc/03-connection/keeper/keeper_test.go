package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	lite "github.com/tendermint/tendermint/lite2"

	"github.com/cosmos/cosmos-sdk/codec"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

const (
	storeKey = host.StoreKey

	testClientIDA     = "testclientida" // chainid for chainA also chainB's clientID for A's liteclient
	testConnectionIDA = "connectionidatob"

	testClientIDB     = "testclientidb" // chainid for chainB also chainA's clientID for B's liteclient
	testConnectionIDB = "connectionidbtoa"

	testClientID3     = "testclientidthree"
	testConnectionID3 = "connectionidthree"
)

type KeeperTestSuite struct {
	suite.Suite

	cdc *codec.Codec

	// ChainA testing fields
	chainA *ibctesting.TestChain

	// ChainB testing fields
	chainB *ibctesting.TestChain
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.chainA = ibctesting.NewTestChain(testClientIDA)
	suite.chainB = ibctesting.NewTestChain(testClientIDB)

	suite.cdc = suite.chainA.App.Codec()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestSetAndGetConnection() {
	_, existed := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetConnection(suite.chainA.GetContext(), testConnectionIDA)
	suite.Require().False(existed)

	counterparty := types.NewCounterparty(testClientIDA, testConnectionIDA, commitmenttypes.NewMerklePrefix(suite.chainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes()))
	expConn := types.NewConnectionEnd(types.INIT, testConnectionIDB, testClientIDB, counterparty, types.GetCompatibleVersions())
	suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainA.GetContext(), testConnectionIDA, expConn)
	conn, existed := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetConnection(suite.chainA.GetContext(), testConnectionIDA)
	suite.Require().True(existed)
	suite.Require().EqualValues(expConn, conn)
}

func (suite *KeeperTestSuite) TestSetAndGetClientConnectionPaths() {
	_, existed := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetClientConnectionPaths(suite.chainA.GetContext(), testClientIDA)
	suite.False(existed)

	suite.chainA.App.IBCKeeper.ConnectionKeeper.SetClientConnectionPaths(suite.chainA.GetContext(), testClientIDB, types.GetCompatibleVersions())
	paths, existed := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetClientConnectionPaths(suite.chainA.GetContext(), testClientIDB)
	suite.True(existed)
	suite.EqualValues(types.GetCompatibleVersions(), paths)
}

func (suite KeeperTestSuite) TestGetAllConnections() {
	// Connection (Counterparty): A(C) -> C(B) -> B(A)
	counterparty1 := types.NewCounterparty(testClientIDA, testConnectionIDA, commitmenttypes.NewMerklePrefix(suite.chainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes()))
	counterparty2 := types.NewCounterparty(testClientIDB, testConnectionIDB, commitmenttypes.NewMerklePrefix(suite.chainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes()))
	counterparty3 := types.NewCounterparty(testClientID3, testConnectionID3, commitmenttypes.NewMerklePrefix(suite.chainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes()))

	conn1 := types.NewConnectionEnd(types.INIT, testConnectionIDA, testClientIDA, counterparty3, types.GetCompatibleVersions())
	conn2 := types.NewConnectionEnd(types.INIT, testConnectionIDB, testClientIDB, counterparty1, types.GetCompatibleVersions())
	conn3 := types.NewConnectionEnd(types.UNINITIALIZED, testConnectionID3, testClientID3, counterparty2, types.GetCompatibleVersions())

	expConnections := []types.ConnectionEnd{conn1, conn2, conn3}

	for i := range expConnections {
		suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainA.GetContext(), expConnections[i].ID, expConnections[i])
	}

	connections := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetAllConnections(suite.chainA.GetContext())
	suite.Require().Len(connections, len(expConnections))
	suite.Require().Equal(expConnections, connections)
}

func (suite KeeperTestSuite) TestGetAllClientConnectionPaths() {
	clients := []clientexported.ClientState{
		ibctmtypes.NewClientState(testClientIDA, lite.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, ibctmtypes.Header{}),
		ibctmtypes.NewClientState(testClientIDB, lite.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, ibctmtypes.Header{}),
		ibctmtypes.NewClientState(testClientID3, lite.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, ibctmtypes.Header{}),
	}

	for i := range clients {
		suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), clients[i])
	}

	expPaths := []types.ConnectionPaths{
		types.NewConnectionPaths(testClientIDA, []string{host.ConnectionPath(testConnectionIDA)}),
		types.NewConnectionPaths(testClientIDB, []string{host.ConnectionPath(testConnectionIDB), host.ConnectionPath(testConnectionID3)}),
	}

	for i := range expPaths {
		suite.chainA.App.IBCKeeper.ConnectionKeeper.SetClientConnectionPaths(suite.chainA.GetContext(), expPaths[i].ClientID, expPaths[i].Paths)
	}

	connPaths := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetAllClientConnectionPaths(suite.chainA.GetContext())
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
			suite.chainA.CreateClient(suite.chainB)
		}, true},
		{"client state not found", func() {}, false},
	}

	for i, tc := range cases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			// create and store a connection to chainB on chainA
			connection := suite.chainA.CreateConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.OPEN)

			actualTimestamp, err := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetTimestampAtHeight(
				suite.chainA.GetContext(), connection, uint64(suite.chainB.Header.Height),
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
				suite.Require().EqualValues(uint64(suite.chainB.Header.Time.UnixNano()), actualTimestamp)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func prefixedClientKey(clientID string, key []byte) []byte {
	return append([]byte("clients/"+clientID+"/"), key...)
}
