package keeper_test

import (
	"fmt"
	"time"

	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
	ibcmock "github.com/cosmos/cosmos-sdk/x/ibc/testing/mock"
)

var defaultTimeoutHeight = clienttypes.NewHeight(0, 100000)

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

			_, clientB, connA, _ := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)

			counterpartyClient, clientProof := suite.chainB.QueryClientStateProof(clientB)
			proofHeight := clienttypes.NewHeight(0, uint64(suite.chainB.GetContext().BlockHeight()-1))

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
				malleateHeight(proofHeight, tc.heightDiff), clientProof, counterpartyClient,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
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
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
		}, true},
		{"client state not found", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)

			changeClientID = true
		}, false},
		{"consensus state not found", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)

			heightDiff = 5
		}, false},
		{"verification failed", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			clientB := connB.ClientID
			clientState := suite.chainB.GetClientState(clientB)

			// give chainB wrong consensus state for chainA
			consState, found := suite.chainB.App.IBCKeeper.ClientKeeper.GetLatestClientConsensusState(suite.chainB.GetContext(), clientB)
			suite.Require().True(found)

			tmConsState, ok := consState.(*ibctmtypes.ConsensusState)
			suite.Require().True(ok)

			tmConsState.Timestamp = time.Now()
			suite.chainB.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainB.GetContext(), clientB, clientState.GetLatestHeight(), tmConsState)

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
			proofHeight := clienttypes.NewHeight(0, uint64(suite.chainB.GetContext().BlockHeight()-1))
			consensusState, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetSelfConsensusState(suite.chainA.GetContext(), consensusHeight)
			suite.Require().True(found)

			err := suite.chainA.App.IBCKeeper.ConnectionKeeper.VerifyClientConsensusState(
				suite.chainA.GetContext(), connection,
				malleateHeight(proofHeight, heightDiff), consensusHeight, proof, consensusState,
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

			_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)

			connection := suite.chainA.GetConnection(connA)
			if tc.changeClientID {
				connection.ClientId = ibctesting.InvalidID
			}
			expectedConnection := suite.chainB.GetConnection(connB)

			connectionKey := host.ConnectionKey(connB.ID)
			proof, proofHeight := suite.chainB.QueryProof(connectionKey)

			if tc.changeConnectionState {
				expectedConnection.State = types.TRYOPEN
			}

			err := suite.chainA.App.IBCKeeper.ConnectionKeeper.VerifyConnectionState(
				suite.chainA.GetContext(), connection,
				malleateHeight(proofHeight, tc.heightDiff), proof, connB.ID, expectedConnection,
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

			_, _, connA, _, _, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, channeltypes.UNORDERED)
			connection := suite.chainA.GetConnection(connA)
			if tc.changeClientID {
				connection.ClientId = ibctesting.InvalidID
			}

			channelKey := host.ChannelKey(channelB.PortID, channelB.ID)
			proof, proofHeight := suite.chainB.QueryProof(channelKey)

			channel := suite.chainB.GetChannel(channelB)
			if tc.changeChannelState {
				channel.State = channeltypes.TRYOPEN
			}

			err := suite.chainA.App.IBCKeeper.ConnectionKeeper.VerifyChannelState(
				suite.chainA.GetContext(), connection, malleateHeight(proofHeight, tc.heightDiff), proof,
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
		delayPeriod                 uint64
		expPass                     bool
	}{
		{"verification success", false, false, 0, 0, true},
		{"verification success: delay period passed", false, false, 0, uint64(1 * time.Second.Nanoseconds()), true},
		{"delay period has not passed", false, false, 0, uint64(1 * time.Hour.Nanoseconds()), false},
		{"client state not found- changed client ID", true, false, 0, 0, false},
		{"consensus state not found - increased proof height", false, false, 5, 0, false},
		{"verification failed - changed packet commitment state", false, true, 0, 0, false},
	}

	for _, tc := range cases {
		tc := tc

		suite.Run(tc.msg, func() {
			suite.SetupTest() // reset

			_, clientB, _, connB, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, channeltypes.UNORDERED)

			connection := suite.chainB.GetConnection(connB)
			connection.DelayPeriod = tc.delayPeriod
			if tc.changeClientID {
				connection.ClientId = ibctesting.InvalidID
			}

			packet := channeltypes.NewPacket(ibctesting.TestHash, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, defaultTimeoutHeight, 0)
			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			commitmentKey := host.PacketCommitmentKey(packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
			proof, proofHeight := suite.chainA.QueryProof(commitmentKey)

			if tc.changePacketCommitmentState {
				packet.Data = []byte(ibctesting.InvalidID)
			}

			commitment := channeltypes.CommitPacket(suite.chainB.App.IBCKeeper.Codec(), packet)
			err = suite.chainB.App.IBCKeeper.ConnectionKeeper.VerifyPacketCommitment(
				suite.chainB.GetContext(), connection, malleateHeight(proofHeight, tc.heightDiff), proof,
				packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence(), commitment,
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
		delayPeriod           uint64
		expPass               bool
	}{
		{"verification success", false, false, 0, 0, true},
		{"verification success: delay period passed", false, false, 0, uint64(1 * time.Second.Nanoseconds()), true},
		{"delay period has not passed", false, false, 0, uint64(1 * time.Hour.Nanoseconds()), false},
		{"client state not found- changed client ID", true, false, 0, 0, false},
		{"consensus state not found - increased proof height", false, false, 5, 0, false},
		{"verification failed - changed acknowledgement", false, true, 0, 0, false},
	}

	for _, tc := range cases {
		tc := tc

		suite.Run(tc.msg, func() {
			suite.SetupTest() // reset

			clientA, clientB, connA, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, channeltypes.UNORDERED)

			connection := suite.chainA.GetConnection(connA)
			connection.DelayPeriod = tc.delayPeriod
			if tc.changeClientID {
				connection.ClientId = ibctesting.InvalidID
			}

			// send and receive packet
			packet := channeltypes.NewPacket(ibctesting.TestHash, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, defaultTimeoutHeight, 0)
			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			// increment receiving chain's (chainB) time by 2 hour to always pass receive
			suite.coordinator.IncrementTimeBy(time.Hour * 2)
			suite.coordinator.CommitBlock(suite.chainB)

			err = suite.coordinator.RecvPacket(suite.chainA, suite.chainB, clientA, packet)
			suite.Require().NoError(err)

			packetAckKey := host.PacketAcknowledgementKey(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
			proof, proofHeight := suite.chainB.QueryProof(packetAckKey)

			ack := ibcmock.MockAcknowledgement
			if tc.changeAcknowledgement {
				ack = []byte(ibctesting.InvalidID)
			}

			err = suite.chainA.App.IBCKeeper.ConnectionKeeper.VerifyPacketAcknowledgement(
				suite.chainA.GetContext(), connection, malleateHeight(proofHeight, tc.heightDiff), proof,
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

// TestVerifyPacketReceiptAbsence has chainA verify the receipt
// absence on channelB. The channels on chainA and chainB are fully opened and
// a packet is sent from chainA to chainB and not received.
func (suite *KeeperTestSuite) TestVerifyPacketReceiptAbsence() {
	cases := []struct {
		msg            string
		changeClientID bool
		recvAck        bool
		heightDiff     uint64
		delayPeriod    uint64
		expPass        bool
	}{
		{"verification success", false, false, 0, 0, true},
		{"verification success: delay period passed", false, false, 0, uint64(1 * time.Second.Nanoseconds()), true},
		{"delay period has not passed", false, false, 0, uint64(1 * time.Hour.Nanoseconds()), false},
		{"client state not found - changed client ID", true, false, 0, 0, false},
		{"consensus state not found - increased proof height", false, false, 5, 0, false},
		{"verification failed - acknowledgement was received", false, true, 0, 0, false},
	}

	for _, tc := range cases {
		tc := tc

		suite.Run(tc.msg, func() {
			suite.SetupTest() // reset

			clientA, clientB, connA, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, channeltypes.UNORDERED)

			connection := suite.chainA.GetConnection(connA)
			connection.DelayPeriod = tc.delayPeriod
			if tc.changeClientID {
				connection.ClientId = ibctesting.InvalidID
			}

			// send, only receive if specified
			packet := channeltypes.NewPacket(ibctesting.TestHash, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, defaultTimeoutHeight, 0)
			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			if tc.recvAck {
				// increment receiving chain's (chainB) time by 2 hour to always pass receive
				suite.coordinator.IncrementTimeBy(time.Hour * 2)
				suite.coordinator.CommitBlock(suite.chainB)

				err = suite.coordinator.RecvPacket(suite.chainA, suite.chainB, clientA, packet)
				suite.Require().NoError(err)
			} else {
				// need to update height to prove absence
				suite.coordinator.CommitBlock(suite.chainA, suite.chainB)
				suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, exported.Tendermint)
			}

			packetReceiptKey := host.PacketReceiptKey(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
			proof, proofHeight := suite.chainB.QueryProof(packetReceiptKey)

			err = suite.chainA.App.IBCKeeper.ConnectionKeeper.VerifyPacketReceiptAbsence(
				suite.chainA.GetContext(), connection, malleateHeight(proofHeight, tc.heightDiff), proof,
				packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence(),
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
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
		delayPeriod    uint64
		expPass        bool
	}{
		{"verification success", false, 0, 0, 0, true},
		{"verification success: delay period passed", false, 0, 0, uint64(1 * time.Second.Nanoseconds()), true},
		{"delay period has not passed", false, 0, 0, uint64(1 * time.Hour.Nanoseconds()), false},
		{"client state not found- changed client ID", true, 0, 0, 0, false},
		{"consensus state not found - increased proof height", false, 0, 5, 0, false},
		{"verification failed - wrong expected next seq recv", false, 1, 0, 0, false},
	}

	for _, tc := range cases {
		tc := tc

		suite.Run(tc.msg, func() {
			suite.SetupTest() // reset

			clientA, clientB, connA, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, channeltypes.UNORDERED)

			connection := suite.chainA.GetConnection(connA)
			connection.DelayPeriod = tc.delayPeriod
			if tc.changeClientID {
				connection.ClientId = ibctesting.InvalidID
			}

			// send and receive packet
			packet := channeltypes.NewPacket(ibctesting.TestHash, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, defaultTimeoutHeight, 0)
			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			// increment receiving chain's (chainB) time by 2 hour to always pass receive
			suite.coordinator.IncrementTimeBy(time.Hour * 2)
			suite.coordinator.CommitBlock(suite.chainB)

			err = suite.coordinator.RecvPacket(suite.chainA, suite.chainB, clientA, packet)
			suite.Require().NoError(err)

			nextSeqRecvKey := host.NextSequenceRecvKey(packet.GetDestPort(), packet.GetDestChannel())
			proof, proofHeight := suite.chainB.QueryProof(nextSeqRecvKey)

			err = suite.chainA.App.IBCKeeper.ConnectionKeeper.VerifyNextSequenceRecv(
				suite.chainA.GetContext(), connection, malleateHeight(proofHeight, tc.heightDiff), proof,
				packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence()+tc.offsetSeq,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func malleateHeight(height exported.Height, diff uint64) exported.Height {
	return clienttypes.NewHeight(height.GetRevisionNumber(), height.GetRevisionHeight()+diff)
}
