package keeper_test

import (
	"fmt"

	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

var (
	validPacketData          = []byte("VALID PACKET DATA")
	disabledTimeoutTimestamp = uint64(0)
	disabledTimeoutHeight    = clienttypes.ZeroHeight()
	timeoutHeight            = clienttypes.NewHeight(0, 100)

	// for when the testing package cannot be used
	clientIDA  = "clientA"
	clientIDB  = "clientB"
	connIDA    = "connA"
	connIDB    = "connB"
	portID     = "portid"
	channelIDA = "channelidA"
	channelIDB = "channelidB"
)

// TestSendPacket tests SendPacket from chainA to chainB
func (suite *KeeperTestSuite) TestSendPacket() {
	var (
		packet     exported.PacketI
		channelCap *capabilitytypes.Capability
	)

	testCases := []testCase{
		{"success: UNORDERED channel", func() {
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, true},
		{"success: ORDERED channel", func() {
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.ORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, true},
		{"sending packet out of order on UNORDERED channel", func() {
			// setup creates an unordered channel
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			packet = types.NewPacket(validPacketData, 5, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"sending packet out of order on ORDERED channel", func() {
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.ORDERED)
			packet = types.NewPacket(validPacketData, 5, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"packet basic validation failed, empty packet data", func() {
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			packet = types.NewPacket([]byte{}, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"channel not found", func() {
			// use wrong channel naming
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			packet = types.NewPacket(validPacketData, 1, ibctesting.InvalidID, ibctesting.InvalidID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"channel closed", func() {
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)

			err := suite.coordinator.SetChannelClosed(suite.chainA, suite.chainB, channelA)
			suite.Require().NoError(err)
		}, false},
		{"packet dest port ≠ channel counterparty port", func() {
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			// use wrong port for dest
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, ibctesting.InvalidID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"packet dest channel ID ≠ channel counterparty channel ID", func() {
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			// use wrong channel for dest
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, ibctesting.InvalidID, timeoutHeight, disabledTimeoutTimestamp)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"connection not found", func() {
			channelA := ibctesting.TestChannel{PortID: portID, ID: channelIDA}
			channelB := ibctesting.TestChannel{PortID: portID, ID: channelIDB}
			// pass channel check
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetChannel(
				suite.chainA.GetContext(),
				channelA.PortID, channelA.ID,
				types.NewChannel(types.OPEN, types.ORDERED, types.NewCounterparty(channelB.PortID, channelB.ID), []string{connIDA}, channelA.Version),
			)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			suite.chainA.CreateChannelCapability(channelA.PortID, channelA.ID)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"connection is UNINITIALIZED", func() {
			// set connection as UNINITIALIZED
			counterparty := connectiontypes.NewCounterparty(clientIDB, connIDA, suite.chainB.GetPrefix())
			connection := connectiontypes.NewConnectionEnd(connectiontypes.UNINITIALIZED, clientIDA, counterparty, []string{ibctesting.ConnectionVersion})
			suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainA.GetContext(), connIDA, connection)

			channelA := ibctesting.TestChannel{PortID: portID, ID: channelIDA}
			channelB := ibctesting.TestChannel{PortID: portID, ID: channelIDB}
			// pass channel check
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetChannel(
				suite.chainA.GetContext(),
				channelA.PortID, channelA.ID,
				types.NewChannel(types.OPEN, types.ORDERED, types.NewCounterparty(channelB.PortID, channelB.ID), []string{connIDA}, channelA.Version),
			)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			suite.chainA.CreateChannelCapability(channelA.PortID, channelA.ID)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"client state not found", func() {
			_, _, connA, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)

			// change connection client ID
			connection := suite.chainA.GetConnection(connA)
			connection.ClientId = ibctesting.InvalidID
			suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainA.GetContext(), connA.ID, connection)

			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"timeout height passed", func() {
			clientA, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			// use client state latest height for timeout
			clientState := suite.chainA.GetClientState(clientA)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, clientState.GetLatestHeight().(clienttypes.Height), disabledTimeoutTimestamp)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"timeout timestamp passed", func() {
			clientA, _, connA, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			// use latest time on client state
			clientState := suite.chainA.GetClientState(clientA)
			connection := suite.chainA.GetConnection(connA)
			timestamp, err := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetTimestampAtHeight(suite.chainA.GetContext(), connection, clientState.GetLatestHeight())
			suite.Require().NoError(err)

			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, disabledTimeoutHeight, timestamp)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"next sequence send not found", func() {
			_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			channelA := connA.NextTestChannel(ibctesting.TransferPort)
			channelB := connB.NextTestChannel(ibctesting.TransferPort)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			// manually creating channel prevents next sequence from being set
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetChannel(
				suite.chainA.GetContext(),
				channelA.PortID, channelA.ID,
				types.NewChannel(types.OPEN, types.ORDERED, types.NewCounterparty(channelB.PortID, channelB.ID), []string{connA.ID}, channelA.Version),
			)
			suite.chainA.CreateChannelCapability(channelA.PortID, channelA.ID)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"next sequence wrong", func() {
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.chainA.GetContext(), channelA.PortID, channelA.ID, 5)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"channel capability not found", func() {
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
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
		{"success: ORDERED channel", func() {
			_, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.ORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)
		}, true},
		{"success UNORDERED channel", func() {
			// setup uses an UNORDERED channel
			_, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)
		}, true},
		{"success with out of order packet: UNORDERED channel", func() {
			// setup uses an UNORDERED channel
			_, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)

			// send 2 packets
			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)
			// set sequence to 2
			packet = types.NewPacket(validPacketData, 2, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			err = suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)
			// attempts to receive packet 2 without receiving packet 1
		}, true},
		{"out of order packet failure with ORDERED channel", func() {
			_, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.ORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)

			// send 2 packets
			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)
			// set sequence to 2
			packet = types.NewPacket(validPacketData, 2, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			err = suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)
			// attempts to receive packet 2 without receiving packet 1
		}, false},
		{"channel not found", func() {
			// use wrong channel naming
			_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, ibctesting.InvalidID, ibctesting.InvalidID, timeoutHeight, disabledTimeoutTimestamp)
		}, false},
		{"channel not open", func() {
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)

			err := suite.coordinator.SetChannelClosed(suite.chainB, suite.chainA, channelB)
			suite.Require().NoError(err)
		}, false},
		{"packet source port ≠ channel counterparty port", func() {
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			// use wrong port for dest
			packet = types.NewPacket(validPacketData, 1, ibctesting.InvalidID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
		}, false},
		{"packet source channel ID ≠ channel counterparty channel ID", func() {
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			// use wrong port for dest
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, ibctesting.InvalidID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
		}, false},
		{"connection not found", func() {
			channelA := ibctesting.TestChannel{PortID: portID, ID: channelIDA}
			channelB := ibctesting.TestChannel{PortID: portID, ID: channelIDB}
			// pass channel check
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetChannel(
				suite.chainB.GetContext(),
				channelB.PortID, channelB.ID,
				types.NewChannel(types.OPEN, types.ORDERED, types.NewCounterparty(channelA.PortID, channelA.ID), []string{connIDB}, channelB.Version),
			)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
		}, false},
		{"connection not OPEN", func() {
			clientA, clientB := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
			// connection on chainB is in INIT
			connB, connA, err := suite.coordinator.ConnOpenInit(suite.chainB, suite.chainA, clientB, clientA)
			suite.Require().NoError(err)

			channelA := connA.NextTestChannel(ibctesting.TransferPort)
			channelB := connB.NextTestChannel(ibctesting.TransferPort)
			// pass channel check
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetChannel(
				suite.chainB.GetContext(),
				channelB.PortID, channelB.ID,
				types.NewChannel(types.OPEN, types.ORDERED, types.NewCounterparty(channelA.PortID, channelA.ID), []string{connB.ID}, channelB.Version),
			)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
		}, false},
		{"timeout height passed", func() {
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, clienttypes.GetSelfHeight(suite.chainB.GetContext()), disabledTimeoutTimestamp)
		}, false},
		{"timeout timestamp passed", func() {
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, disabledTimeoutHeight, uint64(suite.chainB.GetContext().BlockTime().UnixNano()))
		}, false},
		{"acknowledgement already received", func() {
			// setup uses an UNORDERED channel
			clientA, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)

			// write packet acknowledgement
			suite.coordinator.ReceiveExecuted(suite.chainB, suite.chainA, packet, clientA)
		}, false},
		{"next receive sequence is not found", func() {
			_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			channelA := connA.NextTestChannel(ibctesting.TransferPort)
			channelB := connB.NextTestChannel(ibctesting.TransferPort)

			// manually creating channel prevents next recv sequence from being set
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetChannel(
				suite.chainB.GetContext(),
				channelB.PortID, channelB.ID,
				types.NewChannel(types.OPEN, types.ORDERED, types.NewCounterparty(channelA.PortID, channelA.ID), []string{connB.ID}, channelB.Version),
			)

			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)

			// manually set packet commitment
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainA.GetContext(), channelA.PortID, channelA.ID, packet.GetSequence(), ibctesting.TestHash)
		}, false},
		{"validation failed", func() {
			// packet commitment not set resulting in invalid proof
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
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

			err := suite.chainB.App.IBCKeeper.ChannelKeeper.RecvPacket(suite.chainB.GetContext(), packet, proof, proofHeight)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}

}

// TestReceiveExecuted tests the ReceiveExecuted call on chainB.
func (suite *KeeperTestSuite) TestReceiveExecuted() {
	var (
		packet     types.Packet
		channelCap *capabilitytypes.Capability
		ack        []byte
	)

	testCases := []testCase{
		{"success: UNORDERED", func() {
			// setup uses an UNORDERED channel
			_, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			channelCap = suite.chainB.GetChannelCapability(channelB.PortID, channelB.ID)
		}, true},
		{"success: ORDERED", func() {
			_, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.ORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			channelCap = suite.chainB.GetChannelCapability(channelB.PortID, channelB.ID)
		}, true},
		{"channel not found", func() {
			// use wrong channel naming
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, ibctesting.InvalidID, ibctesting.InvalidID, timeoutHeight, disabledTimeoutTimestamp)
			channelCap = suite.chainB.GetChannelCapability(channelB.PortID, channelB.ID)
		}, false},
		{"channel not OPEN", func() {
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)

			err := suite.coordinator.SetChannelClosed(suite.chainB, suite.chainA, channelB)
			suite.Require().NoError(err)
		}, false},
		{"next sequence receive not found", func() {
			_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			channelA := connA.NextTestChannel(ibctesting.TransferPort)
			channelB := connB.NextTestChannel(ibctesting.TransferPort)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			// manually creating channel prevents next sequence receive from being set
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetChannel(
				suite.chainB.GetContext(),
				channelB.PortID, channelB.ID,
				types.NewChannel(types.OPEN, types.ORDERED, types.NewCounterparty(channelA.PortID, channelA.ID), []string{connB.ID}, channelB.Version),
			)
			suite.chainB.CreateChannelCapability(channelB.PortID, channelB.ID)
			channelCap = suite.chainB.GetChannelCapability(channelB.PortID, channelB.ID)
		}, false},
		{"capability not found", func() {
			_, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.ORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)

			channelCap = capabilitytypes.NewCapability(3)
		}, false},
		{"acknowledgement is empty", func() {
			_, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			channelCap = suite.chainB.GetChannelCapability(channelB.PortID, channelB.ID)

			ack = []byte{}
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest()         // reset
			ack = ibctesting.TestHash // must explicitly be changed in malleate

			tc.malleate()

			err := suite.chainB.App.IBCKeeper.ChannelKeeper.ReceiveExecuted(suite.chainB.GetContext(), channelCap, packet, ack)

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

// TestAcknowledgePacket tests the call AcknowledgePacket on chainA.
func (suite *KeeperTestSuite) TestAcknowledgePacket() {
	var (
		packet types.Packet
		ack    = ibctesting.TestHash
	)

	testCases := []testCase{
		{"success on ordered channel", func() {
			clientA, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.ORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			// create packet commitment
			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			// create packet acknowledgement
			err = suite.coordinator.ReceiveExecuted(suite.chainB, suite.chainA, packet, clientA)
			suite.Require().NoError(err)
		}, true},
		{"success on unordered channel", func() {
			// setup uses an UNORDERED channel
			clientA, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)

			// create packet commitment
			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			// create packet acknowledgement
			err = suite.coordinator.ReceiveExecuted(suite.chainB, suite.chainA, packet, clientA)
			suite.Require().NoError(err)
		}, true},
		{"channel not found", func() {
			// use wrong channel naming
			_, _, _, _, _, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			packet = types.NewPacket(validPacketData, 1, ibctesting.InvalidID, ibctesting.InvalidID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
		}, false},
		{"channel not open", func() {
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)

			err := suite.coordinator.SetChannelClosed(suite.chainA, suite.chainB, channelA)
			suite.Require().NoError(err)
		}, false},
		{"packet destination port ≠ channel counterparty port", func() {
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			// use wrong port for dest
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, ibctesting.InvalidID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
		}, false},
		{"packet destination channel ID ≠ channel counterparty channel ID", func() {
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			// use wrong channel for dest
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, ibctesting.InvalidID, timeoutHeight, disabledTimeoutTimestamp)
		}, false},
		{"connection not found", func() {
			channelA := ibctesting.TestChannel{PortID: portID, ID: channelIDA}
			channelB := ibctesting.TestChannel{PortID: portID, ID: channelIDB}
			// pass channel check
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetChannel(
				suite.chainB.GetContext(),
				channelB.PortID, channelB.ID,
				types.NewChannel(types.OPEN, types.ORDERED, types.NewCounterparty(channelA.PortID, channelA.ID), []string{connIDB}, channelB.Version),
			)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
		}, false},
		{"connection not OPEN", func() {
			clientA, clientB := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
			// connection on chainA is in INIT
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			channelA := connA.NextTestChannel(ibctesting.TransferPort)
			channelB := connB.NextTestChannel(ibctesting.TransferPort)
			// pass channel check
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetChannel(
				suite.chainA.GetContext(),
				channelA.PortID, channelA.ID,
				types.NewChannel(types.OPEN, types.ORDERED, types.NewCounterparty(channelB.PortID, channelB.ID), []string{connA.ID}, channelA.Version),
			)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
		}, false},
		{"packet hasn't been sent", func() {
			// packet commitment never written
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
		}, false},
		{"packet ack verification failed", func() {
			// ack never written
			_, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)

			// create packet commitment
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
		}, false},
		{"next ack sequence not found", func() {
			_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			channelA := connA.NextTestChannel(ibctesting.TransferPort)
			channelB := connB.NextTestChannel(ibctesting.TransferPort)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			// manually creating channel prevents next sequence acknowledgement from being set
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetChannel(
				suite.chainA.GetContext(),
				channelA.PortID, channelA.ID,
				types.NewChannel(types.OPEN, types.ORDERED, types.NewCounterparty(channelB.PortID, channelB.ID), []string{connA.ID}, channelA.Version),
			)
			// manually set packet commitment
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainA.GetContext(), channelA.PortID, channelA.ID, packet.GetSequence(), ibctesting.TestHash)

			// manually set packet acknowledgement
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(suite.chainB.GetContext(), channelB.PortID, channelB.ID, packet.GetSequence(), ibctesting.TestHash)
		}, false},
		{"next ack sequence mismatch", func() {
			clientA, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.ORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			// create packet commitment
			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			// create packet acknowledgement
			err = suite.coordinator.ReceiveExecuted(suite.chainB, suite.chainA, packet, clientA)
			suite.Require().NoError(err)

			// set next sequence ack wrong
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceAck(suite.chainA.GetContext(), channelA.PortID, channelA.ID, 10)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest() // reset
			tc.malleate()

			packetKey := host.KeyPacketAcknowledgement(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
			proof, proofHeight := suite.chainB.QueryProof(packetKey)

			err := suite.chainA.App.IBCKeeper.ChannelKeeper.AcknowledgePacket(suite.chainA.GetContext(), packet, ack, proof, proofHeight)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestAcknowledgementExectued verifies that packet commitments are deleted on chainA
// after capabilities are verified.
func (suite *KeeperTestSuite) TestAcknowledgementExecuted() {
	var (
		packet  types.Packet
		chanCap *capabilitytypes.Capability
	)

	testCases := []testCase{
		{"success ORDERED", func() {
			clientA, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.ORDERED)

			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)

			// create packet commitment
			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			// create packet acknowledgement
			err = suite.coordinator.ReceiveExecuted(suite.chainB, suite.chainA, packet, clientA)
			suite.Require().NoError(err)

			chanCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)

		}, true},
		{"success UNORDERED", func() {
			// setup uses an UNORDERED channel
			clientA, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)

			// create packet commitment
			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			// create packet acknowledgement
			err = suite.coordinator.ReceiveExecuted(suite.chainB, suite.chainA, packet, clientA)
			suite.Require().NoError(err)

			chanCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, true},
		{"channel not found", func() {
			// use wrong channel naming
			_, _, _, _, _, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			packet = types.NewPacket(validPacketData, 1, ibctesting.InvalidID, ibctesting.InvalidID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
		}, false},
		{"incorrect capability", func() {
			_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)

			chanCap = capabilitytypes.NewCapability(100)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest() // reset

			tc.malleate()

			err := suite.chainA.App.IBCKeeper.ChannelKeeper.AcknowledgementExecuted(suite.chainA.GetContext(), chanCap, packet)
			pc := suite.chainA.App.IBCKeeper.ChannelKeeper.GetPacketCommitment(suite.chainA.GetContext(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())

			channelA, _ := suite.chainA.App.IBCKeeper.ChannelKeeper.GetChannel(suite.chainA.GetContext(), packet.GetSourcePort(), packet.GetSourceChannel())
			sequenceAck, _ := suite.chainA.App.IBCKeeper.ChannelKeeper.GetNextSequenceAck(suite.chainA.GetContext(), packet.GetSourcePort(), packet.GetSourceChannel())

			if tc.expPass {
				suite.NoError(err)
				suite.Nil(pc)

				if channelA.Ordering == types.ORDERED {
					suite.Require().Equal(packet.GetSequence()+1, sequenceAck, "sequence not incremented in ordered channel")
				} else {
					suite.Require().Equal(uint64(1), sequenceAck, "sequence incremented for UNORDERED channel")
				}
			} else {
				suite.Error(err)
			}
		})
	}
}
