package keeper

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
)

func (suite *KeeperTestSuite) TestSetAndGetConnection() {
	gaia := suite.apps[CosmosChainId]

	conn, existed := gaia.connKeeper.GetConnection(gaia.ctx, ConnectionToIris)
	suite.False(existed)

	counterparty := types.NewCounterparty(ClientToGaia, ConnectionToGaia, gaia.connKeeper.GetCommitmentPrefix())
	expConn := types.ConnectionEnd{
		State:        types.INIT,
		ClientID:     ClientToIris,
		Counterparty: counterparty,
		Versions:     types.GetCompatibleVersions(),
	}
	gaia.connKeeper.SetConnection(gaia.ctx, ConnectionToIris, expConn)
	conn, existed = gaia.connKeeper.GetConnection(gaia.ctx, ConnectionToIris)
	suite.True(existed)
	suite.EqualValues(expConn, conn)
}

func (suite *KeeperTestSuite) TestSetAndGetClientConnectionPaths() {
	gaia := suite.apps[CosmosChainId]

	paths, existed := gaia.connKeeper.GetClientConnectionPaths(gaia.ctx, ClientToGaia)
	suite.False(existed)

	gaia.connKeeper.SetClientConnectionPaths(gaia.ctx, ClientToGaia, types.GetCompatibleVersions())
	paths, existed = gaia.connKeeper.GetClientConnectionPaths(gaia.ctx, ClientToGaia)
	suite.True(existed)
	suite.EqualValues(types.GetCompatibleVersions(), paths)
}

func (suite *KeeperTestSuite) TestAddAndRemoveConnectionToClient() {
	gaia := suite.apps[CosmosChainId]

	//add connection to client
	err := gaia.connKeeper.addConnectionToClient(gaia.ctx, ClientToGaia, ConnectionToGaia)
	suite.Nil(err)
	expConns := []string{ConnectionToGaia}
	conns, existed := gaia.connKeeper.GetClientConnectionPaths(gaia.ctx, ClientToGaia)
	suite.True(existed)
	suite.EqualValues(expConns, conns)

	//add connection to client once again
	newConn := ConnectionToGaia + "1"
	err = gaia.connKeeper.addConnectionToClient(gaia.ctx, ClientToGaia, newConn)
	suite.Nil(err)
	expConns = append(expConns, newConn)
	conns, existed = gaia.connKeeper.GetClientConnectionPaths(gaia.ctx, ClientToGaia)
	suite.True(existed)
	suite.EqualValues(expConns, conns)

	//remove connection from client
	expConns = expConns[0:1]
	err = gaia.connKeeper.removeConnectionFromClient(gaia.ctx, ClientToGaia, newConn)
	suite.Nil(err)
	conns, existed = gaia.connKeeper.GetClientConnectionPaths(gaia.ctx, ClientToGaia)
	suite.True(existed)
	suite.EqualValues(expConns, conns)

	//remove connection from client again
	err = gaia.connKeeper.removeConnectionFromClient(gaia.ctx, ClientToGaia, newConn)
	suite.NotNil(err)

	err = gaia.connKeeper.removeConnectionFromClient(gaia.ctx, ClientToGaia, ConnectionToGaia)
	suite.Nil(err)
	conns, existed = gaia.connKeeper.GetClientConnectionPaths(gaia.ctx, ClientToGaia)
	suite.True(existed)
	suite.Nil(conns)
}
