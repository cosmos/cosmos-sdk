package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
)

// define constants used for testing
const (
	testChainID    = "test-chain-id"
	testClient     = "test-client"
	testClientType = clientexported.Tendermint

	testConnection = "testconnection"
	testPort1      = "firstport"
	testPort2      = "secondport"
	testChannel1   = "firstchannel"
	testChannel2   = "secondchannel"

	testChannelOrder   = types.ORDERED
	testChannelVersion = "1.0"
)

type KeeperTestSuite struct {
	suite.Suite

	cdc *codec.Codec
	ctx sdk.Context
	app *simapp.SimApp
}

func (suite *KeeperTestSuite) SetupTest() {
	isCheckTx := false
	app := simapp.Setup(isCheckTx)

	suite.cdc = app.Codec()
	suite.ctx = app.BaseApp.NewContext(isCheckTx, abci.Header{})
	suite.app = app

	suite.createClient()
}

func (suite *KeeperTestSuite) TestSetChannel() {
	_, found := suite.app.IBCKeeper.ChannelKeeper.GetChannel(suite.ctx, testPort1, testChannel1)
	suite.False(found)

	channel := types.Channel{
		State:    types.OPEN,
		Ordering: testChannelOrder,
		Counterparty: types.Counterparty{
			PortID:    testPort1,
			ChannelID: testChannel1,
		},
		ConnectionHops: []string{testConnection},
		Version:        testChannelVersion,
	}
	suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, testPort1, testChannel1, channel)

	storedChannel, found := suite.app.IBCKeeper.ChannelKeeper.GetChannel(suite.ctx, testPort1, testChannel1)
	suite.True(found)
	suite.Equal(channel, storedChannel)
}

func (suite *KeeperTestSuite) TestSetChannelCapability() {
	_, found := suite.app.IBCKeeper.ChannelKeeper.GetChannelCapability(suite.ctx, testPort1, testChannel1)
	suite.False(found)

	channelCap := "test-channel-capability"
	suite.app.IBCKeeper.ChannelKeeper.SetChannelCapability(suite.ctx, testPort1, testChannel1, channelCap)

	storedChannelCap, found := suite.app.IBCKeeper.ChannelKeeper.GetChannelCapability(suite.ctx, testPort1, testChannel1)
	suite.True(found)
	suite.Equal(channelCap, storedChannelCap)
}

func (suite *KeeperTestSuite) TestSetSequence() {
	_, found := suite.app.IBCKeeper.ChannelKeeper.GetNextSequenceSend(suite.ctx, testPort1, testChannel1)
	suite.False(found)

	_, found = suite.app.IBCKeeper.ChannelKeeper.GetNextSequenceRecv(suite.ctx, testPort1, testChannel1)
	suite.False(found)

	nextSeqSend, nextSeqRecv := uint64(10), uint64(10)
	suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, testPort1, testChannel1, nextSeqSend)
	suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.ctx, testPort1, testChannel1, nextSeqRecv)

	storedNextSeqSend, found := suite.app.IBCKeeper.ChannelKeeper.GetNextSequenceSend(suite.ctx, testPort1, testChannel1)
	suite.True(found)
	suite.Equal(nextSeqSend, storedNextSeqSend)

	storedNextSeqRecv, found := suite.app.IBCKeeper.ChannelKeeper.GetNextSequenceSend(suite.ctx, testPort1, testChannel1)
	suite.True(found)
	suite.Equal(nextSeqRecv, storedNextSeqRecv)
}

func (suite *KeeperTestSuite) TestPackageCommitment() {
	seq := uint64(10)
	storedCommitment := suite.app.IBCKeeper.ChannelKeeper.GetPacketCommitment(suite.ctx, testPort1, testChannel1, seq)
	suite.Equal([]byte(nil), storedCommitment)

	commitment := []byte("commitment")
	suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort1, testChannel1, seq, commitment)

	storedCommitment = suite.app.IBCKeeper.ChannelKeeper.GetPacketCommitment(suite.ctx, testPort1, testChannel1, seq)
	suite.Equal(commitment, storedCommitment)
}

func (suite *KeeperTestSuite) TestSetPacketAcknowledgement() {
	seq := uint64(10)

	storedAckHash, found := suite.app.IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(suite.ctx, testPort1, testChannel1, seq)
	suite.False(found)
	suite.Nil(storedAckHash)

	ackHash := []byte("ackhash")
	suite.app.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(suite.ctx, testPort1, testChannel1, seq, ackHash)

	storedAckHash, found = suite.app.IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(suite.ctx, testPort1, testChannel1, seq)
	suite.True(found)
	suite.Equal(ackHash, storedAckHash)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
