package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

const (
	clientType = clientexported.Tendermint
	storeKey   = "ibc"

	ChainIDGaia1 = "gaia-1"
	ChainIDGaia2 = "gaia-2"

	ClientToGaia2 = "clienttogaia2"
	ClientToGaia1 = "clienttogaia1"

	ConnectionToGaia1 = "connectiontogaia1"
	ConnectionToGaia2 = "connectiontogaia2"
)

type KeeperTestSuite struct {
	suite.Suite
	apps map[string]App
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.apps = map[string]App{
		ChainIDGaia1: NewApp(ChainIDGaia1),
		ChainIDGaia2: NewApp(ChainIDGaia2),
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestSetAndGetConnection() {
	gaia := suite.apps[ChainIDGaia1]

	_, existed := gaia.connKeeper.GetConnection(gaia.ctx, ConnectionToGaia2)
	suite.False(existed)

	counterparty := types.NewCounterparty(ClientToGaia2, ConnectionToGaia1, gaia.connKeeper.GetCommitmentPrefix())
	expConn := types.ConnectionEnd{
		State:        types.INIT,
		ClientID:     ClientToGaia1,
		Counterparty: counterparty,
		Versions:     types.GetCompatibleVersions(),
	}
	gaia.connKeeper.SetConnection(gaia.ctx, ConnectionToGaia2, expConn)
	conn, existed := gaia.connKeeper.GetConnection(gaia.ctx, ConnectionToGaia2)
	suite.True(existed)
	suite.EqualValues(expConn, conn)
}

func (suite *KeeperTestSuite) TestSetAndGetClientConnectionPaths() {
	gaia := suite.apps[ChainIDGaia1]

	_, existed := gaia.connKeeper.GetClientConnectionPaths(gaia.ctx, ClientToGaia2)
	suite.False(existed)

	gaia.connKeeper.SetClientConnectionPaths(gaia.ctx, ClientToGaia2, types.GetCompatibleVersions())
	paths, existed := gaia.connKeeper.GetClientConnectionPaths(gaia.ctx, ClientToGaia2)
	suite.True(existed)
	suite.EqualValues(types.GetCompatibleVersions(), paths)
}

func (suite *KeeperTestSuite) TestAddAndRemoveConnectionToClient() {
	gaia := suite.apps[ChainIDGaia1]

	//add connection to client
	err := gaia.connKeeper.addConnectionToClient(gaia.ctx, ClientToGaia2, ConnectionToGaia1)
	suite.Nil(err)
	expConns := []string{ConnectionToGaia1}
	conns, existed := gaia.connKeeper.GetClientConnectionPaths(gaia.ctx, ClientToGaia2)
	suite.True(existed)
	suite.EqualValues(expConns, conns)

	//add connection to client once again
	newConn := ConnectionToGaia1 + "1"
	err = gaia.connKeeper.addConnectionToClient(gaia.ctx, ClientToGaia2, newConn)
	suite.Nil(err)
	expConns = append(expConns, newConn)
	conns, existed = gaia.connKeeper.GetClientConnectionPaths(gaia.ctx, ClientToGaia2)
	suite.True(existed)
	suite.EqualValues(expConns, conns)

	//remove connection from client
	expConns = expConns[0:1]
	err = gaia.connKeeper.removeConnectionFromClient(gaia.ctx, ClientToGaia2, newConn)
	suite.Nil(err)
	conns, existed = gaia.connKeeper.GetClientConnectionPaths(gaia.ctx, ClientToGaia2)
	suite.True(existed)
	suite.EqualValues(expConns, conns)

	//remove connection from client again
	err = gaia.connKeeper.removeConnectionFromClient(gaia.ctx, ClientToGaia2, newConn)
	suite.NotNil(err)

	err = gaia.connKeeper.removeConnectionFromClient(gaia.ctx, ClientToGaia2, ConnectionToGaia1)
	suite.Nil(err)
	conns, existed = gaia.connKeeper.GetClientConnectionPaths(gaia.ctx, ClientToGaia2)
	suite.True(existed)
	suite.Nil(conns)
}

type App struct {
	chainID string
	ctx     sdk.Context
	cdc     *codec.Codec
	store   sdk.CommitMultiStore
	IBCKeeper
}

type IBCKeeper struct {
	connKeeper   Keeper
	clientKeeper client.Keeper
}

func NewApp(chainID string) App {
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
	ctx := sdk.NewContext(ms, abci.Header{ChainID: chainID, Height: 0}, false, log.NewNopLogger())

	return App{
		chainID: chainID,
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
