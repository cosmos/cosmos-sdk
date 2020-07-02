package keeper_test

import (
	"fmt"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// TestConnOpenInit - Chain A (ID #1) initializes (INIT state) a connection with
// Chain B (ID #2) which is yet UNINITIALIZED
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

// TestConnOpenTry - Chain B (ID #2) calls ConnOpenTry to verify the state of
// connection on Chain A (ID #1) is INIT
func (suite *KeeperTestSuite) TestConnOpenTry() {
	// counterparty for A on B
	counterparty := connectiontypes.NewCounterparty(
		testClientIDB, testConnectionIDA, commitmenttypes.NewMerklePrefix(suite.oldchainB.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes()),
	)

	testCases := []struct {
		msg      string
		malleate func() uint64
		expPass  bool
	}{
		{"success", func() uint64 {
			suite.oldchainB.CreateClient(suite.oldchainA)
			suite.oldchainA.CreateClient(suite.oldchainB)
			suite.oldchainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.INIT)
			suite.oldchainB.updateClient(suite.oldchainA)
			suite.oldchainA.updateClient(suite.oldchainB)
			suite.oldchainB.updateClient(suite.oldchainA)
			suite.oldchainA.updateClient(suite.oldchainB)
			return suite.oldchainB.Header.GetHeight() - 1
		}, true},
		{"consensus height > latest height", func() uint64 {
			return 0
		}, false},
		{"self consensus state not found", func() uint64 {
			//suite.ctx = suite.ctx.WithBlockHeight(100)
			return 100
		}, false},
		{"connection state verification invalid", func() uint64 {
			suite.oldchainB.CreateClient(suite.oldchainA)
			suite.oldchainA.CreateClient(suite.oldchainB)
			suite.oldchainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.UNINITIALIZED)
			suite.oldchainB.updateClient(suite.oldchainA)
			return 0
		}, false},
		{"consensus state verification invalid", func() uint64 {
			suite.oldchainB.CreateClient(suite.oldchainA)
			suite.oldchainA.CreateClient(suite.oldchainB)
			suite.oldchainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.INIT)
			suite.oldchainB.updateClient(suite.oldchainA)
			suite.oldchainA.updateClient(suite.oldchainB)
			return suite.oldchainB.Header.GetHeight()
		}, false},
		{"invalid previous connection", func() uint64 {
			suite.oldchainB.CreateClient(suite.oldchainA)
			suite.oldchainA.CreateClient(suite.oldchainB)
			suite.oldchainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, types.UNINITIALIZED)
			suite.oldchainB.updateClient(suite.oldchainA)
			suite.oldchainA.updateClient(suite.oldchainB)
			return 0
		}, false},
		{"couldn't add connection to client", func() uint64 {
			suite.oldchainB.CreateClient(suite.oldchainA)
			suite.oldchainA.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, types.UNINITIALIZED)
			suite.oldchainB.updateClient(suite.oldchainA)
			suite.oldchainA.updateClient(suite.oldchainB)
			return 0
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			consensusHeight := tc.malleate()

			connectionKey := host.KeyConnection(testConnectionIDA)
			proofInit, proofHeight := queryProof(suite.oldchainA, connectionKey)

			consensusKey := prefixedClientKey(testClientIDB, host.KeyConsensusState(consensusHeight))
			proofConsensus, _ := queryProof(suite.oldchainA, consensusKey)

			err := suite.oldchainB.App.IBCKeeper.ConnectionKeeper.ConnOpenTry(
				suite.oldchainB.GetContext(), testConnectionIDB, counterparty, testClientIDA,
				connectiontypes.GetCompatibleVersions(), proofInit, proofConsensus,
				proofHeight+1, consensusHeight,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed with consensus height %d and proof height %d: %s", i, consensusHeight, proofHeight, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed with consensus height %d and proof height %d: %s", i, consensusHeight, proofHeight, tc.msg)
			}
		})
	}
}

// TestConnOpenAck - Chain A (ID #1) calls TestConnOpenAck to acknowledge (ACK state)
// the initialization (TRYINIT) of the connection on  Chain B (ID #2).
func (suite *KeeperTestSuite) TestConnOpenAck() {
	version := connectiontypes.GetCompatibleVersions()[0]

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
	testCases := []testCase{
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

type testCase = struct {
	msg      string
	malleate func()
	expPass  bool
}
