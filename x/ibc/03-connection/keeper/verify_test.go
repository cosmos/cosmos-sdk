package keeper_test

import (
	"fmt"
	"time"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

// TestVerifyClientConsensusState verifies that the consensus state of
// chainA stored on clientB (which is on chainB) matches the consensus
// state for chainA at that height.
func (suite *KeeperTestSuite) TestVerifyClientConsensusState() {
	var (
		connA          *ibctesting.TestConnection
		connB          *ibctesting.TestConnection
		changeClientID bool
		heightDiff     uint64
	)
	cases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"verification success", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
		}, true},
		{"client state not found", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)

			changeClientID = true
		}, false},
		{"consensus state not found", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)

			heightDiff = 5
		}, false},
		{"verification failed", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			clientB := connB.ClientID

			// give chainB wrong consensus state for chainA
			consState, found := suite.chainB.App.IBCKeeper.ClientKeeper.GetLatestClientConsensusState(suite.chainB.GetContext(), clientB)
			suite.Require().True(found)

			tmConsState, ok := consState.(ibctmtypes.ConsensusState)
			suite.Require().True(ok)

			tmConsState.Timestamp = time.Now()
			suite.chainB.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainB.GetContext(), clientB, tmConsState.Height, tmConsState)

			suite.coordinator.CommitBlock(suite.chainB)
		}, false},
	}

	for _, tc := range cases {
		tc := tc

		suite.Run(tc.msg, func() {
			suite.SetupTest()      // reset
			heightDiff = 0         // must be explicitly changed in malleate
			changeClientID = false // must be explicitly changed in malleate

			tc.malleate()

			connection := suite.chainA.GetConnection(connA)
			if changeClientID {
				connection.ClientID = ibctesting.InvalidID
			}

			proof, consensusHeight := suite.chainB.QueryConsensusStateProof(connB.ClientID)
			proofHeight := uint64(suite.chainA.GetContext().BlockHeight() - 1)
			consensusState, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetSelfConsensusState(suite.chainA.GetContext(), consensusHeight)
			suite.Require().True(found)

			err := suite.chainA.App.IBCKeeper.ConnectionKeeper.VerifyClientConsensusState(
				suite.chainA.GetContext(), connection,
				proofHeight+heightDiff, consensusHeight, proof, consensusState,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestVerifyConnectionState verifies the connection state of the connection
// on chainB.
func (suite *KeeperTestSuite) TestVerifyConnectionState() {
	var (
		connA                 *ibctesting.TestConnection
		connB                 *ibctesting.TestConnection
		changeClientID        bool
		changeConnectionState bool
		heightDiff            uint64
	)
	cases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"verification success", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
		}, true},
		{"client state not found", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)

			changeClientID = true
		}, false},
		{"consensus state not found", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)

			heightDiff = 5
		}, false},
		{"verification failed", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)

			changeConnectionState = true
		}, false},
	}

	for _, tc := range cases {
		tc := tc

		suite.Run(tc.msg, func() {
			suite.SetupTest()             // reset
			changeClientID = false        // must be explicitly changed in malleate
			changeConnectionState = false // must be explicitly changed in malleate
			heightDiff = 0                // must be explicitly changed in malleate

			tc.malleate()

			connection := suite.chainA.GetConnection(connA)
			if changeClientID {
				connection.ClientID = ibctesting.InvalidID
			}
			expectedConnection := suite.chainB.GetConnection(connB)

			connectionKey := host.KeyConnection(connB.ID)
			proof, proofHeight := suite.chainB.QueryProof(connectionKey)

			if changeConnectionState {
				expectedConnection.State = types.TRYOPEN
			}

			err := suite.chainA.App.IBCKeeper.ConnectionKeeper.VerifyConnectionState(
				suite.chainA.GetContext(), connection,
				proofHeight+heightDiff, proof, connB.ID, expectedConnection,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
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
