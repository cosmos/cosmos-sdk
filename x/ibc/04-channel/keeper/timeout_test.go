package keeper_test

import (
	"fmt"

	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

func (suite *KeeperTestSuite) TestTimeoutPacket() {

	var proof commitment.ProofI
	var nextSequenceRecv uint64
	var msgPacket types.MsgPacket
	var proofHeight uint64

	testCases := []testCase{
		{"Channel not found", func() {
			proof = invalidProof{}
			nextSequenceRecv = 3
			proofHeight = 1
			msgPacket = types.NewMsgPacket(invalidPacket, validProof{}, 3, addr1)
		}, false},
		{"Channel state CLOSED", func() {
			suite.createChannel(testPort1, testChannel1, testPort1, testChannel1, exported.CLOSED, exported.ORDERED, testConnectionID1)
		}, false},
		{"packet destination port doesn't match the counterparty's port", func() {
			suite.createChannel(testPort1, testChannel1, testPort1, testChannel1, exported.OPEN, exported.ORDERED, testConnectionID1)
		}, false},
		{"packet destination channel doesn't match the counterparty's port", func() {
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel1, exported.OPEN, exported.ORDERED, testConnectionID1)
		}, false},
		{"No connection found", func() {
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
		}, false},
		{"packet timeout", func() {
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
		}, false},
		{"Packet hasn't been sent", func() {
			proofHeight = 101
			msgPacket = types.NewMsgPacket(timeoutPacket, validProof{}, 1, addr1)
		}, false},
		{"packet already received", func() {
			msgPacket = types.NewMsgPacket(timeoutPacket, validProof{}, 1, addr1)
		}, false},
		{"invalid acknowledgement on counterparty chain", func() {
			nextSequenceRecv = 1
			suite.createClient(testClientID1)
			suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort1, testChannel1, 1, types.CommitPacket(msgPacket.GetData()))
		}, false},
		{"ordered channel: invalid acknowledgement on counterparty chain", func() {
			suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort1, testChannel1, 2, types.CommitPacket(msgPacket.GetData()))
		}, false},
		{"unordered channel: invalid acknowledgement on counterparty chain", func() {
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.UNORDERED, testConnectionID1)
		}, false},
		{"Success", func() {
			proof = validProof{}
			proofHeight = uint64(suite.app.LastBlockHeight())
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
		}, true},
	}

	suite.SetupTest() // reset

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {

			tc.malleate()
			packetOut, err := suite.app.IBCKeeper.ChannelKeeper.TimeoutPacket(suite.ctx, msgPacket, proof, proofHeight, nextSequenceRecv)
			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
				suite.Require().NotNil(packetOut)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
				suite.Require().Nil(packetOut)
			}
		})
	}

}

func (suite *KeeperTestSuite) TestTimeoutExecuted() {

	var msgPacket types.MsgPacket

	testCases := []testCase{
		{"Channel not found", func() {
			msgPacket = types.NewMsgPacket(invalidPacket, validProof{}, 3, addr1)
		}, false},
		{"invalid acknowledgement on counterparty chain", func() {
			suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort1, testChannel1, 2, types.CommitPacket(msgPacket.GetData()))
		}, false},
		{"Success", func() {
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
		}, true},
	}

	suite.SetupTest() // reset

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {

			tc.malleate()
			err := suite.app.IBCKeeper.ChannelKeeper.TimeoutExecuted(suite.ctx, msgPacket)
			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}

}

func (suite *KeeperTestSuite) TestTimeoutOnClose() {

	var proofClosed commitment.ProofI
	var proofNonMembership commitment.ProofI
	var nextSequenceRecv uint64
	var proofHeight uint64

	testCases := []testCase{
		{"Channel not found", func() {
			proofClosed = invalidProof{}
			nextSequenceRecv = 3
			proofHeight = 1
		}, false},
		{"Channel state CLOSED", func() {
			suite.createChannel(testPort1, testChannel1, testPort1, testChannel1, exported.CLOSED, exported.ORDERED, testConnectionID1)
		}, false},
		{"packet destination port doesn't match the counterparty's port", func() {
			suite.createChannel(testPort1, testChannel1, testPort1, testChannel1, exported.OPEN, exported.ORDERED, testConnectionID1)
		}, false},
		{"packet destination channel doesn't match the counterparty's port", func() {
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel1, exported.OPEN, exported.ORDERED, testConnectionID1)
		}, false},
		{"No connection found", func() {
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
		}, false},
		{"packet hasn't been sent", func() {
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
		}, false},
		{"channel state verification failed", func() {
			nextSequenceRecv = 1
			suite.createClient(testClientID1)
			suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort1, testChannel1, 2, types.CommitPacket(timeoutPacket.GetData()))
		}, false},
		{"Success: Unordered Channel", func() {
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.UNORDERED, testConnectionID1)
			proofClosed = validProof{}
			proofHeight = uint64(suite.app.LastBlockHeight())
		}, true},
		{"Success: Ordered Channel", func() {
			suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort1, testChannel1, 2, types.CommitPacket(timeoutPacket.GetData()))
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.ORDERED, testConnectionID1)
			proofClosed = validProof{}
			proofHeight = uint64(suite.app.LastBlockHeight())
		}, true},
	}

	suite.SetupTest() // reset

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {

			tc.malleate()
			packetOut, err := suite.app.IBCKeeper.ChannelKeeper.TimeoutOnClose(suite.ctx, timeoutPacket, proofNonMembership, proofClosed, proofHeight, nextSequenceRecv)
			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
				suite.Require().NotNil(packetOut)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
				suite.Require().Nil(packetOut)
			}
		})
	}
}

var _ exported.PacketDataI = timeoutPacketT{}

type timeoutPacketT struct{}

func (timeoutPacketT) GetBytes() []byte {
	return []byte("testdata")
}

func (timeoutPacketT) GetTimeoutHeight() uint64 {
	return 1
}

func (timeoutPacketT) ValidateBasic() error {
	return nil
}

func (timeoutPacketT) Type() string {
	return "valid"
}

var (
	timeoutPacket = types.NewPacket(timeoutPacketT{}, 2, testPort1, testChannel1, testPort2, testChannel2)
)
