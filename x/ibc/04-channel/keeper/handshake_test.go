package keeper_test

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypestm "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

func (suite *KeeperTestSuite) createClient() {
	suite.app.Commit()
	commitID := suite.app.LastCommitID()

	suite.app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: suite.app.LastBlockHeight() + 1}})
	suite.ctx = suite.app.BaseApp.NewContext(false, abci.Header{})

	consensusState := clienttypestm.ConsensusState{
		ChainID:          testChainID,
		Height:           uint64(commitID.Version),
		Root:             commitment.NewRoot(commitID.Hash),
		ValidatorSet:     suite.valSet,
		NextValidatorSet: suite.valSet,
	}

	_ = consensusState.GetCommitter()

	_, err := suite.app.IBCKeeper.ClientKeeper.CreateClient(suite.ctx, testClient, testClientType, consensusState)
	suite.NoError(err)
}

func (suite *KeeperTestSuite) updateClient() {
	// always commit and begin a new block on updateClient
	suite.app.Commit()
	commitID := suite.app.LastCommitID()

	suite.app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: suite.app.LastBlockHeight() + 1}})
	suite.ctx = suite.app.BaseApp.NewContext(false, abci.Header{})

	state := clienttypestm.ConsensusState{
		ChainID: testChainID,
		Height:  uint64(commitID.Version),
		Root:    commitment.NewRoot(commitID.Hash),
	}

	suite.app.IBCKeeper.ClientKeeper.SetConsensusState(suite.ctx, testClient, state)
	suite.app.IBCKeeper.ClientKeeper.SetVerifiedRoot(suite.ctx, testClient, state.GetHeight(), state.GetRoot())
}

func (suite *KeeperTestSuite) createConnection(state connection.State) {
	connection := connection.ConnectionEnd{
		State:    state,
		ClientID: testClient,
		Counterparty: connection.Counterparty{
			ClientID:     testClient,
			ConnectionID: testConnection,
			Prefix:       suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(),
		},
		Versions: connection.GetCompatibleVersions(),
	}

	suite.app.IBCKeeper.ConnectionKeeper.SetConnection(suite.ctx, testConnection, connection)
}

func (suite *KeeperTestSuite) createChannel(portID string, chanID string, connID string, counterpartyPort string, counterpartyChan string, state types.State) {
	channel := types.Channel{
		State:    state,
		Ordering: testChannelOrder,
		Counterparty: types.Counterparty{
			PortID:    counterpartyPort,
			ChannelID: counterpartyChan,
		},
		ConnectionHops: []string{connID},
		Version:        testChannelVersion,
	}

	suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, portID, chanID, channel)
}

func (suite *KeeperTestSuite) deleteChannel(portID string, chanID string) {
	store := suite.ctx.KVStore(suite.app.GetKey(ibctypes.StoreKey))
	store.Delete(types.KeyChannel(portID, chanID))
}

func (suite *KeeperTestSuite) bindPort(portID string) sdk.CapabilityKey {
	return suite.app.IBCKeeper.PortKeeper.BindPort(portID)
}

func (suite *KeeperTestSuite) queryProof(key []byte) (proof commitment.Proof, height int64) {
	res := suite.app.Query(abci.RequestQuery{
		Path:  fmt.Sprintf("store/%s/key", ibctypes.StoreKey),
		Data:  key,
		Prove: true,
	})

	height = res.Height
	proof = commitment.Proof{
		Proof: res.Proof,
	}

	return
}

func (suite *KeeperTestSuite) TestChanOpenInit() {
	counterparty := types.NewCounterparty(testPort2, testChannel2)

	suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, types.INIT)
	err := suite.app.IBCKeeper.ChannelKeeper.ChanOpenInit(suite.ctx, testChannelOrder, []string{testConnection}, testPort1, testChannel1, counterparty, testChannelVersion)
	suite.Error(err) // channel has already exist

	suite.deleteChannel(testPort1, testChannel1)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanOpenInit(suite.ctx, testChannelOrder, []string{testConnection}, testPort1, testChannel1, counterparty, testChannelVersion)
	suite.Error(err) // connection does not exist

	suite.createConnection(connection.UNINITIALIZED)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanOpenInit(suite.ctx, testChannelOrder, []string{testConnection}, testPort1, testChannel1, counterparty, testChannelVersion)
	suite.Error(err) // invalid connection state

	suite.createConnection(connection.OPEN)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanOpenInit(suite.ctx, testChannelOrder, []string{testConnection}, testPort1, testChannel1, counterparty, testChannelVersion)
	suite.NoError(err) // successfully executed

	channel, found := suite.app.IBCKeeper.ChannelKeeper.GetChannel(suite.ctx, testPort1, testChannel1)
	suite.True(found)
	suite.Equal(types.INIT.String(), channel.State.String(), "invalid channel state")
}

func (suite *KeeperTestSuite) TestChanOpenTry() {
	counterparty := types.NewCounterparty(testPort1, testChannel1)
	suite.bindPort(testPort2)
	channelKey := types.KeyChannel(testPort1, testChannel1)

	suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, types.INIT)
	suite.createChannel(testPort2, testChannel2, testConnection, testPort1, testChannel1, types.INIT)
	suite.updateClient()
	proofInit, proofHeight := suite.queryProof(channelKey)
	err := suite.app.IBCKeeper.ChannelKeeper.ChanOpenTry(suite.ctx, testChannelOrder, []string{testConnection}, testPort2, testChannel2, counterparty, testChannelVersion, testChannelVersion, proofInit, uint64(proofHeight))
	suite.Error(err) // channel has already exist

	suite.deleteChannel(testPort2, testChannel2)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanOpenTry(suite.ctx, testChannelOrder, []string{testConnection}, testPort1, testChannel2, counterparty, testChannelVersion, testChannelVersion, proofInit, uint64(proofHeight))
	suite.Error(err) // unauthenticated port

	err = suite.app.IBCKeeper.ChannelKeeper.ChanOpenTry(suite.ctx, testChannelOrder, []string{testConnection}, testPort2, testChannel2, counterparty, testChannelVersion, testChannelVersion, proofInit, uint64(proofHeight))
	suite.Error(err) // connection does not exist

	suite.createConnection(connection.UNINITIALIZED)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanOpenTry(suite.ctx, testChannelOrder, []string{testConnection}, testPort2, testChannel2, counterparty, testChannelVersion, testChannelVersion, proofInit, uint64(proofHeight))
	suite.Error(err) // invalid connection state

	suite.createConnection(connection.OPEN)
	suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, types.TRYOPEN)
	suite.updateClient()
	proofInit, proofHeight = suite.queryProof(channelKey)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanOpenTry(suite.ctx, testChannelOrder, []string{testConnection}, testPort2, testChannel2, counterparty, testChannelVersion, testChannelVersion, proofInit, uint64(proofHeight))
	suite.Error(err) // channel membership verification failed due to invalid counterparty

	suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, types.INIT)
	suite.updateClient()
	proofInit, proofHeight = suite.queryProof(channelKey)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanOpenTry(suite.ctx, testChannelOrder, []string{testConnection}, testPort2, testChannel2, counterparty, testChannelVersion, testChannelVersion, proofInit, uint64(proofHeight))
	suite.NoError(err) // successfully executed

	channel, found := suite.app.IBCKeeper.ChannelKeeper.GetChannel(suite.ctx, testPort2, testChannel2)
	suite.True(found)
	suite.Equal(types.TRYOPEN.String(), channel.State.String(), "invalid channel state")
}

func (suite *KeeperTestSuite) TestChanOpenAck() {
	suite.bindPort(testPort1)
	channelKey := types.KeyChannel(testPort2, testChannel2)

	suite.createChannel(testPort2, testChannel2, testConnection, testPort1, testChannel1, types.TRYOPEN)
	suite.updateClient()
	proofTry, proofHeight := suite.queryProof(channelKey)
	err := suite.app.IBCKeeper.ChannelKeeper.ChanOpenAck(suite.ctx, testPort1, testChannel1, testChannelVersion, proofTry, uint64(proofHeight))
	suite.Error(err) // channel does not exist

	suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, types.CLOSED)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanOpenAck(suite.ctx, testPort1, testChannel1, testChannelVersion, proofTry, uint64(proofHeight))
	suite.Error(err) // invalid channel state

	suite.createChannel(testPort2, testChannel1, testConnection, testPort1, testChannel2, types.INIT)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanOpenAck(suite.ctx, testPort2, testChannel1, testChannelVersion, proofTry, uint64(proofHeight))
	suite.Error(err) // unauthenticated port

	suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, types.INIT)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanOpenAck(suite.ctx, testPort1, testChannel1, testChannelVersion, proofTry, uint64(proofHeight))
	suite.Error(err) // connection does not exist

	suite.createConnection(connection.UNINITIALIZED)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanOpenAck(suite.ctx, testPort1, testChannel1, testChannelVersion, proofTry, uint64(proofHeight))
	suite.Error(err) // invalid connection state

	suite.createConnection(connection.OPEN)
	suite.createChannel(testPort2, testChannel2, testConnection, testPort1, testChannel1, types.OPEN)
	suite.updateClient()
	proofTry, proofHeight = suite.queryProof(channelKey)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanOpenAck(suite.ctx, testPort1, testChannel1, testChannelVersion, proofTry, uint64(proofHeight))
	suite.Error(err) // channel membership verification failed due to invalid counterparty

	suite.createChannel(testPort2, testChannel2, testConnection, testPort1, testChannel1, types.TRYOPEN)
	suite.updateClient()
	proofTry, proofHeight = suite.queryProof(channelKey)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanOpenAck(suite.ctx, testPort1, testChannel1, testChannelVersion, proofTry, uint64(proofHeight))
	suite.NoError(err) // successfully executed

	channel, found := suite.app.IBCKeeper.ChannelKeeper.GetChannel(suite.ctx, testPort1, testChannel1)
	suite.True(found)
	suite.Equal(types.OPEN.String(), channel.State.String(), "invalid channel state")
}

func (suite *KeeperTestSuite) TestChanOpenConfirm() {
	suite.bindPort(testPort2)
	channelKey := types.KeyChannel(testPort1, testChannel1)

	suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, types.OPEN)
	suite.updateClient()
	proofAck, proofHeight := suite.queryProof(channelKey)
	err := suite.app.IBCKeeper.ChannelKeeper.ChanOpenConfirm(suite.ctx, testPort2, testChannel2, proofAck, uint64(proofHeight))
	suite.Error(err) // channel does not exist

	suite.createChannel(testPort2, testChannel2, testConnection, testPort1, testChannel1, types.OPEN)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanOpenConfirm(suite.ctx, testPort2, testChannel2, proofAck, uint64(proofHeight))
	suite.Error(err) // invalid channel state

	suite.createChannel(testPort1, testChannel2, testConnection, testPort2, testChannel1, types.TRYOPEN)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanOpenConfirm(suite.ctx, testPort1, testChannel2, proofAck, uint64(proofHeight))
	suite.Error(err) // unauthenticated port

	suite.createChannel(testPort2, testChannel2, testConnection, testPort1, testChannel1, types.TRYOPEN)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanOpenConfirm(suite.ctx, testPort2, testChannel2, proofAck, uint64(proofHeight))
	suite.Error(err) // connection does not exist

	suite.createConnection(connection.UNINITIALIZED)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanOpenConfirm(suite.ctx, testPort2, testChannel2, proofAck, uint64(proofHeight))
	suite.Error(err) // invalid connection state

	suite.createConnection(connection.OPEN)
	suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, types.TRYOPEN)
	suite.updateClient()
	proofAck, proofHeight = suite.queryProof(channelKey)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanOpenConfirm(suite.ctx, testPort2, testChannel2, proofAck, uint64(proofHeight))
	suite.Error(err) // channel membership verification failed due to invalid counterparty

	suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, types.OPEN)
	suite.updateClient()
	proofAck, proofHeight = suite.queryProof(channelKey)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanOpenConfirm(suite.ctx, testPort2, testChannel2, proofAck, uint64(proofHeight))
	suite.NoError(err) // successfully executed

	channel, found := suite.app.IBCKeeper.ChannelKeeper.GetChannel(suite.ctx, testPort2, testChannel2)
	suite.True(found)
	suite.Equal(types.OPEN.String(), channel.State.String(), "invalid channel state")
}

func (suite *KeeperTestSuite) TestChanCloseInit() {
	suite.bindPort(testPort1)

	err := suite.app.IBCKeeper.ChannelKeeper.ChanCloseInit(suite.ctx, testPort2, testChannel1)
	suite.Error(err) // authenticated port

	err = suite.app.IBCKeeper.ChannelKeeper.ChanCloseInit(suite.ctx, testPort1, testChannel1)
	suite.Error(err) // channel does not exist

	suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, types.CLOSED)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanCloseInit(suite.ctx, testPort1, testChannel1)
	suite.Error(err) // channel is already closed

	suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, types.OPEN)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanCloseInit(suite.ctx, testPort1, testChannel1)
	suite.Error(err) // connection does not exist

	suite.createConnection(connection.TRYOPEN)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanCloseInit(suite.ctx, testPort1, testChannel1)
	suite.Error(err) // invalid connection state

	suite.createConnection(connection.OPEN)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanCloseInit(suite.ctx, testPort1, testChannel1)
	suite.NoError(err) // successfully executed

	channel, found := suite.app.IBCKeeper.ChannelKeeper.GetChannel(suite.ctx, testPort1, testChannel1)
	suite.True(found)
	suite.Equal(types.CLOSED.String(), channel.State.String(), "invalid channel state")
}

func (suite *KeeperTestSuite) TestChanCloseConfirm() {
	suite.bindPort(testPort2)
	channelKey := types.KeyChannel(testPort1, testChannel1)

	suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, types.CLOSED)
	suite.updateClient()
	proofInit, proofHeight := suite.queryProof(channelKey)
	err := suite.app.IBCKeeper.ChannelKeeper.ChanCloseConfirm(suite.ctx, testPort1, testChannel2, proofInit, uint64(proofHeight))
	suite.Error(err) // unauthenticated port

	err = suite.app.IBCKeeper.ChannelKeeper.ChanCloseConfirm(suite.ctx, testPort2, testChannel2, proofInit, uint64(proofHeight))
	suite.Error(err) // channel does not exist

	suite.createChannel(testPort2, testChannel2, testConnection, testPort1, testChannel1, types.CLOSED)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanCloseConfirm(suite.ctx, testPort2, testChannel2, proofInit, uint64(proofHeight))
	suite.Error(err) // channel is already closed

	suite.createChannel(testPort2, testChannel2, testConnection, testPort1, testChannel1, types.OPEN)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanCloseConfirm(suite.ctx, testPort2, testChannel2, proofInit, uint64(proofHeight))
	suite.Error(err) // connection does not exist

	suite.createConnection(connection.TRYOPEN)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanCloseConfirm(suite.ctx, testPort2, testChannel2, proofInit, uint64(proofHeight))
	suite.Error(err) // invalid connection state

	suite.createConnection(connection.OPEN)
	suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, types.OPEN)
	suite.updateClient()
	proofInit, proofHeight = suite.queryProof(channelKey)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanCloseConfirm(suite.ctx, testPort2, testChannel2, proofInit, uint64(proofHeight))
	suite.Error(err) // channel membership verification failed due to invalid counterparty

	suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, types.CLOSED)
	suite.updateClient()
	proofInit, proofHeight = suite.queryProof(channelKey)
	err = suite.app.IBCKeeper.ChannelKeeper.ChanCloseConfirm(suite.ctx, testPort2, testChannel2, proofInit, uint64(proofHeight))
	suite.NoError(err) // successfully executed

	channel, found := suite.app.IBCKeeper.ChannelKeeper.GetChannel(suite.ctx, testPort2, testChannel2)
	suite.True(found)
	suite.Equal(types.CLOSED.String(), channel.State.String(), "invalid channel state")
}
