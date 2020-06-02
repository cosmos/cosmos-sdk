package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

// define constants used for testing
const (
	testClientIDA     = "testclientida"
	testConnectionIDA = "connectionidatob"

	testClientIDB     = "testclientidb"
	testConnectionIDB = "connectionidbtoa"

	testPort1 = "firstport"
	testPort2 = "secondport"
	testPort3 = "thirdport"

	testChannel1 = "firstchannel"
	testChannel2 = "secondchannel"
	testChannel3 = "thirdchannel"

	testChannelOrder = types.ORDERED

	timeoutHeight            = 100
	timeoutTimestamp         = 100
	disabledTimeoutTimestamp = 0
	disabledTimeoutHeight    = 0
)

type KeeperTestSuite struct {
	suite.Suite

	cdc *codec.Codec

	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.chainA = ibctesting.NewTestChain(testClientIDA)
	suite.chainB = ibctesting.NewTestChain(testClientIDB)

	suite.cdc = suite.chainA.App.Codec()
}

func (suite *KeeperTestSuite) TestSetChannel() {
	ctx := suite.chainB.GetContext()
	_, found := suite.chainB.App.IBCKeeper.ChannelKeeper.GetChannel(ctx, testPort1, testChannel1)
	suite.False(found)

	counterparty2 := types.NewCounterparty(testPort2, testChannel2)
	channel := types.NewChannel(
		types.INIT, testChannelOrder,
		counterparty2, []string{testConnectionIDA}, ibctesting.ChannelVersion,
	)
	suite.chainB.App.IBCKeeper.ChannelKeeper.SetChannel(ctx, testPort1, testChannel1, channel)

	storedChannel, found := suite.chainB.App.IBCKeeper.ChannelKeeper.GetChannel(ctx, testPort1, testChannel1)
	suite.True(found)
	suite.Equal(channel, storedChannel)
}

func (suite KeeperTestSuite) TestGetAllChannels() {
	// Channel (Counterparty): A(C) -> C(B) -> B(A)
	counterparty1 := types.NewCounterparty(testPort1, testChannel1)
	counterparty2 := types.NewCounterparty(testPort2, testChannel2)
	counterparty3 := types.NewCounterparty(testPort3, testChannel3)

	channel1 := types.NewChannel(
		types.INIT, testChannelOrder,
		counterparty3, []string{testConnectionIDA}, ibctesting.ChannelVersion,
	)
	channel2 := types.NewChannel(
		types.INIT, testChannelOrder,
		counterparty1, []string{testConnectionIDA}, ibctesting.ChannelVersion,
	)
	channel3 := types.NewChannel(
		types.CLOSED, testChannelOrder,
		counterparty2, []string{testConnectionIDA}, ibctesting.ChannelVersion,
	)

	expChannels := []types.IdentifiedChannel{
		types.NewIdentifiedChannel(testPort1, testChannel1, channel1),
		types.NewIdentifiedChannel(testPort2, testChannel2, channel2),
		types.NewIdentifiedChannel(testPort3, testChannel3, channel3),
	}

	ctx := suite.chainB.GetContext()

	suite.chainB.App.IBCKeeper.ChannelKeeper.SetChannel(ctx, testPort1, testChannel1, channel1)
	suite.chainB.App.IBCKeeper.ChannelKeeper.SetChannel(ctx, testPort2, testChannel2, channel2)
	suite.chainB.App.IBCKeeper.ChannelKeeper.SetChannel(ctx, testPort3, testChannel3, channel3)

	channels := suite.chainB.App.IBCKeeper.ChannelKeeper.GetAllChannels(ctx)
	suite.Require().Len(channels, len(expChannels))
	suite.Require().Equal(expChannels, channels)
}

func (suite KeeperTestSuite) TestGetAllSequences() {
	seq1 := types.NewPacketSequence(testPort1, testChannel1, 1)
	seq2 := types.NewPacketSequence(testPort2, testChannel2, 2)

	expSeqs := []types.PacketSequence{seq1, seq2}

	ctx := suite.chainB.GetContext()

	for _, seq := range expSeqs {
		suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(ctx, seq.PortID, seq.ChannelID, seq.Sequence)
		suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(ctx, seq.PortID, seq.ChannelID, seq.Sequence)
		suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceAck(ctx, seq.PortID, seq.ChannelID, seq.Sequence)
	}

	sendSeqs := suite.chainB.App.IBCKeeper.ChannelKeeper.GetAllPacketSendSeqs(ctx)
	recvSeqs := suite.chainB.App.IBCKeeper.ChannelKeeper.GetAllPacketRecvSeqs(ctx)
	ackSeqs := suite.chainB.App.IBCKeeper.ChannelKeeper.GetAllPacketAckSeqs(ctx)
	suite.Require().Len(sendSeqs, 2)
	suite.Require().Len(recvSeqs, 2)
	suite.Require().Len(ackSeqs, 2)

	suite.Require().Equal(expSeqs, sendSeqs)
	suite.Require().Equal(expSeqs, recvSeqs)
	suite.Require().Equal(expSeqs, ackSeqs)
}

func (suite KeeperTestSuite) TestGetAllCommitmentsAcks() {
	ack1 := types.NewPacketAckCommitment(testPort1, testChannel1, 1, []byte("ack"))
	ack2 := types.NewPacketAckCommitment(testPort1, testChannel1, 2, []byte("ack"))
	comm1 := types.NewPacketAckCommitment(testPort1, testChannel1, 1, []byte("hash"))
	comm2 := types.NewPacketAckCommitment(testPort1, testChannel1, 2, []byte("hash"))

	expAcks := []types.PacketAckCommitment{ack1, ack2}
	expCommitments := []types.PacketAckCommitment{comm1, comm2}

	ctx := suite.chainB.GetContext()

	for i := 0; i < 2; i++ {
		suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(ctx, expAcks[i].PortID, expAcks[i].ChannelID, expAcks[i].Sequence, expAcks[i].Hash)
		suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(ctx, expCommitments[i].PortID, expCommitments[i].ChannelID, expCommitments[i].Sequence, expCommitments[i].Hash)
	}

	acks := suite.chainB.App.IBCKeeper.ChannelKeeper.GetAllPacketAcks(ctx)
	commitments := suite.chainB.App.IBCKeeper.ChannelKeeper.GetAllPacketCommitments(ctx)
	suite.Require().Len(acks, 2)
	suite.Require().Len(commitments, 2)

	suite.Require().Equal(expAcks, acks)
	suite.Require().Equal(expCommitments, commitments)
}

func (suite *KeeperTestSuite) TestSetSequence() {
	ctx := suite.chainB.GetContext()
	_, found := suite.chainB.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(ctx, testPort1, testChannel1)
	suite.False(found)

	_, found = suite.chainB.App.IBCKeeper.ChannelKeeper.GetNextSequenceRecv(ctx, testPort1, testChannel1)
	suite.False(found)

	_, found = suite.chainB.App.IBCKeeper.ChannelKeeper.GetNextSequenceAck(ctx, testPort1, testChannel1)
	suite.False(found)

	nextSeqSend, nextSeqRecv, nextSeqAck := uint64(10), uint64(10), uint64(10)
	suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(ctx, testPort1, testChannel1, nextSeqSend)
	suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(ctx, testPort1, testChannel1, nextSeqRecv)
	suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceAck(ctx, testPort1, testChannel1, nextSeqAck)

	storedNextSeqSend, found := suite.chainB.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(ctx, testPort1, testChannel1)
	suite.True(found)
	suite.Equal(nextSeqSend, storedNextSeqSend)

	storedNextSeqRecv, found := suite.chainB.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(ctx, testPort1, testChannel1)
	suite.True(found)
	suite.Equal(nextSeqRecv, storedNextSeqRecv)

	storedNextSeqAck, found := suite.chainB.App.IBCKeeper.ChannelKeeper.GetNextSequenceAck(ctx, testPort1, testChannel1)
	suite.True(found)
	suite.Equal(nextSeqAck, storedNextSeqAck)
}

func (suite *KeeperTestSuite) TestPackageCommitment() {
	ctx := suite.chainB.GetContext()
	seq := uint64(10)
	storedCommitment := suite.chainB.App.IBCKeeper.ChannelKeeper.GetPacketCommitment(ctx, testPort1, testChannel1, seq)
	suite.Equal([]byte(nil), storedCommitment)

	commitment := []byte("commitment")
	suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(ctx, testPort1, testChannel1, seq, commitment)

	storedCommitment = suite.chainB.App.IBCKeeper.ChannelKeeper.GetPacketCommitment(ctx, testPort1, testChannel1, seq)
	suite.Equal(commitment, storedCommitment)
}

func (suite *KeeperTestSuite) TestSetPacketAcknowledgement() {
	ctx := suite.chainB.GetContext()
	seq := uint64(10)

	storedAckHash, found := suite.chainB.App.IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(ctx, testPort1, testChannel1, seq)
	suite.False(found)
	suite.Nil(storedAckHash)

	ackHash := []byte("ackhash")
	suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(ctx, testPort1, testChannel1, seq, ackHash)

	storedAckHash, found = suite.chainB.App.IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(ctx, testPort1, testChannel1, seq)
	suite.True(found)
	suite.Equal(ackHash, storedAckHash)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// Mocked types

type mockSuccessPacket struct{}

// GetBytes returns the serialised packet data
func (mp mockSuccessPacket) GetBytes() []byte { return []byte("THIS IS A SUCCESS PACKET") }

type mockFailPacket struct{}

// GetBytes returns the serialised packet data (without timeout)
func (mp mockFailPacket) GetBytes() []byte { return []byte("THIS IS A FAILURE PACKET") }
