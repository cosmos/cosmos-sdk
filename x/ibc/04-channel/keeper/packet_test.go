package keeper_test

// import (
// 	"fmt"

// 	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
// 	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
// 	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
// 	transfertypes "github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
// 	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
// )

// func (suite *KeeperTestSuite) TestSendPacket() {
// 	counterparty := types.NewCounterparty(testPort2, testChannel2)
// 	var packet exported.PacketI

// 	testCases := []testCase{
// 		{"success", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createClient(testClientID1)
// 			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
// 			suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, testPort1, testChannel1, 1)
// 		}, true},
// 		{"packet basic validation failed", func() {
// 			packet = types.NewPacket(mockFailPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 		}, false},
// 		{"channel not found", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 		}, false},
// 		{"channel closed", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.CLOSED, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"packet dest port ≠ channel counterparty port", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, testPort3, counterparty.GetChannelID())
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"packet dest channel ID ≠ channel counterparty channel ID", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), testChannel3)
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"connection not found", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"connection is UNINITIALIZED", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.UNINITIALIZED)
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"client state not found", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"timeout height passed", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.commitNBlocks(10)
// 			suite.createClient(testClientID1)
// 			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"next sequence send not found", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createClient(testClientID1)
// 			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"next sequence wrong", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createClient(testClientID1)
// 			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
// 			suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, testPort1, testChannel1, 5)
// 		}, false},
// 	}

// 	for i, tc := range testCases {
// 		tc := tc
// 		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
// 			suite.SetupTest() // reset
// 			tc.malleate()

// 			err := suite.app.IBCKeeper.ChannelKeeper.SendPacket(suite.ctx, packet)

// 			if tc.expPass {
// 				suite.Require().NoError(err)
// 			} else {
// 				suite.Require().Error(err)
// 			}
// 		})
// 	}

// }

// func (suite *KeeperTestSuite) TestRecvPacket() {
// 	counterparty := types.NewCounterparty(testPort2, testChannel2)
// 	var packet exported.PacketI

// 	testCases := []testCase{
// 		{"success", func() {
// 			suite.createClient(testClientID1)
// 			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
// 			suite.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.OPEN, exported.ORDERED, testConnectionID1)
// 			suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, testPort1, testChannel1, 1)
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.updateClient()
// 		}, true},
// 		{"channel not found", func() {}, false},
// 		{"channel not open", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.INIT, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"packet source port ≠ channel counterparty port", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createChannel(testPort2, testChannel2, testPort3, testChannel1, exported.OPEN, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"packet source channel ID ≠ channel counterparty channel ID", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createChannel(testPort2, testChannel2, testPort1, testChannel3, exported.OPEN, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"connection not found", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.OPEN, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"connection not OPEN", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.INIT)
// 			suite.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.OPEN, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"timeout passed", func() {
// 			suite.commitNBlocks(10)
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
// 			suite.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.OPEN, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"validation failed", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
// 			suite.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.OPEN, exported.ORDERED, testConnectionID1)
// 		}, false},
// 	}

// 	for i, tc := range testCases {
// 		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
// 			suite.SetupTest() // reset
// 			tc.malleate()

// 			var err error
// 			if tc.expPass {
// 				packet, err = suite.app.IBCKeeper.ChannelKeeper.RecvPacket(suite.ctx, packet, ibctypes.ValidProof{}, uint64(suite.ctx.BlockHeader().Height))
// 				suite.Require().NoError(err)
// 			} else {
// 				packet, err = suite.app.IBCKeeper.ChannelKeeper.RecvPacket(suite.ctx, packet, ibctypes.InvalidProof{}, uint64(suite.ctx.BlockHeader().Height))
// 				suite.Require().Error(err)
// 			}
// 		})
// 	}

// }

// func (suite *KeeperTestSuite) TestPacketExecuted() {
// 	counterparty := types.NewCounterparty(testPort2, testChannel2)
// 	var packet types.Packet

// 	testCases := []testCase{
// 		{"success: UNORDERED", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.OPEN, exported.UNORDERED, testConnectionID1)
// 			suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.ctx, testPort2, testChannel2, 1)
// 		}, true},
// 		{"success: ORDERED", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.OPEN, exported.ORDERED, testConnectionID1)
// 			suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.ctx, testPort2, testChannel2, 1)
// 		}, true},
// 		{"channel not found", func() {}, false},
// 		{"channel not OPEN", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.CLOSED, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"next sequence receive not found", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.OPEN, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"packet sequence ≠ next sequence receive", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.OPEN, exported.ORDERED, testConnectionID1)
// 			suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.ctx, testPort2, testChannel2, 5)
// 		}, false},
// 	}

// 	for i, tc := range testCases {
// 		tc := tc
// 		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
// 			suite.SetupTest() // reset
// 			tc.malleate()

// 			err := suite.app.IBCKeeper.ChannelKeeper.PacketExecuted(suite.ctx, packet, mockSuccessPacket{})

// 			if tc.expPass {
// 				suite.Require().NoError(err)
// 			} else {
// 				suite.Require().Error(err)
// 			}
// 		})
// 	}
// }

// func (suite *KeeperTestSuite) TestAcknowledgePacket() {
// 	counterparty := types.NewCounterparty(testPort2, testChannel2)
// 	var packet types.Packet

// 	ack := transfertypes.AckDataTransfer{}

// 	testCases := []testCase{
// 		{"success", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createClient(testClientID1)
// 			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
// 			suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort1, testChannel1, 1, types.CommitPacket(packet.Data))
// 		}, true},
// 		{"channel not found", func() {}, false},
// 		{"channel not open", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.CLOSED, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"packet source port ≠ channel counterparty port", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createChannel(testPort1, testChannel1, testPort3, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"packet source channel ID ≠ channel counterparty channel ID", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel3, exported.OPEN, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"connection not found", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"connection not OPEN", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.INIT)
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"packet hasn't been sent", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"packet ack verification failed", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
// 			suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort1, testChannel1, 1, types.CommitPacket(packet.Data))
// 		}, false},
// 	}

// 	for i, tc := range testCases {
// 		tc := tc
// 		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
// 			suite.SetupTest() // reset
// 			tc.malleate()

// 			if tc.expPass {
// 				packetOut, err := suite.app.IBCKeeper.ChannelKeeper.AcknowledgePacket(suite.ctx, packet, ack, ibctypes.ValidProof{}, uint64(suite.ctx.BlockHeader().Height))
// 				suite.Require().NoError(err)
// 				suite.Require().NotNil(packetOut)
// 			} else {
// 				packetOut, err := suite.app.IBCKeeper.ChannelKeeper.AcknowledgePacket(suite.ctx, packet, ack, ibctypes.InvalidProof{}, uint64(suite.ctx.BlockHeader().Height))
// 				suite.Require().Error(err)
// 				suite.Require().Nil(packetOut)
// 			}
// 		})
// 	}
// }

// func (suite *KeeperTestSuite) TestCleanupPacket() {
// 	counterparty := types.NewCounterparty(testPort2, testChannel2)
// 	var (
// 		packet      types.Packet
// 		nextSeqRecv uint64
// 	)

// 	ack := []byte("ack")

// 	testCases := []testCase{
// 		{"success", func() {
// 			nextSeqRecv = 10
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createClient(testClientID1)
// 			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.UNORDERED, testConnectionID1)
// 			suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort1, testChannel1, 1, types.CommitPacket(packet.Data))
// 		}, true},
// 		{"channel not found", func() {}, false},
// 		{"channel not open", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.CLOSED, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"packet source port ≠ channel counterparty port", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createChannel(testPort1, testChannel1, testPort3, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"packet source channel ID ≠ channel counterparty channel ID", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel3, exported.OPEN, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"connection not found", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"connection not OPEN", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.INIT)
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"packet already received ", func() {
// 			packet = types.NewPacket(mockSuccessPacket{}, 10, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"packet hasn't been sent", func() {
// 			nextSeqRecv = 10
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
// 		}, false},
// 		{"next seq receive verification failed", func() {
// 			nextSeqRecv = 10
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
// 			suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort1, testChannel1, 1, types.CommitPacket(packet.Data))
// 		}, false},
// 		{"packet ack verification failed", func() {
// 			nextSeqRecv = 10
// 			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
// 			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
// 			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.UNORDERED, testConnectionID1)
// 			suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort1, testChannel1, 1, types.CommitPacket(packet.Data))
// 		}, false},
// 	}

// 	for i, tc := range testCases {
// 		tc := tc
// 		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
// 			suite.SetupTest() // reset
// 			tc.malleate()

// 			if tc.expPass {
// 				packetOut, err := suite.app.IBCKeeper.ChannelKeeper.CleanupPacket(suite.ctx, packet, ibctypes.ValidProof{}, uint64(suite.ctx.BlockHeader().Height), nextSeqRecv, ack)
// 				suite.Require().NoError(err)
// 				suite.Require().NotNil(packetOut)
// 			} else {
// 				packetOut, err := suite.app.IBCKeeper.ChannelKeeper.CleanupPacket(suite.ctx, packet, ibctypes.InvalidProof{}, uint64(suite.ctx.BlockHeader().Height), nextSeqRecv, ack)
// 				suite.Require().Error(err)
// 				suite.Require().Nil(packetOut)
// 			}
// 		})
// 	}
// }

// type mockSuccessPacket struct{}

// // GetBytes returns the serialised packet data (without timeout)
// func (mp mockSuccessPacket) GetBytes() []byte { return []byte("THIS IS A SUCCESS PACKET") }

// // GetTimeoutHeight returns the timeout height defined specifically for each packet data type instance
// func (mp mockSuccessPacket) GetTimeoutHeight() uint64 { return 10 }

// // ValidateBasic validates basic properties of the packet data, implements sdk.Msg
// func (mp mockSuccessPacket) ValidateBasic() error { return nil }

// // Type returns human readable identifier, implements sdk.Msg
// func (mp mockSuccessPacket) Type() string { return "mock/packet/success" }

// type mockFailPacket struct{}

// // GetBytes returns the serialised packet data (without timeout)
// func (mp mockFailPacket) GetBytes() []byte { return []byte("THIS IS A FAILURE PACKET") }

// // GetTimeoutHeight returns the timeout height defined specifically for each packet data type instance
// func (mp mockFailPacket) GetTimeoutHeight() uint64 { return 10000 }

// // ValidateBasic validates basic properties of the packet data, implements sdk.Msg
// func (mp mockFailPacket) ValidateBasic() error { return fmt.Errorf("Packet failed validate basic") }

// // Type returns human readable identifier, implements sdk.Msg
// func (mp mockFailPacket) Type() string { return "mock/packet/failure" }
