package keeper_test

import (
	"fmt"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

const (
	testPort1 = "firstport"
	testPort2 = "secondport"

	testChannel1 = "firstchannel"
	testChannel2 = "secondchannel"
)

func (suite *KeeperTestSuite) TestVerifyClientConsensusState() {
	// create connection on chainA to chainB
	counterparty := types.NewCounterparty(
		testClientIDA, testConnectionIDA,
		commitmenttypes.NewMerklePrefix(suite.oldchainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes()),
	)
	connection1 := types.NewConnectionEnd(
		types.UNINITIALIZED, testConnectionIDB, testClientIDB, counterparty,
		types.GetCompatibleVersions(),
	)

	cases := []struct {
		msg        string
		connection types.ConnectionEnd
		malleate   func() clientexported.ConsensusState
		expPass    bool
	}{
		{"verification success", connection1, func() clientexported.ConsensusState {
			suite.oldchainA.CreateClient(suite.oldchainB)
			suite.oldchainB.CreateClient(suite.oldchainA)
			consState := suite.oldchainA.Header.ConsensusState()
			return consState
		}, true},
		{"client state not found", connection1, func() clientexported.ConsensusState {
			return suite.oldchainB.Header.ConsensusState()
		}, false},
		{"verification failed", connection1, func() clientexported.ConsensusState {
			suite.oldchainA.CreateClient(suite.oldchainA)
			return suite.oldchainA.Header.ConsensusState()
		}, false},
	}

	// Create Client of chain B on Chain App
	// Check that we can verify B's consensus state on chain A
	for i, tc := range cases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			consState := tc.malleate()

			// perform a couple updates of chain B on chain A
			suite.oldchainA.updateClient(suite.oldchainB)
			suite.oldchainA.updateClient(suite.oldchainB)

			// TODO: is this the right consensus height
			consensusHeight := suite.oldchainA.Header.GetHeight()
			consensusKey := prefixedClientKey(testClientIDA, host.KeyConsensusState(consensusHeight))

			// get proof that chainB stored chainA' consensus state
			proof, proofHeight := queryProof(suite.oldchainB, consensusKey)

			err := suite.oldchainA.App.IBCKeeper.ConnectionKeeper.VerifyClientConsensusState(
				suite.oldchainA.GetContext(), tc.connection, proofHeight+1, consensusHeight, proof, consState,
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
	connectionKey := host.KeyConnection(testConnectionIDA)
	var invalidProofHeight uint64
	cases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"verification success", func() {
			suite.oldchainA.CreateClient(suite.oldchainB)
			suite.oldchainB.CreateClient(suite.oldchainA)
			invalidProofHeight = 0 // don't use this
		}, true},
		{"client state not found", func() {}, false},
		{"verification failed", func() {
			suite.oldchainA.CreateClient(suite.oldchainB)
			suite.oldchainB.CreateClient(suite.oldchainA)
			invalidProofHeight = 10 // make proofHeight incorrect
		}, false},
	}

	// Chains A and B create clients for each other
	// A creates connectionEnd for chain B and stores it in state
	// Check that B can verify connection is stored after some updates
	for i, tc := range cases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			// create and store connection on chain A
			expectedConnection := suite.oldchainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.OPEN)

			// // create expected connection
			// TODO: why is this commented
			// expectedConnection := types.NewConnectionEnd(types.INIT, testClientIDB, counterparty, []string{"1.0.0"})

			// perform a couple updates of chain A on chain B
			suite.oldchainB.updateClient(suite.oldchainA)
			suite.oldchainB.updateClient(suite.oldchainA)
			proof, proofHeight := queryProof(suite.oldchainA, connectionKey)
			// if invalidProofHeight has been set, use that value instead
			if invalidProofHeight != 0 {
				proofHeight = invalidProofHeight
			}

			// Create B's connection to A
			counterparty := types.NewCounterparty(testClientIDB, testConnectionIDB, commitmenttypes.NewMerklePrefix([]byte("ibc")))
			connection := types.NewConnectionEnd(types.UNINITIALIZED, testConnectionIDA, testClientIDA, counterparty, []string{"1.0.0"})
			// Ensure chain B can verify connection exists in chain A
			err := suite.oldchainB.App.IBCKeeper.ConnectionKeeper.VerifyConnectionState(
				suite.oldchainB.GetContext(), connection, proofHeight+1, proof, testConnectionIDA, expectedConnection,
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
	channelKey := host.KeyChannel(testPort1, testChannel1)

	// create connection of chainB to pass into verify function
	counterparty := types.NewCounterparty(
		testClientIDB, testConnectionIDB,
		commitmenttypes.NewMerklePrefix(suite.oldchainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes()),
	)

	connection := types.NewConnectionEnd(
		types.UNINITIALIZED, testConnectionIDA, testClientIDA, counterparty,
		types.GetCompatibleVersions(),
	)

	cases := []struct {
		msg         string
		proofHeight uint64
		malleate    func()
		expPass     bool
	}{
		{"verification success", 0, func() {
			suite.oldchainB.CreateClient(suite.oldchainA)
		}, true},
		{"client state not found", 0, func() {}, false},
		{"consensus state not found", 100, func() {
			suite.oldchainB.CreateClient(suite.oldchainA)
		}, false},
		{"verification failed", 7, func() {
			suite.oldchainB.CreateClient(suite.oldchainB)
		}, false},
	}

	// Chain A creates channel for chain B and stores in its state
	// Check that chainB can verify channel is stored in chain A
	for i, tc := range cases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			// Create and store channel on chain A
			channel := suite.oldchainA.createChannel(
				testPort1, testChannel1, testPort2, testChannel2,
				channeltypes.OPEN, channeltypes.ORDERED, testConnectionIDA,
			)

			// Update chainA client on chainB
			suite.oldchainB.updateClient(suite.oldchainA)

			// Check that Chain B can verify channel is stored on chainA
			proof, proofHeight := queryProof(suite.oldchainA, channelKey)
			// if testcase proofHeight is not 0, replace proofHeight with this value
			if tc.proofHeight != 0 {
				proofHeight = tc.proofHeight
			}

			err := suite.oldchainB.App.IBCKeeper.ConnectionKeeper.VerifyChannelState(
				suite.oldchainB.GetContext(), connection, proofHeight+1, proof, testPort1,
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
	commitmentKey := host.KeyPacketCommitment(testPort1, testChannel1, 1)
	commitmentBz := []byte("commitment")

	cases := []struct {
		msg         string
		proofHeight uint64
		malleate    func()
		expPass     bool
	}{
		{"verification success", 0, func() {
			suite.oldchainB.CreateClient(suite.oldchainA)
		}, true},
		{"client state not found", 0, func() {}, false},
		{"consensus state not found", 100, func() {
			suite.oldchainB.CreateClient(suite.oldchainA)
		}, false},
	}

	// ChainA sets packet commitment on channel with chainB in its state
	// Check that ChainB can verify the PacketCommitment
	for i, tc := range cases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			// Set PacketCommitment on chainA
			connection := suite.oldchainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, types.OPEN)
			suite.oldchainA.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.oldchainA.GetContext(), testPort1, testChannel1, 1, commitmentBz)

			// Update ChainA client on chainB
			suite.oldchainB.updateClient(suite.oldchainA)

			// Check that ChainB can verify PacketCommitment stored in chainA
			proof, proofHeight := queryProof(suite.oldchainA, commitmentKey)
			// if testcase proofHeight is not 0, replace proofHeight with this value
			if tc.proofHeight != 0 {
				proofHeight = tc.proofHeight
			}

			err := suite.oldchainB.App.IBCKeeper.ConnectionKeeper.VerifyPacketCommitment(
				suite.oldchainB.GetContext(), connection, proofHeight+1, proof, testPort1,
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
	packetAckKey := host.KeyPacketAcknowledgement(testPort1, testChannel1, 1)
	ack := []byte("acknowledgement")

	cases := []struct {
		msg         string
		proofHeight uint64
		malleate    func()
		expPass     bool
	}{
		{"verification success", 0, func() {
			suite.oldchainB.CreateClient(suite.oldchainA)
		}, true},
		{"client state not found", 0, func() {}, false},
		{"consensus state not found", 100, func() {
			suite.oldchainB.CreateClient(suite.oldchainA)
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			connection := suite.oldchainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, types.OPEN)
			suite.oldchainA.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(suite.oldchainA.GetContext(), testPort1, testChannel1, 1, channeltypes.CommitAcknowledgement(ack))
			suite.oldchainB.updateClient(suite.oldchainA)

			// TODO check this proof height
			proof, proofHeight := queryProof(suite.oldchainA, packetAckKey)
			// if testcase proofHeight is not 0, replace proofHeight with this value
			if tc.proofHeight != 0 {
				proofHeight = tc.proofHeight
			}

			err := suite.oldchainB.App.IBCKeeper.ConnectionKeeper.VerifyPacketAcknowledgement(
				suite.oldchainB.GetContext(), connection, proofHeight+1, proof, testPort1,
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
	packetAckKey := host.KeyPacketAcknowledgement(testPort1, testChannel1, 1)

	cases := []struct {
		msg         string
		proofHeight uint64
		malleate    func()
		expPass     bool
	}{
		{"verification success", 0, func() {
			suite.oldchainB.CreateClient(suite.oldchainA)
		}, true},
		{"client state not found", 0, func() {}, false},
		{"consensus state not found", 100, func() {
			suite.oldchainB.CreateClient(suite.oldchainA)
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			connection := suite.oldchainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, types.OPEN)
			suite.oldchainB.updateClient(suite.oldchainA)

			proof, proofHeight := queryProof(suite.oldchainA, packetAckKey)
			// if testcase proofHeight is not 0, replace proofHeight with this value
			if tc.proofHeight != 0 {
				proofHeight = tc.proofHeight
			}

			err := suite.oldchainB.App.IBCKeeper.ConnectionKeeper.VerifyPacketAcknowledgementAbsence(
				suite.oldchainB.GetContext(), connection, proofHeight+1, proof, testPort1,
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
	nextSeqRcvKey := host.KeyNextSequenceRecv(testPort1, testChannel1)

	cases := []struct {
		msg         string
		proofHeight uint64
		malleate    func()
		expPass     bool
	}{
		{"verification success", uint64(0), func() {
			suite.oldchainB.CreateClient(suite.oldchainA)
		}, true},
		{"client state not found", uint64(0), func() {}, false},
		{"consensus state not found", uint64(100), func() {
			suite.oldchainB.CreateClient(suite.oldchainA)
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			connection := suite.oldchainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, types.OPEN)
			suite.oldchainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.oldchainA.GetContext(), testPort1, testChannel1, 1)
			suite.oldchainB.updateClient(suite.oldchainA)

			proof, proofHeight := queryProof(suite.oldchainA, nextSeqRcvKey)
			// if testcase proofHeight is not 0, replace proofHeight with this value
			if tc.proofHeight != 0 {
				proofHeight = tc.proofHeight
			}

			err := suite.oldchainB.App.IBCKeeper.ConnectionKeeper.VerifyNextSequenceRecv(
				suite.oldchainB.GetContext(), connection, proofHeight+1, proof, testPort1,
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
