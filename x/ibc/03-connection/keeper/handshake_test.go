package keeper_test

import (
	"fmt"

	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
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

	counterparty := connection.NewCounterparty(testClientIDB, testConnectionIDB, suite.chainB.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			err := suite.chainA.App.IBCKeeper.ConnectionKeeper.ConnOpenInit(suite.chainA.GetContext(), testConnectionIDA, testClientIDA, counterparty)

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
		testClientIDA, testConnectionIDA, suite.chainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(),
	)

	var (
		consensusHeight int64 = 0
		proofHeight     int64 = 0
	)
	testCases := []struct {
		msg            string
		proofInit      commitment.ProofI
		proofConsensus commitment.ProofI
		malleate       func()
		expPass        bool
	}{
		{"success", ibctypes.ValidProof{}, ibctypes.ValidProof{}, func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, exported.INIT)
			proofHeight = suite.chainA.Header.Height
			suite.chainA.CreateClient(suite.chainB)
			consensusHeight = suite.chainB.Header.Height
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
		}, true},
		{"consensus height > latest height", ibctypes.ValidProof{}, ibctypes.ValidProof{}, func() {
			consensusHeight = 100
		}, false},
		// {"self consensus state not found", ibctypes.ValidProof{}, ibctypes.ValidProof{}, func() {
		// 	consensusHeight = 100
		// 	suite.ctx = suite.ctx.WithBlockHeight(100)
		// }, false},
		{"connection state verification invalid", ibctypes.InvalidProof{}, ibctypes.ValidProof{}, func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			consensusHeight = suite.chainB.Header.Height
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, exported.UNINITIALIZED)
			suite.chainB.updateClient(suite.chainA)
		}, false},
		{"consensus state verification invalid", ibctypes.ValidProof{}, ibctypes.InvalidProof{}, func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			consensusHeight = suite.chainB.Header.Height
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, exported.INIT)
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
		}, false},
		{"invalid previous connection", ibctypes.ValidProof{}, ibctypes.ValidProof{}, func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			consensusHeight = suite.chainB.Header.Height
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, exported.UNINITIALIZED)
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
		}, false},
		{"couldn't add connection to client", ibctypes.ValidProof{}, ibctypes.ValidProof{}, func() {
			suite.chainB.CreateClient(suite.chainA)
			consensusHeight = suite.chainB.Header.Height
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

			// connectionKey := ibctypes.KeyConnection(testConnectionIDA)
			// proofInit, proofHeight := suite.queryProof(connectionKey)

			// consensusKey := ibctypes.KeyConsensusState(testClientIDA, uint64(proofHeight))
			// proofConsensus, consensusHeight := suite.queryProof(consensusKey)

			err := suite.chainB.App.IBCKeeper.ConnectionKeeper.ConnOpenTry(
				suite.chainB.GetContext(), testConnectionIDB, counterparty, testClientIDB,
				connection.GetCompatibleVersions(), tc.proofInit, tc.proofConsensus,
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

	var (
		consensusHeight int64 = 0
		proofHeight     int64 = 0
	)

	testCases := []struct {
		msg            string
		version        string
		proofTry       commitment.ProofI
		proofConsensus commitment.ProofI
		malleate       func()
		expPass        bool
	}{
		{"success", version, ibctypes.ValidProof{}, ibctypes.ValidProof{}, func() {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainA.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, exported.TRYOPEN)
			consensusHeight = suite.chainB.Header.Height
			suite.chainB.CreateClient(suite.chainA)
			proofHeight = suite.chainA.Header.Height
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, exported.INIT)
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
		}, true},
		{"consensus height > latest height", version, ibctypes.ValidProof{}, ibctypes.ValidProof{}, func() {
			consensusHeight = 100
		}, false},
		{"connection not found", version, ibctypes.ValidProof{}, ibctypes.ValidProof{}, func() {
			consensusHeight = suite.chainB.Header.Height
		}, false},
		{"connection state is not INIT", version, ibctypes.ValidProof{}, ibctypes.ValidProof{}, func() {
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, exported.UNINITIALIZED)
			suite.chainB.updateClient(suite.chainA)
		}, false},
		{"incompatible IBC versions", "2.0", ibctypes.ValidProof{}, ibctypes.ValidProof{}, func() {
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, exported.INIT)
			suite.chainB.updateClient(suite.chainA)
		}, false},
		{"self consensus state not found", version, ibctypes.ValidProof{}, ibctypes.ValidProof{}, func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, exported.INIT)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, exported.TRYOPEN)
			suite.chainB.updateClient(suite.chainA)
			consensusHeight = 100
		}, false},
		{"connection state verification failed", version, ibctypes.InvalidProof{}, ibctypes.ValidProof{}, func() {
			suite.chainB.CreateClient(suite.chainA)
			consensusHeight = suite.chainB.Header.Height
			suite.chainA.CreateClient(suite.chainB)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, exported.INIT)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, exported.UNINITIALIZED)
			suite.chainB.updateClient(suite.chainA)
		}, false},
		{"consensus state verification failed", version, ibctypes.ValidProof{}, ibctypes.InvalidProof{}, func() {
			suite.chainB.CreateClient(suite.chainA)
			consensusHeight = suite.chainB.Header.Height
			suite.chainA.CreateClient(suite.chainB)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, exported.INIT)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, exported.UNINITIALIZED)
			suite.chainB.updateClient(suite.chainA)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			// connectionKey := ibctypes.KeyConnection(testConnectionIDB)
			// proofTry, proofHeight := suite.queryProof(connectionKey)

			// consensusKey := ibctypes.KeyConsensusState(testClientIDB, uint64(proofHeight))
			// proofConsensus, consensusHeight := suite.queryProof(consensusKey)

			err := suite.chainA.App.IBCKeeper.ConnectionKeeper.ConnOpenAck(
				suite.chainA.GetContext(), testConnectionIDA, tc.version, tc.proofTry, tc.proofConsensus,
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
	// consensusHeight := int64(0)
	proofHeight := int64(0)

	testCases := []testCase{
		{"success", func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			proofHeight = suite.chainB.Header.Height
			// consensusHeight = suite.chainA.Header.Height
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, exported.OPEN)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, exported.TRYOPEN)
			suite.chainB.updateClient(suite.chainA)
		}, true},
		{"connection not found", func() {}, false},
		{"chain B's connection state is not TRYOPEN", func() {
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, exported.UNINITIALIZED)
		}, false},
		{"consensus state not found", func() {
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, exported.TRYOPEN)
			suite.chainA.updateClient(suite.chainB)
		}, false},
		{"connection state verification failed", func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			// consensusHeight = suite.chainA.Header.Height
			proofHeight = suite.chainB.Header.Height
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, exported.INIT)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, exported.TRYOPEN)
			suite.chainA.updateClient(suite.chainA)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			// connectionKey := ibctypes.KeyConnection(testConnectionIDB)
			// proofAck, proofHeight := suite.queryProof(connectionKey)

			if tc.expPass {
				err := suite.chainB.App.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(
					suite.chainB.GetContext(), testConnectionIDB, ibctypes.ValidProof{}, uint64(proofHeight),
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.chainB.App.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(
					suite.chainB.GetContext(), testConnectionIDB, ibctypes.InvalidProof{}, uint64(proofHeight),
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
