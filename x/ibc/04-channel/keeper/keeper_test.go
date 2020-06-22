package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

// KeeperTestSuite is a testing suite to test keeper functions.
type KeeperTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
}

// TestKeeperTestSuite runs all the tests within this package.
func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// SetupTest creates a coordinator with 2 test chains.
func (suite *KeeperTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))
}

// TestSetChannel create clients and connections on both chains. It tests for the non-existence
// and existence of a channel in INIT on chainA.
func (suite *KeeperTestSuite) TestSetChannel() {
	// create client and connections on both chains
	_, _, connA, connB := suite.coordinator.CreateClientsAndConnections(suite.chainA.ChainID, suite.chainB.ChainID, clientexported.Tendermint)

	// check for channel to be created on chainB
	channelA := connA.NextTestChannel()
	_, found := suite.chainA.App.IBCKeeper.ChannelKeeper.GetChannel(suite.chainA.GetContext(), channelA.PortID, channelA.ChannelID)
	suite.False(found)

	// init channel
	channelA, channelB, err := suite.coordinator.CreateChannelInit(suite.chainA, suite.chainB, connA, connB, types.ORDERED)
	suite.NoError(err)

	storedChannel, found := suite.chainA.App.IBCKeeper.ChannelKeeper.GetChannel(suite.chainA.GetContext(), channelA.PortID, channelA.ChannelID)
	expectedCounterparty := types.NewCounterparty(channelB.PortID, channelB.ChannelID)

	suite.True(found)
	suite.Equal(types.INIT, storedChannel.State)
	suite.Equal(types.ORDERED, storedChannel.Ordering)
	suite.Equal(expectedCounterparty, storedChannel.Counterparty)
}

// TestGetAllChannels creates multiple channels on chain A through various connections
// and tests their retrieval. 2 channels are on connA0 and 1 channel is on connA1
func (suite KeeperTestSuite) TestGetAllChannels() {
	clientA, clientB, connA0, connB0 := suite.coordinator.Setup(suite.chainA.ChainID, suite.chainB.ChainID)
	// channel0 on first connection on chainA
	testchannel0 := connA0.Channels[0]
	counterparty0 := types.Counterparty{
		PortID:    connB0.Channels[0].PortID,
		ChannelID: connB0.Channels[0].ChannelID,
	}

	// channel1 is second channel on first connection on chainA
	testchannel1, _ := suite.coordinator.CreateChannel(suite.chainA.ChainID, suite.chainB.ChainID, connA0, connB0, types.ORDERED)
	counterparty1 := types.Counterparty{
		PortID:    connB0.Channels[1].PortID,
		ChannelID: connB0.Channels[1].ChannelID,
	}

	connA1, connB1 := suite.coordinator.CreateConnection(suite.chainA.ChainID, suite.chainB.ChainID, clientA, clientB)

	// channel2 is on a second connection on chainA
	testchannel2, _, err := suite.coordinator.CreateChannelInit(suite.chainA, suite.chainB, connA1, connB1, types.UNORDERED)
	suite.Require().NoError(err)

	counterparty2 := types.Counterparty{
		PortID:    connB1.Channels[0].PortID,
		ChannelID: connB1.Channels[0].ChannelID,
	}

	channel0 := types.NewChannel(
		types.OPEN, types.UNORDERED,
		counterparty0, []string{connB0.ID}, ibctesting.ChannelVersion,
	)
	channel1 := types.NewChannel(
		types.OPEN, types.ORDERED,
		counterparty1, []string{connB0.ID}, ibctesting.ChannelVersion,
	)
	channel2 := types.NewChannel(
		types.INIT, types.UNORDERED,
		counterparty2, []string{connB1.ID}, ibctesting.ChannelVersion,
	)

	expChannels := []types.IdentifiedChannel{
		types.NewIdentifiedChannel(testchannel0.PortID, testchannel0.ChannelID, channel0),
		types.NewIdentifiedChannel(testchannel1.PortID, testchannel1.ChannelID, channel1),
		types.NewIdentifiedChannel(testchannel2.PortID, testchannel2.ChannelID, channel2),
	}

	ctxA := suite.chainA.GetContext()

	channels := suite.chainA.App.IBCKeeper.ChannelKeeper.GetAllChannels(ctxA)
	suite.Require().Len(channels, len(expChannels))
	suite.Require().Equal(expChannels, channels)
}

// TestGetAllSequences sets all packet sequences for two different channels on chain A and
// tests their retrieval.
func (suite KeeperTestSuite) TestGetAllSequences() {
	_, _, connA, connB := suite.coordinator.Setup(suite.chainA.ChainID, suite.chainB.ChainID)
	channelA0 := connA.Channels[0]

	channelA1, _ := suite.coordinator.CreateChannel(suite.chainA.ChainID, suite.chainB.ChainID, connA, connB, types.UNORDERED)

	seq1 := types.NewPacketSequence(channelA0.PortID, channelA0.ChannelID, 1)
	seq2 := types.NewPacketSequence(channelA0.PortID, channelA0.ChannelID, 2)
	seq3 := types.NewPacketSequence(channelA1.PortID, channelA1.ChannelID, 3)

	// seq1 should be overwritten by seq2
	expSeqs := []types.PacketSequence{seq2, seq3}

	ctxA := suite.chainA.GetContext()

	for _, seq := range []types.PacketSequence{seq1, seq2, seq3} {
		suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(ctxA, seq.PortID, seq.ChannelID, seq.Sequence)
		suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(ctxA, seq.PortID, seq.ChannelID, seq.Sequence)
		suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceAck(ctxA, seq.PortID, seq.ChannelID, seq.Sequence)
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

// TestGetAllCommitmentsAcks creates a set of acks and packet commitments on two different
// channels on chain A and tests their retrieval.
func (suite KeeperTestSuite) TestGetAllCommitmentsAcks() {
	_, _, connA, connB := suite.coordinator.Setup(suite.chainA.ChainID, suite.chainB.ChainID)
	channelA0 := connA.Channels[0]

	channelA1, _ := suite.coordinator.CreateChannel(suite.chainA.ChainID, suite.chainB.ChainID, connA, connB, types.UNORDERED)

	// channel 0 acks
	ack1 := types.NewPacketAckCommitment(channelA0.PortID, channelA0.ChannelID, 1, []byte("ack"))
	ack2 := types.NewPacketAckCommitment(channelA0.PortID, channelA0.ChannelID, 2, []byte("ack"))

	// duplicate ack
	ack2dup := types.NewPacketAckCommitment(channelA0.PortID, channelA0.ChannelID, 2, []byte("ack"))

	// channel 1 acks
	ack3 := types.NewPacketAckCommitment(channelA1.PortID, channelA1.ChannelID, 1, []byte("ack"))

	// channel 0 packet commitments
	comm1 := types.NewPacketAckCommitment(channelA0.PortID, channelA0.ChannelID, 1, []byte("hash"))
	comm2 := types.NewPacketAckCommitment(channelA0.PortID, channelA0.ChannelID, 2, []byte("hash"))

	// channel 1 packet commitments
	comm3 := types.NewPacketAckCommitment(channelA1.PortID, channelA1.ChannelID, 1, []byte("hash"))
	comm4 := types.NewPacketAckCommitment(channelA1.PortID, channelA1.ChannelID, 2, []byte("hash"))

	expAcks := []types.PacketAckCommitment{ack1, ack2, ack3}
	expCommitments := []types.PacketAckCommitment{comm1, comm2, comm3, comm4}

	ctxA := suite.chainA.GetContext()

	// set acknowledgements
	for _, ack := range []types.PacketAckCommitment{ack1, ack2, ack2dup, ack3} {
		suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(ctxA, ack.PortID, ack.ChannelID, ack.Sequence, ack.Hash)
	}

	// set packet commitments
	for _, comm := range expCommitments {
		suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(ctxA, comm.PortID, comm.ChannelID, comm.Sequence, comm.Hash)
	}

	acks := suite.chainA.App.IBCKeeper.ChannelKeeper.GetAllPacketAcks(ctxA)
	commitments := suite.chainA.App.IBCKeeper.ChannelKeeper.GetAllPacketCommitments(ctxA)
	suite.Require().Len(acks, len(expAcks))
	suite.Require().Len(commitments, len(expCommitments))

	suite.Require().Equal(expAcks, acks)
	suite.Require().Equal(expCommitments, commitments)
}

// TestSetSequence verifies that the keeper correctly sets the sequence counters.
func (suite *KeeperTestSuite) TestSetSequence() {
	_, _, connA, _ := suite.coordinator.Setup(suite.chainA.ChainID, suite.chainB.ChainID)
	channelA := connA.Channels[0]

	ctxA := suite.chainA.GetContext()
	one := uint64(1)

	// initialized channel has next send seq of 1
	seq, found := suite.chainA.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(ctxA, channelA.PortID, channelA.ChannelID)
	suite.True(found)
	suite.Equal(one, seq)

	// initialized channel has next seq recv of 1
	seq, found = suite.chainA.App.IBCKeeper.ChannelKeeper.GetNextSequenceRecv(ctxA, channelA.PortID, channelA.ChannelID)
	suite.True(found)
	suite.Equal(one, seq)

	// initialized channel has next seq ack of
	seq, found = suite.chainA.App.IBCKeeper.ChannelKeeper.GetNextSequenceAck(ctxA, channelA.PortID, channelA.ChannelID)
	suite.True(found)
	suite.Equal(one, seq)

	nextSeqSend, nextSeqRecv, nextSeqAck := uint64(10), uint64(10), uint64(10)
	suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(ctxA, channelA.PortID, channelA.ChannelID, nextSeqSend)
	suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(ctxA, channelA.PortID, channelA.ChannelID, nextSeqRecv)
	suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceAck(ctxA, channelA.PortID, channelA.ChannelID, nextSeqAck)

	storedNextSeqSend, found := suite.chainA.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(ctxA, channelA.PortID, channelA.ChannelID)
	suite.True(found)
	suite.Equal(nextSeqSend, storedNextSeqSend)

	storedNextSeqRecv, found := suite.chainA.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(ctxA, channelA.PortID, channelA.ChannelID)
	suite.True(found)
	suite.Equal(nextSeqRecv, storedNextSeqRecv)

	storedNextSeqAck, found := suite.chainA.App.IBCKeeper.ChannelKeeper.GetNextSequenceAck(ctxA, channelA.PortID, channelA.ChannelID)
	suite.True(found)
	suite.Equal(nextSeqAck, storedNextSeqAck)
}

// TestGetAllPacketCommitmentsAtChannel verifies that the keeper returns all stored packet
// commitments for a specific channel. The test will store consecutive commitments up to the
// value of "seq" and then add non-consecutive up to the value of "maxSeq". A final commitment
// with the value maxSeq + 1 is set on a different channel.
func (suite *KeeperTestSuite) TestGetAllPacketCommitmentsAtChannel() {
	_, _, connA, connB := suite.coordinator.Setup(suite.chainA.ChainID, suite.chainB.ChainID)
	channelA := connA.Channels[0]

	// create second channel
	channelA1, _ := suite.coordinator.CreateChannel(suite.chainA.ChainID, suite.chainB.ChainID, connA, connB, types.UNORDERED)

	ctxA := suite.chainA.GetContext()
	expectedSeqs := make(map[uint64]bool)
	hash := []byte("commitment")

	seq := uint64(15)
	maxSeq := uint64(25)
	suite.Require().Greater(maxSeq, seq)

	// create consecutive commitments
	for i := uint64(1); i < seq; i++ {
		suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(ctxA, channelA.PortID, channelA.ChannelID, i, hash)
		expectedSeqs[i] = true
	}

	// add non-consecutive commitments
	for i := seq; i < maxSeq; i += 2 {
		suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(ctxA, channelA.PortID, channelA.ChannelID, i, hash)
		expectedSeqs[i] = true
	}

	// add sequence on different channel/port
	suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(ctxA, channelA1.PortID, channelA1.ChannelID, maxSeq+1, hash)

	commitments := suite.chainA.App.IBCKeeper.ChannelKeeper.GetAllPacketCommitmentsAtChannel(ctxA, channelA.PortID, channelA.ChannelID)

	suite.Equal(len(expectedSeqs), len(commitments))
	// ensure above for loops occurred
	suite.NotEqual(0, len(commitments))

	// verify that all the packet commitments were stored
	for _, packet := range commitments {
		suite.True(expectedSeqs[packet.Sequence])
		suite.Equal(channelA.PortID, packet.PortID)
		suite.Equal(channelA.ChannelID, packet.ChannelID)
		suite.Equal(hash, packet.Hash)

		// prevent duplicates from passing checks
		expectedSeqs[packet.Sequence] = false
	}
}

// TestSetPacketAcknowledgement verifies that packet acknowledgements are correctly
// set in the keeper.
func (suite *KeeperTestSuite) TestSetPacketAcknowledgement() {
	_, _, connA, _ := suite.coordinator.Setup(suite.chainA.ChainID, suite.chainB.ChainID)
	channelA := connA.Channels[0]

	ctxA := suite.chainA.GetContext()
	seq := uint64(10)

	storedAckHash, found := suite.chainA.App.IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(ctxA, channelA.PortID, channelA.ChannelID, seq)
	suite.False(found)
	suite.Nil(storedAckHash)

	ackHash := []byte("ackhash")
	suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(ctxA, channelA.PortID, channelA.ChannelID, seq, ackHash)

	storedAckHash, found = suite.chainA.App.IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(ctxA, channelA.PortID, channelA.ChannelID, seq)
	suite.True(found)
	suite.Equal(ackHash, storedAckHash)
}
