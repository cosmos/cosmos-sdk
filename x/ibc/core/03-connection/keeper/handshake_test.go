package keeper_test

import (
	"time"

	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

// TestConnOpenInit - chainA initializes (INIT state) a connection with
// chainB which is yet UNINITIALIZED
func (suite *KeeperTestSuite) TestConnOpenInit() {
	var (
		clientA      string
		clientB      string
		version      string
		emptyConnBID bool
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"success", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
		}, true},
		{"success with empty counterparty identifier", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			emptyConnBID = true
		}, true},
		{"success with non empty version", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			version = types.GetCompatibleEncodedVersions()[0]
		}, true},
		{"connection already exists", func() {
			clientA, clientB, _, _ = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, ibctesting.Tendermint)
		}, false},
		{"invalid version", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			version = "bad version"
		}, false},
		{"couldn't add connection to client", func() {
			// swap client identifiers to result in client that does not exist
			clientB, clientA = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.msg, func() {
			suite.SetupTest()    // reset
			emptyConnBID = false // must be explicitly changed
			version = ""         // must be explicitly changed

			tc.malleate()

			connA := suite.chainA.GetFirstTestConnection(clientA, clientB)
			connB := suite.chainB.GetFirstTestConnection(clientB, clientA)
			if emptyConnBID {
				connB.ID = ""
			}
			counterparty := types.NewCounterparty(clientB, connB.ID, suite.chainB.GetPrefix())

			err := suite.chainA.App.IBCKeeper.ConnectionKeeper.ConnOpenInit(suite.chainA.GetContext(), connA.ID, clientA, counterparty, version)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestConnOpenTry - chainB calls ConnOpenTry to verify the state of
// connection on chainA is INIT
func (suite *KeeperTestSuite) TestConnOpenTry() {
	var (
		clientA            string
		clientB            string
		versions           []string
		consensusHeight    exported.Height
		counterpartyClient exported.ClientState
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"success", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			_, _, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// retrieve client state of chainA to pass as counterpartyClient
			counterpartyClient = suite.chainA.GetClientState(clientA)
		}, true},
		{"success with empty counterpartyChosenConnectionID", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			connA, _, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// modify connA to set counterparty connection identifier to empty string
			connection, found := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetConnection(suite.chainA.GetContext(), connA.ID)
			suite.Require().True(found)

			connection.Counterparty.ConnectionId = ""

			suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainA.GetContext(), connA.ID, connection)

			err = suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, ibctesting.Tendermint)
			suite.Require().NoError(err)

			err = suite.coordinator.UpdateClient(suite.chainB, suite.chainA, clientB, ibctesting.Tendermint)
			suite.Require().NoError(err)

			// retrieve client state of chainA to pass as counterpartyClient
			counterpartyClient = suite.chainA.GetClientState(clientA)
		}, true},
		{"counterpartyChosenConnectionID does not match desiredConnectionID", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			connA, _, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// modify connA to set counterparty connection identifier to invalid identifier
			connection, found := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetConnection(suite.chainA.GetContext(), connA.ID)
			suite.Require().True(found)

			connection.Counterparty.ConnectionId = "badidentifier"

			suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainA.GetContext(), connA.ID, connection)

			err = suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, ibctesting.Tendermint)
			suite.Require().NoError(err)

			err = suite.coordinator.UpdateClient(suite.chainB, suite.chainA, clientB, ibctesting.Tendermint)
			suite.Require().NoError(err)

			// retrieve client state of chainA to pass as counterpartyClient
			counterpartyClient = suite.chainA.GetClientState(clientA)
		}, false},
		{"invalid counterparty client", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			_, _, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// retrieve client state of chainB to pass as counterpartyClient counterpartyClient = suite.chainA.GetClientState(clientA)

			// Set an invalid client of chainA on chainB
			tmClient, ok := counterpartyClient.(*ibctmtypes.ClientState)
			suite.Require().True(ok)
			tmClient.ChainId = "wrongchainid"

			suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), clientA, tmClient)
		}, false},
		{"consensus height >= latest height", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			_, _, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// retrieve client state of chainA to pass as counterpartyClient
			counterpartyClient = suite.chainA.GetClientState(clientA)

			consensusHeight = clienttypes.GetSelfHeight(suite.chainB.GetContext())
		}, false},
		{"self consensus state not found", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			_, _, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// retrieve client state of chainA to pass as counterpartyClient
			counterpartyClient = suite.chainA.GetClientState(clientA)

			consensusHeight = clienttypes.NewHeight(0, 1)
		}, false},
		{"counterparty versions is empty", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			_, _, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// retrieve client state of chainA to pass as counterpartyClient
			counterpartyClient = suite.chainA.GetClientState(clientA)

			versions = nil
		}, false},
		{"counterparty versions don't have a match", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			_, _, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// retrieve client state of chainA to pass as counterpartyClient
			counterpartyClient = suite.chainA.GetClientState(clientA)

			version, err := types.NewVersion("0.0", nil).Encode()
			suite.Require().NoError(err)
			versions = []string{version}
		}, false},
		{"connection state verification failed", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			// chainA connection not created

			// retrieve client state of chainA to pass as counterpartyClient
			counterpartyClient = suite.chainA.GetClientState(clientA)
		}, false},
		{"client state verification failed", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			_, _, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// retrieve client state of chainA to pass as counterpartyClient
			counterpartyClient = suite.chainA.GetClientState(clientA)

			// modify counterparty client without setting in store so it still passes validate but fails proof verification
			tmClient, ok := counterpartyClient.(*ibctmtypes.ClientState)
			suite.Require().True(ok)
			tmClient.LatestHeight = tmClient.LatestHeight.Increment()
		}, false},
		{"consensus state verification failed", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)

			// retrieve client state of chainA to pass as counterpartyClient
			counterpartyClient = suite.chainA.GetClientState(clientA)

			// give chainA wrong consensus state for chainB
			consState, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetLatestClientConsensusState(suite.chainA.GetContext(), clientA)
			suite.Require().True(found)

			tmConsState, ok := consState.(*ibctmtypes.ConsensusState)
			suite.Require().True(ok)

			tmConsState.Timestamp = time.Now()
			suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), clientA, counterpartyClient.GetLatestHeight(), tmConsState)

			_, _, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)
		}, false},
		{"invalid previous connection is in TRYOPEN", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)

			// open init chainA
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// open try chainB
			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)

			err = suite.coordinator.UpdateClient(suite.chainB, suite.chainA, clientB, ibctesting.Tendermint)
			suite.Require().NoError(err)

			// retrieve client state of chainA to pass as counterpartyClient
			counterpartyClient = suite.chainA.GetClientState(clientA)
		}, false},
		{"invalid previous connection has invalid versions", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)

			// open init chainA
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// open try chainB
			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)

			// modify connB to be in INIT with incorrect versions
			connection, found := suite.chainB.App.IBCKeeper.ConnectionKeeper.GetConnection(suite.chainB.GetContext(), connB.ID)
			suite.Require().True(found)

			connection.State = types.INIT
			connection.Versions = []string{"invalid version"}

			suite.chainB.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainB.GetContext(), connB.ID, connection)

			err = suite.coordinator.UpdateClient(suite.chainB, suite.chainA, clientB, ibctesting.Tendermint)
			suite.Require().NoError(err)

			// retrieve client state of chainA to pass as counterpartyClient
			counterpartyClient = suite.chainA.GetClientState(clientA)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.msg, func() {
			suite.SetupTest()                               // reset
			consensusHeight = clienttypes.ZeroHeight()      // must be explicitly changed in malleate
			versions = types.GetCompatibleEncodedVersions() // must be explicitly changed in malleate

			tc.malleate()

			connA := suite.chainA.GetFirstTestConnection(clientA, clientB)
			connB := suite.chainB.GetFirstTestConnection(clientB, clientA)
			counterparty := types.NewCounterparty(clientA, connA.ID, suite.chainA.GetPrefix())

			// get counterpartyChosenConnectionID
			var counterpartyChosenConnectionID string
			connection, found := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetConnection(suite.chainA.GetContext(), connA.ID)
			if found {
				counterpartyChosenConnectionID = connection.Counterparty.ConnectionId
			}

			connectionKey := host.KeyConnection(connA.ID)
			proofInit, proofHeight := suite.chainA.QueryProof(connectionKey)

			if consensusHeight.IsZero() {
				// retrieve consensus state height to provide proof for
				consensusHeight = counterpartyClient.GetLatestHeight()
			}
			consensusKey := host.FullKeyClientPath(clientA, host.KeyConsensusState(consensusHeight))
			proofConsensus, _ := suite.chainA.QueryProof(consensusKey)

			// retrieve proof of counterparty clientstate on chainA
			clientKey := host.FullKeyClientPath(clientA, host.KeyClientState())
			proofClient, _ := suite.chainA.QueryProof(clientKey)

			err := suite.chainB.App.IBCKeeper.ConnectionKeeper.ConnOpenTry(
				suite.chainB.GetContext(), connB.ID, counterpartyChosenConnectionID, counterparty, clientB, counterpartyClient,
				versions, proofInit, proofClient, proofConsensus,
				proofHeight, consensusHeight,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestConnOpenAck - Chain A (ID #1) calls TestConnOpenAck to acknowledge (ACK state)
// the initialization (TRYINIT) of the connection on  Chain B (ID #2).
func (suite *KeeperTestSuite) TestConnOpenAck() {
	var (
		clientA                  string
		clientB                  string
		counterpartyConnectionID string
		consensusHeight          exported.Height
		version                  string
		counterpartyClient       exported.ClientState
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"success", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)

			// retrieve client state of chainB to pass as counterpartyClient
			counterpartyClient = suite.chainB.GetClientState(clientB)
		}, true},
		{"success with empty stored counterparty connection ID", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)

			// modify connA to set counterparty connection identifier to empty string
			connection, found := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetConnection(suite.chainA.GetContext(), connA.ID)
			suite.Require().True(found)

			connection.Counterparty.ConnectionId = ""
			// use some other identifier
			counterpartyConnectionID = connB.ID

			suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainA.GetContext(), connA.ID, connection)

			err = suite.coordinator.UpdateClient(suite.chainB, suite.chainA, clientB, ibctesting.Tendermint)
			suite.Require().NoError(err)

			err = suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, ibctesting.Tendermint)
			suite.Require().NoError(err)

			// retrieve client state of chainB to pass as counterpartyClient
			counterpartyClient = suite.chainB.GetClientState(clientB)
		}, true},
		{"success from tryopen", func() {
			// chainA is in TRYOPEN, chainB is in TRYOPEN
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			connB, connA, err := suite.coordinator.ConnOpenInit(suite.chainB, suite.chainA, clientB, clientA)
			suite.Require().NoError(err)

			err = suite.coordinator.ConnOpenTry(suite.chainA, suite.chainB, connA, connB)
			suite.Require().NoError(err)

			// set chainB to TRYOPEN
			connection := suite.chainB.GetConnection(connB)
			connection.State = types.TRYOPEN
			suite.chainB.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainB.GetContext(), connB.ID, connection)
			// update clientB so state change is committed
			suite.coordinator.UpdateClient(suite.chainB, suite.chainA, clientB, ibctesting.Tendermint)

			suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, ibctesting.Tendermint)

			// retrieve client state of chainB to pass as counterpartyClient
			counterpartyClient = suite.chainB.GetClientState(clientB)
		}, true},
		{"success from tryopen with empty stored connection id", func() {
			// chainA is in TRYOPEN, chainB is in TRYOPEN
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			connB, connA, err := suite.coordinator.ConnOpenInit(suite.chainB, suite.chainA, clientB, clientA)
			suite.Require().NoError(err)

			err = suite.coordinator.ConnOpenTry(suite.chainA, suite.chainB, connA, connB)
			suite.Require().NoError(err)

			// set chainB to TRYOPEN
			connection := suite.chainB.GetConnection(connB)
			connection.State = types.TRYOPEN

			suite.chainB.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainB.GetContext(), connB.ID, connection)

			// set connA to use empty string
			connection = suite.chainA.GetConnection(connA)

			// set counterparty connection identifier to empty string
			connection.Counterparty.ConnectionId = ""

			suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainA.GetContext(), connA.ID, connection)

			// update clientB so state change is committed
			suite.coordinator.UpdateClient(suite.chainB, suite.chainA, clientB, ibctesting.Tendermint)

			suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, ibctesting.Tendermint)

			// retrieve client state of chainB to pass as counterpartyClient
			counterpartyClient = suite.chainB.GetClientState(clientB)
		}, true},
		{"invalid counterparty client", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// retrieve client state of chainB to pass as counterpartyClient
			counterpartyClient = suite.chainB.GetClientState(clientB)

			// Set an invalid client of chainA on chainB
			tmClient, ok := counterpartyClient.(*ibctmtypes.ClientState)
			suite.Require().True(ok)
			tmClient.ChainId = "wrongchainid"

			suite.chainB.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainB.GetContext(), clientB, tmClient)

			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)
		}, false},
		{"consensus height >= latest height", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// retrieve client state of chainB to pass as counterpartyClient
			counterpartyClient = suite.chainB.GetClientState(clientB)

			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)

			consensusHeight = clienttypes.GetSelfHeight(suite.chainA.GetContext())
		}, false},
		{"connection not found", func() {
			// connections are never created
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)

			// retrieve client state of chainB to pass as counterpartyClient
			counterpartyClient = suite.chainB.GetClientState(clientB)
		}, false},
		{"invalid counterparty connection ID", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// retrieve client state of chainB to pass as counterpartyClient
			counterpartyClient = suite.chainB.GetClientState(clientB)

			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)

			// modify connB to set counterparty connection identifier to wrong identifier
			connection, found := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetConnection(suite.chainA.GetContext(), connA.ID)
			suite.Require().True(found)

			connection.Counterparty.ConnectionId = "badconnectionid"

			suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainA.GetContext(), connA.ID, connection)

			err = suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, ibctesting.Tendermint)
			suite.Require().NoError(err)

			err = suite.coordinator.UpdateClient(suite.chainB, suite.chainA, clientB, ibctesting.Tendermint)
			suite.Require().NoError(err)
		}, false},
		{"connection state is not INIT", func() {
			// connection state is already OPEN on chainA
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// retrieve client state of chainB to pass as counterpartyClient
			counterpartyClient = suite.chainB.GetClientState(clientB)

			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)

			err = suite.coordinator.ConnOpenAck(suite.chainA, suite.chainB, connA, connB)
			suite.Require().NoError(err)
		}, false},
		{"connection is in INIT but the proposed version is invalid", func() {
			// chainA is in INIT, chainB is in TRYOPEN
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// retrieve client state of chainB to pass as counterpartyClient
			counterpartyClient = suite.chainB.GetClientState(clientB)

			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)

			version = "2.0"
		}, false},
		{"connection is in TRYOPEN but the set version in the connection is invalid", func() {
			// chainA is in TRYOPEN, chainB is in TRYOPEN
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			connB, connA, err := suite.coordinator.ConnOpenInit(suite.chainB, suite.chainA, clientB, clientA)
			suite.Require().NoError(err)

			err = suite.coordinator.ConnOpenTry(suite.chainA, suite.chainB, connA, connB)
			suite.Require().NoError(err)

			// set chainB to TRYOPEN
			connection := suite.chainB.GetConnection(connB)
			connection.State = types.TRYOPEN
			suite.chainB.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainB.GetContext(), connB.ID, connection)

			// update clientB so state change is committed
			suite.coordinator.UpdateClient(suite.chainB, suite.chainA, clientB, ibctesting.Tendermint)
			suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, ibctesting.Tendermint)

			// retrieve client state of chainB to pass as counterpartyClient
			counterpartyClient = suite.chainB.GetClientState(clientB)

			version = "2.0"
		}, false},
		{"incompatible IBC versions", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// retrieve client state of chainB to pass as counterpartyClient
			counterpartyClient = suite.chainB.GetClientState(clientB)

			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)

			// set version to a non-compatible version
			version = "(2.0,[])"
		}, false},
		{"empty version", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// retrieve client state of chainB to pass as counterpartyClient
			counterpartyClient = suite.chainB.GetClientState(clientB)

			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)

			version = ""
		}, false},
		{"feature set verification failed - unsupported feature", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// retrieve client state of chainB to pass as counterpartyClient
			counterpartyClient = suite.chainB.GetClientState(clientB)

			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)

			version, err = types.NewVersion(types.DefaultIBCVersionIdentifier, []string{"ORDER_ORDERED", "ORDER_UNORDERED", "ORDER_DAG"}).Encode()
			suite.Require().NoError(err)
		}, false},
		{"self consensus state not found", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// retrieve client state of chainB to pass as counterpartyClient
			counterpartyClient = suite.chainB.GetClientState(clientB)

			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)

			consensusHeight = clienttypes.NewHeight(0, 1)
		}, false},
		{"connection state verification failed", func() {
			// chainB connection is not in INIT
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			_, _, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// retrieve client state of chainB to pass as counterpartyClient
			counterpartyClient = suite.chainB.GetClientState(clientB)
		}, false},
		{"client state verification failed", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// retrieve client state of chainB to pass as counterpartyClient
			counterpartyClient = suite.chainB.GetClientState(clientB)

			// modify counterparty client without setting in store so it still passes validate but fails proof verification
			tmClient, ok := counterpartyClient.(*ibctmtypes.ClientState)
			suite.Require().True(ok)
			tmClient.LatestHeight = tmClient.LatestHeight.Increment()

			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)
		}, false},
		{"consensus state verification failed", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// retrieve client state of chainB to pass as counterpartyClient
			counterpartyClient = suite.chainB.GetClientState(clientB)

			// give chainB wrong consensus state for chainA
			consState, found := suite.chainB.App.IBCKeeper.ClientKeeper.GetLatestClientConsensusState(suite.chainB.GetContext(), clientB)
			suite.Require().True(found)

			tmConsState, ok := consState.(*ibctmtypes.ConsensusState)
			suite.Require().True(ok)

			tmConsState.Timestamp = time.Now()
			suite.chainB.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainB.GetContext(), clientB, counterpartyClient.GetLatestHeight(), tmConsState)

			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.msg, func() {
			suite.SetupTest()                                 // reset
			version = types.GetCompatibleEncodedVersions()[0] // must be explicitly changed in malleate
			consensusHeight = clienttypes.ZeroHeight()        // must be explicitly changed in malleate
			counterpartyConnectionID = ""                     // must be explicitly changed in malleate

			tc.malleate()

			connA := suite.chainA.GetFirstTestConnection(clientA, clientB)
			connB := suite.chainB.GetFirstTestConnection(clientB, clientA)

			if counterpartyConnectionID == "" {
				counterpartyConnectionID = connB.ID
			}

			connectionKey := host.KeyConnection(connB.ID)
			proofTry, proofHeight := suite.chainB.QueryProof(connectionKey)

			if consensusHeight.IsZero() {
				// retrieve consensus state height to provide proof for
				clientState := suite.chainB.GetClientState(clientB)
				consensusHeight = clientState.GetLatestHeight()
			}
			consensusKey := host.FullKeyClientPath(clientB, host.KeyConsensusState(consensusHeight))
			proofConsensus, _ := suite.chainB.QueryProof(consensusKey)

			// retrieve proof of counterparty clientstate on chainA
			clientKey := host.FullKeyClientPath(clientB, host.KeyClientState())
			proofClient, _ := suite.chainB.QueryProof(clientKey)

			err := suite.chainA.App.IBCKeeper.ConnectionKeeper.ConnOpenAck(
				suite.chainA.GetContext(), connA.ID, counterpartyClient, version, counterpartyConnectionID,
				proofTry, proofClient, proofConsensus, proofHeight, consensusHeight,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestConnOpenConfirm - chainB calls ConnOpenConfirm to confirm that
// chainA state is now OPEN.
func (suite *KeeperTestSuite) TestConnOpenConfirm() {
	var (
		clientA string
		clientB string
	)
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"success", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)

			err = suite.coordinator.ConnOpenAck(suite.chainA, suite.chainB, connA, connB)
			suite.Require().NoError(err)
		}, true},
		{"connection not found", func() {
			// connections are never created
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
		}, false},
		{"chain B's connection state is not TRYOPEN", func() {
			// connections are OPEN
			clientA, clientB, _, _ = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, ibctesting.Tendermint)
		}, false},
		{"connection state verification failed", func() {
			// chainA is in INIT
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.msg, func() {
			suite.SetupTest() // reset

			tc.malleate()

			connA := suite.chainA.GetFirstTestConnection(clientA, clientB)
			connB := suite.chainB.GetFirstTestConnection(clientB, clientA)

			connectionKey := host.KeyConnection(connA.ID)
			proofAck, proofHeight := suite.chainA.QueryProof(connectionKey)

			err := suite.chainB.App.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(
				suite.chainB.GetContext(), connB.ID, proofAck, proofHeight,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
