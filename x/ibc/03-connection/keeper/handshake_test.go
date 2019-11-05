package keeper

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

const (
	clientType = exported.Tendermint
	storeKey   = "ibc"

	ChainIdGaia1 = "gaia-1"
	ChainIdGaia2 = "gaia-2"

	ClientToGaia2 = "clienttogaia2"
	ClientToGaia1 = "clienttogaia1"

	ConnectionToGaia1 = "connectiontogaia1"
	ConnectionToGaia2 = "connectiontogaia2"
)

type KeeperTestSuite struct {
	suite.Suite
	apps map[string]App
}

type App struct {
	chainId string
	ctx     sdk.Context
	cdc     *codec.Codec
	store   sdk.CommitMultiStore
	IBCKeeper
}

type IBCKeeper struct {
	connKeeper   Keeper
	clientKeeper client.Keeper
}

func NewApp(chainId string) App {
	var codespaceType sdk.CodespaceType = storeKey
	storeKey := sdk.NewKVStoreKey(storeKey)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(storeKey, sdk.StoreTypeIAVL, db)
	if err := ms.LoadLatestVersion(); err != nil {
		panic(err)
	}

	cdc := MakeCdc()

	clientKeeper := client.NewKeeper(cdc, storeKey, codespaceType)
	connKeeper := NewKeeper(cdc, storeKey, codespaceType, clientKeeper)
	ctx := sdk.NewContext(ms, abci.Header{ChainID: chainId, Height: 0}, false, log.NewNopLogger())

	return App{
		chainId: chainId,
		ctx:     ctx,
		cdc:     cdc,
		store:   ms,
		IBCKeeper: IBCKeeper{
			connKeeper:   connKeeper,
			clientKeeper: clientKeeper,
		},
	}

}

func MakeCdc() *codec.Codec {
	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	client.RegisterCodec(cdc)
	types.RegisterCodec(cdc)
	commitment.RegisterCodec(cdc)
	return cdc
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.apps = map[string]App{
		ChainIdGaia1: NewApp(ChainIdGaia1),
		ChainIdGaia2: NewApp(ChainIdGaia2),
	}
}

func (suite *KeeperTestSuite) getConsensusState(chainId string) tendermint.ConsensusState {
	app := suite.apps[chainId]
	commitId := app.store.Commit()
	state := tendermint.ConsensusState{
		ChainID: app.chainId,
		Height:  uint64(commitId.Version),
		Root:    commitment.NewRoot(commitId.Hash),
	}
	app.ctx.WithBlockHeight(commitId.Version)
	return state
}

func (suite *KeeperTestSuite) queryProof(chainId string, key string) (proof commitment.Proof, height int64) {
	app := suite.apps[chainId]
	store := app.store.(*rootmulti.Store)
	res := store.Query(abci.RequestQuery{
		Path:  fmt.Sprintf("/%s/key", storeKey),
		Data:  []byte(key),
		Prove: true,
	})

	height = res.Height
	proof = commitment.Proof{
		Proof: res.Proof,
	}
	return
}

func (suite *KeeperTestSuite) createClient(chainId string, clientID string,
	clientType exported.ClientType, state tendermint.ConsensusState) {
	app := suite.apps[chainId]
	_, err := app.clientKeeper.CreateClient(app.ctx, clientID, clientType, state)
	if err != nil {
		panic(err)
	}
	commitId := app.store.Commit()
	app.ctx.WithBlockHeight(commitId.Version)
}

func (suite *KeeperTestSuite) updateClient(chainId string, clientID string) {
	otherChainId := ChainIdGaia1
	if chainId == ChainIdGaia1 {
		otherChainId = ChainIdGaia2
	}
	consensusState := suite.getConsensusState(otherChainId)

	//update client consensus state
	app := suite.apps[chainId]
	app.clientKeeper.SetConsensusState(app.ctx, clientID, consensusState)
	app.clientKeeper.SetVerifiedRoot(app.ctx, clientID, consensusState.GetHeight(), consensusState.GetRoot())
	commitId := app.store.Commit()
	app.ctx.WithBlockHeight(commitId.Version)
}

func (suite *KeeperTestSuite) connOpenInit(chainId string, connectionID, clientID, counterpartyClientID, counterpartyConnID string) {
	app := suite.apps[chainId]
	counterparty := types.NewCounterparty(counterpartyClientID, counterpartyConnID, app.connKeeper.GetCommitmentPrefix())

	err := app.connKeeper.ConnOpenInit(app.ctx, connectionID, clientID, counterparty)
	suite.Nil(err)

	conn, existed := app.connKeeper.GetConnection(app.ctx, connectionID)
	suite.True(existed)

	expectConn := types.ConnectionEnd{
		State:        types.INIT,
		ClientID:     clientID,
		Counterparty: counterparty,
		Versions:     types.GetCompatibleVersions(),
	}
	suite.EqualValues(expectConn, conn)
	commitId := app.store.Commit()
	app.ctx.WithBlockHeight(commitId.Version)
}

func (suite *KeeperTestSuite) connOpenTry(chainId string, connectionID, clientID, counterpartyClientID, counterpartyConnID string) {
	app := suite.apps[chainId]
	counterparty := types.NewCounterparty(counterpartyClientID, counterpartyConnID, app.connKeeper.GetCommitmentPrefix())

	connKey := fmt.Sprintf("%s/%s", types.SubModuleName, types.ConnectionPath(counterpartyConnID))
	otherChainId := ChainIdGaia1
	if chainId == ChainIdGaia1 {
		otherChainId = ChainIdGaia2
	}
	proof, h := suite.queryProof(otherChainId, connKey)

	err := app.connKeeper.ConnOpenTry(app.ctx, connectionID, counterparty, clientID, types.GetCompatibleVersions(), proof, uint64(h), 0)
	suite.Nil(err)

	commitId := app.store.Commit()
	app.ctx.WithBlockHeight(commitId.Version)

	//check connection state
	conn, existed := app.connKeeper.GetConnection(app.ctx, connectionID)
	suite.True(existed)
	suite.Equal(types.TRYOPEN, conn.State)
}

func (suite *KeeperTestSuite) connOpenAck(chainId string, connectionID, counterpartyConnID string) {
	app := suite.apps[chainId]
	connKey := fmt.Sprintf("%s/%s", types.SubModuleName, types.ConnectionPath(counterpartyConnID))
	otherChainId := ChainIdGaia1
	if chainId == ChainIdGaia1 {
		otherChainId = ChainIdGaia2
	}
	proof, h := suite.queryProof(otherChainId, connKey)

	err := app.connKeeper.ConnOpenAck(app.ctx, connectionID, types.GetCompatibleVersions()[0], proof, uint64(h), 0)
	suite.Nil(err)

	commitId := app.store.Commit()
	app.ctx.WithBlockHeight(commitId.Version)

	//check connection state
	conn, existed := app.connKeeper.GetConnection(app.ctx, connectionID)
	suite.True(existed)
	suite.Equal(types.OPEN, conn.State)
}

func (suite *KeeperTestSuite) connOpenConfirm(chainId string, connectionID, counterpartyConnID string) {
	app := suite.apps[chainId]
	connKey := fmt.Sprintf("%s/%s", types.SubModuleName, types.ConnectionPath(counterpartyConnID))
	otherChainId := ChainIdGaia1
	if chainId == ChainIdGaia1 {
		otherChainId = ChainIdGaia2
	}
	proof, h := suite.queryProof(otherChainId, connKey)

	err := app.connKeeper.ConnOpenConfirm(app.ctx, connectionID, proof, uint64(h))
	suite.Nil(err)

	commitId := app.store.Commit()
	app.ctx.WithBlockHeight(commitId.Version)

	//check connection state
	conn, existed := app.connKeeper.GetConnection(app.ctx, connectionID)
	suite.True(existed)
	suite.Equal(types.OPEN, conn.State)
}

func (suite *KeeperTestSuite) TestHandshake() {
	//get gaia consensusState
	state := suite.getConsensusState(ChainIdGaia1)
	//create client on iris
	suite.createClient(ChainIdGaia2, ClientToGaia2, clientType, state)

	//get iris consensusState
	state1 := suite.getConsensusState(ChainIdGaia2)
	// create client on gaia
	suite.createClient(ChainIdGaia1, ClientToGaia1, clientType, state1)

	//===========OpenInit on iris===========
	suite.connOpenInit(ChainIdGaia2, ConnectionToGaia1, ClientToGaia2, ClientToGaia1, ConnectionToGaia2)

	//===========OpenTry on gaia===========
	// update gaia consensusState(should be UpdateClient)
	suite.updateClient(ChainIdGaia1, ClientToGaia1)
	// open-try on gaia
	suite.connOpenTry(ChainIdGaia1, ConnectionToGaia2, ClientToGaia1, ClientToGaia2, ConnectionToGaia1)

	//===========ConnOpenAck on iris===========
	// update iris consensusState(should be UpdateClient)
	suite.updateClient(ChainIdGaia2, ClientToGaia2)
	suite.connOpenAck(ChainIdGaia2, ConnectionToGaia1, ConnectionToGaia2)

	//===========ConnOpenConfirm on gaia===========
	// update gaia consensusState(should be UpdateClient)
	suite.updateClient(ChainIdGaia1, ClientToGaia1)
	suite.connOpenConfirm(ChainIdGaia1, ConnectionToGaia2, ConnectionToGaia1)

}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
