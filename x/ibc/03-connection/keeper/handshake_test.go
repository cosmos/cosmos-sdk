package keeper_test

import (
	"fmt"

	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// TestConnOpenInit - Chain A (ID #1) initializes (INIT state) a connection with
// Chain B (ID #2) which is yet UNINITIALIZED
func (suite *KeeperTestSuite) TestConnOpenInit() {
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"success", func() {
			suite.chainA.CreateClient(suite.chainB)
		}, true},
		{"connection already exists", func() {
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, types.INIT)
		}, false},
		{"couldn't add connection to client", func() {}, false},
	}

	prefix := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(commitmentexported.Merkle)
	suite.Require().NotNil(prefix)
	counterparty, err := connection.NewCounterparty(testClientIDB, testConnectionIDB, prefix)
	suite.Require().NoError(err)

	for i, tc := range testCases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			err := suite.chainA.App.IBCKeeper.ConnectionKeeper.ConnOpenInit(suite.chainA.GetContext(), testConnectionIDA, testClientIDB, counterparty)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

// TestConnOpenTry - Chain B (ID #2) calls ConnOpenTry to verify the state of
// connection on Chain A (ID #1) is INIT
func (suite *KeeperTestSuite) TestConnOpenTry() {
	// counterparty for A on B
	prefix := suite.chainB.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(commitmentexported.Merkle)
	suite.Require().NotNil(prefix)
	counterparty, err := connection.NewCounterparty(testClientIDB, testConnectionIDA, prefix)
	suite.Require().NoError(err)

	testCases := []struct {
		msg      string
		malleate func() uint64
		expPass  bool
	}{
		{"success", func() uint64 {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.INIT)
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
			return suite.chainB.Header.GetHeight() - 1
		}, true},
		{"consensus height > latest height", func() uint64 {
			return 0
		}, false},
		{"self consensus state not found", func() uint64 {
			//suite.ctx = suite.ctx.WithBlockHeight(100)
			return 100
		}, false},
		{"connection state verification invalid", func() uint64 {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.UNINITIALIZED)
			suite.chainB.updateClient(suite.chainA)
			return 0
		}, false},
		{"consensus state verification invalid", func() uint64 {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.INIT)
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
			return suite.chainB.Header.GetHeight()
		}, false},
		{"invalid previous connection", func() uint64 {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, types.UNINITIALIZED)
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
			return 0
		}, false},
		{"couldn't add connection to client", func() uint64 {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, types.UNINITIALIZED)
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
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
			proofInit, proofHeight := queryProof(suite.chainA, connectionKey)

			consensusKey := prefixedClientKey(testClientIDB, host.KeyConsensusState(consensusHeight))
			proofConsensus, _ := queryProof(suite.chainA, consensusKey)

			err := suite.chainB.App.IBCKeeper.ConnectionKeeper.ConnOpenTry(
				suite.chainB.GetContext(), testConnectionIDB, counterparty, testClientIDA,
				connection.GetCompatibleVersions(), proofInit, proofConsensus,
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
	version := connection.GetCompatibleVersions()[0]

	testCases := []struct {
		msg      string
		version  string
		malleate func() uint64
		expPass  bool
	}{
		{"success", version, func() uint64 {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, types.TRYOPEN)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.INIT)
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
			return suite.chainB.Header.GetHeight()
		}, true},
		{"success from tryopen", version, func() uint64 {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, types.TRYOPEN)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.TRYOPEN)
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
			return suite.chainB.Header.GetHeight()
		}, true},
		{"consensus height > latest height", version, func() uint64 {
			return 10
		}, false},
		{"connection not found", version, func() uint64 {
			return 2
		}, false},
		{"connection state is not INIT", version, func() uint64 {
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, types.UNINITIALIZED)
			suite.chainB.updateClient(suite.chainA)
			return suite.chainB.Header.GetHeight()
		}, false},
		{"incompatible IBC versions", "2.0", func() uint64 {
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, types.INIT)
			suite.chainB.updateClient(suite.chainA)
			return suite.chainB.Header.GetHeight()
		}, false},
		{"self consensus state not found", version, func() uint64 {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.INIT)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, types.TRYOPEN)
			suite.chainB.updateClient(suite.chainA)
			return suite.chainB.Header.GetHeight()
		}, false},
		{"connection state verification failed", version, func() uint64 {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.INIT)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, types.UNINITIALIZED)
			suite.chainB.updateClient(suite.chainA)
			return suite.chainB.Header.GetHeight()
		}, false},
		{"consensus state verification failed", version, func() uint64 {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.INIT)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, types.UNINITIALIZED)
			suite.chainB.updateClient(suite.chainA)
			return suite.chainB.Header.GetHeight()
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			consensusHeight := tc.malleate()

			connectionKey := host.KeyConnection(testConnectionIDB)
			proofTry, proofHeight := queryProof(suite.chainB, connectionKey)

			consensusKey := prefixedClientKey(testClientIDA, host.KeyConsensusState(consensusHeight))
			proofConsensus, _ := queryProof(suite.chainB, consensusKey)

			err := suite.chainA.App.IBCKeeper.ConnectionKeeper.ConnOpenAck(
				suite.chainA.GetContext(), testConnectionIDA, tc.version, proofTry, proofConsensus,
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
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.OPEN)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, types.TRYOPEN)
			suite.chainB.updateClient(suite.chainA)
		}, true},
		{"connection not found", func() {}, false},
		{"chain B's connection state is not TRYOPEN", func() {
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, types.UNINITIALIZED)
			suite.chainA.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, types.OPEN)
			suite.chainA.updateClient(suite.chainB)
		}, false},
		{"connection state verification failed", func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, types.INIT)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, types.TRYOPEN)
			suite.chainA.updateClient(suite.chainA)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			connectionKey := host.KeyConnection(testConnectionIDA)
			proofAck, proofHeight := queryProof(suite.chainA, connectionKey)

			if tc.expPass {
				err := suite.chainB.App.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(
					suite.chainB.GetContext(), testConnectionIDB, proofAck, proofHeight+1,
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.chainB.App.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(
					suite.chainB.GetContext(), testConnectionIDB, proofAck, proofHeight+1,
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
