package keeper_test

import (
	"fmt"

	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

func (suite *KeeperTestSuite) TestSendPacket() {
	counterparty := types.NewCounterparty(testPort2, testChannel2)
	var packet exported.PacketI

	testCases := []testCase{
		{"success", func() {
			suite.createClient(testClientID1)
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.INIT)
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
			suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, testPort1, testChannel1, 1)
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.updateClient()
		}, true},
		{"fail-validate-basic", func() {
			packet = types.NewPacket(mockFailPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
		}, false},
		{"channel-not-found", func() {
			suite.createClient(testClientID1)
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.INIT)
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
		}, false},
		{"channel-closed", func() {
			suite.createClient(testClientID1)
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.INIT)
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.CLOSED, exported.ORDERED, testConnectionID1)
		}, false},
		{"next-sequence-not-found", func() {
			suite.createClient(testClientID1)
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.INIT)
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.updateClient()
		}, false},
		{"next-sequence-wrong", func() {
			suite.createClient(testClientID1)
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.INIT)
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
			suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, testPort1, testChannel1, 5)
			packet = types.NewPacket(mockSuccessPacket{}, 1, testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID())
			suite.updateClient()
		}, false},
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			var err error
			suite.SetupTest() // reset
			tc.malleate()
			if tc.expPass {
				err = suite.app.IBCKeeper.ChannelKeeper.SendPacket(suite.ctx, packet)
				suite.Require().NoError(err)
			} else {
				err = suite.app.IBCKeeper.ChannelKeeper.SendPacket(suite.ctx, packet)
				suite.Require().Error(err)
			}
		})
	}

}

func (suite *KeeperTestSuite) TestRecvPacket() {
	// Channel found/not found
	// Channel closed/not CLOSED
	// if packet.GetSourcePort() != channel.Counterparty.PortID {}
	// if packet.GetSourceChannel() != channel.Counterparty.ChannelID {}
	// Connection found/not found
	// Connection initiated/uninitialized
	// if uint64(ctx.BlockHeight()) >= packet.GetTimeoutHeight() {}
	// Client state found/not found
	// Success/fail on verify packet commitment
}

func (suite *KeeperTestSuite) TestPacketExecuted() {
	// Channel found/not found
	// Channel closed/not CLOSED
	// if acknowledgement != nil || channel.Ordering == exported.UNORDERED {}
	// if channel.Ordering == exported.ORDERED {
	// Ensure next sequence recieve is found
	// if packet.GetSequence() != nextSequenceRecv {}
	// }
	// Success, packet recieved and acknowledged
}

func (suite *KeeperTestSuite) TestAcknowledgePacket() {
	// Channel found/not found
	// Channel closed/not CLOSED
	// if packet.GetSourcePort() != channel.Counterparty.PortID {}
	// if packet.GetSourceChannel() != channel.Counterparty.ChannelID {}
	// Connection found/not found
	// Connection initiated/uninitialized
	// if !bytes.Equal(commitment, types.CommitPacket(packet.GetData())) {}
	// Client state found/not found
	// Success/fail on verify packet commitment
}

func (suite *KeeperTestSuite) TestAcknowledgementExecuted() {
	// Delete non existent packet commitment
	// Create packet commitment
	// Delete that packet commitment
	// Ensure packet commitment deleted
}

func (suite *KeeperTestSuite) TestCleanupPacket() {
	// Channel found/not found
	// Channel closed/not CLOSED
	// if packet.GetSourcePort() != channel.Counterparty.PortID {}
	// if packet.GetSourceChannel() != channel.Counterparty.ChannelID {}
	// Connection found/not found
	// Connection initiated/uninitialized
	// if nextSequenceRecv <= packet.GetSequence() {}
	// if !bytes.Equal(commitment, types.CommitPacket(packet.GetData())) {}
	// Client state found/not found
	// Success/fail on verify ORDERED packet commitment
	// Success/fail on verify UNORDERED packet commitment
	// Invalid ordering packet failure
}

type mockSuccessPacket struct{}

// GetBytes returns the serialised packet data (without timeout)
func (mp mockSuccessPacket) GetBytes() []byte { return []byte("THIS IS A SUCCESS PACKET") }

// GetTimeoutHeight returns the timeout height defined specifically for each packet data type instance
func (mp mockSuccessPacket) GetTimeoutHeight() uint64 { return 10000 }

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

type mockTimeoutPacket struct{}

// GetBytes returns the serialised packet data (without timeout)
func (mp mockTimeoutPacket) GetBytes() []byte { return []byte("THIS IS A TIMEOUT PACKET") }

// GetTimeoutHeight returns the timeout height defined specifically for each packet data type instance
func (mp mockTimeoutPacket) GetTimeoutHeight() uint64 { return 10 }

// ValidateBasic validates basic properties of the packet data, implements sdk.Msg
func (mp mockTimeoutPacket) ValidateBasic() error { return nil }

// Type returns human readable identifier, implements sdk.Msg
func (mp mockTimeoutPacket) Type() string { return "mock/packet/timeout" }
