package keeper_test

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

func (suite *KeeperTestSuite) TestConnOpenInit() {
	suite.createClient(testClientID1)
	counterparty := connection.NewCounterparty(testClientID1, testConnectionID1, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())

	success := func() error {
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenInit(suite.ctx, testConnectionID1, testClientID1, counterparty)

		conn, existed := suite.app.IBCKeeper.ConnectionKeeper.GetConnection(suite.ctx, testConnectionID1)
		suite.True(existed)

		expectConn := connection.ConnectionEnd{
			State:        connection.INIT,
			ClientID:     testClientID1,
			Counterparty: counterparty,
			Versions:     connection.GetCompatibleVersions(),
		}
		suite.EqualValues(expectConn, conn)

		return err
	}

	connectionExists := func() error {
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenInit(suite.ctx, testConnectionID1, testClientID1, counterparty)
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
	suite.createClient(testClientID2)
	suite.createClient(testClientID1)
	suite.createConnection(testConnectionID2, testConnectionID1, testClientID2, testClientID1, connection.INIT)

	connectionKey := connection.KeyConnection(testConnectionID2)
	consensusKey := client.KeyConsensusState(testClientID2)

	invalidProof := func() error {
		proofInit, proofHeight := suite.queryProof(connectionKey)
		proofConsensus, consensusHeight := suite.queryProof(consensusKey)

		counterparty := connection.NewCounterparty(testClientID2, testConnectionID2, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenTry(suite.ctx,
			testConnectionID1, counterparty, testClientID1,
			connection.GetCompatibleVersions(),
			proofInit, proofConsensus,
			uint64(proofHeight), uint64(consensusHeight))
		return err
	}

	success := func() error {
		suite.updateClient(testClientID1)
		proofInit, proofHeight := suite.queryProof(connectionKey)
		proofConsensus, consensusHeight := suite.queryProof(consensusKey)

		counterparty := connection.NewCounterparty(testClientID2, testConnectionID2, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenTry(suite.ctx,
			testConnectionID1, counterparty, testClientID1,
			connection.GetCompatibleVersions(),
			proofInit, proofConsensus,
			uint64(proofHeight), uint64(consensusHeight))
		suite.NoError(err)

		//check connection state
		conn, existed := suite.app.IBCKeeper.ConnectionKeeper.GetConnection(suite.ctx, testConnectionID1)
		suite.True(existed)
		suite.Equal(connection.TRYOPEN.String(), conn.State.String(), "invalid connection state")
		return err
	}

	connectionExists := func() error {
		suite.updateClient(testClientID1)
		proofInit, proofHeight := suite.queryProof(connectionKey)
		proofConsensus, consensusHeight := suite.queryProof(consensusKey)

		counterparty := connection.NewCounterparty(testClientID2, testConnectionID2, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenTry(suite.ctx,
			testConnectionID1, counterparty, testClientID1,
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
	suite.createClient(testClientID2)
	suite.createClient(testClientID1)

	suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connection.TRYOPEN)
	connectionKey := connection.KeyConnection(testConnectionID1)
	consensusKey := client.KeyConsensusState(testClientID1)

	connectionNotFound := func() error {
		//suite.updateClient(testClientID2)

		proofTry, proofHeight := suite.queryProof(connectionKey)
		proofConsensus, consensusHeight := suite.queryProof(consensusKey)
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenAck(suite.ctx, testConnectionID2, connection.GetCompatibleVersions()[0], proofTry, proofConsensus, uint64(proofHeight), uint64(consensusHeight))
		return err
	}

	invalidConnectionState := func() error {
		suite.createConnection(testConnectionID2, testConnectionID1, testClientID2, testClientID1, connection.UNINITIALIZED)
		//suite.updateClient(testClientID2)

		proofTry, proofHeight := suite.queryProof(connectionKey)
		proofConsensus, consensusHeight := suite.queryProof(consensusKey)
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenAck(suite.ctx, testConnectionID2, connection.GetCompatibleVersions()[0], proofTry, proofConsensus, uint64(proofHeight), uint64(consensusHeight))
		return err
	}

	invalidVersion := func() error {
		suite.createConnection(testConnectionID2, testConnectionID1, testClientID2, testClientID1, connection.INIT)
		//suite.updateClient(testClientID2)

		proofTry, proofHeight := suite.queryProof(connectionKey)
		proofConsensus, consensusHeight := suite.queryProof(consensusKey)
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenAck(suite.ctx, testConnectionID2, "1.0.1", proofTry, proofConsensus, uint64(proofHeight), uint64(consensusHeight))
		return err
	}

	invalidProof := func() error {
		suite.createConnection(testConnectionID2, testConnectionID1, testClientID2, testClientID1, connection.INIT)
		//suite.updateClient(testClientID2)

		proofTry, proofHeight := suite.queryProof(connectionKey)
		proofConsensus, consensusHeight := suite.queryProof(consensusKey)
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenAck(suite.ctx, testConnectionID2, connection.GetCompatibleVersions()[0], proofTry, proofConsensus, uint64(proofHeight), uint64(consensusHeight))
		return err
	}

	success := func() error {
		suite.createConnection(testConnectionID2, testConnectionID1, testClientID2, testClientID1, connection.INIT)
		suite.updateClient(testClientID2)

		proofTry, proofHeight := suite.queryProof(connectionKey)
		proofConsensus, consensusHeight := suite.queryProof(consensusKey)
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenAck(suite.ctx, testConnectionID2, connection.GetCompatibleVersions()[0], proofTry, proofConsensus, uint64(proofHeight), uint64(consensusHeight))

		//check connection state
		conn, existed := suite.app.IBCKeeper.ConnectionKeeper.GetConnection(suite.ctx, testConnectionID2)
		suite.True(existed)
		suite.Equal(connection.OPEN.String(), conn.State.String(), "invalid connection state")
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
	suite.createClient(testClientID2)
	suite.createClient(testClientID1)
	suite.createConnection(testConnectionID2, testConnectionID1, testClientID2, testClientID1, connection.OPEN)

	connKey := connection.KeyConnection(testConnectionID2)
	proof, h := suite.queryProof(connKey)

	connectionNotFound := func() error {
		//ErrConnectionNotFound
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(suite.ctx, testConnectionID1, proof, uint64(h))
		return err
	}

	invalidConnectionState := func() error {
		suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connection.INIT)
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(suite.ctx, testConnectionID1, proof, uint64(h))
		return err
	}

	invalidProof := func() error {
		//Error proof
		suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connection.TRYOPEN)
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(suite.ctx, testConnectionID1, proof, uint64(h))
		return err
	}

	success := func() error {
		//Success
		suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, connection.TRYOPEN)
		suite.updateClient(testClientID1)
		proof, h = suite.queryProof(connKey)
		err := suite.app.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(suite.ctx, testConnectionID1, proof, uint64(h))

		//check connection state
		conn, existed := suite.app.IBCKeeper.ConnectionKeeper.GetConnection(suite.ctx, testConnectionID1)
		suite.True(existed)
		suite.Equal(connection.OPEN.String(), conn.State.String(), "invalid connection state")
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

func (suite *KeeperTestSuite) queryProof(key []byte) (proof commitment.Proof, height int64) {
	res := suite.app.Query(abci.RequestQuery{
		Path:  fmt.Sprintf("store/%s/key", storeKey),
		Data:  key,
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
		ChainID:          chainID,
		Height:           uint64(commitID.Version),
		Root:             commitment.NewRoot(commitID.Hash),
		ValidatorSet:     suite.valSet,
		NextValidatorSet: suite.valSet,
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
		ChainID: chainID,
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
