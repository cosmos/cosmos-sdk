package keeper_test

import (
	"fmt"
	"time"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// TestConnOpenInit - chainA initializes (INIT state) a connection with
// chainB which is yet UNINITIALIZED
func (suite *KeeperTestSuite) TestConnOpenInit() {
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
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
		}, true},
		{"connection already exists", func() {
			clientA, clientB, _, _ = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
		}, false},
		{"couldn't add connection to client", func() {
			// swap client identifiers to result in client that does not exist
			clientB, clientA = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.msg, func() {
			suite.SetupTest() // reset

			tc.malleate()

			connA := suite.chainA.GetFirstTestConnection(clientA, clientB)
			connB := suite.chainB.GetFirstTestConnection(clientB, clientA)
			counterparty := types.NewCounterparty(clientB, connB.ID, suite.chainB.GetPrefix())

			err := suite.chainA.App.IBCKeeper.ConnectionKeeper.ConnOpenInit(suite.chainA.GetContext(), connA.ID, clientA, counterparty)

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
		clientA             string
		clientB             string
		consensusHeightDiff uint64
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"success", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
			_, _, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)
		}, true},
		{"consensus height > latest height", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
			_, _, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			consensusHeightDiff = 20
		}, false},
		{"self consensus state not found", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
			_, _, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// consensus height is retreived using latest client height, so incrementing by one is enough
			consensusHeightDiff = 1
		}, false},
		{"connection state verification failed", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
			// chainA connection not created
		}, false},
		{"consensus state verification failed", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
			// give chainA wrong consensus state for chainB
			consState, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetLatestClientConsensusState(suite.chainA.GetContext(), clientA)
			suite.Require().True(found)

			tmConsState, ok := consState.(ibctmtypes.ConsensusState)
			suite.Require().True(ok)

			tmConsState.Timestamp = time.Now()
			suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), clientA, tmConsState.Height, tmConsState)

			_, _, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)
		}, false},
		{"invalid previous connection is in TRYOPEN", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)

			// open init chainA
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// open try chainB
			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.msg, func() {
			suite.SetupTest()       // reset
			consensusHeightDiff = 0 // must be explicity changed in malleate

			tc.malleate()

			connA := suite.chainA.GetFirstTestConnection(clientA, clientB)
			connB := suite.chainB.GetFirstTestConnection(clientB, clientA)
			counterparty := types.NewCounterparty(clientA, connA.ID, suite.chainA.GetPrefix())

			connectionKey := host.KeyConnection(connA.ID)
			proofInit, proofHeight := suite.chainA.QueryProof(connectionKey)

			// retrieve consensus state to provide proof for
			consState, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetLatestClientConsensusState(suite.chainA.GetContext(), clientA)
			suite.Require().True(found)

			consensusHeight := consState.GetHeight()
			consensusKey := host.FullKeyClientPath(clientA, host.KeyConsensusState(consensusHeight))
			proofConsensus, _ := suite.chainA.QueryProof(consensusKey)

			err := suite.chainB.App.IBCKeeper.ConnectionKeeper.ConnOpenTry(
				suite.chainB.GetContext(), connB.ID, counterparty, clientB,
				types.GetCompatibleVersions(), proofInit, proofConsensus,
				proofHeight, consensusHeight+consensusHeightDiff,
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
	version := types.GetCompatibleVersions()[0]

	testCases := []struct {
		msg      string
		version  string
		malleate func() uint64
		expPass  bool
	}{
		{"success", version, func() uint64 {
			suite.oldchainA.CreateClient(suite.oldchainB)
			suite.oldchainB.CreateClient(suite.oldchainA)
			suite.oldchainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, types.TRYOPEN)
			suite.oldchainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.INIT)
			suite.oldchainB.updateClient(suite.oldchainA)
			suite.oldchainA.updateClient(suite.oldchainB)
			return suite.oldchainB.Header.GetHeight()
		}, true},
		{"success from tryopen", version, func() uint64 {
			suite.oldchainA.CreateClient(suite.oldchainB)
			suite.oldchainB.CreateClient(suite.oldchainA)
			suite.oldchainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, types.TRYOPEN)
			suite.oldchainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.TRYOPEN)
			suite.oldchainB.updateClient(suite.oldchainA)
			suite.oldchainA.updateClient(suite.oldchainB)
			return suite.oldchainB.Header.GetHeight()
		}, true},
		{"consensus height > latest height", version, func() uint64 {
			return 10
		}, false},
		{"connection not found", version, func() uint64 {
			return 2
		}, false},
		{"connection state is not INIT", version, func() uint64 {
			suite.oldchainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, types.UNINITIALIZED)
			suite.oldchainB.updateClient(suite.oldchainA)
			return suite.oldchainB.Header.GetHeight()
		}, false},
		{"incompatible IBC versions", "2.0", func() uint64 {
			suite.oldchainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, types.INIT)
			suite.oldchainB.updateClient(suite.oldchainA)
			return suite.oldchainB.Header.GetHeight()
		}, false},
		{"self consensus state not found", version, func() uint64 {
			suite.oldchainB.CreateClient(suite.oldchainA)
			suite.oldchainA.CreateClient(suite.oldchainB)
			suite.oldchainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.INIT)
			suite.oldchainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, types.TRYOPEN)
			suite.oldchainB.updateClient(suite.oldchainA)
			return suite.oldchainB.Header.GetHeight()
		}, false},
		{"connection state verification failed", version, func() uint64 {
			suite.oldchainB.CreateClient(suite.oldchainA)
			suite.oldchainA.CreateClient(suite.oldchainB)
			suite.oldchainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.INIT)
			suite.oldchainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, types.UNINITIALIZED)
			suite.oldchainB.updateClient(suite.oldchainA)
			return suite.oldchainB.Header.GetHeight()
		}, false},
		{"consensus state verification failed", version, func() uint64 {
			suite.oldchainB.CreateClient(suite.oldchainA)
			suite.oldchainA.CreateClient(suite.oldchainB)
			suite.oldchainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.INIT)
			suite.oldchainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, types.UNINITIALIZED)
			suite.oldchainB.updateClient(suite.oldchainA)
			return suite.oldchainB.Header.GetHeight()
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			consensusHeight := tc.malleate()

			connectionKey := host.KeyConnection(testConnectionIDB)
			proofTry, proofHeight := queryProof(suite.oldchainB, connectionKey)

			consensusKey := prefixedClientKey(testClientIDA, host.KeyConsensusState(consensusHeight))
			proofConsensus, _ := queryProof(suite.oldchainB, consensusKey)

			err := suite.oldchainA.App.IBCKeeper.ConnectionKeeper.ConnOpenAck(
				suite.oldchainA.GetContext(), testConnectionIDA, tc.version, proofTry, proofConsensus,
				proofHeight+1, consensusHeight,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed with consensus height %d: %s", i, consensusHeight, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed with consensus height %d: %s", i, consensusHeight, tc.msg)
			}
		})
	}
}

// TestConnOpenConfirm - Chain B (ID #2) calls ConnOpenConfirm to confirm that
// Chain A (ID #1) state is now OPEN.
func (suite *KeeperTestSuite) TestConnOpenConfirm() {
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"success", func() {
			suite.oldchainB.CreateClient(suite.oldchainA)
			suite.oldchainA.CreateClient(suite.oldchainB)
			suite.oldchainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.OPEN)
			suite.oldchainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, types.TRYOPEN)
			suite.oldchainB.updateClient(suite.oldchainA)
		}, true},
		{"connection not found", func() {}, false},
		{"chain B's connection state is not TRYOPEN", func() {
			suite.oldchainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, types.UNINITIALIZED)
			suite.oldchainA.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, types.OPEN)
			suite.oldchainA.updateClient(suite.oldchainB)
		}, false},
		{"connection state verification failed", func() {
			suite.oldchainB.CreateClient(suite.oldchainA)
			suite.oldchainA.CreateClient(suite.oldchainB)
			suite.oldchainB.updateClient(suite.oldchainA)
			suite.oldchainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, types.INIT)
			suite.oldchainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, types.TRYOPEN)
			suite.oldchainA.updateClient(suite.oldchainA)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			connectionKey := host.KeyConnection(testConnectionIDA)
			proofAck, proofHeight := queryProof(suite.oldchainA, connectionKey)

			if tc.expPass {
				err := suite.oldchainB.App.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(
					suite.oldchainB.GetContext(), testConnectionIDB, proofAck, proofHeight+1,
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.oldchainB.App.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(
					suite.oldchainB.GetContext(), testConnectionIDB, proofAck, proofHeight+1,
				)
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}
