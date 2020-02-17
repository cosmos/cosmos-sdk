package keeper_test

import (
	"fmt"

	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

func (suite *KeeperTestSuite) TestTimeoutPacket() {
	counterparty := types.NewCounterparty(testPort2, testChannel2)
	var packet types.Packet

	testCases := []testCase{
		{"success", func() {
			suite.createClient(testClientID1)
			suite.createClient(testClientID2)
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
			suite.createChannel(testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), exported.OPEN, exported.ORDERED, testConnectionID1)
			suite.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.OPEN, exported.ORDERED, testConnectionID1)
			suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, testPort1, testChannel1, 1)
			packet = types.NewPacket(newMockTimeoutPacket(3), 1, testPort1, testChannel1, testPort2, testChannel2)
			err := suite.app.IBCKeeper.ChannelKeeper.SendPacket(suite.ctx, packet)
			suite.Require().NoError(err)
			suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.ctx, testPort2, testChannel2, 0)
			suite.updateClient()
		}, true},
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest() // reset
			tc.malleate()
			if tc.expPass {
				nextRecv, ok := suite.app.IBCKeeper.ChannelKeeper.GetNextSequenceRecv(suite.ctx, packet.GetDestPort(), packet.GetDestChannel())
				suite.Require().True(ok)
				out, err := suite.app.IBCKeeper.ChannelKeeper.TimeoutPacket(suite.ctx, packet, ibctypes.ValidProof{}, 3, nextRecv)
				suite.Require().NoError(err)
				suite.Require().NotNil(out)
			} else {
				out, err := suite.app.IBCKeeper.ChannelKeeper.AcknowledgePacket(suite.ctx, packet, []byte("ack"), ibctypes.ValidProof{}, uint64(suite.ctx.BlockHeader().Height))
				fmt.Println(tc.msg, err)
				suite.Require().NotNil(out)
				suite.Require().Error(err)
			}
		})
	}

	// Create packet with appropriate channel and port, test wrong port/channel
	// Test with an unopened channel
	// if packet.GetDestPort() != channel.Counterparty.PortID {}
	// if packet.GetDestChannel() != channel.Counterparty.ChannelID {}
	// Ensure connection assocated with channel exists, test fail/pass
	// if nextSequenceRecv >= packet.GetSequence() {}
	// if !bytes.Equal(commitment, types.CommitPacket(packet.GetData())) {}
	// Ensure that consensus state is found
	// Test ordered channel
	// Test unordered channel

}

func (suite *KeeperTestSuite) TestTimeoutExecuted() {
	counterparty := types.NewCounterparty(testPort2, testChannel2)
	var packet types.Packet

	testCases := []testCase{
		{"success", func() {
			suite.createClient(testClientID1)
			suite.createClient(testClientID2)
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
			suite.createChannel(testPort1, testChannel1, counterparty.GetPortID(), counterparty.GetChannelID(), exported.OPEN, exported.ORDERED, testConnectionID1)
			suite.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.OPEN, exported.ORDERED, testConnectionID1)
			suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, testPort1, testChannel1, 1)
			packet = types.NewPacket(newMockTimeoutPacket(3), 1, testPort1, testChannel1, testPort2, testChannel2)
			err := suite.app.IBCKeeper.ChannelKeeper.SendPacket(suite.ctx, packet)
			suite.Require().NoError(err)
			suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.ctx, testPort2, testChannel2, 0)
			suite.updateClient()
			nextRecv, ok := suite.app.IBCKeeper.ChannelKeeper.GetNextSequenceRecv(suite.ctx, packet.GetDestPort(), packet.GetDestChannel())
			suite.Require().True(ok)
			out, err := suite.app.IBCKeeper.ChannelKeeper.TimeoutPacket(suite.ctx, packet, ibctypes.ValidProof{}, 3, nextRecv)
			suite.Require().NoError(err)
			suite.Require().NotNil(out)
		}, true},
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tc.msg, i, len(testCases)), func() {
			suite.SetupTest() // reset
			tc.malleate()
			if tc.expPass {
				err := suite.app.IBCKeeper.ChannelKeeper.TimeoutExecuted(suite.ctx, packet)
				suite.Require().NoError(err)
			} else {
				err := suite.app.IBCKeeper.ChannelKeeper.TimeoutExecuted(suite.ctx, packet)
				suite.Require().Error(err)
			}
		})
	}

	// Create packet with appropriate channel and port, test wrong port/channel
	// Ensure that ordered channel that is created is closed once TimeoutExecuted is called
}

func (suite *KeeperTestSuite) TestTimeoutOnClose() {
	// Create packet with appropriate channel and port, test wrong port/channel
	// if packet.GetDestPort() != channel.Counterparty.PortID {}
	// if packet.GetDestChannel() != channel.Counterparty.ChannelID {}
	// Ensure connection assocated with channel exists, test fail/pass
	// if !bytes.Equal(commitment, types.CommitPacket(packet.GetData())) {}
	// Ensure that consensus state is found
	// Ensure that opposite channel has closed
	// Test ordered channel
	// Test unordered channel
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
