package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
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
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestSetAndGetConnection() {
	_, existed := suite.app.IBCKeeper.ConnectionKeeper.GetConnection(suite.ctx, TestConnectionID1)
	suite.False(existed)

	counterparty := types.NewCounterparty(TestClientID1, TestConnectionID1, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())
	expConn := types.ConnectionEnd{
		State:        types.INIT,
		ClientID:     TestClientID1,
		Counterparty: counterparty,
		Versions:     types.GetCompatibleVersions(),
	}
	suite.app.IBCKeeper.ConnectionKeeper.SetConnection(suite.ctx, TestConnectionID1, expConn)
	conn, existed := suite.app.IBCKeeper.ConnectionKeeper.GetConnection(suite.ctx, TestConnectionID1)
	suite.True(existed)
	suite.EqualValues(expConn, conn)
}

func (suite *KeeperTestSuite) TestSetAndGetClientConnectionPaths() {

	_, existed := suite.app.IBCKeeper.ConnectionKeeper.GetClientConnectionPaths(suite.ctx, TestClientID1)
	suite.False(existed)

	suite.app.IBCKeeper.ConnectionKeeper.SetClientConnectionPaths(suite.ctx, TestClientID1, types.GetCompatibleVersions())
	paths, existed := suite.app.IBCKeeper.ConnectionKeeper.GetClientConnectionPaths(suite.ctx, TestClientID1)
	suite.True(existed)
	suite.EqualValues(types.GetCompatibleVersions(), paths)
}
