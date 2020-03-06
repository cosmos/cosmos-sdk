package keeper_test

import (
	"fmt"

	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"

	transfertypes "github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

func (suite *KeeperTestSuite) TestSendPacket() {
	counterparty := types.NewCounterparty(testPort2, testChannel2)
	var packet exported.PacketI

	testCases := []testCase{
		{"success", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connectionexported.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionIDA)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.chainB.GetContext(), testPort1, testChannel1, 1)
		}, true},
		{"packet basic validation failed", func() {
			packet = types.NewPacket(mockFailPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
		}, false},
		{"channel not found", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
		}, false},
		{"channel closed", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.CLOSED, exported.ORDERED, testConnectionIDA)
		}, false},
		{"packet dest port ≠ channel counterparty port", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, testPort3, counterparty.GetChannelID())
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionIDA)
		}, false},
		{"packet dest channel ID ≠ channel counterparty channel ID", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), testChannel3)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionIDA)
		}, false},
		{"connection not found", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionIDA)
		}, false},
		{"connection is UNINITIALIZED", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connectionexported.UNINITIALIZED)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionIDA)
		}, false},
		{"client state not found", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connectionexported.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionIDA)
		}, false},
		{"timeout height passed", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			commitNBlocks(suite.chainB, 10)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connectionexported.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionIDA)
		}, false},
		{"next sequence send not found", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connectionexported.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionIDA)
		}, false},
		{"next sequence wrong", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connectionexported.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionIDA)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.chainB.GetContext(), testPort1, testChannel1, 5)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest() // reset
			tc.malleate()

			err := suite.chainB.App.IBCKeeper.ChannelKeeper.SendPacket(suite.chainB.GetContext(), packet)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}

}

func (suite *KeeperTestSuite) TestRecvPacket() {
	counterparty := types.NewCounterparty(testPort2, testChannel2)
	packetKey := ibctypes.KeyPacketCommitment(testPort1, testChannel1, 1)

	var packet exported.PacketI

	testCases := []testCase{
		{"success", func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connectionexported.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionIDA)
			suite.chainB.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.OPEN, exported.ORDERED, testConnectionIDA)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.chainB.GetContext(), testPort1, testChannel1, 1)
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 1, types.CommitPacket(packet.(types.Packet).Data))

			suite.chainB.updateClient(suite.chainA)
		}, true},
		{"channel not found", func() {}, false},
		{"channel not open", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.INIT, exported.ORDERED, testConnectionIDA)
		}, false},
		{"packet source port ≠ channel counterparty port", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createChannel(testPort2, testChannel2, testPort3, testChannel1, exported.OPEN, exported.ORDERED, testConnectionIDA)
		}, false},
		{"packet source channel ID ≠ channel counterparty channel ID", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createChannel(testPort2, testChannel2, testPort1, testChannel3, exported.OPEN, exported.ORDERED, testConnectionIDA)
		}, false},
		{"connection not found", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.OPEN, exported.ORDERED, testConnectionIDA)
		}, false},
		{"connection not OPEN", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connectionexported.INIT)
			suite.chainB.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.OPEN, exported.ORDERED, testConnectionIDA)
		}, false},
		{"timeout passed", func() {
			commitNBlocks(suite.chainB, 10)
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connectionexported.OPEN)
			suite.chainB.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.OPEN, exported.ORDERED, testConnectionIDA)
		}, false},
		{"validation failed", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connectionexported.OPEN)
			suite.chainB.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.OPEN, exported.ORDERED, testConnectionIDA)
		}, false},
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest() // reset
			tc.malleate()

			ctx := suite.chainB.GetContext()
			proof, proofHeight := queryProof(suite.chainA, packetKey)
			var err error
			if tc.expPass {
				packet, err = suite.chainB.App.IBCKeeper.ChannelKeeper.RecvPacket(ctx, packet, proof, proofHeight+1)
				suite.Require().NoError(err)
			} else {
				packet, err = suite.chainB.App.IBCKeeper.ChannelKeeper.RecvPacket(ctx, packet, ibctypes.InvalidProof{}, proofHeight)
				suite.Require().Error(err)
			}
		})
	}

}

func (suite *KeeperTestSuite) TestPacketExecuted() {
	counterparty := types.NewCounterparty(testPort2, testChannel2)
	var packet types.Packet

	testCases := []testCase{
		{"success: UNORDERED", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainA.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.OPEN, exported.UNORDERED, testConnectionIDA)
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.chainA.GetContext(), testPort2, testChannel2, 1)
		}, true},
		{"success: ORDERED", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainA.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.OPEN, exported.ORDERED, testConnectionIDA)
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.chainA.GetContext(), testPort2, testChannel2, 1)
		}, true},
		{"channel not found", func() {}, false},
		{"channel not OPEN", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainA.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.CLOSED, exported.ORDERED, testConnectionIDA)
		}, false},
		{"next sequence receive not found", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainA.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.OPEN, exported.ORDERED, testConnectionIDA)
		}, false},
		{"packet sequence ≠ next sequence receive", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainA.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.OPEN, exported.ORDERED, testConnectionIDA)
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.chainA.GetContext(), testPort2, testChannel2, 5)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest() // reset
			tc.malleate()

			err := suite.chainA.App.IBCKeeper.ChannelKeeper.PacketExecuted(suite.chainA.GetContext(), packet, mockSuccessPacket{})

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
	packetKey := ibctypes.KeyPacketCommitment(testPort1, testChannel1, 1)

	ack := transfertypes.AckDataTransfer{}

	testCases := []testCase{
		{"success", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connectionexported.OPEN)
			suite.chainA.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, connectionexported.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionIDA)
			suite.chainA.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.OPEN, exported.ORDERED, testConnectionIDB)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 1, types.CommitPacket(packet.Data))
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(suite.chainA.GetContext(), testPort2, testChannel2, 1, ack.GetBytes())
		}, true},
		{"channel not found", func() {}, false},
		{"channel not open", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.CLOSED, exported.ORDERED, testConnectionIDA)
		}, false},
		{"packet source port ≠ channel counterparty port", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createChannel(testPort1, testChannel1, testPort3, testChannel2, exported.OPEN, exported.ORDERED, testConnectionIDA)
		}, false},
		{"packet source channel ID ≠ channel counterparty channel ID", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel3, exported.OPEN, exported.ORDERED, testConnectionIDA)
		}, false},
		{"connection not found", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionIDA)
		}, false},
		{"connection not OPEN", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connectionexported.INIT)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionIDA)
		}, false},
		{"packet hasn't been sent", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connectionexported.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionIDA)
		}, false},
		{"packet ack verification failed", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connectionexported.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionIDA)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 1, types.CommitPacket(packet.Data))
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
			if tc.expPass {
				packetOut, err := suite.chainB.App.IBCKeeper.ChannelKeeper.AcknowledgePacket(ctx, packet, ack, proof, proofHeight+1)
				suite.Require().NoError(err)
				suite.Require().NotNil(packetOut)
			} else {
				packetOut, err := suite.chainB.App.IBCKeeper.ChannelKeeper.AcknowledgePacket(ctx, packet, ack, proof, proofHeight+1)
				suite.Require().Error(err)
				suite.Require().Nil(packetOut)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestCleanupPacket() {
	counterparty := types.NewCounterparty(testPort2, testChannel2)
	packetKey := ibctypes.KeyPacketCommitment(testPort1, testChannel1, 1)
	var (
		packet      types.Packet
		nextSeqRecv uint64
	)

	ack := []byte("ack")

	testCases := []testCase{
		{"success", func() {
			nextSeqRecv = 10
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connectionexported.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.UNORDERED, testConnectionIDA)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 1, types.CommitPacket(packet.Data))
		}, true},
		{"channel not found", func() {}, false},
		{"channel not open", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.CLOSED, exported.ORDERED, testConnectionIDA)
		}, false},
		{"packet source port ≠ channel counterparty port", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createChannel(testPort1, testChannel1, testPort3, testChannel2, exported.OPEN, exported.ORDERED, testConnectionIDA)
		}, false},
		{"packet source channel ID ≠ channel counterparty channel ID", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel3, exported.OPEN, exported.ORDERED, testConnectionIDA)
		}, false},
		{"connection not found", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionIDA)
		}, false},
		{"connection not OPEN", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connectionexported.INIT)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionIDA)
		}, false},
		{"packet already received ", func() {
			packet = types.NewPacket(mockSuccessPacket{}, 10, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connectionexported.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionIDA)
		}, false},
		{"packet hasn't been sent", func() {
			nextSeqRecv = 10
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connectionexported.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionIDA)
		}, false},
		{"next seq receive verification failed", func() {
			nextSeqRecv = 10
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connectionexported.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionIDA)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 1, types.CommitPacket(packet.Data))
		}, false},
		{"packet ack verification failed", func() {
			nextSeqRecv = 10
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.chainB.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connectionexported.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.UNORDERED, testConnectionIDA)
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), testPort1, testChannel1, 1, types.CommitPacket(packet.Data))
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest() // reset
			tc.malleate()

			ctx := suite.chainB.GetContext()

			proof, proofHeight := queryProof(suite.chainA, packetKey)
			if tc.expPass {
				packetOut, err := suite.chainB.App.IBCKeeper.ChannelKeeper.CleanupPacket(ctx, packet, proof, proofHeight+1, nextSeqRecv, ack)
				suite.Require().NoError(err)
				suite.Require().NotNil(packetOut)
			} else {
				packetOut, err := suite.chainB.App.IBCKeeper.ChannelKeeper.CleanupPacket(ctx, packet, ibctypes.InvalidProof{}, proofHeight+1, nextSeqRecv, ack)
				suite.Require().Error(err)
				suite.Require().Nil(packetOut)
			}
		})
	}
}

type mockSuccessPacket struct{}

// GetBytes returns the serialised packet data (without timeout)
func (mp mockSuccessPacket) GetBytes() []byte { return []byte("THIS IS A SUCCESS PACKET") }

// GetTimeoutHeight returns the timeout height defined specifically for each packet data type instance
func (mp mockSuccessPacket) GetTimeoutHeight() uint64 { return 10 }

// ValidateBasic validates basic properties of the packet data, implements sdk.Msg
func (mp mockSuccessPacket) ValidateBasic() error { return nil }

// Type returns human readable identifier, implements sdk.Msg
func (mp mockSuccessPacket) Type() string { return "mock/packet/success" }

type mockFailPacket struct{}

// GetBytes returns the serialised packet data (without timeout)
func (mp mockFailPacket) GetBytes() []byte { return []byte("THIS IS A FAILURE PACKET") }

// GetTimeoutHeight returns the timeout height defined specifically for each packet data type instance
func (mp mockFailPacket) GetTimeoutHeight() uint64 { return 10000 }

// ValidateBasic validates basic properties of the packet data, implements sdk.Msg
func (mp mockFailPacket) ValidateBasic() error { return fmt.Errorf("Packet failed validate basic") }

// Type returns human readable identifier, implements sdk.Msg
func (mp mockFailPacket) Type() string { return "mock/packet/failure" }
