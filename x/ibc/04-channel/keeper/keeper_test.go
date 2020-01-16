package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// define constants used for testing
const (
	testChainID    = "test-chain-id"
	testClient     = "test-client"
	testClientType = clientexported.Tendermint

	testConnection = "testconnection"

	testPort1 = "firstport"
	testPort2 = "secondport"
	testPort3 = "thirdport"

	testChannel1 = "firstchannel"
	testChannel2 = "secondchannel"
	testChannel3 = "thirdchannel"

	testChannelOrder   = exported.ORDERED
	testChannelVersion = "1.0"
)

type KeeperTestSuite struct {
	suite.Suite

	cdc    *codec.Codec
	ctx    sdk.Context
	app    *simapp.SimApp
	valSet *tmtypes.ValidatorSet
}

func (suite *KeeperTestSuite) SetupTest() {
	isCheckTx := false
	app := simapp.Setup(isCheckTx)

	suite.cdc = app.Codec()
	suite.ctx = app.BaseApp.NewContext(isCheckTx, abci.Header{})
	suite.app = app

	privVal := tmtypes.NewMockPV()

	validator := tmtypes.NewValidator(privVal.GetPubKey(), 1)
	suite.valSet = tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})

	suite.createClient()
}

func (suite *KeeperTestSuite) TestSetChannel() {
	_, found := suite.app.IBCKeeper.ChannelKeeper.GetChannel(suite.ctx, testPort1, testChannel1)
	suite.False(found)

	channel := types.Channel{
		State:    exported.OPEN,
		Ordering: testChannelOrder,
		Counterparty: types.Counterparty{
			PortID:    testPort2,
			ChannelID: testChannel2,
		},
		ConnectionHops: []string{testConnection},
		Version:        testChannelVersion,
	}
	suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, testPort1, testChannel1, channel)

	storedChannel, found := suite.app.IBCKeeper.ChannelKeeper.GetChannel(suite.ctx, testPort1, testChannel1)
	suite.True(found)
	suite.Equal(channel, storedChannel)
}

func (suite KeeperTestSuite) TestGetAllChannels() {
	// Channel (Counterparty): A(C) -> C(B) -> B(A)
	counterparty1 := types.NewCounterparty(testPort1, testChannel1)
	counterparty2 := types.NewCounterparty(testPort2, testChannel2)
	counterparty3 := types.NewCounterparty(testPort3, testChannel3)

	channel1 := types.Channel{
		State:          exported.INIT,
		Ordering:       testChannelOrder,
		Counterparty:   counterparty3,
		ConnectionHops: []string{testConnection},
		Version:        testChannelVersion,
	}

	channel2 := types.Channel{
		State:          exported.INIT,
		Ordering:       testChannelOrder,
		Counterparty:   counterparty1,
		ConnectionHops: []string{testConnection},
		Version:        testChannelVersion,
	}

	channel3 := types.Channel{
		State:          exported.CLOSED,
		Ordering:       testChannelOrder,
		Counterparty:   counterparty2,
		ConnectionHops: []string{testConnection},
		Version:        testChannelVersion,
	}

	expChannels := []types.Channel{channel1, channel2, channel3}

	suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, testPort1, testChannel1, expChannels[0])
	suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, testPort2, testChannel2, expChannels[1])
	suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, testPort3, testChannel3, expChannels[2])

	channels := suite.app.IBCKeeper.ChannelKeeper.GetAllChannels(suite.ctx)
	suite.Require().Len(channels, len(expChannels))
	suite.Require().Equal(expChannels, channels)
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
