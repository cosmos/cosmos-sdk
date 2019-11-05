package keeper

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"

	"github.com/cosmos/cosmos-sdk/codec"
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

	TestChannelOrder = types.ORDERED
	TestVersion      = ""
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
	suite.createConnection()
	suite.bindPorts()
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

func (suite *KeeperTestSuite) createConnection() {
	connection := connection.ConnectionEnd{
		State:    connection.OPEN,
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
		Version:        TestVersion,
	}

	suite.channelKeeper.SetChannel(suite.ctx, portID, chanID, channel)
}

func (suite *KeeperTestSuite) bindPorts() {
	suite.portKeeper.BindPort(TestPort1)
	suite.portKeeper.BindPort(TestPort2)
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

	err := suite.channelKeeper.ChanOpenInit(suite.ctx, TestChannelOrder, []string{TestConnection}, TestPort1, TestChannel1, counterparty, TestVersion, capabilityKey)
	suite.Nil(err)

	channel, found := suite.channelKeeper.GetChannel(suite.ctx, TestPort1, TestChannel1)
	suite.True(found)
	suite.Equal(types.INIT, channel.State)
}

func (suite *KeeperTestSuite) TestChanOpenTry() {
	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.INIT)
	suite.updateClient()

	chanKey := fmt.Sprintf("%s/%s", types.SubModuleName, types.ChannelPath(TestPort1, TestChannel1))
	proofInit, proofHeight := suite.queryProof(chanKey)

	counterparty := types.NewCounterparty(TestPort1, TestChannel1)
	capabilityKey := sdk.NewKVStoreKey(TestPort2)

	err := suite.channelKeeper.ChanOpenTry(suite.ctx, TestChannelOrder, []string{TestConnection}, TestPort2, TestChannel2, counterparty, TestVersion, TestVersion, proofInit, uint64(proofHeight), capabilityKey)
	suite.Nil(err)

	channel, found := suite.channelKeeper.GetChannel(suite.ctx, TestPort2, TestChannel2)
	suite.True(found)
	suite.Equal(types.OPENTRY, channel.State)
}

func (suite *KeeperTestSuite) TestChanOpenAck() {
	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.INIT)
	suite.createChannel(TestPort2, TestChannel2, TestConnection, TestPort1, TestChannel1, types.OPENTRY)
	suite.updateClient()

	chanKey := fmt.Sprintf("%s/%s", types.SubModuleName, types.ChannelPath(TestPort2, TestChannel2))
	proofTry, proofHeight := suite.queryProof(chanKey)

	capabilityKey := sdk.NewKVStoreKey(TestPort1)

	err := suite.channelKeeper.ChanOpenAck(suite.ctx, TestPort1, TestChannel1, TestVersion, proofTry, uint64(proofHeight), capabilityKey)
	suite.Nil(err)

	channel, found := suite.channelKeeper.GetChannel(suite.ctx, TestPort1, TestChannel1)
	suite.True(found)
	suite.Equal(types.OPEN, channel.State)
}

func (suite *KeeperTestSuite) TestChanOpenConfirm() {
	suite.createChannel(TestPort2, TestChannel2, TestConnection, TestPort1, TestChannel1, types.OPENTRY)
	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.OPEN)
	suite.updateClient()

	chanKey := fmt.Sprintf("%s/%s", types.SubModuleName, types.ChannelPath(TestPort1, TestChannel1))
	proofAck, proofHeight := suite.queryProof(chanKey)

	capabilityKey := sdk.NewKVStoreKey(TestPort2)

	err := suite.channelKeeper.ChanOpenConfirm(suite.ctx, TestPort2, TestChannel2, proofAck, uint64(proofHeight), capabilityKey)
	suite.Nil(err)

	channel, found := suite.channelKeeper.GetChannel(suite.ctx, TestPort2, TestChannel2)
	suite.True(found)
	suite.Equal(types.OPEN, channel.State)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
