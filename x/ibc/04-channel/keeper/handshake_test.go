package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypestm "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

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

func (suite *KeeperTestSuite) bindPort(portID string) sdk.CapabilityKey {
	return suite.portKeeper.BindPort(portID)
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

	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.INIT)
	err := suite.channelKeeper.ChanOpenInit(suite.ctx, TestChannelOrder, []string{TestConnection}, TestPort1, TestChannel1, counterparty, TestChannelVersion)
	suite.NotNil(err) // channel has already exist

	suite.deleteChannel(TestPort1, TestChannel1)
	err = suite.channelKeeper.ChanOpenInit(suite.ctx, TestChannelOrder, []string{TestConnection}, TestPort1, TestChannel1, counterparty, TestChannelVersion)
	suite.NotNil(err) // connection does not exist

	suite.createConnection(connection.NONE)
	err = suite.channelKeeper.ChanOpenInit(suite.ctx, TestChannelOrder, []string{TestConnection}, TestPort1, TestChannel1, counterparty, TestChannelVersion)
	suite.NotNil(err) // invalid connection state

	suite.createConnection(connection.OPEN)
	err = suite.channelKeeper.ChanOpenInit(suite.ctx, TestChannelOrder, []string{TestConnection}, TestPort1, TestChannel1, counterparty, TestChannelVersion)
	suite.Nil(err) // successfully executed

	channel, found := suite.channelKeeper.GetChannel(suite.ctx, TestPort1, TestChannel1)
	suite.True(found)
	suite.Equal(types.INIT, channel.State)
}

func (suite *KeeperTestSuite) TestChanOpenTry() {
	counterparty := types.NewCounterparty(TestPort1, TestChannel1)
	_ = suite.bindPort(TestPort2)
	chanKey := fmt.Sprintf("%s/%s", types.SubModuleName, types.ChannelPath(TestPort1, TestChannel1))

	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.INIT)
	suite.createChannel(TestPort2, TestChannel2, TestConnection, TestPort1, TestChannel1, types.INIT)
	suite.updateClient()
	proofInit, proofHeight := suite.queryProof(chanKey)
	err := suite.channelKeeper.ChanOpenTry(suite.ctx, TestChannelOrder, []string{TestConnection}, TestPort2, TestChannel2, counterparty, TestChannelVersion, TestChannelVersion, proofInit, uint64(proofHeight))
	suite.NotNil(err) // channel has already exist

	suite.deleteChannel(TestPort2, TestChannel2)
	err = suite.channelKeeper.ChanOpenTry(suite.ctx, TestChannelOrder, []string{TestConnection}, TestPort1, TestChannel2, counterparty, TestChannelVersion, TestChannelVersion, proofInit, uint64(proofHeight))
	suite.NotNil(err) // unauthenticated port

	err = suite.channelKeeper.ChanOpenTry(suite.ctx, TestChannelOrder, []string{TestConnection}, TestPort2, TestChannel2, counterparty, TestChannelVersion, TestChannelVersion, proofInit, uint64(proofHeight))
	suite.NotNil(err) // connection does not exist

	suite.createConnection(connection.NONE)
	err = suite.channelKeeper.ChanOpenTry(suite.ctx, TestChannelOrder, []string{TestConnection}, TestPort2, TestChannel2, counterparty, TestChannelVersion, TestChannelVersion, proofInit, uint64(proofHeight))
	suite.NotNil(err) // invalid connection state

	suite.createConnection(connection.OPEN)
	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.OPENTRY)
	suite.updateClient()
	proofInit, proofHeight = suite.queryProof(chanKey)
	err = suite.channelKeeper.ChanOpenTry(suite.ctx, TestChannelOrder, []string{TestConnection}, TestPort2, TestChannel2, counterparty, TestChannelVersion, TestChannelVersion, proofInit, uint64(proofHeight))
	suite.NotNil(err) // channel membership verification failed due to invalid counterparty

	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.INIT)
	suite.updateClient()
	proofInit, proofHeight = suite.queryProof(chanKey)
	err = suite.channelKeeper.ChanOpenTry(suite.ctx, TestChannelOrder, []string{TestConnection}, TestPort2, TestChannel2, counterparty, TestChannelVersion, TestChannelVersion, proofInit, uint64(proofHeight))
	suite.Nil(err) // successfully executed

	channel, found := suite.channelKeeper.GetChannel(suite.ctx, TestPort2, TestChannel2)
	suite.True(found)
	suite.Equal(types.OPENTRY, channel.State)
}

// func (suite *KeeperTestSuite) TestChanOpenAck() {
// 	suite.bindPort(TestPort1)
// 	chanKey := fmt.Sprintf("%s/%s", types.SubModuleName, types.ChannelPath(TestPort2, TestChannel2))

// 	suite.createChannel(TestPort2, TestChannel2, TestConnection, TestPort1, TestChannel1, types.OPENTRY)
// 	suite.updateClient()
// 	proofTry, proofHeight := suite.queryProof(chanKey)
// 	err := suite.channelKeeper.ChanOpenAck(suite.ctx, TestPort1, TestChannel1, TestChannelVersion, proofTry, uint64(proofHeight))
// 	suite.NotNil(err) // channel does not exist

// 	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.CLOSED)
// 	err = suite.channelKeeper.ChanOpenAck(suite.ctx, TestPort1, TestChannel1, TestChannelVersion, proofTry, uint64(proofHeight))
// 	suite.NotNil(err) // invalid channel state

// 	suite.createChannel(TestPort2, TestChannel1, TestConnection, TestPort1, TestChannel2, types.INIT)
// 	err = suite.channelKeeper.ChanOpenAck(suite.ctx, TestPort2, TestChannel1, TestChannelVersion, proofTry, uint64(proofHeight))
// 	suite.NotNil(err) // unauthenticated port

// 	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.INIT)
// 	err = suite.channelKeeper.ChanOpenAck(suite.ctx, TestPort1, TestChannel1, TestChannelVersion, proofTry, uint64(proofHeight))
// 	suite.NotNil(err) // connection does not exist

// 	suite.createConnection(connection.NONE)
// 	err = suite.channelKeeper.ChanOpenAck(suite.ctx, TestPort1, TestChannel1, TestChannelVersion, proofTry, uint64(proofHeight))
// 	suite.NotNil(err) // invalid connection state

// 	suite.createConnection(connection.OPEN)
// 	suite.createChannel(TestPort2, TestChannel2, TestConnection, TestPort1, TestChannel1, types.OPEN)
// 	suite.updateClient()
// 	proofTry, proofHeight = suite.queryProof(chanKey)
// 	err = suite.channelKeeper.ChanOpenAck(suite.ctx, TestPort1, TestChannel1, TestChannelVersion, proofTry, uint64(proofHeight))
// 	suite.NotNil(err) // channel membership verification failed due to invalid counterparty

// 	suite.createChannel(TestPort2, TestChannel2, TestConnection, TestPort1, TestChannel1, types.OPENTRY)
// 	suite.updateClient()
// 	proofTry, proofHeight = suite.queryProof(chanKey)
// 	err = suite.channelKeeper.ChanOpenAck(suite.ctx, TestPort1, TestChannel1, TestChannelVersion, proofTry, uint64(proofHeight))
// 	suite.Nil(err) // successfully executed

// 	channel, found := suite.channelKeeper.GetChannel(suite.ctx, TestPort1, TestChannel1)
// 	suite.True(found)
// 	suite.Equal(types.OPEN, channel.State)
// }

// func (suite *KeeperTestSuite) TestChanOpenConfirm() {
// 	suite.bindPort(TestPort2)
// 	chanKey := fmt.Sprintf("%s/%s", types.SubModuleName, types.ChannelPath(TestPort1, TestChannel1))

// 	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.OPEN)
// 	suite.updateClient()
// 	proofAck, proofHeight := suite.queryProof(chanKey)
// 	err := suite.channelKeeper.ChanOpenConfirm(suite.ctx, TestPort2, TestChannel2, proofAck, uint64(proofHeight))
// 	suite.NotNil(err) // channel does not exist

// 	suite.createChannel(TestPort2, TestChannel2, TestConnection, TestPort1, TestChannel1, types.OPEN)
// 	err = suite.channelKeeper.ChanOpenConfirm(suite.ctx, TestPort2, TestChannel2, proofAck, uint64(proofHeight))
// 	suite.NotNil(err) // invalid channel state

// 	suite.createChannel(TestPort1, TestChannel2, TestConnection, TestPort2, TestChannel1, types.OPENTRY)
// 	err = suite.channelKeeper.ChanOpenConfirm(suite.ctx, TestPort1, TestChannel2, proofAck, uint64(proofHeight))
// 	suite.NotNil(err) // unauthenticated port

// 	suite.createChannel(TestPort2, TestChannel2, TestConnection, TestPort1, TestChannel1, types.OPENTRY)
// 	err = suite.channelKeeper.ChanOpenConfirm(suite.ctx, TestPort2, TestChannel2, proofAck, uint64(proofHeight))
// 	suite.NotNil(err) // connection does not exist

// 	suite.createConnection(connection.NONE)
// 	err = suite.channelKeeper.ChanOpenConfirm(suite.ctx, TestPort2, TestChannel2, proofAck, uint64(proofHeight))
// 	suite.NotNil(err) // invalid connection state

// 	suite.createConnection(connection.OPEN)
// 	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.OPENTRY)
// 	suite.updateClient()
// 	proofAck, proofHeight = suite.queryProof(chanKey)
// 	err = suite.channelKeeper.ChanOpenConfirm(suite.ctx, TestPort2, TestChannel2, proofAck, uint64(proofHeight))
// 	suite.NotNil(err) // channel membership verification failed due to invalid counterparty

// 	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.OPEN)
// 	suite.updateClient()
// 	proofAck, proofHeight = suite.queryProof(chanKey)
// 	err = suite.channelKeeper.ChanOpenConfirm(suite.ctx, TestPort2, TestChannel2, proofAck, uint64(proofHeight))
// 	suite.Nil(err) // successfully executed

// 	channel, found := suite.channelKeeper.GetChannel(suite.ctx, TestPort2, TestChannel2)
// 	suite.True(found)
// 	suite.Equal(types.OPEN, channel.State)
// }

// func (suite *KeeperTestSuite) TestChanCloseInit() {
// 	suite.bindPort(TestPort1)

// 	err := suite.channelKeeper.ChanCloseInit(suite.ctx, TestPort2, TestChannel1)
// 	suite.NotNil(err) // authenticated port

// 	err = suite.channelKeeper.ChanCloseInit(suite.ctx, TestPort1, TestChannel1)
// 	suite.NotNil(err) // channel does not exist

// 	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.CLOSED)
// 	err = suite.channelKeeper.ChanCloseInit(suite.ctx, TestPort1, TestChannel1)
// 	suite.NotNil(err) // channel is already closed

// 	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.OPEN)
// 	err = suite.channelKeeper.ChanCloseInit(suite.ctx, TestPort1, TestChannel1)
// 	suite.NotNil(err) // connection does not exist

// 	suite.createConnection(connection.TRYOPEN)
// 	err = suite.channelKeeper.ChanCloseInit(suite.ctx, TestPort1, TestChannel1)
// 	suite.NotNil(err) // invalid connection state

// 	suite.createConnection(connection.OPEN)
// 	err = suite.channelKeeper.ChanCloseInit(suite.ctx, TestPort1, TestChannel1)
// 	suite.Nil(err) // successfully executed

// 	channel, found := suite.channelKeeper.GetChannel(suite.ctx, TestPort1, TestChannel1)
// 	suite.True(found)
// 	suite.Equal(types.CLOSED, channel.State)
// }

// func (suite *KeeperTestSuite) TestChanCloseConfirm() {
// 	suite.bindPort(TestPort2)
// 	chanKey := fmt.Sprintf("%s/%s", types.SubModuleName, types.ChannelPath(TestPort1, TestChannel1))

// 	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.CLOSED)
// 	suite.updateClient()
// 	proofInit, proofHeight := suite.queryProof(chanKey)
// 	err := suite.channelKeeper.ChanCloseConfirm(suite.ctx, TestPort1, TestChannel2, proofInit, uint64(proofHeight))
// 	suite.NotNil(err) // unauthenticated port

// 	err = suite.channelKeeper.ChanCloseConfirm(suite.ctx, TestPort2, TestChannel2, proofInit, uint64(proofHeight))
// 	suite.NotNil(err) // channel does not exist

// 	suite.createChannel(TestPort2, TestChannel2, TestConnection, TestPort1, TestChannel1, types.CLOSED)
// 	err = suite.channelKeeper.ChanCloseConfirm(suite.ctx, TestPort2, TestChannel2, proofInit, uint64(proofHeight))
// 	suite.NotNil(err) // channel is already closed

// 	suite.createChannel(TestPort2, TestChannel2, TestConnection, TestPort1, TestChannel1, types.OPEN)
// 	err = suite.channelKeeper.ChanCloseConfirm(suite.ctx, TestPort2, TestChannel2, proofInit, uint64(proofHeight))
// 	suite.NotNil(err) // connection does not exist

// 	suite.createConnection(connection.TRYOPEN)
// 	err = suite.channelKeeper.ChanCloseConfirm(suite.ctx, TestPort2, TestChannel2, proofInit, uint64(proofHeight))
// 	suite.NotNil(err) // invalid connection state

// 	suite.createConnection(connection.OPEN)
// 	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.OPEN)
// 	suite.updateClient()
// 	proofInit, proofHeight = suite.queryProof(chanKey)
// 	err = suite.channelKeeper.ChanCloseConfirm(suite.ctx, TestPort2, TestChannel2, proofInit, uint64(proofHeight))
// 	suite.NotNil(err) // channel membership verification failed due to invalid counterparty

// 	suite.createChannel(TestPort1, TestChannel1, TestConnection, TestPort2, TestChannel2, types.CLOSED)
// 	suite.updateClient()
// 	proofInit, proofHeight = suite.queryProof(chanKey)
// 	err = suite.channelKeeper.ChanCloseConfirm(suite.ctx, TestPort2, TestChannel2, proofInit, uint64(proofHeight))
// 	suite.Nil(err) // successfully executed

// 	channel, found := suite.channelKeeper.GetChannel(suite.ctx, TestPort2, TestChannel2)
// 	suite.True(found)
// 	suite.Equal(types.CLOSED, channel.State)
// }
