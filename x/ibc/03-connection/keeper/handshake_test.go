package keeper_test

// import (
// 	"fmt"

// 	abci "github.com/tendermint/tendermint/abci/types"

// 	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
// 	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
// 	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
// 	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
// 	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
// )

// func (suite *KeeperTestSuite) TestConnOpenInit() {
// 	suite.createClient(testClientID1)
// 	counterparty := connection.NewCounterparty(testClientID1, testConnectionID1, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())

// 	success := func() error {
// 		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenInit(suite.ctx, testConnectionID1, testClientID1, counterparty)
// 		suite.NoError(err)

// 		conn, existed := suite.app.IBCKeeper.ConnectionKeeper.GetConnection(suite.ctx, testConnectionID1)
// 		suite.True(existed)

// 		expectConn := connection.ConnectionEnd{
// 			State:        exported.INIT,
// 			ClientID:     testClientID1,
// 			Counterparty: counterparty,
// 			Versions:     connection.GetCompatibleVersions(),
// 		}
// 		suite.EqualValues(expectConn, conn)

// 		return nil
// 	}

// 	connectionExists := func() error {
// 		return suite.app.IBCKeeper.ConnectionKeeper.ConnOpenInit(suite.ctx, testConnectionID1, testClientID1, counterparty)
// 	}

// 	var testCases = []testCase{
// 		{success, true, "success"},
// 		{connectionExists, false, "connection already exists"},
// 	}

// 	for _, tc := range testCases {
// 		if tc.expectPass {
// 			suite.Require().NoError(tc.fun(), tc.msg)
// 		} else {
// 			suite.Require().Error(tc.fun(), tc.msg)
// 		}
// 	}
// }

// func (suite *KeeperTestSuite) TestConnOpenTry() {
// 	suite.createClient(testClientID2)
// 	suite.createClient(testClientID1)
// 	suite.createConnection(testConnectionID2, testConnectionID1, testClientID2, testClientID1, exported.INIT)

// 	connectionKey := ibctypes.KeyConnection(testConnectionID2)

// 	proofInit, proofHeight := suite.queryProof(connectionKey)
// 	consensusKey := ibctypes.KeyConsensusState(testClientID2, uint64(proofHeight))
// 	proofConsensus, consensusHeight := suite.queryProof(consensusKey)

// 	invalidProof := func() error {
// 		counterparty := connection.NewCounterparty(testClientID2, testConnectionID2, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())
// 		return suite.app.IBCKeeper.ConnectionKeeper.ConnOpenTry(suite.ctx,
// 			testConnectionID1, counterparty, testClientID1,
// 			connection.GetCompatibleVersions(),
// 			proofInit, proofConsensus,
// 			uint64(proofHeight), uint64(consensusHeight))
// 	}

// 	success := func() error {
// 		suite.updateClient(testClientID1)

// 		counterparty := connection.NewCounterparty(testClientID2, testConnectionID2, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())
// 		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenTry(suite.ctx,
// 			testConnectionID1, counterparty, testClientID1,
// 			connection.GetCompatibleVersions(),
// 			proofInit, proofConsensus,
// 			uint64(proofHeight), uint64(consensusHeight))
// 		suite.NoError(err)

// 		//check connection state
// 		conn, existed := suite.app.IBCKeeper.ConnectionKeeper.GetConnection(suite.ctx, testConnectionID1)
// 		suite.True(existed)
// 		suite.Equal(exported.TRYOPEN.String(), conn.State.String(), "invalid connection state")
// 		return nil
// 	}

// 	connectionExists := func() error {
// 		suite.updateClient(testClientID1)
// 		counterparty := connection.NewCounterparty(testClientID2, testConnectionID2, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())
// 		return suite.app.IBCKeeper.ConnectionKeeper.ConnOpenTry(suite.ctx,
// 			testConnectionID1, counterparty, testClientID1,
// 			connection.GetCompatibleVersions(),
// 			proofInit, proofConsensus,
// 			uint64(proofHeight), uint64(consensusHeight))
// 	}

// 	var testCases = []testCase{
// 		{invalidProof, false, "invalid proof"},
// 		{connectionExists, false, "connection already exists"},
// 		{success, true, "success"},
// 	}

// 	for _, tc := range testCases {
// 		if tc.expectPass {
// 			suite.Require().NoError(tc.fun(), tc.msg)
// 		} else {
// 			suite.Require().Error(tc.fun(), tc.msg)
// 		}
// 	}

// }

// func (suite *KeeperTestSuite) TestConnOpenAck() {
// 	suite.createClient(testClientID2)
// 	suite.createClient(testClientID1)

// 	suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.TRYOPEN)
// 	connectionKey := ibctypes.KeyConnection(testConnectionID1)

// 	proofTry, proofHeight := suite.queryProof(connectionKey)
// 	consensusKey := ibctypes.KeyConsensusState(testClientID1, uint64(proofHeight))
// 	proofConsensus, consensusHeight := suite.queryProof(consensusKey)

// 	connectionNotFound := func() error {
// 		return suite.app.IBCKeeper.ConnectionKeeper.ConnOpenAck(suite.ctx, testConnectionID2, connection.GetCompatibleVersions()[0], proofTry, proofConsensus, uint64(proofHeight), uint64(consensusHeight))
// 	}

// 	invalidConnectionState := func() error {
// 		suite.createConnection(testConnectionID2, testConnectionID1, testClientID2, testClientID1, exported.UNINITIALIZED)
// 		return suite.app.IBCKeeper.ConnectionKeeper.ConnOpenAck(suite.ctx, testConnectionID2, connection.GetCompatibleVersions()[0], proofTry, proofConsensus, uint64(proofHeight), uint64(consensusHeight))
// 	}

// 	invalidVersion := func() error {
// 		suite.createConnection(testConnectionID2, testConnectionID1, testClientID2, testClientID1, exported.INIT)
// 		return suite.app.IBCKeeper.ConnectionKeeper.ConnOpenAck(suite.ctx, testConnectionID2, "1.0.1", proofTry, proofConsensus, uint64(proofHeight), uint64(consensusHeight))
// 	}

// 	invalidProof := func() error {
// 		suite.createConnection(testConnectionID2, testConnectionID1, testClientID2, testClientID1, exported.INIT)
// 		return suite.app.IBCKeeper.ConnectionKeeper.ConnOpenAck(suite.ctx, testConnectionID2, connection.GetCompatibleVersions()[0], proofTry, proofConsensus, uint64(proofHeight), uint64(consensusHeight))
// 	}

// 	success := func() error {
// 		suite.createConnection(testConnectionID2, testConnectionID1, testClientID2, testClientID1, exported.INIT)
// 		suite.updateClient(testClientID2)
// 		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenAck(suite.ctx, testConnectionID2, connection.GetCompatibleVersions()[0], proofTry, proofConsensus, uint64(proofHeight), uint64(consensusHeight))
// 		suite.NoError(err)

// 		//check connection state
// 		conn, existed := suite.app.IBCKeeper.ConnectionKeeper.GetConnection(suite.ctx, testConnectionID2)
// 		suite.True(existed)
// 		suite.Equal(exported.OPEN.String(), conn.State.String(), "invalid connection state")
// 		return nil
// 	}

// 	var testCases = []testCase{
// 		{connectionNotFound, false, "connection not exists"},
// 		{invalidConnectionState, false, "invalid connection state"},
// 		{invalidVersion, false, "invalid version"},
// 		{invalidProof, false, "invalid proof"},
// 		{success, true, ""},
// 	}

// 	for _, tc := range testCases {
// 		if tc.expectPass {
// 			suite.Require().NoError(tc.fun(), tc.msg)
// 		} else {
// 			suite.Require().Error(tc.fun(), tc.msg)
// 		}
// 	}
// }

// func (suite *KeeperTestSuite) TestConnOpenConfirm() {
// 	suite.createClient(testClientID2)
// 	suite.createClient(testClientID1)
// 	suite.createConnection(testConnectionID2, testConnectionID1, testClientID2, testClientID1, exported.OPEN)

// 	connKey := ibctypes.KeyConnection(testConnectionID2)
// 	proof, h := suite.queryProof(connKey)

// 	connectionNotFound := func() error {
// 		return suite.app.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(suite.ctx, testConnectionID1, proof, uint64(h))
// 	}

// 	invalidConnectionState := func() error {
// 		suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.INIT)
// 		return suite.app.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(suite.ctx, testConnectionID1, proof, uint64(h))
// 	}

// 	invalidProof := func() error {
// 		suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.TRYOPEN)
// 		return suite.app.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(suite.ctx, testConnectionID1, proof, uint64(h))
// 	}

// 	success := func() error {
// 		suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.TRYOPEN)
// 		suite.updateClient(testClientID1)
// 		proof, h = suite.queryProof(connKey)
// 		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(suite.ctx, testConnectionID1, proof, uint64(h))
// 		suite.NoError(err)

// 		conn, existed := suite.app.IBCKeeper.ConnectionKeeper.GetConnection(suite.ctx, testConnectionID1)
// 		suite.True(existed)
// 		suite.Equal(exported.OPEN.String(), conn.State.String(), "invalid connection state")
// 		return nil
// 	}

// 	var testCases = []testCase{
// 		{connectionNotFound, false, "connection not exists"},
// 		{invalidConnectionState, false, "invalid connection state"},
// 		{invalidProof, false, "invalid proof"},
// 		{success, true, ""},
// 	}

// 	for _, tc := range testCases {
// 		if tc.expectPass {
// 			suite.Require().NoError(tc.fun(), tc.msg)
// 		} else {
// 			suite.Require().Error(tc.fun(), tc.msg)
// 		}
// 	}
// }

// type testCase = struct {
// 	fun        func() error
// 	expectPass bool
// 	msg        string
// }
