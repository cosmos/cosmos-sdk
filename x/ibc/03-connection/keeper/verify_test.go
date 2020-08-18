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

// TestVerifyClientState verifies a client state of chainA
// stored on clientB (which is on chainB)
func (suite *KeeperTestSuite) TestVerifyClientState() {
	cases := []struct {
		msg                  string
		changeClientID       bool
		heightDiff           uint64
		malleateCounterparty bool
		expPass              bool
	}{
		{"verification success", false, 0, false, true},
		{"client state not found", true, 0, false, false},
		{"consensus state for proof height not found", false, 5, false, false},
		{"verification failed", false, 0, true, false},
	}

	for _, tc := range cases {
		tc := tc

		suite.Run(tc.msg, func() {
			suite.SetupTest() // reset

			_, clientB, connA, _ := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)

			counterpartyClient, clientProof := suite.chainB.QueryClientStateProof(clientB)
			proofHeight := uint64(suite.chainB.GetContext().BlockHeight() - 1)

			if tc.malleateCounterparty {
				tmClient, _ := counterpartyClient.(*ibctmtypes.ClientState)
				tmClient.ChainId = "wrongChainID"
			}

			connection := suite.chainA.GetConnection(connA)
			if tc.changeClientID {
				connection.ClientId = ibctesting.InvalidID
			}

			err := suite.chainA.App.IBCKeeper.ConnectionKeeper.VerifyClientState(
				suite.chainA.GetContext(), connection,
				proofHeight+tc.heightDiff, clientProof, counterpartyClient,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case: %s failed, error: %v", tc.msg, err)
			} else {
				suite.Require().Error(err, "invalid test case: %s passed", tc.msg)
			}
		})
	}
}

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

			tmConsState, ok := consState.(*ibctmtypes.ConsensusState)
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
				connection.ClientId = ibctesting.InvalidID
			}

			proof, consensusHeight := suite.chainB.QueryConsensusStateProof(connB.ClientID)
			proofHeight := uint64(suite.chainB.GetContext().BlockHeight() - 1)
			consensusState, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetSelfConsensusState(suite.chainA.GetContext(), consensusHeight)
			suite.Require().True(found)

			err := suite.chainA.App.IBCKeeper.ConnectionKeeper.VerifyClientConsensusState(
				suite.chainA.GetContext(), connection,
				proofHeight+heightDiff, consensusHeight, proof, consensusState,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case: %s failed, error: %v", tc.msg, err)
			} else {
				suite.Require().Error(err, "invalid test case: %s passed", tc.msg)
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
				connection.ClientId = ibctesting.InvalidID
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
				suite.Require().NoError(err, "valid test case: %s failed, error: %v", tc.msg, err)
			} else {
				suite.Require().Error(err, "invalid test case: %s passed", tc.msg)
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
				connection.ClientId = ibctesting.InvalidID
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
				suite.Require().NoError(err, "valid test case: %s failed, error: %v", tc.msg, err)
			} else {
				suite.Require().Error(err, "invalid test case: %s passed", tc.msg)
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
				connection.ClientId = ibctesting.InvalidID
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
				suite.Require().NoError(err, "valid test case: %s failed, error: %v", tc.msg, err)
			} else {
				suite.Require().Error(err, "invalid test case: %s passed", tc.msg)
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
				connection.ClientId = ibctesting.InvalidID
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
				suite.Require().NoError(err, "valid test case: %s failed, error: %v", tc.msg, err)
			} else {
				suite.Require().Error(err, "invalid test case: %s passed", tc.msg)
			}
		})
	}
}

// TestVerifyPacketAcknowledgementAbsence has chainA verify the acknowledgement
// absence on channelB. The channels on chainA and chainB are fully opened and
// a packet is sent from chainA to chainB and not received.
func (suite *KeeperTestSuite) TestVerifyPacketAcknowledgementAbsence() {
	cases := []struct {
		msg            string
		changeClientID bool
		recvAck        bool
		heightDiff     uint64
		expPass        bool
	}{
		{"verification success", false, false, 0, true},
		{"client state not found - changed client ID", true, false, 0, false},
		{"consensus state not found - increased proof height", false, false, 5, false},
		{"verification failed - acknowledgement was received", false, true, 0, false},
	}

	for _, tc := range cases {
		tc := tc

		suite.Run(tc.msg, func() {
			suite.SetupTest() // reset

			clientA, clientB, connA, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			connection := suite.chainA.GetConnection(connA)
			if tc.changeClientID {
				connection.ClientId = ibctesting.InvalidID
			}

			// send, only receive if specified
			packet := channeltypes.NewPacket(ibctesting.TestHash, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, 100000, 0)
			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			if tc.recvAck {
				err = suite.coordinator.PacketExecuted(suite.chainB, suite.chainA, packet, clientA)
				suite.Require().NoError(err)
			} else {
				// need to update height to prove absence
				suite.coordinator.CommitBlock(suite.chainA, suite.chainB)
				suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, clientexported.Tendermint)
			}

			packetAckKey := host.KeyPacketAcknowledgement(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
			proof, proofHeight := suite.chainB.QueryProof(packetAckKey)

			err = suite.chainA.App.IBCKeeper.ConnectionKeeper.VerifyPacketAcknowledgementAbsence(
				suite.chainA.GetContext(), connection, proofHeight+tc.heightDiff, proof,
				packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence(),
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case: %s failed, error: %v", tc.msg, err)
			} else {
				suite.Require().Error(err, "invalid test case: %s passed", tc.msg)
			}
		})
	}
}

// TestVerifyNextSequenceRecv has chainA verify the next sequence receive on
// channelB. The channels on chainA and chainB are fully opened and a packet
// is sent from chainA to chainB and received.
func (suite *KeeperTestSuite) TestVerifyNextSequenceRecv() {
	cases := []struct {
		msg            string
		changeClientID bool
		offsetSeq      uint64
		heightDiff     uint64
		expPass        bool
	}{
		{"verification success", false, 0, 0, true},
		{"client state not found- changed client ID", true, 0, 0, false},
		{"consensus state not found - increased proof height", false, 0, 5, false},
		{"verification failed - wrong expected next seq recv", false, 1, 0, false},
	}

	for _, tc := range cases {
		tc := tc

		suite.Run(tc.msg, func() {
			suite.SetupTest() // reset

			clientA, clientB, connA, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			connection := suite.chainA.GetConnection(connA)
			if tc.changeClientID {
				connection.ClientId = ibctesting.InvalidID
			}

			// send and receive packet
			packet := channeltypes.NewPacket(ibctesting.TestHash, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, 100000, 0)
			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			err = suite.coordinator.PacketExecuted(suite.chainB, suite.chainA, packet, clientA)
			suite.Require().NoError(err)

			nextSeqRecvKey := host.KeyNextSequenceRecv(packet.GetDestPort(), packet.GetDestChannel())
			proof, proofHeight := suite.chainB.QueryProof(nextSeqRecvKey)

			err = suite.chainA.App.IBCKeeper.ConnectionKeeper.VerifyNextSequenceRecv(
				suite.chainA.GetContext(), connection, proofHeight+tc.heightDiff, proof,
				packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence()+tc.offsetSeq,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case: %s failed, error: %v", tc.msg, err)
			} else {
				suite.Require().Error(err, "invalid test case: %s passed", tc.msg)
			}
		})
	}
}
