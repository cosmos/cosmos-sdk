package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

// KeeperTestSuite is a testing suite to test keeper functions.
type KeeperTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// ID's of testing chains used for convience and readability
	chainA string
	chainB string
}

// TestKeeperTestSuite runs all the tests within this package.
func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// SetupTest creates a coordinator with 2 test chains.
func (suite *KeeperTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = ibctesting.GetChainID(0)
	suite.chainB = ibctesting.GetChainID(1)
}

// TestSetChannel create clients and connections on both chains. It tests for the non-existence
// and existence of a channel in INIT on chainA.
func (suite *KeeperTestSuite) TestSetChannel() {
	chainA := suite.coordinator.GetChain(suite.chainA)
	chainB := suite.coordinator.GetChain(suite.chainB)

	// create client and connections on both chains
	_, _, connA, connB := suite.coordinator.CreateClientsAndConnections(suite.chainA, suite.chainB, clientexported.Tendermint)

	// check for channel to be created on chainB
	channelA := connA.NextTestChannel()
	_, found := chainA.App.IBCKeeper.ChannelKeeper.GetChannel(chainA.GetContext(), channelA.PortID, channelA.ChannelID)
	suite.False(found)

	// init channel
	channelA, channelB, err := suite.coordinator.CreateChannelInit(chainA, chainB, connA, connB, channeltypes.ORDERED)
	suite.NoError(err)

	storedChannel, found := chainA.App.IBCKeeper.ChannelKeeper.GetChannel(chainA.GetContext(), channelA.PortID, channelA.ChannelID)
	expectedCounterparty := channeltypes.NewCounterparty(channelB.PortID, channelB.ChannelID)

	suite.True(found)
	suite.Equal(channeltypes.INIT, storedChannel.State)
	suite.Equal(channeltypes.ORDERED, storedChannel.Ordering)
	suite.Equal(expectedCounterparty, storedChannel.Counterparty)
}

/*
// TestGetAllChannels creates multiple channels on chain A through various connections
// and tests their retrieval. 2 channels are on connA0 and 1 channel is on connA1
func (suite KeeperTestSuite) TestGetAllChannels() {
	clientA, clientB, connA0, connB0 := suite.coordinator.Setup(suite.chainA, suite.chainB)
	testchannel0 := connA0.Channels[0]

	channel2, _, err := suite.cooordinator.CreateChannel(connA0, connB0)
	suite.NoError(err)

	connA1, connB1, err := suite.coordinator.CreateConnection(clientA, clientB)
	suite.No(err)

	channel3, _, err := suite.coordinator.CreateChannelInit(connA1, connB1)

	channel0 := types.NewChannel(
		types.OPEN, types.ORDERED,
		counterparty0, []string{connB0.ID}, ibctesting.ChannelVersion,
	)
	channel2 := types.NewChannel(
		types.OPEN, types.ORDERED,
		counterparty2, []string{connB0.ID}, ibctesting.ChannelVersion,
	)
	channel3 := types.NewChannel(
		types.INIT, types.ORDERED,
		counterparty3, []string{connB1.ID}, ibctesting.ChannelVersion,
	)

	expChannels := []types.IdentifiedChannel{
		types.NewIdentifiedChannel(testchannel0.PortID, testchannel0.ChannelID, channel0),
		types.NewIdentifiedChannel(testchannel1.PortID, testChannel1.ChannelID, channel1),
		types.NewIdentifiedChannel(testchannel2.PortID, testChannel2.ChannelID, channel2),
	}

	chainA := suite.coordinator.GetChain(suite.chainA)
	ctxA := chainA.GetContext()

	channels := suite.chainA.App.IBCKeeper.ChannelKeeper.GetAllChannels(ctxA)
	suite.Require().Len(channels, len(expChannels))
	suite.Require().Equal(expChannels, channels)
}

// TestGetAllSequences sets all packet sequences for two different channels on chain A and
// tests their retrieval.
func (suite KeeperTestSuite) TestGetAllSequences() {
	connA, connB, err := suite.coordinator.Setup(suite.chainA, suite.chainB)
	channelA0 := connA.Channels[0]

	channelA1, _, err := suite.cooordinator.CreateChannel(connA, connB)
	suite.NoError(err)

	seq1 := types.NewPacketSequence(channelA0.PortID, channelA0.ChannelID, 1)
	seq2 := types.NewPacketSequence(channelA0.PortID, channelA0.ChannelID, 2)
	seq3 := types.NewPacketSequence(channelA1.PortID, channelA1.ChannelID, 3)

	// seq1 should be overwritten by seq2
	expSeqs := []types.PacketSequence{seq2, seq3}

	chainA := suite.coordinator.GetChain(suite.chainA)
	ctxA := chainA.GetContext()

	for _, seq := range expSeqs {
		chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(ctxA, seq.PortID, seq.ChannelID, seq.Sequence)
		chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(ctxA, seq.PortID, seq.ChannelID, seq.Sequence)
		chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceAck(ctxA, seq.PortID, seq.ChannelID, seq.Sequence)
	}

	sendSeqs := suite.chainA.App.IBCKeeper.ChannelKeeper.GetAllPacketSendSeqs(ctxA)
	recvSeqs := suite.chainA.App.IBCKeeper.ChannelKeeper.GetAllPacketRecvSeqs(ctxA)
	ackSeqs := suite.chainA.App.IBCKeeper.ChannelKeeper.GetAllPacketAckSeqs(ctxA)
	suite.Len(sendSeqs, 2)
	suite.Len(recvSeqs, 2)
	suite.Len(ackSeqs, 2)

	suite.Equal(expSeqs, sendSeqs)
	suite.Equal(expSeqs, recvSeqs)
	suite.Equal(expSeqs, ackSeqs)
}

// TestGetAllCommitmentsAcks
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

// TestSetSequence
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

// TestPacketCommitment does basic verification of setting and getting of packet commitments within
// the Channel Keeper.
func (suite *KeeperTestSuite) TestPacketCommitment() {
	ctx := suite.chainB.GetContext()
	seq := uint64(10)

	storedCommitment := suite.chainB.App.IBCKeeper.ChannelKeeper.GetPacketCommitment(ctx, testPort1, testChannel1, seq)
	suite.Nil(storedCommitment)

	suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(ctx, testPort1, testChannel1, seq, testPacketCommitment)

	storedCommitment = suite.chainB.App.IBCKeeper.ChannelKeeper.GetPacketCommitment(ctx, testPort1, testChannel1, seq)
	suite.Equal(testPacketCommitment, storedCommitment)
}

// TestGetAllPacketCommitmentsAtChannel verifies that iterator returns all stored packet commitments
// for a specific channel.
func (suite *KeeperTestSuite) TestGetAllPacketCommitmentsAtChannel() {
	// setup
	ctx := suite.chainB.GetContext()
	expectedSeqs := make(map[uint64]bool)

	seq := uint64(15)
	maxSeq := uint64(25)
	suite.Require().Greater(maxSeq, seq)

	// create consecutive commitments
	for i := uint64(1); i < seq; i++ {
		suite.chainB.storePacketCommitment(ctx, testPort1, testChannel1, i)
		expectedSeqs[i] = true
	}

	// add non-consecutive commitments
	for i := seq; i < maxSeq; i += 2 {
		suite.chainB.storePacketCommitment(ctx, testPort1, testChannel1, i)
		expectedSeqs[i] = true
	}

	// add sequence on different channel/port
	suite.chainB.storePacketCommitment(ctx, testPort2, testChannel2, maxSeq+1)

	commitments := suite.chainB.App.IBCKeeper.ChannelKeeper.GetAllPacketCommitmentsAtChannel(ctx, testPort1, testChannel1)

	suite.Equal(len(expectedSeqs), len(commitments))
	suite.NotEqual(0, len(commitments))

	// verify that all the packet commitments were stored
	for _, packet := range commitments {
		suite.True(expectedSeqs[packet.Sequence])
		suite.Equal(testPort1, packet.PortID)
		suite.Equal(testChannel1, packet.ChannelID)

		// prevent duplicates from passing checks
		expectedSeqs[packet.Sequence] = false
	}
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

// Mocked types

type mockSuccessPacket struct{}

// GetBytes returns the serialised packet data
func (mp mockSuccessPacket) GetBytes() []byte { return []byte("THIS IS A SUCCESS PACKET") }

type mockFailPacket struct{}

// GetBytes returns the serialised packet data (without timeout)
func (mp mockFailPacket) GetBytes() []byte { return []byte("THIS IS A FAILURE PACKET") }
*/
