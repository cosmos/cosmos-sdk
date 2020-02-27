package keeper_test

import (
	"fmt"

	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
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
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, exported.INIT)
		}, false},
		{"couldn't add connection to client", func() {}, false},
	}

	counterparty := connection.NewCounterparty(testClientIDA, testConnectionIDB, suite.chainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())

	for i, tc := range testCases {
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
	counterparty := connection.NewCounterparty(
		testClientIDB, testConnectionIDA, suite.chainB.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(),
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"success", func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, exported.INIT)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
		}, true},
		{"consensus height > latest height", func() {
		}, false},
		// {"self consensus state not found", func() {
		// 	consensusHeight = 100
		// 	suite.ctx = suite.ctx.WithBlockHeight(100)
		// }, false},
		{"connection state verification invalid", func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, exported.UNINITIALIZED)
			suite.chainB.updateClient(suite.chainA)
		}, false},
		{"consensus state verification invalid", func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			//consensusHeight = suite.chainB.Header.Height
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, exported.INIT)
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
		}, false},
		{"invalid previous connection", func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, exported.UNINITIALIZED)
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
		}, false},
		{"couldn't add connection to client", func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, exported.UNINITIALIZED)
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			connectionKey := ibctypes.KeyConnection(testConnectionIDA)
			proofInit, proofHeight := suite.queryProof(connectionKey)

			consensusKey := ibctypes.KeyConsensusState(testClientIDA, uint64(proofHeight))
			proofConsensus, consensusHeight := suite.queryProof(consensusKey)
			// TODO: This consensus height seems wrong, it should be the height of the last header for B.

			err := suite.chainB.App.IBCKeeper.ConnectionKeeper.ConnOpenTry(
				suite.chainB.GetContext(), testConnectionIDB, counterparty, testClientIDA,
				connection.GetCompatibleVersions(), proofInit, proofConsensus,
				uint64(proofHeight), uint64(consensusHeight),
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
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
		malleate func()
		expPass  bool
	}{
		{"success", version, func() {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, exported.TRYOPEN)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, exported.INIT)
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
		}, true},
		{"consensus height > latest height", version, func() {
		}, false},
		{"connection not found", version, func() {
		}, false},
		{"connection state is not INIT", version, func() {
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, exported.UNINITIALIZED)
			suite.chainB.updateClient(suite.chainA)
		}, false},
		{"incompatible IBC versions", "2.0", func() {
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, exported.INIT)
			suite.chainB.updateClient(suite.chainA)
		}, false},
		{"self consensus state not found", version, func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, exported.INIT)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, exported.TRYOPEN)
			suite.chainB.updateClient(suite.chainA)
		}, false},
		{"connection state verification failed", version, func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, exported.INIT)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, exported.UNINITIALIZED)
			suite.chainB.updateClient(suite.chainA)
		}, false},
		{"consensus state verification failed", version, func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, exported.INIT)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, exported.UNINITIALIZED)
			suite.chainB.updateClient(suite.chainA)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			connectionKey := ibctypes.KeyConnection(testConnectionIDB)
			proofTry, proofHeight := suite.queryProof(connectionKey)

			// TODO: This consensus height seems wrong, it should be the height of the client of A?
			consensusHeight := uint64(1)
			consensusKey := ibctypes.KeyConsensusState(testClientIDB, uint64(consensusHeight))
			proofConsensus, _ := suite.queryProof(consensusKey)

			err := suite.chainA.App.IBCKeeper.ConnectionKeeper.ConnOpenAck(
				suite.chainA.GetContext(), testConnectionIDA, tc.version, proofTry, proofConsensus,
				uint64(proofHeight), uint64(consensusHeight),
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

// TestConnOpenAck - Chain B (ID #2) calls ConnOpenConfirm to confirm that
// Chain A (ID #1) state is now OPEN.
func (suite *KeeperTestSuite) TestConnOpenConfirm() {
	testCases := []testCase{
		{"success", func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, exported.OPEN)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, exported.TRYOPEN)
			suite.chainB.updateClient(suite.chainA)
		}, true},
		{"connection not found", func() {}, false},
		{"chain B's connection state is not TRYOPEN", func() {
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, exported.UNINITIALIZED)
		}, false},
		{"consensus state not found", func() {
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, exported.TRYOPEN)
			suite.chainA.updateClient(suite.chainB)
		}, false},
		{"connection state verification failed", func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, exported.INIT)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, exported.TRYOPEN)
			suite.chainA.updateClient(suite.chainA)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			connectionKey := ibctypes.KeyConnection(testConnectionIDA)
			proofAck, proofHeight := suite.queryProof(connectionKey)

			if tc.expPass {
				err := suite.chainB.App.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(
					suite.chainB.GetContext(), testConnectionIDB, proofAck, uint64(proofHeight),
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.chainB.App.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(
					suite.chainB.GetContext(), testConnectionIDB, proofAck, uint64(proofHeight),
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
