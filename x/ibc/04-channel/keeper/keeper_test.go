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
// and existence of a channel in INIT on chainB.
func (suite *KeeperTestSuite) TestSetChannel() {
	chainA := suite.coordinator.GetChain(suite.chainA)
	chainB := suite.coordinator.GetChain(suite.chainB)

	// create client and connections on both chains
	_, _, connA, connB := suite.coordinator.CreateClientsAndConnections(suite.chainA, suite.chainB, clientexported.Tendermint)

	// check for channel to be created on chainB
	channelB := connB.NextTestChannel()
	_, found := chainB.App.IBCKeeper.ChannelKeeper.GetChannel(chainB.GetContext(), channelB.PortID, channelB.ChannelID)
	suite.False(found)

	// init channel
	channelA, channelB, err := suite.coordinator.CreateChannelInit(chainB, chainA, connB, connA, channeltypes.ORDERED)
	suite.NoError(err)

	storedChannel, found := chainB.App.IBCKeeper.ChannelKeeper.GetChannel(chainB.GetContext(), channelB.PortID, channelB.ChannelID)
	suite.True(found)
	suite.Equal(channeltypes.INIT, storedChannel.State)
	suite.Equal(channeltypes.ORDERED, storedChannel.Ordering)
	suite.Equal(channeltypes.NewCounterparty(channelA.PortID, channelA.ChannelID), storedChannel.Counterparty)
}

/*
func (suite KeeperTestSuite) TestGetAllChannels() {
	// Channel (Counterparty): A(C) -> C(B) -> B(A)
	counterparty1 := types.NewCounterparty(testPort1, testChannel1)
	counterparty2 := types.NewCounterparty(testPort2, testChannel2)
	counterparty3 := types.NewCounterparty(testPort3, testChannel3)

	channel1 := types.NewChannel(
		types.INIT, testChannelOrder,
		counterparty3, []string{testConnectionIDA}, testChannelVersion,
	)
	channel2 := types.NewChannel(
		types.INIT, testChannelOrder,
		counterparty1, []string{testConnectionIDA}, testChannelVersion,
	)
	channel3 := types.NewChannel(
		types.CLOSED, testChannelOrder,
		counterparty2, []string{testConnectionIDA}, testChannelVersion,
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

func commitNBlocks(chain *TestChain, n int) {
	for i := 0; i < n; i++ {
		chain.App.Commit()
		chain.App.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: chain.App.LastBlockHeight() + 1}})
	}
}

// commit current block and start the next block with the provided time
func commitBlockWithNewTimestamp(chain *TestChain, timestamp int64) {
	chain.App.Commit()
	chain.App.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: chain.App.LastBlockHeight() + 1, Time: time.Unix(timestamp, 0)}})
}

// Mocked types

type mockSuccessPacket struct{}

// GetBytes returns the serialised packet data
func (mp mockSuccessPacket) GetBytes() []byte { return []byte("THIS IS A SUCCESS PACKET") }

type mockFailPacket struct{}

// GetBytes returns the serialised packet data (without timeout)
func (mp mockFailPacket) GetBytes() []byte { return []byte("THIS IS A FAILURE PACKET") }
*/
