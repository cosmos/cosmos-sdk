package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"
)

const (
	clientType = clientexported.Tendermint
	storeKey   = "ibc"
	ChainID    = "test"

	TestClientID1     = "testclientid1"
	TestConnectionID1 = "connectionid1"

	TestClientID2     = "testclientid2"
	TestConnectionID2 = "connectionid2"
)

type KeeperTestSuite struct {
	suite.Suite
	ctx              sdk.Context
	cms              sdk.CommitMultiStore
	cdc              *codec.Codec
	clientKeeper     client.Keeper
	connectionKeeper connection.Keeper
}

func (suite *KeeperTestSuite) SetupTest() {
	var codespaceType sdk.CodespaceType = storeKey
	storeKey := sdk.NewKVStoreKey(storeKey)

	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(storeKey, sdk.StoreTypeIAVL, db)
	if err := cms.LoadLatestVersion(); err != nil {
		panic(err)
	}

	cms.Commit()

	cdc := MakeCdc()
	clientKeeper := client.NewKeeper(cdc, storeKey, codespaceType)
	connectionKeeper := connection.NewKeeper(cdc, storeKey, codespaceType, clientKeeper)
	ctx := sdk.NewContext(cms, abci.Header{ChainID: ChainID, Height: 0}, false, log.NewNopLogger())
	suite.ctx = ctx
	suite.clientKeeper = clientKeeper
	suite.connectionKeeper = connectionKeeper
	suite.cms = cms
	suite.cdc = cdc
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestSetAndGetConnection() {
	_, existed := suite.connectionKeeper.GetConnection(suite.ctx, TestConnectionID1)
	suite.False(existed)

	counterparty := types.NewCounterparty(TestClientID1, TestConnectionID1, suite.connectionKeeper.GetCommitmentPrefix())
	expConn := types.ConnectionEnd{
		State:        types.INIT,
		ClientID:     TestClientID1,
		Counterparty: counterparty,
		Versions:     types.GetCompatibleVersions(),
	}
	suite.connectionKeeper.SetConnection(suite.ctx, TestConnectionID1, expConn)
	conn, existed := suite.connectionKeeper.GetConnection(suite.ctx, TestConnectionID1)
	suite.True(existed)
	suite.EqualValues(expConn, conn)
}

func (suite *KeeperTestSuite) TestSetAndGetClientConnectionPaths() {

	_, existed := suite.connectionKeeper.GetClientConnectionPaths(suite.ctx, TestClientID1)
	suite.False(existed)

	suite.connectionKeeper.SetClientConnectionPaths(suite.ctx, TestClientID1, types.GetCompatibleVersions())
	paths, existed := suite.connectionKeeper.GetClientConnectionPaths(suite.ctx, TestClientID1)
	suite.True(existed)
	suite.EqualValues(types.GetCompatibleVersions(), paths)
}

func MakeCdc() *codec.Codec {
	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	client.RegisterCodec(cdc)
	types.RegisterCodec(cdc)
	commitment.RegisterCodec(cdc)
	return cdc
}
