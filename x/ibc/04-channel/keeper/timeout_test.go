package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/capability"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

func (suite *KeeperTestSuite) TestTimeoutPacket() {
	counterparty := types.NewCounterparty(testPort2, testChannel2)
	packetKey := host.KeyPacketAcknowledgement(testPort2, testChannel2, 2)
	var (
		packet      types.Packet
		nextSeqRecv uint64
	)

	testCases := []testCase{
		{"success", func() {
			nextSeqRecv = 1
			packet = types.NewPacket(newMockTimeoutPacket().GetBytes(), 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), 1, disabledTimeoutTimestamp)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainA.CreateConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, connection.OPEN)
			suite.chainB.CreateChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.UNORDERED, testConnectionIDA)
			suite.chainA.CreateChannel(testPort2, testChannel2, testPort1, testChannel1, types.OPEN, types.UNORDERED, testConnectionIDB)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 2, types.CommitPacket(packet))
		}, true},
		{"channel not found", func() {}, false},
		{"channel not open", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateChannel(testPort1, testChannel1, testPort2, testChannel2, types.CLOSED, types.ORDERED, testConnectionIDA)
		}, false},
		{"packet source port ≠ channel counterparty port", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateChannel(testPort1, testChannel1, testPort3, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"packet source channel ID ≠ channel counterparty channel ID", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateChannel(testPort1, testChannel1, testPort2, testChannel3, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"connection not found", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"timeout", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 10, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainB.CreateChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"packet already received ", func() {
			nextSeqRecv = 2
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainB.CreateChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"packet hasn't been sent", func() {
			nextSeqRecv = 1
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainB.CreateChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"next seq receive verification failed", func() {
			nextSeqRecv = 1
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainB.CreateChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 2, types.CommitPacket(packet))
		}, false},
		{"packet ack verification failed", func() {
			nextSeqRecv = 1
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainB.CreateChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.UNORDERED, testConnectionIDA)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 2, types.CommitPacket(packet))
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest() // reset
			tc.malleate()

			ctx := suite.chainB.GetContext()

			suite.chainB.UpdateClient(suite.chainA)
			suite.chainA.UpdateClient(suite.chainB)
			proof, proofHeight := suite.chainA.QueryProof(packetKey)

			if tc.expPass {
				packetOut, err := suite.chainB.App.IBCKeeper.ChannelKeeper.TimeoutPacket(ctx, packet, proof, proofHeight, nextSeqRecv)
				suite.Require().NoError(err)
				suite.Require().NotNil(packetOut)
			} else {
				packetOut, err := suite.chainB.App.IBCKeeper.ChannelKeeper.TimeoutPacket(ctx, packet, proof, proofHeight, nextSeqRecv)
				suite.Require().Error(err)
				suite.Require().Nil(packetOut)
			}
		})
	}
}

// TestTimeoutExectued verifies that packet commitments are deleted after
// capabilities are verified.
func (suite *KeeperTestSuite) TestTimeoutExecuted() {
	var (
		packet  types.Packet
		chanCap *capability.Capability
	)

	testCases := []testCase{
		{"success ORDERED", func() {
			packet = types.NewPacket(newMockTimeoutPacket().GetBytes(), 1, testPort1, testChannel1, testPort2, testChannel2, timeoutHeight, disabledTimeoutTimestamp)
			suite.chainA.CreateChannel(packet.GetSourcePort(), packet.GetSourceChannel(), testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainA.GetContext(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.Sequence, types.CommitPacket(packet))
		}, true},
		{"channel not found", func() {}, false},
		{"incorrect capability", func() {
			packet = types.NewPacket(newMockTimeoutPacket().GetBytes(), 1, testPort1, testChannel1, testPort2, testChannel2, timeoutHeight, disabledTimeoutTimestamp)
			suite.chainA.CreateChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
			chanCap = capability.NewCapability(100)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest() // reset

			var err error
			chanCap, err = suite.chainA.NewCapability(packet.GetSourcePort(), packet.GetSourceChannel())
			suite.Require().NoError(err, "could not create capability")

			tc.malleate()

			err = suite.chainA.App.IBCKeeper.ChannelKeeper.TimeoutExecuted(suite.chainA.GetContext(), chanCap, packet)
			pc := suite.chainA.App.IBCKeeper.ChannelKeeper.GetPacketCommitment(suite.chainA.GetContext(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.Sequence)

			if tc.expPass {
				suite.NoError(err)
				suite.Nil(pc)
			} else {
				suite.Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestTimeoutOnClose() {
	channelKey := host.KeyChannel(testPort2, testChannel2)
	unorderedPacketKey := host.KeyPacketAcknowledgement(testPort2, testChannel2, 2)
	orderedPacketKey := host.KeyNextSequenceRecv(testPort2, testChannel2)

	counterparty := types.NewCounterparty(testPort2, testChannel2)
	var (
		packet      types.Packet
		nextSeqRecv uint64
		ordered     bool
	)

	testCases := []testCase{
		{"success on ordered channel", func() {
			ordered = true
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainA.CreateConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, connection.OPEN)
			suite.chainB.CreateChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
			suite.chainA.CreateChannel(testPort2, testChannel2, testPort1, testChannel1, types.CLOSED, types.ORDERED, testConnectionIDB) // channel on chainA is closed
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 2, types.CommitPacket(packet))
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.chainA.GetContext(), testPort2, testChannel2, nextSeqRecv)
		}, true},
		{"success on unordered channel", func() {
			ordered = false
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainA.CreateConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, connection.OPEN)
			suite.chainB.CreateChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.UNORDERED, testConnectionIDA)
			suite.chainA.CreateChannel(testPort2, testChannel2, testPort1, testChannel1, types.CLOSED, types.UNORDERED, testConnectionIDB) // channel on chainA is closed
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 2, types.CommitPacket(packet))
		}, true},
		{"channel not found", func() {}, false},
		{"packet dest port ≠ channel counterparty port", func() {
			ordered = true
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateChannel(testPort1, testChannel1, testPort3, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"packet dest channel ID ≠ channel counterparty channel ID", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateChannel(testPort1, testChannel1, testPort2, testChannel3, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"connection not found", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"packet hasn't been sent", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainB.CreateChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
		}, false},
		{"channel verification failed", func() {
			ordered = false
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.CreateConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainB.CreateChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.UNORDERED, testConnectionIDA)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 2, types.CommitPacket(packet))
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.chainB.GetContext(), testPort1, testChannel1, nextSeqRecv)
		}, false},
		{"next seq receive verification failed", func() {
			ordered = true
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.CreateConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainB.CreateChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.ORDERED, testConnectionIDA)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 2, types.CommitPacket(packet))
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.chainB.GetContext(), testPort1, testChannel1, nextSeqRecv)
		}, false},
		{"packet ack verification failed", func() {
			ordered = false
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.CreateConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connection.OPEN)
			suite.chainB.CreateChannel(testPort1, testChannel1, testPort2, testChannel2, types.OPEN, types.UNORDERED, testConnectionIDA)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 2, types.CommitPacket(packet))
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.chainB.GetContext(), testPort1, testChannel1, nextSeqRecv)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			var proof commitmentexported.Proof

			suite.SetupTest() // reset
			tc.malleate()

			suite.chainB.UpdateClient(suite.chainA)
			suite.chainA.UpdateClient(suite.chainB)
			proofClosed, proofHeight := suite.chainA.QueryProof(channelKey)

			if ordered {
				proof, _ = suite.chainA.QueryProof(orderedPacketKey)
			} else {
				proof, _ = suite.chainA.QueryProof(unorderedPacketKey)
			}

			ctx := suite.chainB.GetContext()
			cap, err := suite.chainB.NewCapability(packet.GetSourcePort(), packet.GetSourceChannel())
			suite.Require().NoError(err)

			if tc.expPass {
				packetOut, err := suite.chainB.App.IBCKeeper.ChannelKeeper.TimeoutOnClose(ctx, cap, packet, proof, proofClosed, proofHeight, nextSeqRecv)
				suite.Require().NoError(err)
				suite.Require().NotNil(packetOut)
			} else {
				// switch the proofs to invalidate them
				packetOut, err := suite.chainB.App.IBCKeeper.ChannelKeeper.TimeoutOnClose(ctx, cap, packet, proofClosed, proof, proofHeight, nextSeqRecv)
				suite.Require().Error(err)
				suite.Require().Nil(packetOut)
			}
		})
	}

}

type mockTimeoutPacket struct{}

func newMockTimeoutPacket() mockTimeoutPacket {
	return mockTimeoutPacket{}
}

// GetBytes returns the serialised packet data (without timeout)
func (mp mockTimeoutPacket) GetBytes() []byte { return []byte("THIS IS A TIMEOUT PACKET") }
