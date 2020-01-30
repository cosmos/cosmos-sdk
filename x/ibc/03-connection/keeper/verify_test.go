package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

const (
	testPort1 = "firstport"
	testPort2 = "secondport"

	testChannel1 = "firstchannel"
	testChannel2 = "secondchannel"
)

func (suite *KeeperTestSuite) TestVerifyClientConsensusState() {
	counterparty := types.NewCounterparty(
		testClientID2, testConnectionID2,
		suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(),
	)

	connection1 := types.NewConnectionEnd(
		exported.UNINITIALIZED, testClientID1, counterparty,
		types.GetCompatibleVersions(),
	)

	cases := []struct {
		msg        string
		connection types.ConnectionEnd
		proof      commitment.ProofI
		malleate   func()
		expPass    bool
	}{
		{"verification success", connection1, validProof{}, func() {
			suite.createClient(testClientID1)
		}, true},
		{"client state not found", connection1, validProof{}, func() {}, false},
		{"verification failed", connection1, invalidProof{}, func() {
			suite.createClient(testClientID2)
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			proofHeight := suite.ctx.BlockHeight()

			// TODO: remove mocked types and uncomment
			// consensusKey := ibctypes.KeyConsensusState(testClientID1, uint64(suite.app.LastBlockHeight()))
			// proof, proofHeight := suite.queryProof(consensusKey)

			err := suite.app.IBCKeeper.ConnectionKeeper.VerifyClientConsensusState(
				suite.ctx, tc.connection, uint64(proofHeight), tc.proof, suite.consensusState,
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
	// connectionKey := ibctypes.KeyConnection(testConnectionID1)
	cases := []struct {
		msg      string
		proof    commitment.ProofI
		malleate func()
		expPass  bool
	}{
		{"verification success", validProof{}, func() {
			suite.createClient(testClientID1)
			suite.createClient(testClientID2)
		}, true},
		{"client state not found", validProof{}, func() {}, false},
		{"verification failed", invalidProof{}, func() {
			suite.createClient(testClientID1)
			suite.createClient(testClientID2)
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			connection := suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.OPEN)
			suite.updateClient(testClientID1)
			counterparty := types.NewCounterparty(testClientID1, testConnectionID1, commitment.NewPrefix([]byte("ibc")))
			expectedConnection := types.NewConnectionEnd(exported.INIT, testClientID2, counterparty, []string{"1.0.0"})
			suite.updateClient(testClientID1)
			proofHeight := uint64(3)
			// proof, proofHeight := suite.queryProof(connectionKey)

			err := suite.app.IBCKeeper.ConnectionKeeper.VerifyConnectionState(
				suite.ctx, connection, proofHeight, tc.proof, testConnectionID1, expectedConnection,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVerifyChannelState() {
	// channelKey := ibctypes.KeyChannel(testPort1, testChannel1)
	cases := []struct {
		msg         string
		proof       commitment.ProofI
		proofHeight uint64
		malleate    func()
		expPass     bool
	}{
		{"verification success", validProof{}, 2, func() {
			suite.createClient(testClientID1)
		}, true},
		{"client state not found", validProof{}, 2, func() {}, false},
		{"consensus state not found", validProof{}, 100, func() {
			suite.createClient(testClientID1)
		}, false},
		{"verification failed", invalidProof{}, 2, func() {
			suite.createClient(testClientID2)
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			connection := suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.OPEN)
			channel := suite.createChannel(
				testPort1, testChannel1, testPort2, testChannel2,
				channelexported.OPEN, channelexported.ORDERED, testConnectionID1,
			)
			suite.updateClient(testClientID1)

			// proof, proofHeight := suite.queryProof(channelKey)
			err := suite.app.IBCKeeper.ConnectionKeeper.VerifyChannelState(
				suite.ctx, connection, tc.proofHeight, tc.proof, testPort1,
				testChannel1, channel,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVerifyPacketCommitment() {
	// commitmentKey := ibctypes.KeyPacketCommitment(testPort1, testChannel1, 1)
	commitmentBz := []byte("commitment")

	cases := []struct {
		msg         string
		proof       commitment.ProofI
		proofHeight uint64
		malleate    func()
		expPass     bool
	}{
		{"verification success", validProof{}, 2, func() {
			suite.createClient(testClientID1)
		}, true},
		{"client state not found", validProof{}, 2, func() {}, false},
		{"consensus state not found", validProof{}, 100, func() {
			suite.createClient(testClientID1)
		}, false},
		{"verification failed", invalidProof{}, 2, func() {
			suite.createClient(testClientID2)
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			connection := suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.OPEN)
			suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort1, testChannel1, 1, commitmentBz)
			suite.updateClient(testClientID1)

			// proof, proofHeight := suite.queryProof(commitmentKey)
			err := suite.app.IBCKeeper.ConnectionKeeper.VerifyPacketCommitment(
				suite.ctx, connection, tc.proofHeight, tc.proof, testPort1,
				testChannel1, 1, commitmentBz,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVerifyPacketAcknowledgement() {
	// packetAckKey := ibctypes.KeyPacketAcknowledgement(testPort1, testChannel1, 1)
	ack := []byte("acknowledgement")

	cases := []struct {
		msg         string
		proof       commitment.ProofI
		proofHeight uint64
		malleate    func()
		expPass     bool
	}{
		{"verification success", validProof{}, 2, func() {
			suite.createClient(testClientID1)
		}, true},
		{"client state not found", validProof{}, 2, func() {}, false},
		{"consensus state not found", validProof{}, 100, func() {
			suite.createClient(testClientID1)
		}, false},
		{"verification failed", invalidProof{}, 2, func() {
			suite.createClient(testClientID2)
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			connection := suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.OPEN)
			suite.app.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(suite.ctx, testPort1, testChannel1, 1, ack)
			suite.updateClient(testClientID1)
			// proof, proofHeight := suite.queryProof(packetAckKey)

			err := suite.app.IBCKeeper.ConnectionKeeper.VerifyPacketAcknowledgement(
				suite.ctx, connection, tc.proofHeight, tc.proof, testPort1,
				testChannel1, 1, ack,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVerifyPacketAcknowledgementAbsence() {
	// packetAckKey := ibctypes.KeyPacketAcknowledgement(testPort1, testChannel1, 1)

	cases := []struct {
		msg         string
		proof       commitment.ProofI
		proofHeight uint64
		malleate    func()
		expPass     bool
	}{
		{"verification success", validProof{}, 2, func() {
			suite.createClient(testClientID1)
		}, true},
		{"client state not found", validProof{}, 2, func() {}, false},
		{"consensus state not found", validProof{}, 100, func() {
			suite.createClient(testClientID1)
		}, false},
		{"verification failed", invalidProof{}, 2, func() {
			suite.createClient(testClientID2)
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			connection := suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.OPEN)
			suite.updateClient(testClientID1)

			// proof, proofHeight := suite.queryProof(packetAckKey)

			err := suite.app.IBCKeeper.ConnectionKeeper.VerifyPacketAcknowledgementAbsence(
				suite.ctx, connection, tc.proofHeight, tc.proof, testPort1,
				testChannel1, 1,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVerifyNextSequenceRecv() {
	// nextSeqRcvKey := ibctypes.KeyNextSequenceRecv(testPort1, testChannel1)

	cases := []struct {
		msg         string
		proof       commitment.ProofI
		proofHeight uint64
		malleate    func()
		expPass     bool
	}{
		{"verification success", validProof{}, 2, func() {
			suite.createClient(testClientID1)
		}, true},
		{"client state not found", validProof{}, 2, func() {}, false},
		{"consensus state not found", validProof{}, 100, func() {
			suite.createClient(testClientID1)
		}, false},
		{"verification failed", invalidProof{}, 2, func() {
			suite.createClient(testClientID2)
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			connection := suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.OPEN)
			suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.ctx, testPort1, testChannel1, 1)
			suite.updateClient(testClientID1)

			// proof, proofHeight := suite.queryProof(nextSeqRcvKey)
			err := suite.app.IBCKeeper.ConnectionKeeper.VerifyNextSequenceRecv(
				suite.ctx, connection, tc.proofHeight, tc.proof, testPort1,
				testChannel1, 1,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}
