package keeper_test

import (
	"fmt"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
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
	prefix := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(commitmentexported.Merkle)
	suite.Require().NotNil(prefix)
	counterparty, err := types.NewCounterparty(testClientIDA, testConnectionIDA, prefix)
	suite.Require().NoError(err)

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
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
			consState := suite.chainA.Header.ConsensusState()
			return consState
		}, true},
		{"client state not found", connection1, func() clientexported.ConsensusState {
			return suite.chainB.Header.ConsensusState()
		}, false},
		{"verification failed", connection1, func() clientexported.ConsensusState {
			suite.chainA.CreateClient(suite.chainA)
			return suite.chainA.Header.ConsensusState()
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
			suite.chainA.updateClient(suite.chainB)
			suite.chainA.updateClient(suite.chainB)

			// TODO: is this the right consensus height
			consensusHeight := suite.chainA.Header.GetHeight()
			consensusKey := prefixedClientKey(testClientIDA, host.KeyConsensusState(consensusHeight))

			// get proof that chainB stored chainA' consensus state
			proof, proofHeight := queryProof(suite.chainB, consensusKey)

			err := suite.chainA.App.IBCKeeper.ConnectionKeeper.VerifyClientConsensusState(
				suite.chainA.GetContext(), tc.connection, proofHeight+1, consensusHeight, proof, consState,
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
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
			invalidProofHeight = 0 // don't use this
		}, true},
		{"client state not found", func() {}, false},
		{"verification failed", func() {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
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
			expectedConnection := suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.OPEN)

			// // create expected connection
			// TODO: why is this commented
			// expectedConnection := types.NewConnectionEnd(types.INIT, testClientIDB, counterparty, []string{"1.0.0"})

			// perform a couple updates of chain A on chain B
			suite.chainB.updateClient(suite.chainA)
			suite.chainB.updateClient(suite.chainA)
			proof, proofHeight := queryProof(suite.chainA, connectionKey)
			// if invalidProofHeight has been set, use that value instead
			if invalidProofHeight != 0 {
				proofHeight = invalidProofHeight
			}

			// Create B's connection to A
			prefix := suite.chainB.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(commitmentexported.Merkle)
			suite.Require().NotNil(prefix)

			counterparty, err := types.NewCounterparty(testClientIDB, testConnectionIDB, prefix)
			suite.Require().NoError(err)

			connection := types.NewConnectionEnd(types.UNINITIALIZED, testConnectionIDA, testClientIDA, counterparty, []string{"1.0.0"})
			// Ensure chain B can verify connection exists in chain A
			err = suite.chainB.App.IBCKeeper.ConnectionKeeper.VerifyConnectionState(
				suite.chainB.GetContext(), connection, proofHeight+1, proof, testConnectionIDA, expectedConnection,
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
	prefix := suite.chainB.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(commitmentexported.Merkle)
	suite.Require().NotNil(prefix)

	counterparty, err := types.NewCounterparty(testClientIDB, testConnectionIDB, prefix)
	suite.Require().NoError(err)

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
			suite.chainB.CreateClient(suite.chainA)
		}, true},
		{"client state not found", 0, func() {}, false},
		{"consensus state not found", 100, func() {
			suite.chainB.CreateClient(suite.chainA)
		}, false},
		{"verification failed", 7, func() {
			suite.chainB.CreateClient(suite.chainB)
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
			channel := suite.chainA.createChannel(
				testPort1, testChannel1, testPort2, testChannel2,
				channeltypes.OPEN, channeltypes.ORDERED, testConnectionIDA,
			)

			// Update chainA client on chainB
			suite.chainB.updateClient(suite.chainA)

			// Check that Chain B can verify channel is stored on chainA
			proof, proofHeight := queryProof(suite.chainA, channelKey)
			// if testcase proofHeight is not 0, replace proofHeight with this value
			if tc.proofHeight != 0 {
				proofHeight = tc.proofHeight
			}

			err := suite.chainB.App.IBCKeeper.ConnectionKeeper.VerifyChannelState(
				suite.chainB.GetContext(), connection, proofHeight+1, proof, testPort1,
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
			suite.chainB.CreateClient(suite.chainA)
		}, true},
		{"client state not found", 0, func() {}, false},
		{"consensus state not found", 100, func() {
			suite.chainB.CreateClient(suite.chainA)
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
			connection := suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, types.OPEN)
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainA.GetContext(), testPort1, testChannel1, 1, commitmentBz)

			// Update ChainA client on chainB
			suite.chainB.updateClient(suite.chainA)

			// Check that ChainB can verify PacketCommitment stored in chainA
			proof, proofHeight := queryProof(suite.chainA, commitmentKey)
			// if testcase proofHeight is not 0, replace proofHeight with this value
			if tc.proofHeight != 0 {
				proofHeight = tc.proofHeight
			}

			err := suite.chainB.App.IBCKeeper.ConnectionKeeper.VerifyPacketCommitment(
				suite.chainB.GetContext(), connection, proofHeight+1, proof, testPort1,
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
			suite.chainB.CreateClient(suite.chainA)
		}, true},
		{"client state not found", 0, func() {}, false},
		{"consensus state not found", 100, func() {
			suite.chainB.CreateClient(suite.chainA)
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			connection := suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, types.OPEN)
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(suite.chainA.GetContext(), testPort1, testChannel1, 1, channeltypes.CommitAcknowledgement(ack))
			suite.chainB.updateClient(suite.chainA)

			// TODO check this proof height
			proof, proofHeight := queryProof(suite.chainA, packetAckKey)
			// if testcase proofHeight is not 0, replace proofHeight with this value
			if tc.proofHeight != 0 {
				proofHeight = tc.proofHeight
			}

			err := suite.chainB.App.IBCKeeper.ConnectionKeeper.VerifyPacketAcknowledgement(
				suite.chainB.GetContext(), connection, proofHeight+1, proof, testPort1,
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
			suite.chainB.CreateClient(suite.chainA)
		}, true},
		{"client state not found", 0, func() {}, false},
		{"consensus state not found", 100, func() {
			suite.chainB.CreateClient(suite.chainA)
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			connection := suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, types.OPEN)
			suite.chainB.updateClient(suite.chainA)

			proof, proofHeight := queryProof(suite.chainA, packetAckKey)
			// if testcase proofHeight is not 0, replace proofHeight with this value
			if tc.proofHeight != 0 {
				proofHeight = tc.proofHeight
			}

			err := suite.chainB.App.IBCKeeper.ConnectionKeeper.VerifyPacketAcknowledgementAbsence(
				suite.chainB.GetContext(), connection, proofHeight+1, proof, testPort1,
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
			suite.chainB.CreateClient(suite.chainA)
		}, true},
		{"client state not found", uint64(0), func() {}, false},
		{"consensus state not found", uint64(100), func() {
			suite.chainB.CreateClient(suite.chainA)
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			connection := suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, types.OPEN)
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.chainA.GetContext(), testPort1, testChannel1, 1)
			suite.chainB.updateClient(suite.chainA)

			proof, proofHeight := queryProof(suite.chainA, nextSeqRcvKey)
			// if testcase proofHeight is not 0, replace proofHeight with this value
			if tc.proofHeight != 0 {
				proofHeight = tc.proofHeight
			}

			err := suite.chainB.App.IBCKeeper.ConnectionKeeper.VerifyNextSequenceRecv(
				suite.chainB.GetContext(), connection, proofHeight+1, proof, testPort1,
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
