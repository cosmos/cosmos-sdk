package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (suite *KeeperTestSuite) getConsensusState() tendermint.ConsensusState {
	commitID := suite.cms.Commit()
	state := tendermint.ConsensusState{
		ChainID: ChainID,
		Height:  uint64(commitID.Version),
		Root:    commitment.NewRoot(commitID.Hash),
	}
	return state
}

func (suite *KeeperTestSuite) queryProof(key string) (proof commitment.Proof, height int64) {
	store := suite.cms.(*rootmulti.Store)
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

func (suite *KeeperTestSuite) createClient(clientID string) {
	commitID := suite.cms.LastCommitID()
	state := tendermint.ConsensusState{
		ChainID: ChainID,
		Height:  uint64(commitID.Version),
		Root:    commitment.NewRoot(commitID.Hash),
	}
	_, err := suite.clientKeeper.CreateClient(suite.ctx, clientID, clientType, state)
	if err != nil {
		panic(err)
	}
}

func (suite *KeeperTestSuite) updateClient(clientID string) {
	consensusState := suite.getConsensusState()
	//update client consensus state
	suite.clientKeeper.SetConsensusState(suite.ctx, clientID, consensusState)
	suite.clientKeeper.SetVerifiedRoot(suite.ctx, clientID, consensusState.GetHeight(), consensusState.GetRoot())
}

func (suite *KeeperTestSuite) createConnection(connID, counterpartyConnID string, clientID, counterpartyClientID string, state types.State) {
	counterparty := types.NewCounterparty(counterpartyClientID, counterpartyConnID, suite.connectionKeeper.GetCommitmentPrefix())
	connection := types.ConnectionEnd{
		State:        state,
		ClientID:     clientID,
		Counterparty: counterparty,
		Versions:     types.GetCompatibleVersions(),
	}
	suite.connectionKeeper.SetConnection(suite.ctx, connID, connection)
}

func (suite *KeeperTestSuite) TestConnOpenInit() {
	suite.createClient(TestClientID1)
	counterparty := types.NewCounterparty(TestClientID1, TestConnectionID1, suite.connectionKeeper.GetCommitmentPrefix())

	success := func() {
		err := suite.connectionKeeper.ConnOpenInit(suite.ctx, TestConnectionID1, TestClientID1, counterparty)
		suite.Nil(err)

		conn, existed := suite.connectionKeeper.GetConnection(suite.ctx, TestConnectionID1)
		suite.True(existed)

		expectConn := types.ConnectionEnd{
			State:        types.INIT,
			ClientID:     TestClientID1,
			Counterparty: counterparty,
			Versions:     types.GetCompatibleVersions(),
		}
		suite.EqualValues(expectConn, conn)
	}

	connectionExists := func() {
		err := suite.connectionKeeper.ConnOpenInit(suite.ctx, TestConnectionID1, TestClientID1, counterparty)
		suite.NotNil(err)
	}

	var testCases = []func(){
		success,
		connectionExists,
	}

	for _, fun := range testCases {
		fun()
	}
}

func (suite *KeeperTestSuite) TestConnOpenTry() {
	suite.createClient(TestClientID2)
	suite.createClient(TestClientID1)
	suite.createConnection(TestConnectionID2, TestConnectionID1, TestClientID2, TestClientID1, types.INIT)

	connectionKey := types.ConnectionPath(TestConnectionID2)
	consensusKey := string(client.KeyConsensusState(TestClientID2))

	invalidProof := func() {
		proofInit, proofHeight := suite.queryProof(connectionKey)
		proofConsensus, consensusHeight := suite.queryProof(consensusKey)

		counterparty := types.NewCounterparty(TestClientID2, TestConnectionID2, suite.connectionKeeper.GetCommitmentPrefix())
		err := suite.connectionKeeper.ConnOpenTry(suite.ctx,
			TestConnectionID1, counterparty, TestClientID1,
			types.GetCompatibleVersions(),
			proofInit, proofConsensus,
			uint64(proofHeight), uint64(consensusHeight))
		suite.NotNil(err)
	}

	success := func() {
		suite.updateClient(TestClientID1)
		proofInit, proofHeight := suite.queryProof(connectionKey)
		proofConsensus, consensusHeight := suite.queryProof(consensusKey)

		counterparty := types.NewCounterparty(TestClientID2, TestConnectionID2, suite.connectionKeeper.GetCommitmentPrefix())
		err := suite.connectionKeeper.ConnOpenTry(suite.ctx,
			TestConnectionID1, counterparty, TestClientID1,
			types.GetCompatibleVersions(),
			proofInit, proofConsensus,
			uint64(proofHeight), uint64(consensusHeight))
		suite.Nil(err)

		//check connection state
		conn, existed := suite.connectionKeeper.GetConnection(suite.ctx, TestConnectionID1)
		suite.True(existed)
		suite.Equal(types.TRYOPEN, conn.State)
	}

	connectionExists := func() {
		suite.updateClient(TestClientID1)
		proofInit, proofHeight := suite.queryProof(connectionKey)
		proofConsensus, consensusHeight := suite.queryProof(consensusKey)

		counterparty := types.NewCounterparty(TestClientID2, TestConnectionID2, suite.connectionKeeper.GetCommitmentPrefix())
		err := suite.connectionKeeper.ConnOpenTry(suite.ctx,
			TestConnectionID1, counterparty, TestClientID1,
			types.GetCompatibleVersions(),
			proofInit, proofConsensus,
			uint64(proofHeight), uint64(consensusHeight))
		suite.NotNil(err)
	}

	var testCases = []func(){
		invalidProof,
		success,
		connectionExists,
	}

	for _, fun := range testCases {
		fun()
	}

}

func (suite *KeeperTestSuite) TestConnOpenAck() {
	suite.createClient(TestClientID2)
	suite.createClient(TestClientID1)

	suite.createConnection(TestConnectionID1, TestConnectionID2, TestClientID1, TestClientID2, types.TRYOPEN)
	connectionKey := types.ConnectionPath(TestConnectionID1)
	consensusKey := string(client.KeyConsensusState(TestClientID1))

	connectionNotFound := func() {
		//suite.updateClient(TestClientID2)

		proofTry, proofHeight := suite.queryProof(connectionKey)
		proofConsensus, consensusHeight := suite.queryProof(consensusKey)
		err := suite.connectionKeeper.ConnOpenAck(suite.ctx, TestConnectionID2, types.GetCompatibleVersions()[0], proofTry, proofConsensus, uint64(proofHeight), uint64(consensusHeight))
		suite.NotNil(err)
	}

	invalidConnectionState := func() {
		suite.createConnection(TestConnectionID2, TestConnectionID1, TestClientID2, TestClientID1, types.NONE)
		//suite.updateClient(TestClientID2)

		proofTry, proofHeight := suite.queryProof(connectionKey)
		proofConsensus, consensusHeight := suite.queryProof(consensusKey)
		err := suite.connectionKeeper.ConnOpenAck(suite.ctx, TestConnectionID2, types.GetCompatibleVersions()[0], proofTry, proofConsensus, uint64(proofHeight), uint64(consensusHeight))
		suite.NotNil(err)
	}

	invalidVersion := func() {
		suite.createConnection(TestConnectionID2, TestConnectionID1, TestClientID2, TestClientID1, types.INIT)
		//suite.updateClient(TestClientID2)

		proofTry, proofHeight := suite.queryProof(connectionKey)
		proofConsensus, consensusHeight := suite.queryProof(consensusKey)
		err := suite.connectionKeeper.ConnOpenAck(suite.ctx, TestConnectionID2, "1.0.1", proofTry, proofConsensus, uint64(proofHeight), uint64(consensusHeight))
		suite.NotNil(err)
	}

	invalidProof := func() {
		suite.createConnection(TestConnectionID2, TestConnectionID1, TestClientID2, TestClientID1, types.INIT)
		//suite.updateClient(TestClientID2)

		proofTry, proofHeight := suite.queryProof(connectionKey)
		proofConsensus, consensusHeight := suite.queryProof(consensusKey)
		err := suite.connectionKeeper.ConnOpenAck(suite.ctx, TestConnectionID2, types.GetCompatibleVersions()[0], proofTry, proofConsensus, uint64(proofHeight), uint64(consensusHeight))
		suite.NotNil(err)
	}

	success := func() {
		suite.createConnection(TestConnectionID2, TestConnectionID1, TestClientID2, TestClientID1, types.INIT)
		suite.updateClient(TestClientID2)

		proofTry, proofHeight := suite.queryProof(connectionKey)
		proofConsensus, consensusHeight := suite.queryProof(consensusKey)
		err := suite.connectionKeeper.ConnOpenAck(suite.ctx, TestConnectionID2, types.GetCompatibleVersions()[0], proofTry, proofConsensus, uint64(proofHeight), uint64(consensusHeight))
		suite.Nil(err)

		//check connection state
		conn, existed := suite.connectionKeeper.GetConnection(suite.ctx, TestConnectionID2)
		suite.True(existed)
		suite.Equal(types.OPEN, conn.State)

	}

	var testCases = []func(){
		connectionNotFound,
		invalidConnectionState,
		invalidVersion,
		invalidProof,
		success,
	}

	for _, fun := range testCases {
		fun()
	}

}

func (suite *KeeperTestSuite) TestConnOpenConfirm() {
	suite.createClient(TestClientID2)
	suite.createClient(TestClientID1)
	suite.createConnection(TestConnectionID2, TestConnectionID1, TestClientID2, TestClientID1, types.OPEN)

	connKey := types.ConnectionPath(TestConnectionID2)
	proof, h := suite.queryProof(connKey)

	connectionNotFound := func() {
		//ErrConnectionNotFound
		err := suite.connectionKeeper.ConnOpenConfirm(suite.ctx, TestConnectionID1, proof, uint64(h))
		suite.NotNil(err)
	}

	invalidConnectionState := func() {
		suite.createConnection(TestConnectionID1, TestConnectionID2, TestClientID1, TestClientID2, types.INIT)
		err := suite.connectionKeeper.ConnOpenConfirm(suite.ctx, TestConnectionID1, proof, uint64(h))
		suite.NotNil(err)
	}

	invalidProof := func() {
		//Error proof
		suite.createConnection(TestConnectionID1, TestConnectionID2, TestClientID1, TestClientID2, types.TRYOPEN)
		err := suite.connectionKeeper.ConnOpenConfirm(suite.ctx, TestConnectionID1, proof, uint64(h))
		suite.NotNil(err)
	}

	success := func() {
		//Success
		suite.createConnection(TestConnectionID1, TestConnectionID2, TestClientID1, TestClientID2, types.TRYOPEN)
		suite.updateClient(TestClientID1)
		proof, h = suite.queryProof(connKey)
		err := suite.connectionKeeper.ConnOpenConfirm(suite.ctx, TestConnectionID1, proof, uint64(h))
		suite.Nil(err)
		//check connection state
		conn, existed := suite.connectionKeeper.GetConnection(suite.ctx, TestConnectionID1)
		suite.True(existed)
		suite.Equal(types.OPEN, conn.State)
	}

	var testCases = []func(){
		connectionNotFound,
		invalidConnectionState,
		invalidProof,
		success,
	}

	for _, fun := range testCases {
		fun()
	}
}
