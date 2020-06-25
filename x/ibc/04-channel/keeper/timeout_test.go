package keeper_test

import (
	"fmt"

	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

// TestTimeoutPacket test the TimeoutPacket call on chainA by ensuring the timeout has passed
// on chainB, but that no ack has been written yet. Test cases expected to reach proof
// verification must specify which proof to use using the ordered bool.
func (suite *KeeperTestSuite) TestTimeoutPacket() {
	var (
		packet      types.Packet
		nextSeqRecv uint64
		ordered     bool
	)

	testCases := []testCase{
		{"success: ORDERED", func() {
			nextSeqRecv = 1
			ordered = true

			clientA, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, types.ORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, uint64(suite.chainB.GetContext().BlockHeight()), uint64(suite.chainB.GetContext().BlockTime().UnixNano()))
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			// need to update chainA's client representing chainB to prove missing ack
			suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, clientexported.Tendermint)
		}, true},
		{"success: UNORDERED", func() {
			nextSeqRecv = 1
			ordered = false

			clientA, clientB, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, uint64(suite.chainB.GetContext().BlockHeight()), disabledTimeoutTimestamp)
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			// need to update chainA's client representing chainB to prove missing ack
			suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, clientexported.Tendermint)
		}, true},
		{"channel not found", func() {
			// use wrong channel naming
			_, _, _, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelB := connB.Channels[0]
			packet = types.NewPacket(validPacketData, 1, "invalidport", "invalidchannel", channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
		}, false},
		{"channel not open", func() {
			_, _, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)

			err := suite.coordinator.SetChannelClosed(suite.chainA, suite.chainB, channelA)
			suite.Require().NoError(err)
		}, false},
		{"packet destination port ≠ channel counterparty port", func() {
			_, _, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			// use wrong port for dest
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, "invalidport", channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
		}, false},
		{"packet destination channel ID ≠ channel counterparty channel ID", func() {
			_, _, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			// use wrong channel for dest
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, "invalidchannel", timeoutHeight, disabledTimeoutTimestamp)
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
		}, false},
		{"timeout", func() {
			_, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, types.ORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
		}, false},
		{"packet already received ", func() {
			ordered = true
			nextSeqRecv = 2

			_, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, types.ORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
		}, false},
		{"packet hasn't been sent", func() {
			_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, types.ORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
		}, false},
		{"next seq receive verification failed", func() {
			nextSeqRecv = 1
			// set ordered to false resulting in wrong proof provided
			ordered = false

			clientA, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, types.ORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, clientexported.Tendermint)
		}, false},
		{"packet ack verification failed", func() {
			nextSeqRecv = 1
			// set ordered to true resulting in wrong proof provided
			ordered = true

			clientA, clientB, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, uint64(suite.chainB.GetContext().BlockHeight()), disabledTimeoutTimestamp)
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, clientexported.Tendermint)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			var (
				proof       []byte
				proofHeight uint64
			)

			suite.SetupTest() // reset
			nextSeqRecv = 1
			tc.malleate()

			orderedPacketKey := host.KeyNextSequenceRecv(packet.GetDestPort(), packet.GetDestChannel())
			unorderedPacketKey := host.KeyPacketAcknowledgement(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())

			if ordered {
				proof, proofHeight = suite.chainB.QueryProof(orderedPacketKey)
			} else {
				proof, proofHeight = suite.chainB.QueryProof(unorderedPacketKey)
			}

			_, err := suite.chainA.App.IBCKeeper.ChannelKeeper.TimeoutPacket(suite.chainA.GetContext(), packet, proof, proofHeight+1, nextSeqRecv)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestTimeoutExectued verifies that packet commitments are deleted on chainA after the
// channel capabilities are verified.
func (suite *KeeperTestSuite) TestTimeoutExecuted() {
	var (
		packet  types.Packet
		chanCap *capabilitytypes.Capability
	)

	testCases := []testCase{
		{"success ORDERED", func() {
			_, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, types.ORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, uint64(suite.chainB.GetContext().BlockHeight()), uint64(suite.chainB.GetContext().BlockTime().UnixNano()))
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)

			chanCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, true},
		{"channel not found", func() {
			// use wrong channel naming
			_, _, _, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelB := connB.Channels[0]
			packet = types.NewPacket(validPacketData, 1, "invalidport", "invalidchannel", channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
		}, false},
		{"incorrect capability", func() {
			_, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, types.ORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, uint64(suite.chainB.GetContext().BlockHeight()), uint64(suite.chainB.GetContext().BlockTime().UnixNano()))
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)

			chanCap = capabilitytypes.NewCapability(100)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest() // reset

			tc.malleate()

			err := suite.chainA.App.IBCKeeper.ChannelKeeper.TimeoutExecuted(suite.chainA.GetContext(), chanCap, packet)
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

// TestTimeoutOnClose tests the call TimeoutOnClose on chainA by closing the corresponding
// channel on chainB after the packet commitment has been created.
func (suite *KeeperTestSuite) TestTimeoutOnClose() {
	var (
		packet      types.Packet
		chanCap     *capabilitytypes.Capability
		nextSeqRecv uint64
		ordered     bool
	)

	testCases := []testCase{
		{"success: ORDERED", func() {
			nextSeqRecv = 1
			ordered = true
			clientA, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, types.ORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, uint64(suite.chainB.GetContext().BlockHeight()), uint64(suite.chainB.GetContext().BlockTime().UnixNano()))
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.coordinator.SetChannelClosed(suite.chainB, suite.chainA, channelB)
			// need to update chainA's client representing chainB to prove missing ack
			suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, clientexported.Tendermint)

			chanCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, true},
		{"success: UNORDERED", func() {
			nextSeqRecv = 1
			ordered = false
			clientA, clientB, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, uint64(suite.chainB.GetContext().BlockHeight()), disabledTimeoutTimestamp)
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.coordinator.SetChannelClosed(suite.chainB, suite.chainA, channelB)
			// need to update chainA's client representing chainB to prove missing ack
			suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, clientexported.Tendermint)

			chanCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, true},
		{"channel not found", func() {
			// use wrong channel naming
			_, _, _, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelB := connB.Channels[0]
			packet = types.NewPacket(validPacketData, 1, "invalidport", "invalidchannel", channelB.PortID, channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
		}, false},
		{"packet dest port ≠ channel counterparty port", func() {
			_, _, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			// use wrong port for dest
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, "invalidport", channelB.ID, timeoutHeight, disabledTimeoutTimestamp)
			chanCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"packet dest channel ID ≠ channel counterparty channel ID", func() {
			_, _, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			// use wrong channel for dest
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, "invalidchannel", timeoutHeight, disabledTimeoutTimestamp)
			chanCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
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
			// create chancap
			suite.chainA.CreateChannelCapability(channelA.PortID, channelA.ID)
			chanCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"packet hasn't been sent", func() {
			_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, types.ORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, uint64(suite.chainB.GetContext().BlockHeight()), uint64(suite.chainB.GetContext().BlockTime().UnixNano()))
			chanCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"channel verification failed", func() {
			ordered = true
			_, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, types.ORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, uint64(suite.chainB.GetContext().BlockHeight()), uint64(suite.chainB.GetContext().BlockTime().UnixNano()))
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			chanCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"next seq receive verification failed", func() {
			// set ordered to false providing the wrong proof for ORDERED case
			ordered = false
			clientA, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			channelA, channelB := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, types.ORDERED)
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, uint64(suite.chainB.GetContext().BlockHeight()), uint64(suite.chainB.GetContext().BlockTime().UnixNano()))
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.coordinator.SetChannelClosed(suite.chainA, suite.chainB, channelA)
			suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, clientexported.Tendermint)
			chanCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"packet ack verification failed", func() {
			// set ordered to true providing the wrong proof for UNORDERED case
			ordered = true
			clientA, clientB, connA, connB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channelA := connA.Channels[0]
			channelB := connB.Channels[0]
			packet = types.NewPacket(validPacketData, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, uint64(suite.chainB.GetContext().BlockHeight()), disabledTimeoutTimestamp)
			suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.coordinator.SetChannelClosed(suite.chainA, suite.chainB, channelA)
			suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, clientexported.Tendermint)
			chanCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			var proof []byte

			suite.SetupTest() // reset
			tc.malleate()

			channelKey := host.KeyChannel(packet.GetDestPort(), packet.GetDestChannel())
			unorderedPacketKey := host.KeyPacketAcknowledgement(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
			orderedPacketKey := host.KeyNextSequenceRecv(packet.GetDestPort(), packet.GetDestChannel())

			proofClosed, proofHeight := suite.chainB.QueryProof(channelKey)

			if ordered {
				proof, _ = suite.chainB.QueryProof(orderedPacketKey)
			} else {
				proof, _ = suite.chainB.QueryProof(unorderedPacketKey)
			}

			_, err := suite.chainA.App.IBCKeeper.ChannelKeeper.TimeoutOnClose(suite.chainA.GetContext(), chanCap, packet, proof, proofClosed, proofHeight+1, nextSeqRecv)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}

}
