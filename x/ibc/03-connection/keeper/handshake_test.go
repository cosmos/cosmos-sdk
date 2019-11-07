package keeper_test

import (
	"fmt"

	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (suite *KeeperTestSuite) TestConnOpenInit() {
	suite.createClient(TestClientID1)
	counterparty := connection.NewCounterparty(TestClientID1, TestConnectionID1, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())

	success := func() error {
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenInit(suite.ctx, TestConnectionID1, TestClientID1, counterparty)

		conn, existed := suite.app.IBCKeeper.ConnectionKeeper.GetConnection(suite.ctx, TestConnectionID1)
		suite.True(existed)

		expectConn := connection.ConnectionEnd{
			State:        connection.INIT,
			ClientID:     TestClientID1,
			Counterparty: counterparty,
			Versions:     connection.GetCompatibleVersions(),
		}
		suite.EqualValues(expectConn, conn)

		return err
	}

	connectionExists := func() error {
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenInit(suite.ctx, TestConnectionID1, TestClientID1, counterparty)
		return err
	}

	var testCases = []TestCase{
		{success, true, ""},
		{connectionExists, false, "connection already exists"},
	}

	for _, tc := range testCases {
		suite.Equal(tc.expected, tc.fun() == nil, "error: %s", tc.errMsg)
	}
}

func (suite *KeeperTestSuite) TestConnOpenTry() {
	suite.createClient(TestClientID2)
	suite.createClient(TestClientID1)
	suite.createConnection(TestConnectionID2, TestConnectionID1, TestClientID2, TestClientID1, connection.INIT)

	connectionKey := connection.ConnectionPath(TestConnectionID2)
	consensusKey := string(client.KeyConsensusState(TestClientID2))

	invalidProof := func() error {
		proofInit, proofHeight := suite.queryProof(connectionKey)
		proofConsensus, consensusHeight := suite.queryProof(consensusKey)

		counterparty := connection.NewCounterparty(TestClientID2, TestConnectionID2, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenTry(suite.ctx,
			TestConnectionID1, counterparty, TestClientID1,
			connection.GetCompatibleVersions(),
			proofInit, proofConsensus,
			uint64(proofHeight), uint64(consensusHeight))
		return err
	}

	success := func() error {
		suite.updateClient(TestClientID1)
		proofInit, proofHeight := suite.queryProof(connectionKey)
		proofConsensus, consensusHeight := suite.queryProof(consensusKey)

		counterparty := connection.NewCounterparty(TestClientID2, TestConnectionID2, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenTry(suite.ctx,
			TestConnectionID1, counterparty, TestClientID1,
			connection.GetCompatibleVersions(),
			proofInit, proofConsensus,
			uint64(proofHeight), uint64(consensusHeight))
		suite.Nil(err)

		//check connection state
		conn, existed := suite.app.IBCKeeper.ConnectionKeeper.GetConnection(suite.ctx, TestConnectionID1)
		suite.True(existed)
		suite.Equal(connection.TRYOPEN, conn.State)
		return err
	}

	connectionExists := func() error {
		suite.updateClient(TestClientID1)
		proofInit, proofHeight := suite.queryProof(connectionKey)
		proofConsensus, consensusHeight := suite.queryProof(consensusKey)

		counterparty := connection.NewCounterparty(TestClientID2, TestConnectionID2, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenTry(suite.ctx,
			TestConnectionID1, counterparty, TestClientID1,
			connection.GetCompatibleVersions(),
			proofInit, proofConsensus,
			uint64(proofHeight), uint64(consensusHeight))
		return err
	}

	var testCases = []TestCase{
		{invalidProof, false, "invalid proof"},
		{success, true, ""},
		{connectionExists, false, "connection already exists"},
	}

	for _, tc := range testCases {
		suite.Equal(tc.expected, tc.fun() == nil, "error: %s", tc.errMsg)
	}

}

func (suite *KeeperTestSuite) TestConnOpenAck() {
	suite.createClient(TestClientID2)
	suite.createClient(TestClientID1)

	suite.createConnection(TestConnectionID1, TestConnectionID2, TestClientID1, TestClientID2, connection.TRYOPEN)
	connectionKey := connection.ConnectionPath(TestConnectionID1)
	consensusKey := string(client.KeyConsensusState(TestClientID1))

	connectionNotFound := func() error {
		//suite.updateClient(TestClientID2)

		proofTry, proofHeight := suite.queryProof(connectionKey)
		proofConsensus, consensusHeight := suite.queryProof(consensusKey)
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenAck(suite.ctx, TestConnectionID2, connection.GetCompatibleVersions()[0], proofTry, proofConsensus, uint64(proofHeight), uint64(consensusHeight))
		return err
	}

	invalidConnectionState := func() error {
		suite.createConnection(TestConnectionID2, TestConnectionID1, TestClientID2, TestClientID1, connection.NONE)
		//suite.updateClient(TestClientID2)

		proofTry, proofHeight := suite.queryProof(connectionKey)
		proofConsensus, consensusHeight := suite.queryProof(consensusKey)
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenAck(suite.ctx, TestConnectionID2, connection.GetCompatibleVersions()[0], proofTry, proofConsensus, uint64(proofHeight), uint64(consensusHeight))
		return err
	}

	invalidVersion := func() error {
		suite.createConnection(TestConnectionID2, TestConnectionID1, TestClientID2, TestClientID1, connection.INIT)
		//suite.updateClient(TestClientID2)

		proofTry, proofHeight := suite.queryProof(connectionKey)
		proofConsensus, consensusHeight := suite.queryProof(consensusKey)
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenAck(suite.ctx, TestConnectionID2, "1.0.1", proofTry, proofConsensus, uint64(proofHeight), uint64(consensusHeight))
		return err
	}

	invalidProof := func() error {
		suite.createConnection(TestConnectionID2, TestConnectionID1, TestClientID2, TestClientID1, connection.INIT)
		//suite.updateClient(TestClientID2)

		proofTry, proofHeight := suite.queryProof(connectionKey)
		proofConsensus, consensusHeight := suite.queryProof(consensusKey)
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenAck(suite.ctx, TestConnectionID2, connection.GetCompatibleVersions()[0], proofTry, proofConsensus, uint64(proofHeight), uint64(consensusHeight))
		return err
	}

	success := func() error {
		suite.createConnection(TestConnectionID2, TestConnectionID1, TestClientID2, TestClientID1, connection.INIT)
		suite.updateClient(TestClientID2)

		proofTry, proofHeight := suite.queryProof(connectionKey)
		proofConsensus, consensusHeight := suite.queryProof(consensusKey)
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenAck(suite.ctx, TestConnectionID2, connection.GetCompatibleVersions()[0], proofTry, proofConsensus, uint64(proofHeight), uint64(consensusHeight))

		//check connection state
		conn, existed := suite.app.IBCKeeper.ConnectionKeeper.GetConnection(suite.ctx, TestConnectionID2)
		suite.True(existed)
		suite.Equal(connection.OPEN, conn.State)
		return err

	}

	var testCases = []TestCase{
		{connectionNotFound, false, "connection not exists"},
		{invalidConnectionState, false, "invalid connection state"},
		{invalidVersion, false, "invalid version"},
		{invalidProof, false, "invalid proof"},
		{success, true, ""},
	}

	for _, tc := range testCases {
		suite.Equal(tc.expected, tc.fun() == nil, "error: %s", tc.errMsg)
	}

}

func (suite *KeeperTestSuite) TestConnOpenConfirm() {
	suite.createClient(TestClientID2)
	suite.createClient(TestClientID1)
	suite.createConnection(TestConnectionID2, TestConnectionID1, TestClientID2, TestClientID1, connection.OPEN)

	connKey := connection.ConnectionPath(TestConnectionID2)
	proof, h := suite.queryProof(connKey)

	connectionNotFound := func() error {
		//ErrConnectionNotFound
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(suite.ctx, TestConnectionID1, proof, uint64(h))
		return err
	}

	invalidConnectionState := func() error {
		suite.createConnection(TestConnectionID1, TestConnectionID2, TestClientID1, TestClientID2, connection.INIT)
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(suite.ctx, TestConnectionID1, proof, uint64(h))
		return err
	}

	invalidProof := func() error {
		//Error proof
		suite.createConnection(TestConnectionID1, TestConnectionID2, TestClientID1, TestClientID2, connection.TRYOPEN)
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(suite.ctx, TestConnectionID1, proof, uint64(h))
		return err
	}

	success := func() error {
		//Success
		suite.createConnection(TestConnectionID1, TestConnectionID2, TestClientID1, TestClientID2, connection.TRYOPEN)
		suite.updateClient(TestClientID1)
		proof, h = suite.queryProof(connKey)
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(suite.ctx, TestConnectionID1, proof, uint64(h))

		//check connection state
		conn, existed := suite.app.IBCKeeper.ConnectionKeeper.GetConnection(suite.ctx, TestConnectionID1)
		suite.True(existed)
		suite.Equal(connection.OPEN, conn.State)
		return err
	}

	var testCases = []TestCase{
		{connectionNotFound, false, "connection not exists"},
		{invalidConnectionState, false, "invalid connection state"},
		{invalidProof, false, "invalid proof"},
		{success, true, ""},
	}

	for _, tc := range testCases {
		suite.Equal(tc.expected, tc.fun() == nil, "error: %s", tc.errMsg)
	}
}

func (suite *KeeperTestSuite) queryProof(key string) (proof commitment.Proof, height int64) {
	res := suite.app.Query(abci.RequestQuery{
		Path:  fmt.Sprintf("store/%s/key", storeKey),
		Data:  []byte(key),
		Prove: true,
	})

	height = res.Height
	proof = commitment.Proof{
		Proof: res.Proof,
	}

	return
}

func (suite *KeeperTestSuite) createClient(clientID string) {
	suite.app.Commit()
	commitID := suite.app.LastCommitID()

	suite.app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: suite.app.LastBlockHeight() + 1}})
	suite.ctx = suite.app.BaseApp.NewContext(false, abci.Header{})

	consensusState := tendermint.ConsensusState{
		ChainID: ChainID,
		Height:  uint64(commitID.Version),
		Root:    commitment.NewRoot(commitID.Hash),
	}

	_, err := suite.app.IBCKeeper.ClientKeeper.CreateClient(suite.ctx, clientID, clientType, consensusState)
	suite.NoError(err)
}

func (suite *KeeperTestSuite) updateClient(clientID string) {
	// always commit when updateClient and begin a new block
	suite.app.Commit()
	commitID := suite.app.LastCommitID()

	suite.app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: suite.app.LastBlockHeight() + 1}})
	suite.ctx = suite.app.BaseApp.NewContext(false, abci.Header{})

	state := tendermint.ConsensusState{
		ChainID: ChainID,
		Height:  uint64(commitID.Version),
		Root:    commitment.NewRoot(commitID.Hash),
	}

	suite.app.IBCKeeper.ClientKeeper.SetConsensusState(suite.ctx, clientID, state)
	suite.app.IBCKeeper.ClientKeeper.SetVerifiedRoot(suite.ctx, clientID, state.GetHeight(), state.GetRoot())
}

func (suite *KeeperTestSuite) createConnection(connID, counterpartyConnID string, clientID, counterpartyClientID string, state connection.State) {
	counterparty := connection.NewCounterparty(counterpartyClientID, counterpartyConnID, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())
	connection := connection.ConnectionEnd{
		State:        state,
		ClientID:     clientID,
		Counterparty: counterparty,
		Versions:     connection.GetCompatibleVersions(),
	}
	suite.app.IBCKeeper.ConnectionKeeper.SetConnection(suite.ctx, connID, connection)
}

type TestCase = struct {
	fun      func() error
	expected bool
	errMsg   string
}
