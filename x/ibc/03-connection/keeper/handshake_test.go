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
	testCases := []testCase{
		{"success", func() {
			suite.createClient(testClientID1)
		}, true},
		{"connection already exists", func() {
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.INIT)
		}, false},
		{"couldn't add connection to client", func() {}, false},
	}

	counterparty := connection.NewCounterparty(testClientID2, testConnectionID2, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenInit(suite.ctx, testConnectionID1, testClientID1, counterparty)

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
		testClientID1, testConnectionID1, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(),
	)

	consensusHeight := int64(0)

	testCases := []testCase{
		{"success", func() {
			suite.createClient(testClientID1) // height = 2
			suite.createClient(testClientID2)
			consensusHeight = suite.ctx.BlockHeight() // height = 3
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.INIT)
			suite.updateClient(testClientID1)
			suite.updateClient(testClientID2)
		}, true},
		{"consensus state not found", func() {}, false},
		{"connection state verification invalid", func() {
			suite.createClient(testClientID1)
			suite.createClient(testClientID2)
			consensusHeight = suite.ctx.BlockHeight()
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.UNINITIALIZED)
			suite.updateClient(testClientID1)
		}, false},
		{"invalid previous connection", func() {
			suite.createClient(testClientID1)
			suite.createClient(testClientID2)
			consensusHeight = suite.ctx.BlockHeight()
			suite.createConnection(testConnectionID2, testConnectionID1, testClientID2, testClientID1, exported.UNINITIALIZED)
			suite.updateClient(testClientID1)
		}, false},
		{"couldn't add connection to client", func() {
			suite.createClient(testClientID1)
			consensusHeight = suite.ctx.BlockHeight()
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.INIT)
			suite.createConnection(testConnectionID2, testConnectionID1, testClientID2, testClientID1, exported.INIT)
			suite.updateClient(testClientID1)
		}, false},
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			// connectionKey := ibctypes.KeyConnection(testConnectionID1)
			// proofInit, proofHeight := suite.queryProof(connectionKey)

			// consensusKey := ibctypes.KeyConsensusState(testClientID1, uint64(proofHeight))
			// proofConsensus, consensusHeight := suite.queryProof(consensusKey)

			proofHeight := consensusHeight - 1

			if tc.expPass {
				err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenTry(
					suite.ctx, testConnectionID2, counterparty, testClientID2,
					connection.GetCompatibleVersions(), ibctypes.ValidProof{}, ibctypes.ValidProof{},
					uint64(proofHeight), uint64(consensusHeight),
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenTry(
					suite.ctx, testConnectionID2, counterparty, testClientID2,
					connection.GetCompatibleVersions(), ibctypes.InvalidProof{}, ibctypes.ValidProof{},
					uint64(proofHeight), uint64(consensusHeight),
				)
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

// TestConnOpenAck - Chain A (ID #1) calls TestConnOpenAck to acknowledge (ACK state)
// the initialization (TRYINIT) of the connection on  Chain B (ID #2).
func (suite *KeeperTestSuite) TestConnOpenAck() {
	version := connection.GetCompatibleVersions()[0]
	consensusHeight := int64(0)

	testCases := []struct {
		msg      string
		version  string
		malleate func()
		expPass  bool
	}{
		{"success", version, func() {
			suite.createClient(testClientID1)
			consensusHeight = suite.ctx.BlockHeight()
			suite.createClient(testClientID2)
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.INIT)
			suite.createConnection(testConnectionID2, testConnectionID1, testClientID2, testClientID1, exported.TRYOPEN)
			suite.updateClient(testClientID1)
		}, true},
		{"connection not found", version, func() {}, false},
		{"connection state is not INIT", version, func() {
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.UNINITIALIZED)
			suite.updateClient(testClientID1)
		}, false},
		{"incompatible IBC versions", "2.0", func() {
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.INIT)
			suite.updateClient(testClientID1)
		}, false},
		{"consensus state not found", version, func() {
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.INIT)
			suite.updateClient(testClientID1)
		}, false},
		{"connection state verification failed", version, func() {
			suite.createClient(testClientID1)
			consensusHeight = suite.ctx.BlockHeight()
			suite.createClient(testClientID2)
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.INIT)
			suite.createConnection(testConnectionID2, testConnectionID1, testClientID2, testClientID1, exported.UNINITIALIZED)
			suite.updateClient(testClientID1)
		}, false},
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			// connectionKey := ibctypes.KeyConnection(testConnectionID2)
			// proofTry, proofHeight := suite.queryProof(connectionKey)

			// consensusKey := ibctypes.KeyConsensusState(testClientID2, uint64(proofHeight))
			// proofConsensus, consensusHeight := suite.queryProof(consensusKey)

			proofHeight := consensusHeight - 1

			if tc.expPass {
				err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenAck(
					suite.ctx, testConnectionID1, tc.version, ibctypes.ValidProof{}, ibctypes.ValidProof{},
					uint64(proofHeight), uint64(consensusHeight),
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenAck(
					suite.ctx, testConnectionID1, tc.version, ibctypes.InvalidProof{}, ibctypes.ValidProof{},
					uint64(proofHeight), uint64(consensusHeight),
				)
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

// TestConnOpenAck - Chain B (ID #2) calls ConnOpenConfirm to confirm that
// Chain A (ID #1) state is now OPEN.
func (suite *KeeperTestSuite) TestConnOpenConfirm() {
	consensusHeight := int64(0)

	testCases := []testCase{
		{"success", func() {
			suite.createClient(testClientID1)
			suite.createClient(testClientID2)
			consensusHeight = suite.ctx.BlockHeight()
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.OPEN)
			suite.createConnection(testConnectionID2, testConnectionID1, testClientID2, testClientID1, exported.TRYOPEN)
			suite.updateClient(testClientID1)
		}, true},
		{"connection not found", func() {}, false},
		{"chain B's connection state is not TRYOPEN", func() {
			suite.createConnection(testConnectionID2, testConnectionID1, testClientID2, testClientID1, exported.UNINITIALIZED)
		}, false},
		{"consensus state not found", func() {
			suite.createConnection(testConnectionID2, testConnectionID1, testClientID2, testClientID1, exported.TRYOPEN)
			suite.updateClient(testClientID2)
		}, false},
		{"connection state verification failed", func() {
			suite.createClient(testClientID1)
			suite.createClient(testClientID2)
			consensusHeight = suite.ctx.BlockHeight()
			suite.updateClient(testClientID1)
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.INIT)
			suite.createConnection(testConnectionID2, testConnectionID1, testClientID2, testClientID1, exported.TRYOPEN)
			suite.updateClient(testClientID1)
		}, false},
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			// connectionKey := ibctypes.KeyConnection(testConnectionID2)
			// proofAck, proofHeight := suite.queryProof(connectionKey)
			proofHeight := consensusHeight - 1

			if tc.expPass {
				err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(
					suite.ctx, testConnectionID2, ibctypes.ValidProof{}, uint64(proofHeight),
					uint64(consensusHeight),
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(
					suite.ctx, testConnectionID2, ibctypes.InvalidProof{}, uint64(proofHeight),
					uint64(consensusHeight),
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
