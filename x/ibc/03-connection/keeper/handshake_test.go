package keeper

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (suite *KeeperTestSuite) getConsensusState(chainId string) tendermint.ConsensusState {
	app := suite.apps[chainId]
	commitId := app.store.Commit()
	state := tendermint.ConsensusState{
		ChainID: app.chainId,
		Height:  uint64(commitId.Version),
		Root:    commitment.NewRoot(commitId.Hash),
	}
	app.ctx.WithBlockHeight(commitId.Version)
	return state
}

func (suite *KeeperTestSuite) queryProof(chainId string, key string) (proof commitment.Proof, height int64) {
	app := suite.apps[chainId]
	store := app.store.(*rootmulti.Store)
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

func (suite *KeeperTestSuite) createClient(chainId string, clientID string,
	clientType exported.ClientType, state tendermint.ConsensusState) {
	app := suite.apps[chainId]
	_, err := app.clientKeeper.CreateClient(app.ctx, clientID, clientType, state)
	if err != nil {
		panic(err)
	}
	commitId := app.store.Commit()
	app.ctx.WithBlockHeight(commitId.Version)
}

func (suite *KeeperTestSuite) updateClient(chainId string, clientID string) {
	otherChainId := ChainIdGaia1
	if chainId == ChainIdGaia1 {
		otherChainId = ChainIdGaia2
	}
	consensusState := suite.getConsensusState(otherChainId)

	//update client consensus state
	app := suite.apps[chainId]
	app.clientKeeper.SetConsensusState(app.ctx, clientID, consensusState)
	app.clientKeeper.SetVerifiedRoot(app.ctx, clientID, consensusState.GetHeight(), consensusState.GetRoot())
	commitId := app.store.Commit()
	app.ctx.WithBlockHeight(commitId.Version)
}

func (suite *KeeperTestSuite) connOpenInit(chainId string, connectionID, clientID, counterpartyClientID, counterpartyConnID string) {
	app := suite.apps[chainId]
	counterparty := types.NewCounterparty(counterpartyClientID, counterpartyConnID, app.connKeeper.GetCommitmentPrefix())

	err := app.connKeeper.ConnOpenInit(app.ctx, connectionID, clientID, counterparty)
	suite.Nil(err)

	conn, existed := app.connKeeper.GetConnection(app.ctx, connectionID)
	suite.True(existed)

	expectConn := types.ConnectionEnd{
		State:        types.INIT,
		ClientID:     clientID,
		Counterparty: counterparty,
		Versions:     types.GetCompatibleVersions(),
	}
	suite.EqualValues(expectConn, conn)
	commitId := app.store.Commit()
	app.ctx.WithBlockHeight(commitId.Version)
}

func (suite *KeeperTestSuite) connOpenTry(chainId string, connectionID, clientID, counterpartyClientID, counterpartyConnID string) {
	app := suite.apps[chainId]
	counterparty := types.NewCounterparty(counterpartyClientID, counterpartyConnID, app.connKeeper.GetCommitmentPrefix())

	connKey := fmt.Sprintf("%s/%s", types.SubModuleName, types.ConnectionPath(counterpartyConnID))
	otherChainId := ChainIdGaia1
	if chainId == ChainIdGaia1 {
		otherChainId = ChainIdGaia2
	}
	proof, h := suite.queryProof(otherChainId, connKey)

	err := app.connKeeper.ConnOpenTry(app.ctx, connectionID, counterparty, clientID, types.GetCompatibleVersions(), proof, uint64(h), 0)
	suite.Nil(err)

	commitId := app.store.Commit()
	app.ctx.WithBlockHeight(commitId.Version)

	//check connection state
	conn, existed := app.connKeeper.GetConnection(app.ctx, connectionID)
	suite.True(existed)
	suite.Equal(types.TRYOPEN, conn.State)
}

func (suite *KeeperTestSuite) connOpenAck(chainId string, connectionID, counterpartyConnID string) {
	app := suite.apps[chainId]
	connKey := fmt.Sprintf("%s/%s", types.SubModuleName, types.ConnectionPath(counterpartyConnID))
	otherChainId := ChainIdGaia1
	if chainId == ChainIdGaia1 {
		otherChainId = ChainIdGaia2
	}
	proof, h := suite.queryProof(otherChainId, connKey)

	err := app.connKeeper.ConnOpenAck(app.ctx, connectionID, types.GetCompatibleVersions()[0], proof, uint64(h), 0)
	suite.Nil(err)

	commitId := app.store.Commit()
	app.ctx.WithBlockHeight(commitId.Version)

	//check connection state
	conn, existed := app.connKeeper.GetConnection(app.ctx, connectionID)
	suite.True(existed)
	suite.Equal(types.OPEN, conn.State)
}

func (suite *KeeperTestSuite) connOpenConfirm(chainId string, connectionID, counterpartyConnID string) {
	app := suite.apps[chainId]
	connKey := fmt.Sprintf("%s/%s", types.SubModuleName, types.ConnectionPath(counterpartyConnID))
	otherChainId := ChainIdGaia1
	if chainId == ChainIdGaia1 {
		otherChainId = ChainIdGaia2
	}
	proof, h := suite.queryProof(otherChainId, connKey)

	err := app.connKeeper.ConnOpenConfirm(app.ctx, connectionID, proof, uint64(h))
	suite.Nil(err)

	commitId := app.store.Commit()
	app.ctx.WithBlockHeight(commitId.Version)

	//check connection state
	conn, existed := app.connKeeper.GetConnection(app.ctx, connectionID)
	suite.True(existed)
	suite.Equal(types.OPEN, conn.State)
}

func (suite *KeeperTestSuite) TestHandshake() {
	//get gaia1 consensusState
	state := suite.getConsensusState(ChainIdGaia1)
	//create client on gaia2
	suite.createClient(ChainIdGaia2, ClientToGaia2, clientType, state)

	//get gaia2 consensusState
	state1 := suite.getConsensusState(ChainIdGaia2)
	// create client on gaia1
	suite.createClient(ChainIdGaia1, ClientToGaia1, clientType, state1)

	//===========OpenInit on gaia2===========
	suite.connOpenInit(ChainIdGaia2, ConnectionToGaia1, ClientToGaia2, ClientToGaia1, ConnectionToGaia2)

	//===========OpenTry on gaia1===========
	// update gaia1 consensusState(should be UpdateClient)
	suite.updateClient(ChainIdGaia1, ClientToGaia1)
	// open-try on gaia1
	suite.connOpenTry(ChainIdGaia1, ConnectionToGaia2, ClientToGaia1, ClientToGaia2, ConnectionToGaia1)

	//===========ConnOpenAck on gaia2===========
	// update gaia2 consensusState(should be UpdateClient)
	suite.updateClient(ChainIdGaia2, ClientToGaia2)
	suite.connOpenAck(ChainIdGaia2, ConnectionToGaia1, ConnectionToGaia2)

	//===========ConnOpenConfirm on gaia1===========
	// update gaia1 consensusState(should be UpdateClient)
	suite.updateClient(ChainIdGaia1, ClientToGaia1)
	suite.connOpenConfirm(ChainIdGaia1, ConnectionToGaia2, ConnectionToGaia1)

}
