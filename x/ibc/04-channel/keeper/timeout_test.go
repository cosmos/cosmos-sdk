package keeper_test

import (
	"fmt"

	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

func (suite *KeeperTestSuite) TestTimeoutPacket() {
	counterparty := types.NewCounterparty(testPort2, testChannel2)
	var (
		packet      types.Packet
		nextSeqRecv uint64
		proofHeight uint64 = 3
	)

	testCases := []testCase{
		{"success", func() {
			nextSeqRecv = 1
			proofHeight = 1
			packet = types.NewPacket(newMockTimeoutPacket(1), 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.createClient(testClientID1)
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.UNORDERED, testConnectionID1)
			suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort1, testChannel1, 2, types.CommitPacket(packet.Data))
		}, true},
		{"channel not found", func() {}, false},
		{"channel not open", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.CLOSED, exported.ORDERED, testConnectionID1)
		}, false},
		{"packet source port ≠ channel counterparty port", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.createChannel(testPort1, testChannel1, testPort3, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
		}, false},
		{"packet source channel ID ≠ channel counterparty channel ID", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel3, exported.OPEN, exported.ORDERED, testConnectionID1)
		}, false},
		{"connection not found", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
		}, false},
		{"timeout", func() {
			proofHeight = 1
			packet = types.NewPacket(mockSuccessPacket{}, 10, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
		}, false},
		{"packet already received ", func() {
			nextSeqRecv = 2
			proofHeight = 100
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
		}, false},
		{"packet hasn't been sent", func() {
			nextSeqRecv = 1
			proofHeight = 100
			packet = types.NewPacket(mockSuccessPacket{}, 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
		}, false},
		{"next seq receive verification failed", func() {
			nextSeqRecv = 1
			proofHeight = 100
			packet = types.NewPacket(mockSuccessPacket{}, 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
			suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort1, testChannel1, 2, types.CommitPacket(packet.Data))
		}, false},
		{"packet ack verification failed", func() {
			nextSeqRecv = 1
			proofHeight = 100
			packet = types.NewPacket(mockSuccessPacket{}, 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.UNORDERED, testConnectionID1)
			suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort1, testChannel1, 2, types.CommitPacket(packet.Data))
		}, false},
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest() // reset
			tc.malleate()

			if tc.expPass {
				packetOut, err := suite.app.IBCKeeper.ChannelKeeper.TimeoutPacket(suite.ctx, packet, ibctypes.ValidProof{}, proofHeight, nextSeqRecv)
				suite.Require().NoError(err)
				suite.Require().NotNil(packetOut)
			} else {
				packetOut, err := suite.app.IBCKeeper.ChannelKeeper.TimeoutPacket(suite.ctx, packet, ibctypes.InvalidProof{}, proofHeight, nextSeqRecv)
				suite.Require().Error(err)
				suite.Require().Nil(packetOut)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestTimeoutExecuted() {
	var packet types.Packet

	testCases := []testCase{
		{"success ORDERED", func() {
			packet = types.NewPacket(newMockTimeoutPacket(3), 1, testPort1, testChannel1, testPort2, testChannel2)
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
		}, true},
		{"channel not found", func() {}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest() // reset
			tc.malleate()

			err := suite.app.IBCKeeper.ChannelKeeper.TimeoutExecuted(suite.ctx, packet)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestTimeoutOnClose() {
	counterparty := types.NewCounterparty(testPort2, testChannel2)
	var (
		packet      types.Packet
		nextSeqRecv uint64
		proofHeight uint64            = 1
		proof       commitment.ProofI = ibctypes.InvalidProof{}
		proofClosed commitment.ProofI = ibctypes.InvalidProof{}
	)

	testCases := []testCase{
		{"success", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.createClient(testClientID1)
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.UNORDERED, testConnectionID1)
			suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort1, testChannel1, 2, types.CommitPacket(packet.Data))
			suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.ctx, testPort1, testChannel1, nextSeqRecv)
		}, true},
		{"channel not found", func() {}, false},
		{"packet dest port ≠ channel counterparty port", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.createChannel(testPort1, testChannel1, testPort3, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
		}, false},
		{"packet dest channel ID ≠ channel counterparty channel ID", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel3, exported.OPEN, exported.ORDERED, testConnectionID1)
		}, false},
		{"connection not found", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
		}, false},
		{"packet hasn't been sent", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
		}, false},
		{"channel verification failed", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.createClient(testClientID1)
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.UNORDERED, testConnectionID1)
			suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort1, testChannel1, 2, types.CommitPacket(packet.Data))
			suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.ctx, testPort1, testChannel1, nextSeqRecv)
		}, false},
		{"next seq receive verification failed", func() {
			proofClosed = ibctypes.ValidProof{}
			packet = types.NewPacket(mockSuccessPacket{}, 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.createClient(testClientID1)
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
			suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort1, testChannel1, 2, types.CommitPacket(packet.Data))
			suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.ctx, testPort1, testChannel1, nextSeqRecv)
		}, false},
		{"packet ack verification failed", func() {
			proofClosed = ibctypes.ValidProof{}
			packet = types.NewPacket(mockSuccessPacket{}, 2, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.createClient(testClientID1)
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.UNORDERED, testConnectionID1)
			suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort1, testChannel1, 2, types.CommitPacket(packet.Data))
			suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.ctx, testPort1, testChannel1, nextSeqRecv)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest() // reset
			tc.malleate()

			if tc.expPass {
				packetOut, err := suite.app.IBCKeeper.ChannelKeeper.TimeoutOnClose(suite.ctx, packet, ibctypes.ValidProof{}, ibctypes.ValidProof{}, proofHeight, nextSeqRecv)
				suite.Require().NoError(err)
				suite.Require().NotNil(packetOut)
			} else {
				packetOut, err := suite.app.IBCKeeper.ChannelKeeper.TimeoutOnClose(suite.ctx, packet, proof, proofClosed, proofHeight, nextSeqRecv)
				suite.Require().Error(err)
				suite.Require().Nil(packetOut)
			}
		})
	}

}

type mockTimeoutPacket struct {
	timeoutHeight uint64
}

func newMockTimeoutPacket(timeoutHeight uint64) mockTimeoutPacket {
	return mockTimeoutPacket{timeoutHeight}
}

// GetBytes returns the serialised packet data (without timeout)
func (mp mockTimeoutPacket) GetBytes() []byte { return []byte("THIS IS A TIMEOUT PACKET") }

// GetTimeoutHeight returns the timeout height defined specifically for each packet data type instance
func (mp mockTimeoutPacket) GetTimeoutHeight() uint64 { return mp.timeoutHeight }

// ValidateBasic validates basic properties of the packet data, implements sdk.Msg
func (mp mockTimeoutPacket) ValidateBasic() error { return nil }

// Type returns human readable identifier, implements sdk.Msg
func (mp mockTimeoutPacket) Type() string { return "mock/packet/success" }
