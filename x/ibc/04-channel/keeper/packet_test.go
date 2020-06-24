package keeper_test

import (
	"fmt"

	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

var (
	validPacketData          = []byte("VALID PACKET DATA")
	disabledTimeoutTimestamp = uint64(0)
	disabledTimeoutHeight    = uint64(0)
	timeoutHeight            = uint64(100)

	// for when the testing package cannot be used
	clientIDA  = "clientA"
	clientIDB  = "clientB"
	connIDA    = "connA"
	connIDB    = "connB"
	portID     = "portid"
	channelIDA = "channelidA"
	channelIDB = "channelidB"
)

// TestSendPacket tests SendPacket on chainA
func (suite *KeeperTestSuite) TestSendPacket() {
	var (
		packet     exported.PacketI
		channelCap *capabilitytypes.Capability
	)

	testCases := []testCase{
		{"success", func() {
			_, _, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, true},
		{"packet basic validation failed, empty packet data", func() {
			_, _, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			packet = types.NewPacket([]byte{}, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"channel not found", func() {
			// use wrong channel naming
			_, _, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			packet = types.NewPacket(validPacketData, 1, "invalidport", "invalidchannel", channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"channel closed", func() {
			_, _, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)

			err := suite.coordinator.SetChannelClosed(suite.chainA, suite.chainB, channelA)
			suite.Require().NoError(err)
		}, false},
		{"packet dest port ≠ channel counterparty port", func() {
			_, _, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			// use wrong port for dest
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, "invalidport", channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"packet dest channel ID ≠ channel counterparty channel ID", func() {
			_, _, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			// use wrong channel for dest
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, "invalidchannel", timeoutHeight, disabledTimeoutTimestamp)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"connection not found", func() {
			channelA := ibctesting.TestChannel{PortID: portID, ID: channelIDA}
			channelB := ibctesting.TestChannel{PortID: portID, ID: channelIDB}
			// pass channel check
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetChannel(
				suite.chainA.GetContext(),
				channelA.PortID, channelA.ID,
				types.NewChannel(types.OPEN, types.ORDERED, types.NewCounterparty(channelB.PortID, channelB.ID), []string{connIDA}, ibctesting.ChannelVersion),
			)
			suite.chainA.CreateChannelCapability(channelA.PortID, channelA.ID)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"connection is UNINITIALIZED", func() {
			// set connection as UNINITIALIZED
			prefix := commitmenttypes.NewMerklePrefix(suite.chainB.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes())
			counterparty := connectiontypes.NewCounterparty(clientIDB, connIDA, prefix)
			connection := connectiontypes.NewConnectionEnd(connectiontypes.UNINITIALIZED, clientIDA, connIDA, counterparty, []string{ibctesting.ConnectionVersion})
			suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainA.GetContext(), connIDA, connection)

			channelA := ibctesting.TestChannel{PortID: portID, ID: channelIDA}
			channelB := ibctesting.TestChannel{PortID: portID, ID: channelIDB}
			// pass channel check
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetChannel(
				suite.chainA.GetContext(),
				channelA.PortID, channelA.ID,
				types.NewChannel(types.OPEN, types.ORDERED, types.NewCounterparty(channelB.PortID, channelB.ID), []string{connIDA}, ibctesting.ChannelVersion),
			)
			suite.chainA.CreateChannelCapability(channelA.PortID, channelA.ID)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"client state not found", func() {
			_, _, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]

			// change connection client ID
			connection := suite.chainA.GetConnection(connA)
			connection.ClientID = "invalidid"
			suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainA.GetContext(), connA.ID, connection)

			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"timeout height passed", func() {
			clientA, _, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			// use client state latest height for timeout
			clientState := suite.chainA.GetClientState(clientA)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, clientState.GetLatestHeight(), disabledTimeoutTimestamp)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"timeout timestamp passed", func() {
			clientA, _, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]

			// use latest time on client state
			clientState := suite.chainA.GetClientState(clientA)
			connection := suite.chainA.GetConnection(connA)
			timestamp, err := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetTimestampAtHeight(suite.chainA.GetContext(), connection, clientState.GetLatestHeight())
			suite.Require().NoError(err)

			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, disabledTimeoutHeight, timestamp)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"next sequence send not found", func() {
			_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA := connA.NextTestChannel()
			channelB := connB.NextTestChannel()
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			// manually creating channel prevents next sequence from being set
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetChannel(
				suite.chainA.GetContext(),
				channelA.PortID, channelA.ID,
				types.NewChannel(types.OPEN, types.ORDERED, types.NewCounterparty(channelB.PortID, channelB.ID), []string{connA.ID}, ibctesting.ChannelVersion),
			)
			suite.chainA.CreateChannelCapability(channelA.PortID, channelA.ID)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"next sequence wrong", func() {
			_, _, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.chainA.GetContext(), channelA.PortID, channelA.ID, 5)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"channel capability not found", func() {
			_, _, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			channelCap = capabilitytypes.NewCapability(5)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest() // reset

			tc.malleate()

			err := suite.chainA.App.IBCKeeper.ChannelKeeper.SendPacket(suite.chainA.GetContext(), channelCap, packet)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}

}

// TestRecvPacket test RecvPacket on chainB. Since packet commitment verification will always
// occur last (resource instensive), only tests expected to succeed and packet commitment
// verification tests need to simulate sending a packet from chainA to chainB.
func (suite *KeeperTestSuite) TestRecvPacket() {
	var (
		packet exported.PacketI
	)

	testCases := []testCase{
		{"success ordered channel", func() {
			_, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, types.ORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
		}, true},
		{"success unordered channel", func() {
			// setup uses an UNORDERED channel
			_, clientB, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
		}, true},
		{"channel not found", func() {
			// use wrong channel naming
			_, _, connA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, "invalidport", "invalidchannel", timeoutHeight, disabledTimeoutTimestamp)
		}, false},
		{"channel not open", func() {
			_, _, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)

			err := suite.coordinator.SetChannelClosed(suite.chainB, suite.chainA, channelB)
			suite.Require().NoError(err)
		}, false},
		{"packet source port ≠ channel counterparty port", func() {
			_, _, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			// use wrong port for dest
			packet = types.NewPacket(validPacketData, 1, "invalidport", channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
		}, false},
		{"packet source channel ID ≠ channel counterparty channel ID", func() {
			_, _, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			// use wrong port for dest
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, "invalidchannel", channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
		}, false},
		{"connection not found", func() {
			channelA := ibctesting.TestChannel{PortID: portID, ID: channelIDA}
			channelB := ibctesting.TestChannel{PortID: portID, ID: channelIDB}
			// pass channel check
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetChannel(
				suite.chainB.GetContext(),
				channelB.PortID, channelB.ID,
				types.NewChannel(types.OPEN, types.ORDERED, types.NewCounterparty(channelA.PortID, channelA.ID), []string{connIDB}, ibctesting.ChannelVersion),
			)
		}, false},
		{"connection not OPEN", func() {
			clientA, clientB := suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
			// connection on chainB is in INIT
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			channelA := connA.NextTestChannel()
			channelB := connB.NextTestChannel()
			// pass channel check
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetChannel(
				suite.chainB.GetContext(),
				channelB.PortID, channelB.ID,
				types.NewChannel(types.OPEN, types.ORDERED, types.NewCounterparty(channelA.PortID, channelA.ID), []string{connIDB}, ibctesting.ChannelVersion),
			)
		}, false},
		{"timeout height passed", func() {
			_, _, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, uint64(suite.chainB.GetContext().BlockHeight()), disabledTimeoutTimestamp)
		}, false},
		{"timeout timestamp passed", func() {
			_, _, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, disabledTimeoutHeight, uint64(suite.chainB.GetContext().BlockTime().UnixNano()))
		}, false},
		{"acknowledgement already received", func() {
			// setup uses an UNORDERED channel
			clientA, clientB, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)

			// write packet acknowledgement
			suite.coordinator.PacketExecuted(suite.chainB, suite.chainA, packet, clientA)
		}, false},
		{"validation failed", func() {
			// packet commitment not set resulting in invalid proof
			_, _, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest() // reset
			tc.malleate()

			// get proof of packet commitment from chainA
			packetKey := host.KeyPacketCommitment(packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
			proof, proofHeight := suite.chainA.QueryProof(packetKey)

			_, err := suite.chainB.App.IBCKeeper.ChannelKeeper.RecvPacket(suite.chainB.GetContext(), packet, proof, proofHeight+1)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}

}

// TestPacketExecuted tests the PacketExecuted call on chainB.
func (suite *KeeperTestSuite) TestPacketExecuted() {
	var (
		packet     types.Packet
		channelCap *capabilitytypes.Capability
	)

	testCases := []testCase{
		{"success: UNORDERED", func() {
			// setup uses an UNORDERED channel
			_, clientB, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			channelCap = suite.chainB.GetChannelCapability(channelB.PortID, channelB.ID)
		}, true},
		{"success: ORDERED", func() {
			_, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, types.ORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			channelCap = suite.chainB.GetChannelCapability(channelB.PortID, channelB.ID)
		}, true},
		{"channel not found", func() {
			// use wrong channel naming
			_, _, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, "invalidport", "invalidchannel", timeoutHeight, disabledTimeoutTimestamp)
			channelCap = suite.chainB.GetChannelCapability(channelB.PortID, channelB.ID)
		}, false},
		{"channel not OPEN", func() {
			_, _, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)

			err := suite.coordinator.SetChannelClosed(suite.chainB, suite.chainA, channelB)
			suite.Require().NoError(err)
		}, false},
		{"next sequence receive not found", func() {
			_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA := connA.NextTestChannel()
			channelB := connB.NextTestChannel()
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			// manually creating channel prevents next sequence receive from being set
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetChannel(
				suite.chainB.GetContext(),
				channelB.PortID, channelB.ID,
				types.NewChannel(types.OPEN, types.ORDERED, types.NewCounterparty(channelA.PortID, channelA.ID), []string{connB.ID}, ibctesting.ChannelVersion),
			)
			suite.chainB.CreateChannelCapability(channelB.PortID, channelB.ID)
			channelCap = suite.chainB.GetChannelCapability(channelB.PortID, channelB.ID)
		}, false},
		{"packet sequence ≠ next sequence receive", func() {
			clientA, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, types.ORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)

			// increments sequence receive
			suite.coordinator.PacketExecuted(suite.chainB, suite.chainA, packet, clientA)
		}, false},
		{"capability not found", func() {
			_, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, types.ORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)

			channelCap = capabilitytypes.NewCapability(3)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest() // reset

			tc.malleate()

			ack := ibctesting.TestHash
			err := suite.chainB.App.IBCKeeper.ChannelKeeper.PacketExecuted(suite.chainB.GetContext(), channelCap, packet, ack)

			if tc.expPass {
				suite.Require().NoError(err)
				// verify packet ack is written
				actualAck := suite.chainB.GetAcknowledgement(packet)
				suite.Require().Equal(types.CommitAcknowledgement(ack), actualAck)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

/*
func (suite *KeeperTestSuite) TestAcknowledgePacket() {
	counterparty := types.NewCounterparty(testPort2, testChannel2)
	var packet types.Packet
	packetKey := host.KeyPacketAcknowledgement(testPort2, testChannel2, 1)

	ack := transfertypes.FungibleTokenPacketAcknowledgement{
		Success: true,
	}.GetBytes()

	testCases := []testCase{
		{"success on ordered channel", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainA.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, connection.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
			suite.chainA.createChannel(testPort2, testChannel2, testPort1, testChannel1, types.OPEN, types.ORDERED, testConnectionIDB)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 1, types.CommitPacket(packet))
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(suite.chainA.GetContext(), testPort2, testChannel2, 1, types.CommitAcknowledgement(ack))
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceAck(suite.chainB.GetContext(), counterparty.GetPortID(), counterparty.GetChannelID(), 1)
		}, true},
		{"success on unordered channel", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainA.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, connection.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.UNORDERED, testConnectionIDA)
			suite.chainA.createChannel(testPort2, testChannel2, testPort1, testChannel1, types.OPEN, types.UNORDERED, testConnectionIDB)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 1, types.CommitPacket(packet))
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(suite.chainA.GetContext(), testPort2, testChannel2, 1, types.CommitAcknowledgement(ack))
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceAck(suite.chainB.GetContext(), counterparty.GetPortID(), counterparty.GetChannelID(), 1)
		}, true},

		{"channel not found", func() {}, false},
		{"channel not open", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.CLOSED, types.ORDERED, testConnectionIDA)
		}, false},
		{"packet source port ≠ channel counterparty port", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createChannel(testPort1, testChannel1, testPort3, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"packet source channel ID ≠ channel counterparty channel ID", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel3, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"connection not found", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"connection not OPEN", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.INIT)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"packet hasn't been sent", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"packet ack verification failed", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 1, types.CommitPacket(packet))
		}, false},
		{"next ack sequence not found", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainA.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, connection.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
			suite.chainA.createChannel(testPort2, testChannel2, testPort1, testChannel1, types.OPEN, types.ORDERED, testConnectionIDB)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 1, types.CommitPacket(packet))
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(suite.chainA.GetContext(), testPort2, testChannel2, 1, types.CommitAcknowledgement(ack))
		}, false},
		{"next ack sequence mismatch", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainA.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, connection.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
			suite.chainA.createChannel(testPort2, testChannel2, testPort1, testChannel1, types.OPEN, types.ORDERED, testConnectionIDB)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 1, types.CommitPacket(packet))
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(suite.chainA.GetContext(), testPort2, testChannel2, 1, types.CommitAcknowledgement(ack))
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceAck(suite.chainB.GetContext(), counterparty.GetPortID(), counterparty.GetChannelID(), 10)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest() // reset
			tc.malleate()

			suite.chainA.updateClient(suite.chainB)
			suite.chainB.updateClient(suite.chainA)
			proof, proofHeight := queryProof(suite.chainA, packetKey)

			ctx := suite.chainB.GetContext()
			packetOut, err := suite.chainB.App.IBCKeeper.ChannelKeeper.AcknowledgePacket(ctx, packet, ack, proof, proofHeight+1)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(packetOut)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(packetOut)
			}
		})
	}
}

// TestAcknowledgementExectued verifies that packet commitments are deleted after
// capabilities are verified.
func (suite *KeeperTestSuite) TestAcknowledgementExecuted() {
	sequence := uint64(1)
	counterparty := types.NewCounterparty(testPort2, testChannel2)

	var (
		packet  types.Packet
		chanCap *capabilitytypes.Capability
	)

	testCases := []testCase{
		{"success ORDERED", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), sequence, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainA.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainA.GetContext(), packet.GetSourcePort(), packet.GetSourceChannel(), sequence, types.CommitPacket(packet))
		}, true},
		{"channel not found", func() {}, false},
		{"incorrect capability", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), sequence, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainA.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
			chanCap = capabilitytypes.NewCapability(100)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest() // reset

			var err error
			chanCap, err = suite.chainA.App.ScopedIBCKeeper.NewCapability(
				suite.chainA.GetContext(), host.ChannelCapabilityPath(testPort1, testChannel1),
			)
			suite.Require().NoError(err, "could not create capability")

			tc.malleate()

			err = suite.chainA.App.IBCKeeper.ChannelKeeper.AcknowledgementExecuted(suite.chainA.GetContext(), chanCap, packet)
			pc := suite.chainA.App.IBCKeeper.ChannelKeeper.GetPacketCommitment(suite.chainA.GetContext(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())

			if tc.expPass {
				suite.NoError(err)
				suite.Nil(pc)
			} else {
				suite.Error(err)
			}
		})
	}
}
func (suite *KeeperTestSuite) TestCleanupPacket() {
	counterparty := types.NewCounterparty(testPort2, testChannel2)
	unorderedPacketKey := host.KeyPacketAcknowledgement(testPort2, testChannel2, 1)
	orderedPacketKey := host.KeyNextSequenceRecv(testPort2, testChannel2)

	var (
		packet      types.Packet
		nextSeqRecv uint64
		ordered     bool
	)

	ack := []byte("ack")

	testCases := []testCase{
		{"success on ordered channel", func() {
			ordered = true
			nextSeqRecv = 6
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainA.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, connection.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
			suite.chainA.createChannel(testPort2, testChannel2, testPort1, testChannel1, types.OPEN, types.ORDERED, testConnectionIDB)
			// create several packet commitments
			for i := uint64(1); i < nextSeqRecv; i++ {
				packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), i, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
				suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, i, types.CommitPacket(packet))
			}
			// set next sequence recv
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.chainA.GetContext(), testPort2, testChannel2, nextSeqRecv)
		}, true},
		{"success on unordered channel", func() {
			ordered = false
			nextSeqRecv = 10
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainA.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, connection.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.UNORDERED, testConnectionIDA)
			suite.chainA.createChannel(testPort2, testChannel2, testPort1, testChannel1, types.OPEN, types.UNORDERED, testConnectionIDB)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 1, types.CommitPacket(packet))
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(suite.chainA.GetContext(), testPort2, testChannel2, 1, types.CommitAcknowledgement(ack))
		}, true},
		{"channel not found", func() {}, false},
		{"channel not open", func() {
			ordered = true
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.CLOSED, types.ORDERED, testConnectionIDA)
		}, false},
		{"packet source port ≠ channel counterparty port", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createChannel(testPort1, testChannel1, testPort3, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"packet source channel ID ≠ channel counterparty channel ID", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel3, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"connection not found", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"connection not OPEN", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.INIT)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"packet already received ", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 10, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"packet hasn't been sent", func() {
			nextSeqRecv = 10
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"next seq receive verification failed", func() {
			nextSeqRecv = 10
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 1, types.CommitPacket(packet))
		}, false},
		{"packet ack verification failed", func() {
			nextSeqRecv = 10
			ordered = false
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.UNORDERED, testConnectionIDA)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 1, types.CommitPacket(packet))
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			var proof []byte
			var proofHeight uint64

			suite.SetupTest() // reset
			tc.malleate()

			ctx := suite.chainB.GetContext()

			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)

			if ordered {
				proof, proofHeight = queryProof(suite.chainA, orderedPacketKey)
			} else {
				proof, proofHeight = queryProof(suite.chainA, unorderedPacketKey)
			}

			cap, err := suite.chainB.App.ScopedIBCKeeper.NewCapability(ctx, host.ChannelCapabilityPath(testPort1, testChannel1))
			suite.Require().NoError(err)

			if tc.expPass {
				packetOut, err := suite.chainB.App.IBCKeeper.ChannelKeeper.CleanupPacket(ctx, cap, packet, proof, proofHeight+1, nextSeqRecv, ack)
				suite.Require().NoError(err)
				suite.Require().NotNil(packetOut)

				if ordered {
					for i := uint64(1); i < nextSeqRecv; i++ {
						pc := suite.chainB.App.IBCKeeper.ChannelKeeper.GetPacketCommitment(ctx, testPort1, testChannel1, i)
						suite.Require().Nil(pc)
					}
				} else {
					pc := suite.chainB.App.IBCKeeper.ChannelKeeper.GetPacketCommitment(ctx, testPort1, testChannel1, packet.GetSequence())
					suite.Require().Nil(pc)
				}
			} else {
				packetOut, err := suite.chainB.App.IBCKeeper.ChannelKeeper.CleanupPacket(ctx, cap, packet, proof, proofHeight, nextSeqRecv, ack)
				suite.Require().Error(err)
				suite.Require().Nil(packetOut)
			}
		})
	}
}
*/
