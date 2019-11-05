package keeper

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypestm "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
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

func (suite *KeeperTestSuite) createClient() {
	store := suite.ctx.MultiStore().(sdk.CommitMultiStore)
	commitID := store.Commit()

	consensusState := clienttypestm.ConsensusState{
		ChainID: TestChainID,
		Height:  uint64(commitID.Version),
		Root:    commitment.NewRoot(commitID.Hash),
	}

	suite.clientKeeper.CreateClient(suite.ctx, TestClient, TestClientType, consensusState)
}

func (suite *KeeperTestSuite) createConnection(state connection.State) {
	connection := connection.ConnectionEnd{
		State:    state,
		ClientID: TestClient,
		Counterparty: connection.Counterparty{
			ClientID:     TestClient,
			ConnectionID: TestConnection,
			Prefix:       suite.connectionKeeper.GetCommitmentPrefix(),
		},
		Versions: connection.GetCompatibleVersions(),
	}

	suite.connectionKeeper.SetConnection(suite.ctx, TestConnection, connection)
}

func (suite *KeeperTestSuite) createChannel(portID string, chanID string, connID string, counterpartyPort string, counterpartyChan string, state types.State) {
	channel := types.Channel{
		State:    state,
		Ordering: TestChannelOrder,
		Counterparty: types.Counterparty{
			PortID:    counterpartyPort,
			ChannelID: counterpartyChan,
		},
		ConnectionHops: []string{connID},
		Version:        TestChannelVersion,
	}

	suite.channelKeeper.SetChannel(suite.ctx, portID, chanID, channel)
}

func (suite *KeeperTestSuite) deleteChannel(portID string, chanID string) {
	store := prefix.NewStore(suite.ctx.KVStore(suite.storeKey), suite.channelKeeper.prefix)
	store.Delete(types.KeyChannel(portID, chanID))
}

func (suite *KeeperTestSuite) bindPorts(portID string) {
	suite.portKeeper.BindPort(portID)
}

func (suite *KeeperTestSuite) updateClient() {
	store := suite.ctx.MultiStore().(sdk.CommitMultiStore)
	commitID := store.Commit()

	state := clienttypestm.ConsensusState{
		ChainID: TestChainID,
		Height:  uint64(commitID.Version),
		Root:    commitment.NewRoot(commitID.Hash),
	}

	suite.clientKeeper.SetConsensusState(suite.ctx, TestClient, state)
	suite.clientKeeper.SetVerifiedRoot(suite.ctx, TestClient, state.GetHeight(), state.GetRoot())
}

func (suite *KeeperTestSuite) queryProof(key string) (proof commitment.Proof, height int64) {
	store := suite.ctx.MultiStore().(*rootmulti.Store)
	res := store.Query(abci.RequestQuery{
		Path:  fmt.Sprintf("/%s/key", ibctypes.StoreKey),
		Data:  []byte(key),
		Prove: true,
	})

	height = res.Height
	proof = commitment.Proof{
		Proof: res.Proof,
	}

	return
}

func (suite *KeeperTestSuite) TestChanOpenInit() {
	counterparty := types.NewCounterparty(TestPort2, TestChannel2)
	capabilityKey := sdk.NewKVStoreKey(TestPort1)

	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.INIT)
	err := suite.channelKeeper.ChanOpenInit(suite.ctx, TestChannelOrder, []string{TestConnection}, TestPort1, TestChannel1, counterparty, TestChannelVersion, capabilityKey)
	suite.NotNil(err) // channel has already exist

	suite.deleteChannel(TestPort1, TestChannel1)
	err = suite.channelKeeper.ChanOpenInit(suite.ctx, TestChannelOrder, []string{TestConnection}, TestPort1, TestChannel1, counterparty, TestChannelVersion, capabilityKey)
	suite.NotNil(err) // connection does not exist

	suite.createConnection(connection.NONE)
	err = suite.channelKeeper.ChanOpenInit(suite.ctx, TestChannelOrder, []string{TestConnection}, TestPort1, TestChannel1, counterparty, TestChannelVersion, capabilityKey)
	suite.NotNil(err) // invalid connection state

	suite.createConnection(connection.OPEN)
	err = suite.channelKeeper.ChanOpenInit(suite.ctx, TestChannelOrder, []string{TestConnection}, TestPort1, TestChannel1, counterparty, TestChannelVersion, capabilityKey)
	suite.Nil(err) // successfully executedTestChannelVersion

	channel, found := suite.channelKeeper.GetChannel(suite.ctx, TestPort1, TestChannel1)
	suite.True(found)
	suite.Equal(types.INIT, channel.State)
}

func (suite *KeeperTestSuite) TestChanOpenTry() {
	counterparty := types.NewCounterparty(TestPort1, TestChannel1)
	capabilityKey := sdk.NewKVStoreKey(TestPort2)
	chanKey := fmt.Sprintf("%s/%s", types.SubModuleName, types.ChannelPath(TestPort1, TestChannel1))

	suite.createChannel(TestPort2, TestChannel2, TestConnection, TestPort1, TestChannel1, types.INIT)
	proofInit, proofHeight := commitment.Proof{}, int64(0)
	err := suite.channelKeeper.ChanOpenTry(suite.ctx, TestChannelOrder, []string{TestConnection}, TestPort2, TestChannel2, counterparty, TestChannelVersion, TestChannelVersion, proofInit, uint64(proofHeight), capabilityKey)
	suite.NotNil(err) // channel has already exist

	suite.deleteChannel(TestPort2, TestChannel2)
	err = suite.channelKeeper.ChanOpenTry(suite.ctx, TestChannelOrder, []string{TestConnection}, TestPort2, TestChannel2, counterparty, TestChannelVersion, TestChannelVersion, proofInit, uint64(proofHeight), capabilityKey)
	suite.NotNil(err) // connection does not exist

	suite.createConnection(connection.NONE)
	err = suite.channelKeeper.ChanOpenTry(suite.ctx, TestChannelOrder, []string{TestConnection}, TestPort2, TestChannel2, counterparty, TestChannelVersion, TestChannelVersion, proofInit, uint64(proofHeight), capabilityKey)
	suite.NotNil(err) // invalid connection state

	suite.createConnection(connection.OPEN)
	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.OPENTRY)
	suite.updateClient()
	proofInit, proofHeight = suite.queryProof(chanKey)
	err = suite.channelKeeper.ChanOpenTry(suite.ctx, TestChannelOrder, []string{TestConnection}, TestPort2, TestChannel2, counterparty, TestChannelVersion, TestChannelVersion, proofInit, uint64(proofHeight), capabilityKey)
	suite.NotNil(err) // channel membership verification failed due to invalid counterparty

	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.INIT)
	suite.updateClient()
	proofInit, proofHeight = suite.queryProof(chanKey)
	err = suite.channelKeeper.ChanOpenTry(suite.ctx, TestChannelOrder, []string{TestConnection}, TestPort2, TestChannel2, counterparty, TestChannelVersion, TestChannelVersion, proofInit, uint64(proofHeight), capabilityKey)
	suite.Nil(err) // successfully executed

	channel, found := suite.channelKeeper.GetChannel(suite.ctx, TestPort2, TestChannel2)
	suite.True(found)
	suite.Equal(types.OPENTRY, channel.State)
}

func (suite *KeeperTestSuite) TestChanOpenAck() {
	capabilityKey := sdk.NewKVStoreKey(TestPort1)
	chanKey := fmt.Sprintf("%s/%s", types.SubModuleName, types.ChannelPath(TestPort2, TestChannel2))

	suite.createChannel(TestPort2, TestChannel2, TestConnection, TestPort1, TestChannel1, types.OPENTRY)
	suite.updateClient()
	proofTry, proofHeight := suite.queryProof(chanKey)
	err := suite.channelKeeper.ChanOpenAck(suite.ctx, TestPort1, TestChannel1, TestChannelVersion, proofTry, uint64(proofHeight), capabilityKey)
	suite.NotNil(err) // channel does not exist

	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.CLOSED)
	suite.updateClient()
	proofTry, proofHeight = suite.queryProof(chanKey)
	err = suite.channelKeeper.ChanOpenAck(suite.ctx, TestPort1, TestChannel1, TestChannelVersion, proofTry, uint64(proofHeight), capabilityKey)
	suite.NotNil(err) // invalid channel state

	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.INIT)
	suite.updateClient()
	proofTry, proofHeight = suite.queryProof(chanKey)
	err = suite.channelKeeper.ChanOpenAck(suite.ctx, TestPort1, TestChannel1, TestChannelVersion, proofTry, uint64(proofHeight), capabilityKey)
	suite.NotNil(err) // connection does not exist

	suite.createConnection(connection.NONE)
	suite.updateClient()
	proofTry, proofHeight = suite.queryProof(chanKey)
	err = suite.channelKeeper.ChanOpenAck(suite.ctx, TestPort1, TestChannel1, TestChannelVersion, proofTry, uint64(proofHeight), capabilityKey)
	suite.NotNil(err) // invalid connection state

	suite.createConnection(connection.OPEN)
	suite.createChannel(TestPort2, TestChannel2, TestConnection, TestPort1, TestChannel1, types.OPEN)
	suite.updateClient()
	proofTry, proofHeight = suite.queryProof(chanKey)
	err = suite.channelKeeper.ChanOpenAck(suite.ctx, TestPort1, TestChannel1, TestChannelVersion, proofTry, uint64(proofHeight), capabilityKey)
	suite.NotNil(err) // channel membership verification failed due to invalid counterparty

	suite.createChannel(TestPort2, TestChannel2, TestConnection, TestPort1, TestChannel1, types.OPENTRY)
	suite.updateClient()
	proofTry, proofHeight = suite.queryProof(chanKey)
	err = suite.channelKeeper.ChanOpenAck(suite.ctx, TestPort1, TestChannel1, TestChannelVersion, proofTry, uint64(proofHeight), capabilityKey)
	suite.Nil(err) // successfully executed

	channel, found := suite.channelKeeper.GetChannel(suite.ctx, TestPort1, TestChannel1)
	suite.True(found)
	suite.Equal(types.OPEN, channel.State)
}

func (suite *KeeperTestSuite) TestChanOpenConfirm() {
	capabilityKey := sdk.NewKVStoreKey(TestPort2)
	chanKey := fmt.Sprintf("%s/%s", types.SubModuleName, types.ChannelPath(TestPort1, TestChannel1))

	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.OPEN)
	suite.updateClient()
	proofAck, proofHeight := suite.queryProof(chanKey)
	err := suite.channelKeeper.ChanOpenConfirm(suite.ctx, TestPort2, TestChannel2, proofAck, uint64(proofHeight), capabilityKey)
	suite.NotNil(err) // channel does not exist

	suite.createChannel(TestPort2, TestChannel2, TestConnection, TestPort1, TestChannel1, types.OPEN)
	suite.updateClient()
	proofAck, proofHeight = suite.queryProof(chanKey)
	err = suite.channelKeeper.ChanOpenConfirm(suite.ctx, TestPort2, TestChannel2, proofAck, uint64(proofHeight), capabilityKey)
	suite.NotNil(err) // invalid channel state

	suite.createChannel(TestPort2, TestChannel2, TestConnection, TestPort1, TestChannel1, types.OPENTRY)
	suite.updateClient()
	proofAck, proofHeight = suite.queryProof(chanKey)
	err = suite.channelKeeper.ChanOpenConfirm(suite.ctx, TestPort2, TestChannel2, proofAck, uint64(proofHeight), capabilityKey)
	suite.NotNil(err) // connection does not exist

	suite.createConnection(connection.NONE)
	suite.updateClient()
	proofAck, proofHeight = suite.queryProof(chanKey)
	err = suite.channelKeeper.ChanOpenConfirm(suite.ctx, TestPort2, TestChannel2, proofAck, uint64(proofHeight), capabilityKey)
	suite.NotNil(err) // invalid connection state

	suite.createConnection(connection.OPEN)
	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.OPENTRY)
	suite.updateClient()
	proofAck, proofHeight = suite.queryProof(chanKey)
	err = suite.channelKeeper.ChanOpenConfirm(suite.ctx, TestPort2, TestChannel2, proofAck, uint64(proofHeight), capabilityKey)
	suite.NotNil(err) // channel membership verification failed due to invalid counterparty

	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.OPEN)
	suite.updateClient()
	proofAck, proofHeight = suite.queryProof(chanKey)
	err = suite.channelKeeper.ChanOpenConfirm(suite.ctx, TestPort2, TestChannel2, proofAck, uint64(proofHeight), capabilityKey)
	suite.Nil(err) // successfully executed

	channel, found := suite.channelKeeper.GetChannel(suite.ctx, TestPort2, TestChannel2)
	suite.True(found)
	suite.Equal(types.OPEN, channel.State)
}

func (suite *KeeperTestSuite) TestChanCloseInit() {
	capabilityKey := sdk.NewKVStoreKey(TestPort1)

	err := suite.channelKeeper.ChanCloseInit(suite.ctx, TestPort1, TestChannel1, capabilityKey)
	suite.NotNil(err) // channel does not exist

	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.CLOSED)
	err = suite.channelKeeper.ChanCloseInit(suite.ctx, TestPort1, TestChannel1, capabilityKey)
	suite.NotNil(err) // channel is already closed

	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.OPEN)
	err = suite.channelKeeper.ChanCloseInit(suite.ctx, TestPort1, TestChannel1, capabilityKey)
	suite.NotNil(err) // connection does not exist

	suite.createConnection(connection.TRYOPEN)
	err = suite.channelKeeper.ChanCloseInit(suite.ctx, TestPort1, TestChannel1, capabilityKey)
	suite.NotNil(err) // invalid connection state

	suite.createConnection(connection.OPEN)
	err = suite.channelKeeper.ChanCloseInit(suite.ctx, TestPort1, TestChannel1, capabilityKey)
	suite.Nil(err) // successfully executed

	channel, found := suite.channelKeeper.GetChannel(suite.ctx, TestPort1, TestChannel1)
	suite.True(found)
	suite.Equal(types.CLOSED, channel.State)
}

func (suite *KeeperTestSuite) TestChanCloseConfirm() {
	capabilityKey := sdk.NewKVStoreKey(TestPort2)
	chanKey := fmt.Sprintf("%s/%s", types.SubModuleName, types.ChannelPath(TestPort1, TestChannel1))

	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.CLOSED)
	suite.updateClient()
	proofInit, proofHeight := suite.queryProof(chanKey)
	err := suite.channelKeeper.ChanCloseConfirm(suite.ctx, TestPort2, TestChannel2, proofInit, uint64(proofHeight), capabilityKey)
	suite.NotNil(err) // channel does not exist

	suite.createChannel(TestPort2, TestChannel2, TestConnection, TestPort1, TestChannel1, types.CLOSED)
	suite.updateClient()
	proofInit, proofHeight = suite.queryProof(chanKey)
	err = suite.channelKeeper.ChanCloseConfirm(suite.ctx, TestPort2, TestChannel2, proofInit, uint64(proofHeight), capabilityKey)
	suite.NotNil(err) // channel is already closed

	suite.createChannel(TestPort2, TestChannel2, TestConnection, TestPort1, TestChannel1, types.OPEN)
	suite.updateClient()
	proofInit, proofHeight = suite.queryProof(chanKey)
	err = suite.channelKeeper.ChanCloseConfirm(suite.ctx, TestPort2, TestChannel2, proofInit, uint64(proofHeight), capabilityKey)
	suite.NotNil(err) // connection does not exist

	suite.createConnection(connection.TRYOPEN)
	suite.updateClient()
	proofInit, proofHeight = suite.queryProof(chanKey)
	err = suite.channelKeeper.ChanCloseConfirm(suite.ctx, TestPort2, TestChannel2, proofInit, uint64(proofHeight), capabilityKey)
	suite.NotNil(err) // invalid connection state

	suite.createConnection(connection.OPEN)
	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.OPEN)
	suite.updateClient()
	proofInit, proofHeight = suite.queryProof(chanKey)
	err = suite.channelKeeper.ChanCloseConfirm(suite.ctx, TestPort2, TestChannel2, proofInit, uint64(proofHeight), capabilityKey)
	suite.NotNil(err) // channel membership verification failed due to invalid counterparty

	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.CLOSED)
	suite.updateClient()
	proofInit, proofHeight = suite.queryProof(chanKey)
	err = suite.channelKeeper.ChanCloseConfirm(suite.ctx, TestPort2, TestChannel2, proofInit, uint64(proofHeight), capabilityKey)
	suite.Nil(err) // successfully executed

	channel, found := suite.channelKeeper.GetChannel(suite.ctx, TestPort2, TestChannel2)
	suite.True(found)
	suite.Equal(types.CLOSED, channel.State)
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
