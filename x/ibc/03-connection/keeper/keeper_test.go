package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

const (
	clientType = clientexported.Tendermint
	storeKey   = ibctypes.StoreKey
	chainID    = "test"

	testClientID1     = "testclientid1"
	testConnectionID1 = "connectionid1"

	testClientID2     = "testclientid2"
	testConnectionID2 = "connectionid2"

	testClientID3     = "testclientid3"
	testConnectionID3 = "connectionid3"
)

type KeeperTestSuite struct {
	suite.Suite

	cdc    *codec.Codec
	ctx    sdk.Context
	app    *simapp.SimApp
	valSet *tmtypes.ValidatorSet
}

func (suite *KeeperTestSuite) SetupTest() {
	isCheckTx := false
	app := simapp.Setup(isCheckTx)

	suite.cdc = app.Codec()
	suite.ctx = app.BaseApp.NewContext(isCheckTx, abci.Header{})
	suite.app = app

	privVal := tmtypes.NewMockPV()

	validator := tmtypes.NewValidator(privVal.GetPubKey(), 1)
	suite.valSet = tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestSetAndGetConnection() {
	_, existed := suite.app.IBCKeeper.ConnectionKeeper.GetConnection(suite.ctx, testConnectionID1)
	suite.False(existed)

	counterparty := types.NewCounterparty(testClientID1, testConnectionID1, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())
	expConn := types.ConnectionEnd{
		State:        types.INIT,
		ClientID:     testClientID1,
		Counterparty: counterparty,
		Versions:     types.GetCompatibleVersions(),
	}
	suite.app.IBCKeeper.ConnectionKeeper.SetConnection(suite.ctx, testConnectionID1, expConn)
	conn, existed := suite.app.IBCKeeper.ConnectionKeeper.GetConnection(suite.ctx, testConnectionID1)
	suite.True(existed)
	suite.EqualValues(expConn, conn)
}

func (suite *KeeperTestSuite) TestSetAndGetClientConnectionPaths() {
	_, existed := suite.app.IBCKeeper.ConnectionKeeper.GetClientConnectionPaths(suite.ctx, testClientID1)
	suite.False(existed)

	suite.app.IBCKeeper.ConnectionKeeper.SetClientConnectionPaths(suite.ctx, testClientID1, types.GetCompatibleVersions())
	paths, existed := suite.app.IBCKeeper.ConnectionKeeper.GetClientConnectionPaths(suite.ctx, testClientID1)
	suite.True(existed)
	suite.EqualValues(types.GetCompatibleVersions(), paths)
}

func (suite KeeperTestSuite) TestGetAllConnections() {
	// Connection (Counterparty): A(C) -> C(B) -> B(A)
	counterparty1 := types.NewCounterparty(testClientID1, testConnectionID1, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())
	counterparty2 := types.NewCounterparty(testClientID2, testConnectionID2, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())
	counterparty3 := types.NewCounterparty(testClientID3, testConnectionID3, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())

	conn1 := types.ConnectionEnd{
		State:        types.INIT,
		ClientID:     testClientID1,
		Counterparty: counterparty3,
		Versions:     types.GetCompatibleVersions(),
	}

	conn2 := types.ConnectionEnd{
		State:        types.INIT,
		ClientID:     testClientID2,
		Counterparty: counterparty1,
		Versions:     types.GetCompatibleVersions(),
	}

	conn3 := types.ConnectionEnd{
		State:        types.UNINITIALIZED,
		ClientID:     testClientID3,
		Counterparty: counterparty2,
		Versions:     types.GetCompatibleVersions(),
	}

	expConnections := []types.ConnectionEnd{conn1, conn2, conn3}

	suite.app.IBCKeeper.ConnectionKeeper.SetConnection(suite.ctx, testConnectionID1, expConnections[0])
	suite.app.IBCKeeper.ConnectionKeeper.SetConnection(suite.ctx, testConnectionID2, expConnections[1])
	suite.app.IBCKeeper.ConnectionKeeper.SetConnection(suite.ctx, testConnectionID3, expConnections[2])

	connections := suite.app.IBCKeeper.ConnectionKeeper.GetAllConnections(suite.ctx)
	suite.Require().Len(connections, len(expConnections))
	suite.Require().Equal(expConnections, connections)
}
