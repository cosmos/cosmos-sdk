package keeper_test

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

func (suite *KeeperTestSuite) TestSendPacket() {

	var msgPacket types.MsgPacket

	testCases := []testCase{
		{"Invalid Packet", func() {
			msgPacket = types.NewMsgPacket(invalidPacket, validProof{}, proofHeight, addr1)
		}, false},
		{"Channel not found", func() {
			msgPacket = types.NewMsgPacket(validPacket, validProof{}, proofHeight, addr1)
		}, false},
		{"Channel state CLOSED", func() {
			suite.createChannel(testPort1, testChannel1, testPort1, testChannel1, exported.CLOSED, testChannelOrder, testConnectionID1)
		}, false},
		{"Packet destination port and counterparty port is different.", func() {
			suite.createChannel(testPort1, testChannel1, testPort1, testChannel1, exported.OPEN, testChannelOrder, testConnectionID1)
		}, false},
		{"Packet destination channel and counterparty channel is different.", func() {
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel1, exported.OPEN, testChannelOrder, testConnectionID1)
		}, false},
		{"Connection not found", func() {
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, testChannelOrder, testConnectionID1)
		}, false},
		{"Connection state is CLOSED", func() {
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.UNINITIALIZED)
		}, false},
		{"Consensus state not found", func() {
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
		}, false},
		{"Sequence send not found", func() {
			suite.createClient(testClientID1)
		}, false},
		{"Packet timed out", func() {
			msgPacket = types.NewMsgPacket(sendingTimedOutValidPacket, validProof{}, proofHeight, addr1)
		}, false},
		{"Invalid packet sequence", func() {
			msgPacket = types.NewMsgPacket(validPacket, validProof{}, proofHeight, addr1)
			suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, msgPacket.GetSourcePort(), msgPacket.GetSourceChannel(), msgPacket.GetSequence()+1)
		}, false},
		{"Success", func() {
			suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, msgPacket.GetSourcePort(), msgPacket.GetSourceChannel(), msgPacket.GetSequence())
		}, true},
	}

	suite.SetupTest() // reset

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {

			tc.malleate()
			err := suite.app.IBCKeeper.ChannelKeeper.SendPacket(suite.ctx, msgPacket)
			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestReceivePacket() {

	var msgPacket types.MsgPacket

	testCases := []testCase{
		{"Channel not found", func() {
			msgPacket = types.NewMsgPacket(validPacket, invalidProof{}, proofHeight, addr1)
		}, false},
		{"Channel state CLOSED", func() {
			suite.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.CLOSED, testChannelOrder, testConnectionID1)
		}, false},
		{"packet source port doesn't match the counterparty's port", func() {
			suite.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.OPEN, testChannelOrder, testConnectionID1)
			msgPacket.SourcePort = testPort2
		}, false},
		{"packet source channel doesn't match the counterparty's channel", func() {
			msgPacket.SourcePort = testPort1
			msgPacket.SourceChannel = testChannel2
		}, false},
		{"Connection not found", func() {
			msgPacket.SourceChannel = testChannel1
		}, false},
		{"Connection state is not OPEN", func() {
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.UNINITIALIZED)
		}, false},
		{"Packet timed out", func() {
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
			msgPacket = types.NewMsgPacket(receiveTimedOutValidPacket, invalidProof{}, proofHeight, addr1)
		}, false},
		{"Invalid counterparty packet commitment", func() {
			msgPacket = types.NewMsgPacket(validPacket, invalidProof{}, proofHeight, addr1)
		}, false},
		{"Success", func() {
			suite.createClient(testClientID1)
			msgPacket = types.NewMsgPacket(validPacket, validProof{}, proofHeight, addr1)
		}, true},
	}

	suite.SetupTest() // reset

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {

			tc.malleate()
			_, err := suite.app.IBCKeeper.ChannelKeeper.RecvPacket(suite.ctx, msgPacket, msgPacket.Proof, msgPacket.ProofHeight)
			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}

}

func (suite *KeeperTestSuite) TestPacketExecuted() {

	var msgPacket types.MsgPacket
	var packetDataI exported.PacketDataI

	testCases := []testCase{
		{"Channel not found", func() {
			msgPacket = types.NewMsgPacket(validPacket, validProof{}, proofHeight, addr1)
			packetDataI = validPacket.GetData()
		}, false},
		{"Channel state CLOSED", func() {
			suite.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.CLOSED, exported.UNORDERED, testConnectionID1)
		}, false},
		{"Channel is Unordered", func() {
			suite.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.OPEN, exported.UNORDERED, testConnectionID1)
		}, true},
		{"Sequence Receive Not Found", func() {
			suite.createChannel(testPort2, testChannel2, testPort1, testChannel1, exported.OPEN, testChannelOrder, testConnectionID1)
		}, false},
		{"Invalid packet sequence", func() {
			suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.ctx, msgPacket.GetDestPort(), msgPacket.GetDestChannel(), msgPacket.GetSequence()+1)
		}, false},
		{"Success", func() {
			suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.ctx, msgPacket.GetDestPort(), msgPacket.GetDestChannel(), msgPacket.GetSequence())
		}, true},
	}

	suite.SetupTest() // reset

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {

			tc.malleate()
			err := suite.app.IBCKeeper.ChannelKeeper.PacketExecuted(suite.ctx, msgPacket, packetDataI)
			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}

}

func (suite *KeeperTestSuite) TestAcknowledgePacket() {

	var proof commitment.ProofI
	var packetDataI exported.PacketDataI

	testCases := []testCase{
		{"Channel not found", func() {
			proof = invalidProof{}
			packetDataI = ackPacket.GetData()
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
		{"Connection is uninitialized", func() {
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.UNINITIALIZED)
		}, false},
		{"Packet hasn't been sent", func() {
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
		}, false},
		{"invalid acknowledgement on counterparty chain", func() {
			suite.createClient(testClientID1)
			suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort1, testChannel1, 1, types.CommitPacket(packetDataI))
		}, false},
		{"Success", func() {
			proof = validProof{}
		}, true},
	}

	suite.SetupTest() // reset

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {

			tc.malleate()
			packetOut, err := suite.app.IBCKeeper.ChannelKeeper.AcknowledgePacket(suite.ctx, ackPacket, packetDataI, proof, proofHeight)
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

func (suite *KeeperTestSuite) TestCleanupPacket() {

	var proof commitment.ProofI
	var nextSequenceRecv uint64

	testCases := []testCase{
		{"Channel not found", func() {
			proof = invalidProof{}
			nextSequenceRecv = 1
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
		{"packet already received", func() {
			suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connectionexported.OPEN)
		}, false},
		{"Packet hasn't been sent", func() {
			nextSequenceRecv = 2
		}, false},
		{"ordered channel: packet verification failed", func() {
			suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort1, testChannel1, 1, types.CommitPacket(cleanPacket.GetData()))
		}, false},
		{"unordered channel: packet verification failed", func() {
			suite.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.OPEN, exported.UNORDERED, testConnectionID1)
			suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort1, testChannel1, 1, types.CommitPacket(cleanPacket.GetData()))
		}, false},
		{"Success", func() {
			suite.createClient(testClientID1)
			proof = validProof{}
		}, true},
	}

	suite.SetupTest() // reset

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {

			tc.malleate()
			packetOut, err := suite.app.IBCKeeper.ChannelKeeper.CleanupPacket(suite.ctx, cleanPacket, proof, proofHeight, nextSequenceRecv, types.CommitPacket(ackPacket.GetData()))
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

var _ exported.PacketDataI = validPacketT{}

type validPacketT struct{}

func (validPacketT) GetBytes() []byte {
	return []byte("testdata")
}

func (validPacketT) GetTimeoutHeight() uint64 {
	return 100
}

func (validPacketT) ValidateBasic() error {
	return nil
}

func (validPacketT) Type() string {
	return "valid"
}

var _ exported.PacketDataI = invalidPacketT{}

type invalidPacketT struct{}

func (invalidPacketT) GetBytes() []byte {
	return []byte("testdata")
}

func (invalidPacketT) GetTimeoutHeight() uint64 {
	return 100
}

func (invalidPacketT) ValidateBasic() error {
	return errors.New("invalid packet")
}

func (invalidPacketT) Type() string {
	return "invalid"
}

var _ exported.PacketDataI = sendingTimedOutValidPacketT{}

type sendingTimedOutValidPacketT struct{}

func (sendingTimedOutValidPacketT) GetBytes() []byte {
	return []byte("testdata")
}

func (sendingTimedOutValidPacketT) GetTimeoutHeight() uint64 {
	return 1
}

func (sendingTimedOutValidPacketT) ValidateBasic() error {
	return nil
}

func (sendingTimedOutValidPacketT) Type() string {
	return "sendingTimedOutValidPacketT"
}

var _ exported.PacketDataI = receiveTimedOutValidPacketT{}

type receiveTimedOutValidPacketT struct{}

func (receiveTimedOutValidPacketT) GetBytes() []byte {
	return []byte("testdata")
}

func (receiveTimedOutValidPacketT) GetTimeoutHeight() uint64 {
	return 0
}

func (receiveTimedOutValidPacketT) ValidateBasic() error {
	return nil
}

func (receiveTimedOutValidPacketT) Type() string {
	return "receiveTimedOutValidPacketT"
}

// define variables used for testing
var (
	validPacket                = types.NewPacket(validPacketT{}, 1, testPort1, testChannel1, testPort2, testChannel2)
	sendingTimedOutValidPacket = types.NewPacket(sendingTimedOutValidPacketT{}, 1, testPort1, testChannel1, testPort2, testChannel2)
	receiveTimedOutValidPacket = types.NewPacket(receiveTimedOutValidPacketT{}, 1, testPort1, testChannel1, testPort2, testChannel2)
	invalidPacket              = types.NewPacket(invalidPacketT{}, 0, testPort1, testChannel1, testPort2, testChannel2)
	ackPacket                  = types.NewPacket(validPacketT{}, 1, testPort1, testChannel1, testPort2, testChannel2)
	cleanPacket                = types.NewPacket(validPacketT{}, 1, testPort1, testChannel1, testPort2, testChannel2)

	addr1       = sdk.AccAddress("testaddr1")
	emptyAddr   sdk.AccAddress
	proofHeight = uint64(1)
)
