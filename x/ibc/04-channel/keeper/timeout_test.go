package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/capability"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

func (suite *KeeperTestSuite) TestTimeoutPacket() {
	counterparty := types.NewCounterparty(testPort2, testChannel2)
	packetKey := ibctypes.KeyPacketAcknowledgement(testPort2, testChannel2, 2)
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
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, ibctypes.OPEN)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, ibctypes.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, ibctypes.OPEN, ibctypes.UNORDERED, testConnectionIDA)
			suite.chainA.createChannel(testPort2, testChannel2, testPort1, testChannel1, ibctypes.OPEN, ibctypes.UNORDERED, testConnectionIDB)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 2, types.CommitPacket(packet))
		}, true},
		{"channel not found", func() {}, false},
		{"channel not open", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, ibctypes.CLOSED, ibctypes.ORDERED, testConnectionIDA)
		}, false},
		{"packet source port ≠ channel counterparty port", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createChannel(testPort1, testChannel1, testPort3, testChannel2, ibctypes.OPEN, ibctypes.ORDERED, testConnectionIDA)
		}, false},
		{"packet source channel ID ≠ channel counterparty channel ID", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel3, ibctypes.OPEN, ibctypes.ORDERED, testConnectionIDA)
		}, false},
		{"connection not found", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, ibctypes.OPEN, ibctypes.ORDERED, testConnectionIDA)
		}, false},
		{"timeout", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 10, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, ibctypes.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, ibctypes.OPEN, ibctypes.ORDERED, testConnectionIDA)
		}, false},
		{"packet already received ", func() {
			nextSeqRecv = 2
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, ibctypes.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, ibctypes.OPEN, ibctypes.ORDERED, testConnectionIDA)
		}, false},
		{"packet hasn't been sent", func() {
			nextSeqRecv = 1
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, ibctypes.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, ibctypes.OPEN, ibctypes.ORDERED, testConnectionIDA)
		}, false},
		{"next seq receive verification failed", func() {
			nextSeqRecv = 1
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, ibctypes.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, ibctypes.OPEN, ibctypes.ORDERED, testConnectionIDA)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 2, types.CommitPacket(packet))
		}, false},
		{"packet ack verification failed", func() {
			nextSeqRecv = 1
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, ibctypes.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, ibctypes.OPEN, ibctypes.UNORDERED, testConnectionIDA)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 2, types.CommitPacket(packet))
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
				packetOut, err := suite.chainB.App.IBCKeeper.ChannelKeeper.TimeoutPacket(ctx, packet, proof, proofHeight+1, nextSeqRecv)
				suite.Require().NoError(err)
				suite.Require().NotNil(packetOut)
			} else {
				packetOut, err := suite.chainB.App.IBCKeeper.ChannelKeeper.TimeoutPacket(ctx, packet, ibctypes.InvalidProof{}, proofHeight+1, nextSeqRecv)
				suite.Require().Error(err)
				suite.Require().Nil(packetOut)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestTimeoutExecuted() {
	var packet types.Packet

	var chanCap *capability.Capability
	testCases := []testCase{
		{"success ORDERED", func() {
			packet = types.NewPacket(newMockTimeoutPacket().GetBytes(), 1, testPort1, testChannel1, testPort2, testChannel2, timeoutHeight, disabledTimeoutTimestamp)
			suite.chainA.createChannel(testPort1, testChannel1, testPort2, testChannel2, ibctypes.OPEN, ibctypes.ORDERED, testConnectionIDA)
		}, true},
		{"channel not found", func() {}, false},
		{"incorrect capability", func() {
			packet = types.NewPacket(newMockTimeoutPacket().GetBytes(), 1, testPort1, testChannel1, testPort2, testChannel2, timeoutHeight, disabledTimeoutTimestamp)
			suite.chainA.createChannel(testPort1, testChannel1, testPort2, testChannel2, ibctypes.OPEN, ibctypes.ORDERED, testConnectionIDA)
			chanCap = capability.NewCapability(100)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest() // reset

			var err error
			chanCap, err = suite.chainA.App.ScopedIBCKeeper.NewCapability(
				suite.chainA.GetContext(), ibctypes.ChannelCapabilityPath(testPort1, testChannel1),
			)
			suite.Require().NoError(err, "could not create capability")

			tc.malleate()

			err = suite.chainA.App.IBCKeeper.ChannelKeeper.TimeoutExecuted(suite.chainA.GetContext(), chanCap, packet)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestTimeoutOnClose() {
	channelKey := ibctypes.KeyChannel(testPort2, testChannel2)
	packetKey := ibctypes.KeyPacketAcknowledgement(testPort1, testChannel1, 2)
	counterparty := types.NewCounterparty(testPort2, testChannel2)
	var (
		packet      types.Packet
		nextSeqRecv uint64
	)

	testCases := []testCase{
		{"success", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, ibctypes.OPEN)
			suite.chainA.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, ibctypes.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, ibctypes.OPEN, ibctypes.UNORDERED, testConnectionIDA)
			suite.chainA.createChannel(testPort2, testChannel2, testPort1, testChannel1, ibctypes.CLOSED, ibctypes.UNORDERED, testConnectionIDB) // channel on chainA is closed
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 2, types.CommitPacket(packet))
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.chainB.GetContext(), testPort1, testChannel1, nextSeqRecv)
		}, true},
		{"channel not found", func() {}, false},
		{"packet dest port ≠ channel counterparty port", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createChannel(testPort1, testChannel1, testPort3, testChannel2, ibctypes.OPEN, ibctypes.ORDERED, testConnectionIDA)
		}, false},
		{"packet dest channel ID ≠ channel counterparty channel ID", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel3, ibctypes.OPEN, ibctypes.ORDERED, testConnectionIDA)
		}, false},
		{"connection not found", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, ibctypes.OPEN, ibctypes.ORDERED, testConnectionIDA)
		}, false},
		{"packet hasn't been sent", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, ibctypes.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, ibctypes.OPEN, ibctypes.ORDERED, testConnectionIDA)
		}, false},
		{"channel verification failed", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, ibctypes.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, ibctypes.OPEN, ibctypes.UNORDERED, testConnectionIDA)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 2, types.CommitPacket(packet))
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.chainB.GetContext(), testPort1, testChannel1, nextSeqRecv)
		}, false},
		{"next seq receive verification failed", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, ibctypes.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, ibctypes.OPEN, ibctypes.ORDERED, testConnectionIDA)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 2, types.CommitPacket(packet))
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.chainB.GetContext(), testPort1, testChannel1, nextSeqRecv)
		}, false},
		{"packet ack verification failed", func() {
			packet = types.NewPacket(mockSuccessPacket{}.GetBytes(), 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), timeoutHeight, disabledTimeoutTimestamp)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, ibctypes.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, ibctypes.OPEN, ibctypes.UNORDERED, testConnectionIDA)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 2, types.CommitPacket(packet))
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.chainB.GetContext(), testPort1, testChannel1, nextSeqRecv)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest() // reset
			tc.malleate()

			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
			proofClosed, proofHeight := queryProof(suite.chainA, channelKey)
			proofAckAbsence, _ := queryProof(suite.chainA, packetKey)

			ctx := suite.chainB.GetContext()
			if tc.expPass {
				packetOut, err := suite.chainB.App.IBCKeeper.ChannelKeeper.TimeoutOnClose(ctx, packet, proofAckAbsence, proofClosed, proofHeight+1, nextSeqRecv)
				suite.Require().NoError(err)
				suite.Require().NotNil(packetOut)
			} else {
				packetOut, err := suite.chainB.App.IBCKeeper.ChannelKeeper.TimeoutOnClose(ctx, packet, invalidProof{}, invalidProof{}, proofHeight+1, nextSeqRecv)
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
