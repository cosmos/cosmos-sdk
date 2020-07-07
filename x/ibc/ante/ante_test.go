package ante_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/ante"
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

func newTx(msg sdk.Msg) sdk.Tx {
	return authtypes.StdTx{
		Msgs: []sdk.Msg{msg},
	}
}

// tests the ante handler receiving a packet on ordered and unordered channels.
// It verifies that no state changes occur as the storing of an acknowledgement
// should occur in the 'PacketExecuted' function. It test high level properties
// like ordering and basic sanity checks. More rigorous testing of 'RecvPacket'
// can be found in the 04-channel/keeper/packet_test.go.
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
			packet = channeltypes.NewPacket(ibctesting.TestHash, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)
		}, true},
		{"success: UNORDERED", func() {
			_, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			packet = channeltypes.NewPacket(ibctesting.TestHash, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)
		}, true},
		{"success: UNORDERED out of order packet", func() {
			// setup uses an UNORDERED channel
			_, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)

			// attempts to receive packet with sequence 10 without receiving packet with sequence 1
			for i := uint64(1); i < 10; i++ {
				packet = channeltypes.NewPacket(ibctesting.TestHash, i, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

				err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
				suite.Require().NoError(err)
			}
		}, true},
		{"failure: ORDERED out of order packet", func() {
			_, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, channeltypes.ORDERED)

			// attempts to receive packet with sequence 10 without receiving packet with sequence 1
			for i := uint64(1); i < 10; i++ {
				packet = channeltypes.NewPacket(ibctesting.TestHash, i, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

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
			packet = channeltypes.NewPacket(ibctesting.TestHash, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)
		}, false},
		{"ORDERED: packet already received (replay)", func() {
			clientA, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, channeltypes.ORDERED)
			packet = channeltypes.NewPacket(ibctesting.TestHash, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			err = suite.coordinator.PacketExecuted(suite.chainB, suite.chainA, packet, clientA)
			suite.Require().NoError(err)
		}, false},
		{"UNORDERED: packet already received (replay)", func() {
			clientA, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)

			packet = channeltypes.NewPacket(ibctesting.TestHash, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

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

			handler := sdk.ChainAnteDecorators(ante.NewProofVerificationDecorator(
				suite.chainB.App.IBCKeeper.ClientKeeper,
				suite.chainB.App.IBCKeeper.ChannelKeeper,
			))

			tc.malleate()

			// get proof of packet commitment from chainA
			packetKey := host.KeyPacketCommitment(packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
			proof, proofHeight := suite.chainA.QueryProof(packetKey)

			msg := channeltypes.NewMsgPacket(packet, proof, proofHeight, suite.chainB.SenderAccount.GetAddress())

			// ante-handle RecvPacket
			_, err := handler(suite.chainB.GetContext(), newTx(msg), false)

			if tc.expPass {
				suite.Require().NoError(err)
				// replay should return same result since there is no state changes
				_, err := handler(suite.chainB.GetContext(), newTx(msg), false)
				suite.Require().NoError(err)

				// verify ack was not written
				ack, found := suite.chainB.App.IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(suite.chainB.GetContext(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
				suite.Require().Nil(ack)
				suite.Require().False(found)

			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// tests the ante handler acknowledging a packet on ordered and unordered
// channels. It verifies that no state changes occur as the deleting of packet
// commitments from state should occur in the 'AcknowledgementExecuted'
// function. It test high level properties like ordering and basic sanity
// checks. More rigorous testing of 'AcknowledgePacket' can be found in
// the 04-channel/keeper/packet_test.go.
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
			packet = channeltypes.NewPacket(ibctesting.TestHash, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			err = suite.coordinator.PacketExecuted(suite.chainB, suite.chainA, packet, clientA)
			suite.Require().NoError(err)
		}, true},
		{"success: UNORDERED", func() {
			clientA, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			packet = channeltypes.NewPacket(ibctesting.TestHash, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

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
				packet = channeltypes.NewPacket(ibctesting.TestHash, i, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

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
				packet = channeltypes.NewPacket(ibctesting.TestHash, i, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

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
			packet = channeltypes.NewPacket(ibctesting.TestHash, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)
		}, false},
		{"ORDERED: packet already acknowledged (replay)", func() {
			clientA, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, channeltypes.ORDERED)
			packet = channeltypes.NewPacket(ibctesting.TestHash, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			err = suite.coordinator.PacketExecuted(suite.chainB, suite.chainA, packet, clientA)
			suite.Require().NoError(err)

			err = suite.coordinator.AcknowledgementExecuted(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)
		}, false},
		{"UNORDERED: packet already received (replay)", func() {
			clientA, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)

			packet = channeltypes.NewPacket(ibctesting.TestHash, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)

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

			handler := sdk.ChainAnteDecorators(ante.NewProofVerificationDecorator(
				suite.chainA.App.IBCKeeper.ClientKeeper,
				suite.chainA.App.IBCKeeper.ChannelKeeper,
			))

			tc.malleate()

			packetKey := host.KeyPacketAcknowledgement(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
			proof, proofHeight := suite.chainB.QueryProof(packetKey)

			ack := ibctesting.TestHash

			msg := channeltypes.NewMsgAcknowledgement(packet, ack, proof, proofHeight, suite.chainA.SenderAccount.GetAddress())

			// ante-handle RecvPacket
			_, err := handler(suite.chainA.GetContext(), newTx(msg), false)

			if tc.expPass {
				suite.Require().NoError(err)
				// replay should return same result since there is no state changes
				_, err := handler(suite.chainA.GetContext(), newTx(msg), false)
				suite.Require().NoError(err)

				// verify packet commitment was not deleted
				has := suite.chainA.App.IBCKeeper.ChannelKeeper.HasPacketCommitment(suite.chainA.GetContext(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
				suite.Require().True(has)

			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// tests the ante handler timing out a packet on ordered and unordered channels.
// It verifies that no state changes occur as the deleting of packet
// commitments from state should occur in the 'TimeoutExecuted' function. It
// tests high level properties like ordering and basic sanity
// checks. More rigorous testing of 'TimeoutPacket' can be found in the
// 04-channel/keeper/timeout_test.go.
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
			packet = channeltypes.NewPacket(ibctesting.TestHash, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, uint64(suite.chainB.GetContext().BlockHeight()), uint64(suite.chainB.GetContext().BlockTime().UnixNano()))
			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			// need to update chainA client to prove missing ack
			suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, clientexported.Tendermint)

			packetKey = host.KeyNextSequenceRecv(packet.GetDestPort(), packet.GetDestChannel())
		}, true},
		{"success: UNORDERED", func() {
			clientA, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			packet = channeltypes.NewPacket(ibctesting.TestHash, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, uint64(suite.chainB.GetContext().BlockHeight()), uint64(suite.chainB.GetContext().BlockTime().UnixNano()))
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
				packet = channeltypes.NewPacket(ibctesting.TestHash, i, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, uint64(suite.chainB.GetContext().BlockHeight()), 0)

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
				packet = channeltypes.NewPacket(ibctesting.TestHash, i, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, uint64(suite.chainB.GetContext().BlockHeight()), 0)

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
			packet = channeltypes.NewPacket(ibctesting.TestHash, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)
			packetKey = host.KeyPacketAcknowledgement(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
		}, false},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			handler := sdk.ChainAnteDecorators(ante.NewProofVerificationDecorator(
				suite.chainA.App.IBCKeeper.ClientKeeper,
				suite.chainA.App.IBCKeeper.ChannelKeeper,
			))

			tc.malleate()

			proof, proofHeight := suite.chainB.QueryProof(packetKey)

			msg := channeltypes.NewMsgTimeout(packet, 1, proof, proofHeight, suite.chainA.SenderAccount.GetAddress())

			// ante-handle RecvPacket
			_, err := handler(suite.chainA.GetContext(), newTx(msg), false)

			if tc.expPass {
				suite.Require().NoError(err)
				// replay should return same result since there is no state changes
				_, err := handler(suite.chainA.GetContext(), newTx(msg), false)
				suite.Require().NoError(err)

				// verify packet commitment was not deleted
				has := suite.chainA.App.IBCKeeper.ChannelKeeper.HasPacketCommitment(suite.chainA.GetContext(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
				suite.Require().True(has)

			} else {
				suite.Require().Error(err)
			}
		})
	}
}
func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
