package ibc_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/x/ibc"
	ibctransfertypes "github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

var (
	timeoutHeight = uint64(10000)
)

type HandlerTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
}

// SetupTest creates a coordinator with 2 test chains.
func (suite *HandlerTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

// tests the IBC handler receiving a packet on ordered and unordered channels.
// It verifies that the storing of an acknowledgement on success occurs. It
// tests high level properties like ordering and basic sanity checks. More
// rigorous testing of 'RecvPacket' and 'PacketExecuted' can be found in the
// 04-channel/keeper/packet_test.go.
func (suite *HandlerTestSuite) TestHandleRecvPacket() {
	var (
		packet channeltypes.Packet
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{"success: ORDERED", func() {
			_, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, channeltypes.ORDERED)
			packet = channeltypes.NewPacket(suite.chainA.GetPacketData(suite.chainB), 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)
		}, true},
		{"success: UNORDERED", func() {
			_, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			packet = channeltypes.NewPacket(suite.chainA.GetPacketData(suite.chainB), 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)
		}, true},
		{"success: UNORDERED out of order packet", func() {
			// setup uses an UNORDERED channel
			_, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)

			// attempts to receive packet with sequence 10 without receiving packet with sequence 1
			for i := uint64(1); i < 10; i++ {
				packet = channeltypes.NewPacket(suite.chainA.GetPacketData(suite.chainB), i, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

				err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
				suite.Require().NoError(err)
			}
		}, true},
		{"failure: ORDERED out of order packet", func() {
			_, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, channeltypes.ORDERED)

			// attempts to receive packet with sequence 10 without receiving packet with sequence 1
			for i := uint64(1); i < 10; i++ {
				packet = channeltypes.NewPacket(suite.chainA.GetPacketData(suite.chainB), i, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

				err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
				suite.Require().NoError(err)
			}
		}, false},
		{"channel does not exist", func() {
			// any non-nil value of packet is valid
			suite.Require().NotNil(packet)
		}, false},
		{"packet not sent", func() {
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			packet = channeltypes.NewPacket(suite.chainA.GetPacketData(suite.chainB), 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)
		}, false},
		{"ORDERED: packet already received (replay)", func() {
			clientA, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, channeltypes.ORDERED)
			packet = channeltypes.NewPacket(suite.chainA.GetPacketData(suite.chainB), 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			err = suite.coordinator.PacketExecuted(suite.chainB, suite.chainA, packet, clientA)
			suite.Require().NoError(err)
		}, false},
		{"UNORDERED: packet already received (replay)", func() {
			clientA, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)

			packet = channeltypes.NewPacket(suite.chainA.GetPacketData(suite.chainB), 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			err = suite.coordinator.PacketExecuted(suite.chainB, suite.chainA, packet, clientA)
			suite.Require().NoError(err)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			handler := ibc.NewHandler(*suite.chainB.App.IBCKeeper)

			tc.malleate()

			// get proof of packet commitment from chainA
			packetKey := host.KeyPacketCommitment(packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
			proof, proofHeight := suite.chainA.QueryProof(packetKey)

			msg := channeltypes.NewMsgRecvPacket(packet, proof, proofHeight, suite.chainB.SenderAccount.GetAddress())

			// ante-handle RecvPacket
			_, err := handler(suite.chainB.GetContext(), msg)

			if tc.expPass {
				suite.Require().NoError(err)

				// replay should fail since state changes occur
				_, err := handler(suite.chainB.GetContext(), msg)
				suite.Require().Error(err)

				// verify ack was written
				ack, found := suite.chainB.App.IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(suite.chainB.GetContext(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
				suite.Require().NotNil(ack)
				suite.Require().True(found)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// tests the IBC handler acknowledgement of a packet on ordered and unordered
// channels. It verifies that the deletion of packet commitments from state
// occurs. It test high level properties like ordering and basic sanity
// checks. More rigorous testing of 'AcknowledgePacket' and 'AcknowledgementExecuted'
// can be found in the 04-channel/keeper/packet_test.go.
func (suite *HandlerTestSuite) TestHandleAcknowledgePacket() {
	var (
		packet channeltypes.Packet
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{"success: ORDERED", func() {
			clientA, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, channeltypes.ORDERED)
			packet = channeltypes.NewPacket(suite.chainA.GetPacketData(suite.chainB), 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			err = suite.coordinator.PacketExecuted(suite.chainB, suite.chainA, packet, clientA)
			suite.Require().NoError(err)
		}, true},
		{"success: UNORDERED", func() {
			clientA, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			packet = channeltypes.NewPacket(suite.chainA.GetPacketData(suite.chainB), 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			err = suite.coordinator.PacketExecuted(suite.chainB, suite.chainA, packet, clientA)
			suite.Require().NoError(err)
		}, true},
		{"success: UNORDERED acknowledge out of order packet", func() {
			// setup uses an UNORDERED channel
			clientA, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)

			// attempts to acknowledge ack with sequence 10 without acknowledging ack with sequence 1 (removing packet commitment)
			for i := uint64(1); i < 10; i++ {
				packet = channeltypes.NewPacket(suite.chainA.GetPacketData(suite.chainB), i, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

				err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
				suite.Require().NoError(err)

				err = suite.coordinator.PacketExecuted(suite.chainB, suite.chainA, packet, clientA)
				suite.Require().NoError(err)

			}
		}, true},
		{"failure: ORDERED acknowledge out of order packet", func() {
			clientA, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, channeltypes.ORDERED)

			// attempts to acknowledge ack with sequence 10 without acknowledging ack with sequence 1 (removing packet commitment
			for i := uint64(1); i < 10; i++ {
				packet = channeltypes.NewPacket(suite.chainA.GetPacketData(suite.chainB), i, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

				err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
				suite.Require().NoError(err)

				err = suite.coordinator.PacketExecuted(suite.chainB, suite.chainA, packet, clientA)
				suite.Require().NoError(err)
			}
		}, false},
		{"channel does not exist", func() {
			// any non-nil value of packet is valid
			suite.Require().NotNil(packet)
		}, false},
		{"packet not received", func() {
			_, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			packet = channeltypes.NewPacket(suite.chainA.GetPacketData(suite.chainB), 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)
		}, false},
		{"ORDERED: packet already acknowledged (replay)", func() {
			clientA, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, channeltypes.ORDERED)
			packet = channeltypes.NewPacket(suite.chainA.GetPacketData(suite.chainB), 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			err = suite.coordinator.PacketExecuted(suite.chainB, suite.chainA, packet, clientA)
			suite.Require().NoError(err)

			err = suite.coordinator.AcknowledgementExecuted(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)
		}, false},
		{"UNORDERED: packet already received (replay)", func() {
			clientA, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)

			packet = channeltypes.NewPacket(suite.chainA.GetPacketData(suite.chainB), 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			err = suite.coordinator.PacketExecuted(suite.chainB, suite.chainA, packet, clientA)
			suite.Require().NoError(err)

			err = suite.coordinator.AcknowledgementExecuted(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			ibctesting.TestHash = ibctransfertypes.FungibleTokenPacketAcknowledgement{true, ""}.GetBytes()

			handler := ibc.NewHandler(*suite.chainA.App.IBCKeeper)

			tc.malleate()

			packetKey := host.KeyPacketAcknowledgement(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
			proof, proofHeight := suite.chainB.QueryProof(packetKey)

			ack := ibctesting.TestHash

			msg := channeltypes.NewMsgAcknowledgement(packet, ack, proof, proofHeight, suite.chainA.SenderAccount.GetAddress())

			_, err := handler(suite.chainA.GetContext(), msg)

			if tc.expPass {
				suite.Require().NoError(err)

				// replay should an error
				_, err := handler(suite.chainA.GetContext(), msg)
				suite.Require().Error(err)

				// verify packet commitment was deleted
				has := suite.chainA.App.IBCKeeper.ChannelKeeper.HasPacketCommitment(suite.chainA.GetContext(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
				suite.Require().False(has)

			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// tests the IBC handler timing out a packet on ordered and unordered channels.
// It verifies that the deletion of a packet commitment occurs. It tests
// high level properties like ordering and basic sanity checks. More
// rigorous testing of 'TimeoutPacket' and 'TimeoutExecuted' can be found in
// the 04-channel/keeper/timeout_test.go.
func (suite *HandlerTestSuite) TestHandleTimeoutPacket() {
	var (
		packet    channeltypes.Packet
		packetKey []byte
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{"success: ORDERED", func() {
			clientA, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, channeltypes.ORDERED)
			packet = channeltypes.NewPacket(suite.chainA.GetPacketData(suite.chainB), 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, uint64(suite.chainB.GetContext().BlockHeight()), uint64(suite.chainB.GetContext().BlockTime().UnixNano()))
			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			// need to update chainA client to prove missing ack
			suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, clientexported.Tendermint)

			packetKey = host.KeyNextSequenceRecv(packet.GetDestPort(), packet.GetDestChannel())
		}, true},
		{"success: UNORDERED", func() {
			clientA, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			packet = channeltypes.NewPacket(suite.chainA.GetPacketData(suite.chainB), 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, uint64(suite.chainB.GetContext().BlockHeight()), uint64(suite.chainB.GetContext().BlockTime().UnixNano()))
			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			// need to update chainA client to prove missing ack
			suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, clientexported.Tendermint)
			packetKey = host.KeyPacketAcknowledgement(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
		}, true},
		{"success: UNORDERED timeout out of order packet", func() {
			// setup uses an UNORDERED channel
			clientA, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)

			// attempts to timeout packet with sequence 10 without timing out packet with sequence 1
			for i := uint64(1); i < 10; i++ {
				packet = channeltypes.NewPacket(suite.chainA.GetPacketData(suite.chainB), i, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, uint64(suite.chainB.GetContext().BlockHeight()), 0)

				err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
				suite.Require().NoError(err)
			}
			// need to update chainA client to prove missing ack
			suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, clientexported.Tendermint)
			packetKey = host.KeyPacketAcknowledgement(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())

		}, true},
		{"success: ORDERED timeout out of order packet", func() {
			clientA, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, channeltypes.ORDERED)

			// attempts to timeout packet with sequence 10 without timing out packet with sequence 1
			for i := uint64(1); i < 10; i++ {
				packet = channeltypes.NewPacket(suite.chainA.GetPacketData(suite.chainB), i, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, uint64(suite.chainB.GetContext().BlockHeight()), 0)

				err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
				suite.Require().NoError(err)
			}
			// need to update chainA client to prove missing ack
			suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, clientexported.Tendermint)
			packetKey = host.KeyNextSequenceRecv(packet.GetDestPort(), packet.GetDestChannel())

		}, true},
		{"channel does not exist", func() {
			// any non-nil value of packet is valid
			suite.Require().NotNil(packet)

			packetKey = host.KeyNextSequenceRecv(packet.GetDestPort(), packet.GetDestChannel())
		}, false},
		{"UNORDERED: packet not sent", func() {
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			packet = channeltypes.NewPacket(suite.chainA.GetPacketData(suite.chainB), 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)
			packetKey = host.KeyPacketAcknowledgement(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
		}, false},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			handler := ibc.NewHandler(*suite.chainA.App.IBCKeeper)

			tc.malleate()

			proof, proofHeight := suite.chainB.QueryProof(packetKey)

			msg := channeltypes.NewMsgTimeout(packet, 1, proof, proofHeight, suite.chainA.SenderAccount.GetAddress())

			_, err := handler(suite.chainA.GetContext(), msg)

			if tc.expPass {
				suite.Require().NoError(err)

				// replay should return an error
				_, err := handler(suite.chainA.GetContext(), msg)
				suite.Require().Error(err)

				// verify packet commitment was deleted
				has := suite.chainA.App.IBCKeeper.ChannelKeeper.HasPacketCommitment(suite.chainA.GetContext(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
				suite.Require().False(has)

			} else {
				suite.Require().Error(err)
			}
		})
	}
}
