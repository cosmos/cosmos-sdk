package keeper_test

import (
	"fmt"
	"time"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
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
// on chainB. The connections on chainA and chainB are fully opened.
func (suite *KeeperTestSuite) TestVerifyConnectionState() {
	cases := []struct {
		msg                   string
		changeClientID        bool
		changeConnectionState bool
		heightDiff            uint64
		expPass               bool
	}{
		{"verification success", false, false, 0, true},
		{"client state not found - changed client ID", true, false, 0, false},
		{"consensus state not found - increased proof height", false, false, 5, false},
		{"verification failed - connection state is different than proof", false, true, 0, false},
	}

	for _, tc := range cases {
		tc := tc

		suite.Run(tc.msg, func() {
			suite.SetupTest() // reset

			_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)

			connection := suite.chainA.GetConnection(connA)
			if tc.changeClientID {
				connection.ClientID = ibctesting.InvalidID
			}
			expectedConnection := suite.chainB.GetConnection(connB)

			connectionKey := host.KeyConnection(connB.ID)
			proof, proofHeight := suite.chainB.QueryProof(connectionKey)

			if tc.changeConnectionState {
				expectedConnection.State = types.TRYOPEN
			}

			err := suite.chainA.App.IBCKeeper.ConnectionKeeper.VerifyConnectionState(
				suite.chainA.GetContext(), connection,
				proofHeight+tc.heightDiff, proof, connB.ID, expectedConnection,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestVerifyChannelState verifies the channel state of the channel on
// chainB. The channels on chainA and chainB are fully opened.
func (suite *KeeperTestSuite) TestVerifyChannelState() {
	cases := []struct {
		msg                string
		changeClientID     bool
		changeChannelState bool
		heightDiff         uint64
		expPass            bool
	}{
		{"verification success", false, false, 0, true},
		{"client state not found- changed client ID", true, false, 0, false},
		{"consensus state not found - increased proof height", false, false, 5, false},
		{"verification failed - changed channel state", false, true, 0, false},
	}

	for _, tc := range cases {
		tc := tc

		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			_, _, connA, _, _, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			connection := suite.chainA.GetConnection(connA)
			if tc.changeClientID {
				connection.ClientID = ibctesting.InvalidID
			}

			channelKey := host.KeyChannel(channelB.PortID, channelB.ID)
			proof, proofHeight := suite.chainB.QueryProof(channelKey)

			channel := suite.chainB.GetChannel(channelB)
			if tc.changeChannelState {
				channel.State = channeltypes.TRYOPEN
			}

			err := suite.chainA.App.IBCKeeper.ConnectionKeeper.VerifyChannelState(
				suite.chainA.GetContext(), connection, proofHeight+tc.heightDiff, proof,
				channelB.PortID, channelB.ID, channel,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestVerifyPacketCommitmentState has chainB verify the packet commitment
// on channelA. The channels on chainA and chainB are fully opened and a
// packet is sent from chainA to chainB, but has not been received.
func (suite *KeeperTestSuite) TestVerifyPacketCommitment() {
	cases := []struct {
		msg                         string
		changeClientID              bool
		changePacketCommitmentState bool
		heightDiff                  uint64
		expPass                     bool
	}{
		{"verification success", false, false, 0, true},
		{"client state not found- changed client ID", true, false, 0, false},
		{"consensus state not found - increased proof height", false, false, 5, false},
		{"verification failed - changed packet commitment state", false, true, 0, false},
	}

	for _, tc := range cases {
		tc := tc

		suite.Run(tc.msg, func() {
			suite.SetupTest() // reset

			_, clientB, _, connB, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			connection := suite.chainB.GetConnection(connB)
			if tc.changeClientID {
				connection.ClientID = ibctesting.InvalidID
			}

			packet := channeltypes.NewPacket(ibctesting.TestHash, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, 100000, 0)
			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			commitmentKey := host.KeyPacketCommitment(packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
			proof, proofHeight := suite.chainA.QueryProof(commitmentKey)

			if tc.changePacketCommitmentState {
				packet.Data = []byte(ibctesting.InvalidID)
			}

			err = suite.chainB.App.IBCKeeper.ConnectionKeeper.VerifyPacketCommitment(
				suite.chainB.GetContext(), connection, proofHeight+tc.heightDiff, proof,
				packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence(), channeltypes.CommitPacket(packet),
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestVerifyPacketAcknowledgement has chainA verify the acknowledgement on
// channelB. The channels on chainA and chainB are fully opened and a packet
// is sent from chainA to chainB and received.
func (suite *KeeperTestSuite) TestVerifyPacketAcknowledgement() {
	cases := []struct {
		msg                   string
		changeClientID        bool
		changeAcknowledgement bool
		heightDiff            uint64
		expPass               bool
	}{
		{"verification success", false, false, 0, true},
		{"client state not found- changed client ID", true, false, 0, false},
		{"consensus state not found - increased proof height", false, false, 5, false},
		{"verification failed - changed acknowledgement", false, true, 0, false},
	}

	for _, tc := range cases {
		tc := tc

		suite.Run(tc.msg, func() {
			suite.SetupTest() // reset

			clientA, clientB, connA, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			connection := suite.chainA.GetConnection(connA)
			if tc.changeClientID {
				connection.ClientID = ibctesting.InvalidID
			}

			// send and receive packet
			packet := channeltypes.NewPacket(ibctesting.TestHash, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, 100000, 0)
			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			err = suite.coordinator.PacketExecuted(suite.chainB, suite.chainA, packet, clientA)
			suite.Require().NoError(err)

			packetAckKey := host.KeyPacketAcknowledgement(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
			proof, proofHeight := suite.chainB.QueryProof(packetAckKey)

			ack := ibctesting.TestHash
			if tc.changeAcknowledgement {
				ack = []byte(ibctesting.InvalidID)
			}

			err = suite.chainA.App.IBCKeeper.ConnectionKeeper.VerifyPacketAcknowledgement(
				suite.chainA.GetContext(), connection, proofHeight+tc.heightDiff, proof,
				packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence(), ack,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
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
