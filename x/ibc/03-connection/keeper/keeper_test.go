package keeper

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
)

func (suite *KeeperTestSuite) TestSetAndGetConnection() {
	gaia := suite.apps[ChainIdGaia1]

	conn, existed := gaia.connKeeper.GetConnection(gaia.ctx, ConnectionToGaia2)
	suite.False(existed)

	counterparty := types.NewCounterparty(ClientToGaia2, ConnectionToGaia1, gaia.connKeeper.GetCommitmentPrefix())
	expConn := types.ConnectionEnd{
		State:        types.INIT,
		ClientID:     ClientToGaia1,
		Counterparty: counterparty,
		Versions:     types.GetCompatibleVersions(),
	}
	gaia.connKeeper.SetConnection(gaia.ctx, ConnectionToGaia2, expConn)
	conn, existed = gaia.connKeeper.GetConnection(gaia.ctx, ConnectionToGaia2)
	suite.True(existed)
	suite.EqualValues(expConn, conn)
}

func (suite *KeeperTestSuite) TestSetAndGetClientConnectionPaths() {
	gaia := suite.apps[ChainIdGaia1]

	paths, existed := gaia.connKeeper.GetClientConnectionPaths(gaia.ctx, ClientToGaia2)
	suite.False(existed)

	gaia.connKeeper.SetClientConnectionPaths(gaia.ctx, ClientToGaia2, types.GetCompatibleVersions())
	paths, existed = gaia.connKeeper.GetClientConnectionPaths(gaia.ctx, ClientToGaia2)
	suite.True(existed)
	suite.EqualValues(types.GetCompatibleVersions(), paths)
}

func (suite *KeeperTestSuite) TestAddAndRemoveConnectionToClient() {
	gaia := suite.apps[ChainIdGaia1]

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
