package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

const (
	testPort1 = "firstport"
	testPort2 = "secondport"

	testChannel1 = "firstchannel"
	testChannel2 = "secondchannel"
)

func (suite *KeeperTestSuite) TestVerifyClientConsensusState() {
	counterparty := types.NewCounterparty(
		testClientIDB, testConnectionIDB,
		suite.chainB.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(),
	)

	connection1 := types.NewConnectionEnd(
		exported.UNINITIALIZED, testClientIDB, counterparty,
		types.GetCompatibleVersions(),
	)

	cases := []struct {
		msg        string
		connection types.ConnectionEnd
		proof      commitment.ProofI
		malleate   func()
		expPass    bool
	}{
		{"verification success", connection1, ibctypes.ValidProof{}, func() {
			suite.chainA.CreateClient(suite.chainB)
		}, true},
		{"client state not found", connection1, ibctypes.ValidProof{}, func() {}, false},
		{"verification failed", connection1, ibctypes.InvalidProof{}, func() {
			suite.chainA.CreateClient(suite.chainA)
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			proofHeight := uint64(suite.chainB.Header.Height)

			// TODO: remove mocked types and uncomment
			// consensusKey := ibctypes.KeyConsensusState(testClientIDA, uint64(suite.app.LastBlockHeight()))
			// proof, proofHeight := suite.queryProof(consensusKey)

			err := suite.chainA.App.IBCKeeper.ConnectionKeeper.VerifyClientConsensusState(
				suite.chainA.GetContext(), tc.connection, proofHeight, tc.proof, suite.chainB.Header.ConsensusState(),
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVerifyConnectionState() {
	// connectionKey := ibctypes.KeyConnection(testConnectionIDA)
	cases := []struct {
		msg      string
		proof    commitment.ProofI
		malleate func()
		expPass  bool
	}{
		{"verification success", ibctypes.ValidProof{}, func() {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
		}, true},
		{"client state not found", ibctypes.ValidProof{}, func() {}, false},
		{"verification failed", ibctypes.InvalidProof{}, func() {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			connection := suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, exported.OPEN)
			suite.chainB.updateClient(suite.chainA)
			counterparty := types.NewCounterparty(testClientIDA, testConnectionIDA, commitment.NewPrefix([]byte("ibc")))
			expectedConnection := types.NewConnectionEnd(exported.INIT, testClientIDB, counterparty, []string{"1.0.0"})
			suite.chainB.updateClient(suite.chainA)
			proofHeight := uint64(suite.chainA.Header.Height)
			// proof, proofHeight := suite.queryProof(connectionKey)

			err := suite.chainB.App.IBCKeeper.ConnectionKeeper.VerifyConnectionState(
				suite.chainB.GetContext(), connection, proofHeight, tc.proof, testConnectionIDA, expectedConnection,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

// func (suite *KeeperTestSuite) TestVerifyChannelState() {
// 	// channelKey := ibctypes.KeyChannel(testPort1, testChannel1)
// 	cases := []struct {
// 		msg         string
// 		proof       commitment.ProofI
// 		proofHeight uint64
// 		malleate    func()
// 		expPass     bool
// 	}{
// 		{"verification success", ibctypes.ValidProof{}, 2, func() {
// 			suite.createClient(testClientIDA)
// 		}, true},
// 		{"client state not found", ibctypes.ValidProof{}, 2, func() {}, false},
// 		{"consensus state not found", ibctypes.ValidProof{}, 100, func() {
// 			suite.createClient(testClientIDA)
// 		}, false},
// 		{"verification failed", ibctypes.InvalidProof{}, 2, func() {
// 			suite.createClient(testClientIDB)
// 		}, false},
// 	}

// 	for i, tc := range cases {
// 		tc := tc
// 		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
// 			suite.SetupTest() // reset

// 			tc.malleate()
// 			connection := suite.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, exported.OPEN)
// 			channel := suite.createChannel(
// 				testPort1, testChannel1, testPort2, testChannel2,
// 				channelexported.OPEN, channelexported.ORDERED, testConnectionIDA,
// 			)
// 			suite.updateClient(testClientIDA)

// 			// proof, proofHeight := suite.queryProof(channelKey)
// 			err := suite.app.IBCKeeper.ConnectionKeeper.VerifyChannelState(
// 				suite.ctx, connection, tc.proofHeight, tc.proof, testPort1,
// 				testChannel1, channel,
// 			)

// 			if tc.expPass {
// 				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
// 			} else {
// 				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
// 			}
// 		})
// 	}
// }

// func (suite *KeeperTestSuite) TestVerifyPacketCommitment() {
// 	// commitmentKey := ibctypes.KeyPacketCommitment(testPort1, testChannel1, 1)
// 	commitmentBz := []byte("commitment")

// 	cases := []struct {
// 		msg         string
// 		proof       commitment.ProofI
// 		proofHeight uint64
// 		malleate    func()
// 		expPass     bool
// 	}{
// 		{"verification success", ibctypes.ValidProof{}, 2, func() {
// 			suite.createClient(testClientIDA)
// 		}, true},
// 		{"client state not found", ibctypes.ValidProof{}, 2, func() {}, false},
// 		{"consensus state not found", ibctypes.ValidProof{}, 100, func() {
// 			suite.createClient(testClientIDA)
// 		}, false},
// 		{"verification failed", ibctypes.InvalidProof{}, 2, func() {
// 			suite.createClient(testClientIDB)
// 		}, false},
// 	}

// 	for i, tc := range cases {
// 		tc := tc
// 		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
// 			suite.SetupTest() // reset

// 			tc.malleate()
// 			connection := suite.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, exported.OPEN)
// 			suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort1, testChannel1, 1, commitmentBz)
// 			suite.updateClient(testClientIDA)

// 			// proof, proofHeight := suite.queryProof(commitmentKey)
// 			err := suite.app.IBCKeeper.ConnectionKeeper.VerifyPacketCommitment(
// 				suite.ctx, connection, tc.proofHeight, tc.proof, testPort1,
// 				testChannel1, 1, commitmentBz,
// 			)

// 			if tc.expPass {
// 				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
// 			} else {
// 				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
// 			}
// 		})
// 	}
// }

// func (suite *KeeperTestSuite) TestVerifyPacketAcknowledgement() {
// 	// packetAckKey := ibctypes.KeyPacketAcknowledgement(testPort1, testChannel1, 1)
// 	ack := []byte("acknowledgement")

// 	cases := []struct {
// 		msg         string
// 		proof       commitment.ProofI
// 		proofHeight uint64
// 		malleate    func()
// 		expPass     bool
// 	}{
// 		{"verification success", ibctypes.ValidProof{}, 2, func() {
// 			suite.createClient(testClientIDA)
// 		}, true},
// 		{"client state not found", ibctypes.ValidProof{}, 2, func() {}, false},
// 		{"consensus state not found", ibctypes.ValidProof{}, 100, func() {
// 			suite.createClient(testClientIDA)
// 		}, false},
// 		{"verification failed", ibctypes.InvalidProof{}, 2, func() {
// 			suite.createClient(testClientIDB)
// 		}, false},
// 	}

// 	for i, tc := range cases {
// 		tc := tc
// 		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
// 			suite.SetupTest() // reset

// 			tc.malleate()
// 			connection := suite.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, exported.OPEN)
// 			suite.app.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(suite.ctx, testPort1, testChannel1, 1, ack)
// 			suite.updateClient(testClientIDA)
// 			// proof, proofHeight := suite.queryProof(packetAckKey)

// 			err := suite.app.IBCKeeper.ConnectionKeeper.VerifyPacketAcknowledgement(
// 				suite.ctx, connection, tc.proofHeight, tc.proof, testPort1,
// 				testChannel1, 1, ack,
// 			)

// 			if tc.expPass {
// 				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
// 			} else {
// 				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
// 			}
// 		})
// 	}
// }

// func (suite *KeeperTestSuite) TestVerifyPacketAcknowledgementAbsence() {
// 	// packetAckKey := ibctypes.KeyPacketAcknowledgement(testPort1, testChannel1, 1)

// 	cases := []struct {
// 		msg         string
// 		proof       commitment.ProofI
// 		proofHeight uint64
// 		malleate    func()
// 		expPass     bool
// 	}{
// 		{"verification success", ibctypes.ValidProof{}, 2, func() {
// 			suite.createClient(testClientIDA)
// 		}, true},
// 		{"client state not found", ibctypes.ValidProof{}, 2, func() {}, false},
// 		{"consensus state not found", ibctypes.ValidProof{}, 100, func() {
// 			suite.createClient(testClientIDA)
// 		}, false},
// 		{"verification failed", ibctypes.InvalidProof{}, 2, func() {
// 			suite.createClient(testClientIDB)
// 		}, false},
// 	}

// 	for i, tc := range cases {
// 		tc := tc
// 		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
// 			suite.SetupTest() // reset

// 			tc.malleate()
// 			connection := suite.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, exported.OPEN)
// 			suite.updateClient(testClientIDA)

// 			// proof, proofHeight := suite.queryProof(packetAckKey)

// 			err := suite.app.IBCKeeper.ConnectionKeeper.VerifyPacketAcknowledgementAbsence(
// 				suite.ctx, connection, tc.proofHeight, tc.proof, testPort1,
// 				testChannel1, 1,
// 			)

// 			if tc.expPass {
// 				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
// 			} else {
// 				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
// 			}
// 		})
// 	}
// }

// func (suite *KeeperTestSuite) TestVerifyNextSequenceRecv() {
// 	// nextSeqRcvKey := ibctypes.KeyNextSequenceRecv(testPort1, testChannel1)

// 	cases := []struct {
// 		msg         string
// 		proof       commitment.ProofI
// 		proofHeight uint64
// 		malleate    func()
// 		expPass     bool
// 	}{
// 		{"verification success", ibctypes.ValidProof{}, 2, func() {
// 			suite.createClient(testClientIDA)
// 		}, true},
// 		{"client state not found", ibctypes.ValidProof{}, 2, func() {}, false},
// 		{"consensus state not found", ibctypes.ValidProof{}, 100, func() {
// 			suite.createClient(testClientIDA)
// 		}, false},
// 		{"verification failed", ibctypes.InvalidProof{}, 2, func() {
// 			suite.createClient(testClientIDB)
// 		}, false},
// 	}

// 	for i, tc := range cases {
// 		tc := tc
// 		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
// 			suite.SetupTest() // reset

// 			tc.malleate()
// 			connection := suite.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, exported.OPEN)
// 			suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.ctx, testPort1, testChannel1, 1)
// 			suite.updateClient(testClientIDA)

// 			// proof, proofHeight := suite.queryProof(nextSeqRcvKey)
// 			err := suite.app.IBCKeeper.ConnectionKeeper.VerifyNextSequenceRecv(
// 				suite.ctx, connection, tc.proofHeight, tc.proof, testPort1,
// 				testChannel1, 1,
// 			)

// 			if tc.expPass {
// 				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
// 			} else {
// 				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
// 			}
// 		})
// 	}
// }
