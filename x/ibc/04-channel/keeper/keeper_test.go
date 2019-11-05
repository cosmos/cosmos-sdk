package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/store"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	port "github.com/cosmos/cosmos-sdk/x/ibc/05-port"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

// define constants used for testing
const (
	TestChainID    = "test-chain-id"
	TestClient     = "test-client"
	TestClientType = clientexported.Tendermint

	TestConnection = "testconnection"
	TestPort1      = "firsttestport"
	TestPort2      = "secondtestport"
	TestChannel1   = "firstchannel"
	TestChannel2   = "secondchannel"

	TestChannelOrder   = types.ORDERED
	TestChannelVersion = ""
)

type KeeperTestSuite struct {
	suite.Suite

	cdc      *codec.Codec
	ctx      sdk.Context
	storeKey sdk.StoreKey

	clientKeeper     client.Keeper
	connectionKeeper connection.Keeper
	portKeeper       port.Keeper
	channelKeeper    Keeper
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.storeKey = sdk.NewKVStoreKey(ibctypes.StoreKey)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(suite.storeKey, sdk.StoreTypeIAVL, db)
	if err := ms.LoadLatestVersion(); err != nil {
		panic(err)
	}

	suite.cdc = codec.New()
	codec.RegisterCrypto(suite.cdc)
	client.RegisterCodec(suite.cdc)
	types.RegisterCodec(suite.cdc)
	commitment.RegisterCodec(suite.cdc)

	suite.ctx = sdk.NewContext(ms, abci.Header{Height: 0}, false, log.NewNopLogger())

	suite.clientKeeper = client.NewKeeper(suite.cdc, suite.storeKey, client.DefaultCodespace)
	suite.connectionKeeper = connection.NewKeeper(suite.cdc, suite.storeKey, connection.DefaultCodespace, suite.clientKeeper)
	suite.portKeeper = port.NewKeeper(suite.cdc, suite.storeKey, port.DefaultCodespace)
	suite.channelKeeper = NewKeeper(suite.cdc, suite.storeKey, types.DefaultCodespace, suite.clientKeeper, suite.connectionKeeper, suite.portKeeper)

	suite.createClient()
}

func (suite *KeeperTestSuite) TestSetChannel() {
	storedChannel, found := suite.channelKeeper.GetChannel(suite.ctx, TestPort1, TestChannel1)
	suite.False(found)

	channel := types.Channel{
		State:    types.OPEN,
		Ordering: TestChannelOrder,
		Counterparty: types.Counterparty{
			PortID:    TestPort1,
			ChannelID: TestChannel1,
		},
		ConnectionHops: []string{TestConnection},
		Version:        TestChannelVersion,
	}
	suite.channelKeeper.SetChannel(suite.ctx, TestPort1, TestChannel1, channel)

	storedChannel, found = suite.channelKeeper.GetChannel(suite.ctx, TestPort1, TestChannel1)
	suite.True(found)
	suite.Equal(channel, storedChannel)
}

func (suite *KeeperTestSuite) TestSetChannelCapability() {
	storedChannelCap, found := suite.channelKeeper.GetChannelCapability(suite.ctx, TestPort1, TestChannel1)
	suite.False(found)

	channelCap := "test-channel-capability"
	suite.channelKeeper.SetChannelCapability(suite.ctx, TestPort1, TestChannel1, channelCap)

	storedChannelCap, found = suite.channelKeeper.GetChannelCapability(suite.ctx, TestPort1, TestChannel1)
	suite.True(found)
	suite.Equal(channelCap, storedChannelCap)
}

func (suite *KeeperTestSuite) TestSetSequence() {
	storedNextSeqSend, found := suite.channelKeeper.GetNextSequenceSend(suite.ctx, TestPort1, TestChannel1)
	suite.False(found)

	storedNextSeqRecv, found := suite.channelKeeper.GetNextSequenceRecv(suite.ctx, TestPort1, TestChannel1)
	suite.False(found)

	nextSeqSend, nextSeqRecv := uint64(10), uint64(10)
	suite.channelKeeper.SetNextSequenceSend(suite.ctx, TestPort1, TestChannel1, nextSeqSend)
	suite.channelKeeper.SetNextSequenceRecv(suite.ctx, TestPort1, TestChannel1, nextSeqRecv)

	storedNextSeqSend, found = suite.channelKeeper.GetNextSequenceSend(suite.ctx, TestPort1, TestChannel1)
	suite.True(found)
	suite.Equal(nextSeqSend, storedNextSeqSend)

	storedNextSeqRecv, found = suite.channelKeeper.GetNextSequenceSend(suite.ctx, TestPort1, TestChannel1)
	suite.True(found)
	suite.Equal(nextSeqRecv, storedNextSeqRecv)
}

func (suite *KeeperTestSuite) TestPackageCommitment() {
	seq := uint64(10)
	storedCommitment := suite.channelKeeper.GetPacketCommitment(suite.ctx, TestPort1, TestChannel1, seq)
	suite.Equal([]byte(nil), storedCommitment)

	commitment := []byte("commitment")
	suite.channelKeeper.SetPacketCommitment(suite.ctx, TestPort1, TestChannel1, seq, commitment)

	storedCommitment = suite.channelKeeper.GetPacketCommitment(suite.ctx, TestPort1, TestChannel1, seq)
	suite.Equal(commitment, storedCommitment)

	suite.channelKeeper.deletePacketCommitment(suite.ctx, TestPort1, TestChannel1, seq)
	storedCommitment = suite.channelKeeper.GetPacketCommitment(suite.ctx, TestPort1, TestChannel1, seq)
	suite.Equal([]byte(nil), storedCommitment)
}

func (suite *KeeperTestSuite) TestSetPacketAcknowledgement() {
	store := prefix.NewStore(suite.ctx.KVStore(suite.storeKey), suite.channelKeeper.prefix)
	seq := uint64(10)

	storedAckHash := store.Get(types.KeyPacketAcknowledgement(TestPort1, TestChannel1, seq))
	suite.Equal([]byte(nil), storedAckHash)

	ackHash := []byte("ackhash")
	suite.channelKeeper.SetPacketAcknowledgement(suite.ctx, TestPort1, TestChannel1, seq, ackHash)

	storedAckHash = store.Get(types.KeyPacketAcknowledgement(TestPort1, TestChannel1, seq))
	suite.Equal(ackHash, storedAckHash)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
