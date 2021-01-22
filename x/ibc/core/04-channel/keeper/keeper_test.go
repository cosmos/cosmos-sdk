package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

// KeeperTestSuite is a testing suite to test keeper functions.
type KeeperTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
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
	// commit some blocks so that QueryProof returns valid proof (cannot return valid query if height <= 1)
	suite.coordinator.CommitNBlocks(suite.chainA, 2)
	suite.coordinator.CommitNBlocks(suite.chainB, 2)
}

// TestSetChannel create clients and connections on both chains. It tests for the non-existence
// and existence of a channel in INIT on chainA.
func (suite *KeeperTestSuite) TestSetChannel() {
	// create client and connections on both chains
	_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)

	// check for channel to be created on chainA
	channelA := suite.chainA.NextTestChannel(connA, ibctesting.MockPort)
	_, found := suite.chainA.App.IBCKeeper.ChannelKeeper.GetChannel(suite.chainA.GetContext(), channelA.PortID, channelA.ID)
	suite.False(found)

	// init channel
	channelA, channelB, err := suite.coordinator.ChanOpenInit(suite.chainA, suite.chainB, connA, connB, ibctesting.MockPort, ibctesting.MockPort, types.ORDERED)
	suite.NoError(err)

	storedChannel, found := suite.chainA.App.IBCKeeper.ChannelKeeper.GetChannel(suite.chainA.GetContext(), channelA.PortID, channelA.ID)
	// counterparty channel id is empty after open init
	expectedCounterparty := types.NewCounterparty(channelB.PortID, "")

	suite.True(found)
	suite.Equal(types.INIT, storedChannel.State)
	suite.Equal(types.ORDERED, storedChannel.Ordering)
	suite.Equal(expectedCounterparty, storedChannel.Counterparty)
}

// TestGetAllChannels creates multiple channels on chain A through various connections
// and tests their retrieval. 2 channels are on connA0 and 1 channel is on connA1
func (suite KeeperTestSuite) TestGetAllChannels() {
	clientA, clientB, connA0, connB0, testchannel0, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
	// channel0 on first connection on chainA
	counterparty0 := types.Counterparty{
		PortId:    connB0.Channels[0].PortID,
		ChannelId: connB0.Channels[0].ID,
	}

	// channel1 is second channel on first connection on chainA
	testchannel1, _ := suite.coordinator.CreateMockChannels(suite.chainA, suite.chainB, connA0, connB0, types.ORDERED)
	counterparty1 := types.Counterparty{
		PortId:    connB0.Channels[1].PortID,
		ChannelId: connB0.Channels[1].ID,
	}

	connA1, connB1 := suite.coordinator.CreateConnection(suite.chainA, suite.chainB, clientA, clientB)

	// channel2 is on a second connection on chainA
	testchannel2, _, err := suite.coordinator.ChanOpenInit(suite.chainA, suite.chainB, connA1, connB1, ibctesting.MockPort, ibctesting.MockPort, types.UNORDERED)
	suite.Require().NoError(err)

	// counterparty channel id is empty after open init
	counterparty2 := types.Counterparty{
		PortId:    connB1.Channels[0].PortID,
		ChannelId: "",
	}

	channel0 := types.NewChannel(
		types.OPEN, types.UNORDERED,
		counterparty0, []string{connA0.ID}, testchannel0.Version,
	)
	channel1 := types.NewChannel(
		types.OPEN, types.ORDERED,
		counterparty1, []string{connA0.ID}, testchannel1.Version,
	)
	channel2 := types.NewChannel(
		types.INIT, types.UNORDERED,
		counterparty2, []string{connA1.ID}, testchannel2.Version,
	)

	expChannels := []types.IdentifiedChannel{
		types.NewIdentifiedChannel(testchannel0.PortID, testchannel0.ID, channel0),
		types.NewIdentifiedChannel(testchannel1.PortID, testchannel1.ID, channel1),
		types.NewIdentifiedChannel(testchannel2.PortID, testchannel2.ID, channel2),
	}

	ctxA := suite.chainA.GetContext()

	channels := suite.chainA.App.IBCKeeper.ChannelKeeper.GetAllChannels(ctxA)
	suite.Require().Len(channels, len(expChannels))
	suite.Require().Equal(expChannels, channels)
}

// TestGetAllSequences sets all packet sequences for two different channels on chain A and
// tests their retrieval.
func (suite KeeperTestSuite) TestGetAllSequences() {
	_, _, connA, connB, channelA0, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
	channelA1, _ := suite.coordinator.CreateMockChannels(suite.chainA, suite.chainB, connA, connB, types.UNORDERED)

	seq1 := types.NewPacketSequence(channelA0.PortID, channelA0.ID, 1)
	seq2 := types.NewPacketSequence(channelA0.PortID, channelA0.ID, 2)
	seq3 := types.NewPacketSequence(channelA1.PortID, channelA1.ID, 3)

	// seq1 should be overwritten by seq2
	expSeqs := []types.PacketSequence{seq2, seq3}

	ctxA := suite.chainA.GetContext()

	for _, seq := range []types.PacketSequence{seq1, seq2, seq3} {
		suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(ctxA, seq.PortId, seq.ChannelId, seq.Sequence)
		suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(ctxA, seq.PortId, seq.ChannelId, seq.Sequence)
		suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceAck(ctxA, seq.PortId, seq.ChannelId, seq.Sequence)
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

// TestGetAllPacketState creates a set of acks, packet commitments, and receipts on two different
// channels on chain A and tests their retrieval.
func (suite KeeperTestSuite) TestGetAllPacketState() {
	_, _, connA, connB, channelA0, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
	channelA1, _ := suite.coordinator.CreateMockChannels(suite.chainA, suite.chainB, connA, connB, types.UNORDERED)

	// channel 0 acks
	ack1 := types.NewPacketState(channelA0.PortID, channelA0.ID, 1, []byte("ack"))
	ack2 := types.NewPacketState(channelA0.PortID, channelA0.ID, 2, []byte("ack"))

	// duplicate ack
	ack2dup := types.NewPacketState(channelA0.PortID, channelA0.ID, 2, []byte("ack"))

	// channel 1 acks
	ack3 := types.NewPacketState(channelA1.PortID, channelA1.ID, 1, []byte("ack"))

	// create channel 0 receipts
	receipt := string([]byte{byte(1)})
	rec1 := types.NewPacketState(channelA0.PortID, channelA0.ID, 1, []byte(receipt))
	rec2 := types.NewPacketState(channelA0.PortID, channelA0.ID, 2, []byte(receipt))

	// channel 1 receipts
	rec3 := types.NewPacketState(channelA1.PortID, channelA1.ID, 1, []byte(receipt))
	rec4 := types.NewPacketState(channelA1.PortID, channelA1.ID, 2, []byte(receipt))

	// channel 0 packet commitments
	comm1 := types.NewPacketState(channelA0.PortID, channelA0.ID, 1, []byte("hash"))
	comm2 := types.NewPacketState(channelA0.PortID, channelA0.ID, 2, []byte("hash"))

	// channel 1 packet commitments
	comm3 := types.NewPacketState(channelA1.PortID, channelA1.ID, 1, []byte("hash"))
	comm4 := types.NewPacketState(channelA1.PortID, channelA1.ID, 2, []byte("hash"))

	expAcks := []types.PacketState{ack1, ack2, ack3}
	expReceipts := []types.PacketState{rec1, rec2, rec3, rec4}
	expCommitments := []types.PacketState{comm1, comm2, comm3, comm4}

	ctxA := suite.chainA.GetContext()

	// set acknowledgements
	for _, ack := range []types.PacketState{ack1, ack2, ack2dup, ack3} {
		suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(ctxA, ack.PortId, ack.ChannelId, ack.Sequence, ack.Data)
	}

	// set packet receipts
	for _, rec := range expReceipts {
		suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketReceipt(ctxA, rec.PortId, rec.ChannelId, rec.Sequence)
	}

	// set packet commitments
	for _, comm := range expCommitments {
		suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(ctxA, comm.PortId, comm.ChannelId, comm.Sequence, comm.Data)
	}

	acks := suite.chainA.App.IBCKeeper.ChannelKeeper.GetAllPacketAcks(ctxA)
	receipts := suite.chainA.App.IBCKeeper.ChannelKeeper.GetAllPacketReceipts(ctxA)
	commitments := suite.chainA.App.IBCKeeper.ChannelKeeper.GetAllPacketCommitments(ctxA)

	suite.Require().Len(acks, len(expAcks))
	suite.Require().Len(commitments, len(expCommitments))
	suite.Require().Len(receipts, len(expReceipts))

	suite.Require().Equal(expAcks, acks)
	suite.Require().Equal(expReceipts, receipts)
	suite.Require().Equal(expCommitments, commitments)
}

// TestSetSequence verifies that the keeper correctly sets the sequence counters.
func (suite *KeeperTestSuite) TestSetSequence() {
	_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)

	ctxA := suite.chainA.GetContext()
	one := uint64(1)

	// initialized channel has next send seq of 1
	seq, found := suite.chainA.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(ctxA, channelA.PortID, channelA.ID)
	suite.True(found)
	suite.Equal(one, seq)

	// initialized channel has next seq recv of 1
	seq, found = suite.chainA.App.IBCKeeper.ChannelKeeper.GetNextSequenceRecv(ctxA, channelA.PortID, channelA.ID)
	suite.True(found)
	suite.Equal(one, seq)

	// initialized channel has next seq ack of
	seq, found = suite.chainA.App.IBCKeeper.ChannelKeeper.GetNextSequenceAck(ctxA, channelA.PortID, channelA.ID)
	suite.True(found)
	suite.Equal(one, seq)

	nextSeqSend, nextSeqRecv, nextSeqAck := uint64(10), uint64(10), uint64(10)
	suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(ctxA, channelA.PortID, channelA.ID, nextSeqSend)
	suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(ctxA, channelA.PortID, channelA.ID, nextSeqRecv)
	suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceAck(ctxA, channelA.PortID, channelA.ID, nextSeqAck)

	storedNextSeqSend, found := suite.chainA.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(ctxA, channelA.PortID, channelA.ID)
	suite.True(found)
	suite.Equal(nextSeqSend, storedNextSeqSend)

	storedNextSeqRecv, found := suite.chainA.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(ctxA, channelA.PortID, channelA.ID)
	suite.True(found)
	suite.Equal(nextSeqRecv, storedNextSeqRecv)

	storedNextSeqAck, found := suite.chainA.App.IBCKeeper.ChannelKeeper.GetNextSequenceAck(ctxA, channelA.PortID, channelA.ID)
	suite.True(found)
	suite.Equal(nextSeqAck, storedNextSeqAck)
}

// TestGetAllPacketCommitmentsAtChannel verifies that the keeper returns all stored packet
// commitments for a specific channel. The test will store consecutive commitments up to the
// value of "seq" and then add non-consecutive up to the value of "maxSeq". A final commitment
// with the value maxSeq + 1 is set on a different channel.
func (suite *KeeperTestSuite) TestGetAllPacketCommitmentsAtChannel() {
	_, _, connA, connB, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)

	// create second channel
	channelA1, _ := suite.coordinator.CreateMockChannels(suite.chainA, suite.chainB, connA, connB, types.UNORDERED)

	ctxA := suite.chainA.GetContext()
	expectedSeqs := make(map[uint64]bool)
	hash := []byte("commitment")

	seq := uint64(15)
	maxSeq := uint64(25)
	suite.Require().Greater(maxSeq, seq)

	// create consecutive commitments
	for i := uint64(1); i < seq; i++ {
		suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(ctxA, channelA.PortID, channelA.ID, i, hash)
		expectedSeqs[i] = true
	}

	// add non-consecutive commitments
	for i := seq; i < maxSeq; i += 2 {
		suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(ctxA, channelA.PortID, channelA.ID, i, hash)
		expectedSeqs[i] = true
	}

	// add sequence on different channel/port
	suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(ctxA, channelA1.PortID, channelA1.ID, maxSeq+1, hash)

	commitments := suite.chainA.App.IBCKeeper.ChannelKeeper.GetAllPacketCommitmentsAtChannel(ctxA, channelA.PortID, channelA.ID)

	suite.Equal(len(expectedSeqs), len(commitments))
	// ensure above for loops occurred
	suite.NotEqual(0, len(commitments))

	// verify that all the packet commitments were stored
	for _, packet := range commitments {
		suite.True(expectedSeqs[packet.Sequence])
		suite.Equal(channelA.PortID, packet.PortId)
		suite.Equal(channelA.ID, packet.ChannelId)
		suite.Equal(hash, packet.Data)

		// prevent duplicates from passing checks
		expectedSeqs[packet.Sequence] = false
	}
}

// TestSetPacketAcknowledgement verifies that packet acknowledgements are correctly
// set in the keeper.
func (suite *KeeperTestSuite) TestSetPacketAcknowledgement() {
	_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)

	ctxA := suite.chainA.GetContext()
	seq := uint64(10)

	storedAckHash, found := suite.chainA.App.IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(ctxA, channelA.PortID, channelA.ID, seq)
	suite.Require().False(found)
	suite.Require().Nil(storedAckHash)

	ackHash := []byte("ackhash")
	suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(ctxA, channelA.PortID, channelA.ID, seq, ackHash)

	storedAckHash, found = suite.chainA.App.IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(ctxA, channelA.PortID, channelA.ID, seq)
	suite.Require().True(found)
	suite.Require().Equal(ackHash, storedAckHash)
	suite.Require().True(suite.chainA.App.IBCKeeper.ChannelKeeper.HasPacketAcknowledgement(ctxA, channelA.PortID, channelA.ID, seq))
}
