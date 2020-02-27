package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
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
		testClientIDA, testConnectionIDA,
		suite.chainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(),
	)

	connection1 := types.NewConnectionEnd(
		exported.UNINITIALIZED, testClientIDB, counterparty,
		types.GetCompatibleVersions(),
	)

	cases := []struct {
		msg        string
		connection types.ConnectionEnd
		malleate   func()
		expPass    bool
	}{
		{"verification success", connection1, func() {
			suite.chainA.CreateClient(suite.chainB)
		}, true},
		{"client state not found", connection1, func() {}, false},
		{"verification failed", connection1, func() {
			suite.chainA.CreateClient(suite.chainA)
		}, false},
	}

	// Create Client of chain B on Chain App
	// Check that we can verify B's consensus state on chain A
	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			proofHeight := uint64(suite.chainB.Header.Height)

			// TODO: is this the right consensus height
			consensusKey := ibctypes.KeyConsensusState(testClientIDA, uint64(suite.chainA.App.LastBlockHeight()))
			proof, proofHeight := suite.queryProof(consensusKey)

			err := suite.chainA.App.IBCKeeper.ConnectionKeeper.VerifyClientConsensusState(
				suite.chainA.GetContext(), tc.connection, proofHeight, proof, suite.chainB.Header.ConsensusState(),
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
	connectionKey := ibctypes.KeyConnection(testConnectionIDA)
	cases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"verification success", func() {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
		}, true},
		{"client state not found", func() {}, false},
		{"verification failed", func() {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
		}, false},
	}

	// Chains A and B create clients for each other
	// A creates connectionEnd for chain B and stores it in state
	// Check that B can verify connection is stored after some updates
	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			// create and store connection on chain A
			expectedConnection := suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, exported.OPEN)

			// // create expected connection
			// TODO: why is this commented
			// expectedConnection := types.NewConnectionEnd(exported.INIT, testClientIDB, counterparty, []string{"1.0.0"})

			// perform a couple updates of chain A on chain B
			suite.chainB.updateClient(suite.chainA)
			suite.chainB.updateClient(suite.chainA)
			proof, proofHeight := suite.queryProof(connectionKey)

			counterparty := types.NewCounterparty(testClientIDB, testConnectionIDA, commitment.NewPrefix([]byte("ibc")))
			connection := types.NewConnectionEnd(exported.UNINITIALIZED, testClientIDA, counterparty, []string{"1.0.0"})
			// Ensure chain B can verify connection exists in chain A
			err := suite.chainB.App.IBCKeeper.ConnectionKeeper.VerifyConnectionState(
				suite.chainB.GetContext(), connection, proofHeight, proof, testConnectionIDA, expectedConnection,
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
	channelKey := ibctypes.KeyChannel(testPort1, testChannel1)
	cases := []struct {
		msg         string
		proofHeight uint64
		malleate    func()
		expPass     bool
	}{
		{"verification success", 2, func() {
			suite.chainB.CreateClient(suite.chainA)
		}, true},
		{"client state not found", 2, func() {}, false},
		{"consensus state not found", 100, func() {
			suite.chainB.CreateClient(suite.chainA)
		}, false},
		{"verification failed", 2, func() {
			suite.chainB.CreateClient(suite.chainB)
		}, false},
	}

	// Chain A creates channel for chain B and stores in its state
	// Check that chainB can verify channel is stored in chain A
	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			// Create and store channel on chain A
			connection := suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, exported.OPEN)
			channel := suite.chainA.createChannel(
				testPort1, testChannel1, testPort2, testChannel2,
				channelexported.OPEN, channelexported.ORDERED, testConnectionIDA,
			)

			// Update chainA client on chainB
			suite.chainB.updateClient(suite.chainA)

			// Check that Chain B can verify channel is stored on chainA
			proof, proofHeight := suite.queryProof(channelKey)

			err := suite.chainB.App.IBCKeeper.ConnectionKeeper.VerifyChannelState(
				suite.chainB.GetContext(), connection, proofHeight, proof, testPort1,
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
	commitmentKey := ibctypes.KeyPacketCommitment(testPort1, testChannel1, 1)
	commitmentBz := []byte("commitment")

	cases := []struct {
		msg         string
		proofHeight uint64
		malleate    func()
		expPass     bool
	}{
		{"verification success", 2, func() {
			suite.chainB.CreateClient(suite.chainA)
		}, true},
		{"client state not found", 2, func() {}, false},
		{"consensus state not found", 100, func() {
			suite.chainB.CreateClient(suite.chainA)
		}, false},
		{"verification failed", 2, func() {
			suite.chainB.CreateClient(suite.chainA)
		}, false},
	}

	// ChainA sets packet commitment on channel with chainB in its state
	// Check that ChainB can verify the PacketCommitment
	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			// Set PacketCommitment on chainA
			connection := suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, exported.OPEN)
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainA.GetContext(), testPort1, testChannel1, 1, commitmentBz)

			// Update ChainA client on chainB
			suite.chainB.updateClient(suite.chainA)

			// Check that ChainB can verify PacketCommitment stored in chainA
			proof, proofHeight := suite.queryProof(commitmentKey)

			err := suite.chainB.App.IBCKeeper.ConnectionKeeper.VerifyPacketCommitment(
				suite.chainB.GetContext(), connection, proofHeight, proof, testPort1,
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
	packetAckKey := ibctypes.KeyPacketAcknowledgement(testPort1, testChannel1, 1)
	ack := []byte("acknowledgement")

	cases := []struct {
		msg         string
		proofHeight uint64
		malleate    func()
		expPass     bool
	}{
		{"verification success", 2, func() {
			suite.chainB.CreateClient(suite.chainA)
		}, true},
		{"client state not found", 2, func() {}, false},
		{"consensus state not found", 100, func() {
			suite.chainB.CreateClient(suite.chainA)
		}, false},
		{"verification failed", 2, func() {
			suite.chainB.CreateClient(suite.chainB)
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			connection := suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, exported.OPEN)
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(suite.chainA.GetContext(), testPort1, testChannel1, 1, ack)
			suite.chainB.updateClient(suite.chainA)

			// TODO check this proof height
			proof, proofHeight := suite.queryProof(packetAckKey)

			err := suite.chainB.App.IBCKeeper.ConnectionKeeper.VerifyPacketAcknowledgement(
				suite.chainB.GetContext(), connection, proofHeight, proof, testPort1,
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
	packetAckKey := ibctypes.KeyPacketAcknowledgement(testPort1, testChannel1, 1)

	cases := []struct {
		msg         string
		proofHeight uint64
		malleate    func()
		expPass     bool
	}{
		{"verification success", 2, func() {
			suite.chainB.CreateClient(suite.chainA)
		}, true},
		{"client state not found", 2, func() {}, false},
		{"consensus state not found", 100, func() {
			suite.chainB.CreateClient(suite.chainA)
		}, false},
		{"verification failed", 2, func() {
			suite.chainB.CreateClient(suite.chainB)
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			connection := suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, exported.OPEN)
			suite.chainB.updateClient(suite.chainA)

			proof, proofHeight := suite.queryProof(packetAckKey)

			err := suite.chainB.App.IBCKeeper.ConnectionKeeper.VerifyPacketAcknowledgementAbsence(
				suite.chainB.GetContext(), connection, proofHeight, proof, testPort1,
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
	nextSeqRcvKey := ibctypes.KeyNextSequenceRecv(testPort1, testChannel1)

	cases := []struct {
		msg         string
		proofHeight uint64
		malleate    func()
		expPass     bool
	}{
		{"verification success", uint64(2), func() {
			suite.chainB.CreateClient(suite.chainA)
		}, true},
		{"client state not found", uint64(2), func() {}, false},
		{"consensus state not found", uint64(100), func() {
			suite.chainB.CreateClient(suite.chainA)
		}, false},
		{"verification failed", uint64(2), func() {
			suite.chainB.CreateClient(suite.chainB)
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			connection := suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, exported.OPEN)
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.chainA.GetContext(), testPort1, testChannel1, 1)
			suite.chainB.updateClient(suite.chainA)

			proof, proofHeight := suite.queryProof(nextSeqRcvKey)

			err := suite.chainB.App.IBCKeeper.ConnectionKeeper.VerifyNextSequenceRecv(
				suite.chainB.GetContext(), connection, proofHeight, proof, testPort1,
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
