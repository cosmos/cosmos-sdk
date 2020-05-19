package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/capability"
	transfertypes "github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

func (suite *KeeperTestSuite) TestSendPacket() {
	counterparty := types.NewCounterparty(testPort2, testChannel2)
	var packet exported.PacketI

	var channelCap *capability.Capability
	testCases := []testCase{
		{"success", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.chainB.GetContext(), testPort1, testChannel1, 1)
		}, true},
		{"packet basic validation failed", func() {
			packet = types.NewPacket(mockFailPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
		}, false},
		{"channel not found", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
		}, false},
		{"channel closed", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.CLOSED, types.ORDERED, testConnectionIDA)
		}, false},
		{"packet dest port ≠ channel counterparty port", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, testPort3, counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"packet dest channel ID ≠ channel counterparty channel ID", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), testChannel3, timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"connection not found", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"connection is UNINITIALIZED", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.UNINITIALIZED)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"client state not found", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"timeout height passed", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			commitNBlocks(suite.chainB, 10)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"timeout timestamp passed", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), disabledTimeoutHeight, timeoutTimestamp)
			commitBlockWithNewTimestamp(suite.chainB, timeoutTimestamp)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"next sequence send not found", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"next sequence wrong", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.chainB.GetContext(), testPort1, testChannel1, 5)
		}, false},
		{"channel capability not found", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.chainB.GetContext(), testPort1, testChannel1, 1)
			channelCap = capability.NewCapability(3)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest() // reset

			var err error
			channelCap, err = suite.chainB.App.ScopedIBCKeeper.NewCapability(suite.chainB.GetContext(), host.ChannelCapabilityPath(testPort1, testChannel1))
			suite.Require().Nil(err, "could not create capability")

			tc.malleate()

			err = suite.chainB.App.IBCKeeper.ChannelKeeper.SendPacket(suite.chainB.GetContext(), channelCap, packet)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}

}

func (suite *KeeperTestSuite) TestRecvPacket() {
	counterparty := types.NewCounterparty(testPort1, testChannel1)
	packetKey := host.KeyPacketCommitment(testPort2, testChannel2, 1)

	var packet exported.PacketI

	testCases := []testCase{
		{"success", func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainA.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, connection.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
			suite.chainA.createChannel(testPort2, testChannel2, testPort1, testChannel1, types.OPEN, types.ORDERED, testConnectionIDB)
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.chainA.GetContext(), testPort2, testChannel2, 1)
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort2, testChannel2, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainA.GetContext(), testPort2, testChannel2, 1, types.CommitPacket(packet))

		}, true},
		{"channel not found", func() {}, false},
		{"channel not open", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createChannel(testPort2, testChannel2, testPort1, testChannel1, types.INIT, types.ORDERED, testConnectionIDA)
		}, false},
		{"packet source port ≠ channel counterparty port", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createChannel(testPort2, testChannel2, testPort3, testChannel1, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"packet source channel ID ≠ channel counterparty channel ID", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createChannel(testPort2, testChannel2, testPort1, testChannel3, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"connection not found", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createChannel(testPort2, testChannel2, testPort1, testChannel1, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"connection not OPEN", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.INIT)
			suite.chainB.createChannel(testPort2, testChannel2, testPort1, testChannel1, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"timeout height passed", func() {
			commitNBlocks(suite.chainB, 10)
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainB.createChannel(testPort2, testChannel2, testPort1, testChannel1, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"validation failed", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainB.createChannel(testPort2, testChannel2, testPort1, testChannel1, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest() // reset
			tc.malleate()

			ctx := suite.chainB.GetContext()

			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
			proof, proofHeight := queryProof(suite.chainA, packetKey)

			var err error
			if tc.expPass {
				_, err = suite.chainB.App.IBCKeeper.ChannelKeeper.RecvPacket(ctx, packet, proof, proofHeight+1)
				suite.Require().NoError(err)
			} else {
				packet, err = suite.chainB.App.IBCKeeper.ChannelKeeper.RecvPacket(ctx, packet, proof, proofHeight)
				suite.Require().Error(err)
			}
		})
	}

}

func (suite *KeeperTestSuite) TestPacketExecuted() {
	counterparty := types.NewCounterparty(testPort2, testChannel2)
	var packet types.Packet

	var channelCap *capability.Capability
	testCases := []testCase{
		{"success: UNORDERED", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainA.createChannel(testPort2, testChannel2, testPort1, testChannel1, types.OPEN, types.UNORDERED, testConnectionIDA)
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.chainA.GetContext(), testPort2, testChannel2, 1)
		}, true},
		{"success: ORDERED", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainA.createChannel(testPort2, testChannel2, testPort1, testChannel1, types.OPEN, types.ORDERED, testConnectionIDA)
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.chainA.GetContext(), testPort2, testChannel2, 1)
		}, true},
		{"channel not found", func() {}, false},
		{"channel not OPEN", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainA.createChannel(testPort2, testChannel2, testPort1, testChannel1, types.CLOSED, types.ORDERED, testConnectionIDA)
		}, false},
		{"next sequence receive not found", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainA.createChannel(testPort2, testChannel2, testPort1, testChannel1, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"packet sequence ≠ next sequence receive", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainA.createChannel(testPort2, testChannel2, testPort1, testChannel1, types.OPEN, types.ORDERED, testConnectionIDA)
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.chainA.GetContext(), testPort2, testChannel2, 5)
		}, false},
		{"capability not found", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainA.createChannel(testPort2, testChannel2, testPort1, testChannel1, types.OPEN, types.UNORDERED, testConnectionIDA)
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.chainA.GetContext(), testPort2, testChannel2, 1)
			channelCap = capability.NewCapability(3)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest() // reset

			var err error
			channelCap, err = suite.chainA.App.ScopedIBCKeeper.NewCapability(suite.chainA.GetContext(), host.ChannelCapabilityPath(testPort2, testChannel2))
			suite.Require().NoError(err, "could not create capability")

			tc.malleate()

			err = suite.chainA.App.IBCKeeper.ChannelKeeper.PacketExecuted(suite.chainA.GetContext(), channelCap, packet, mockSuccessPacket{}.GetBytes())

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestAcknowledgePacket() {
	counterparty := types.NewCounterparty(testPort2, testChannel2)
	var packet types.Packet
	packetKey := host.KeyPacketAcknowledgement(testPort2, testChannel2, 1)

	ack := transfertypes.FungibleTokenPacketAcknowledgement{
		Success: true,
	}.GetBytes()

	testCases := []testCase{
		{"success", func() {
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

func (suite *KeeperTestSuite) TestCleanupPacket() {
	counterparty := types.NewCounterparty(testPort2, testChannel2)
	packetKey := host.KeyPacketAcknowledgement(testPort2, testChannel2, 1)
	var (
		packet      types.Packet
		nextSeqRecv uint64
	)

	ack := []byte("ack")

	testCases := []testCase{
		{"success", func() {
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
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.UNORDERED, testConnectionIDA)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 1, types.CommitPacket(packet))
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest() // reset
			tc.malleate()

			ctx := suite.chainB.GetContext()

			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
			proof, proofHeight := queryProof(suite.chainA, packetKey)

			if tc.expPass {
				packetOut, err := suite.chainB.App.IBCKeeper.ChannelKeeper.CleanupPacket(ctx, packet, proof, proofHeight+1, nextSeqRecv, ack)
				suite.Require().NoError(err)
				suite.Require().NotNil(packetOut)
			} else {
				packetOut, err := suite.chainB.App.IBCKeeper.ChannelKeeper.CleanupPacket(ctx, packet, proof, proofHeight, nextSeqRecv, ack)
				suite.Require().Error(err)
				suite.Require().Nil(packetOut)
			}
		})
	}
}
