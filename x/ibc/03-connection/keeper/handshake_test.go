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

func (suite *KeeperTestSuite) getConsensusState(chainID string) tendermint.ConsensusState {
	app := suite.apps[chainID]
	commitID := app.store.Commit()
	state := tendermint.ConsensusState{
		ChainID: app.chainID,
		Height:  uint64(commitID.Version),
		Root:    commitment.NewRoot(commitID.Hash),
	}
	app.ctx.WithBlockHeight(commitID.Version)
	return state
}

func (suite *KeeperTestSuite) queryProof(chainID string, key string) (proof commitment.Proof, height int64) {
	app := suite.apps[chainID]
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

func (suite *KeeperTestSuite) createClient(chainID string, clientID string,
	clientType exported.ClientType, state tendermint.ConsensusState) {
	app := suite.apps[chainID]
	_, err := app.clientKeeper.CreateClient(app.ctx, clientID, clientType, state)
	if err != nil {
		panic(err)
	}
	commitID := app.store.Commit()
	app.ctx.WithBlockHeight(commitID.Version)
}

func (suite *KeeperTestSuite) updateClient(chainID string, clientID string) {
	otherChainID := ChainIDGaia1
	if chainID == ChainIDGaia1 {
		otherChainID = ChainIDGaia2
	}
	consensusState := suite.getConsensusState(otherChainID)

	//update client consensus state
	app := suite.apps[chainID]
	app.clientKeeper.SetConsensusState(app.ctx, clientID, consensusState)
	app.clientKeeper.SetVerifiedRoot(app.ctx, clientID, consensusState.GetHeight(), consensusState.GetRoot())
	commitID := app.store.Commit()
	app.ctx.WithBlockHeight(commitID.Version)
}

func (suite *KeeperTestSuite) connOpenInit(chainID string, connectionID, clientID, counterpartyClientID, counterpartyConnID string) {
	app := suite.apps[chainID]
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
	commitID := app.store.Commit()
	app.ctx.WithBlockHeight(commitID.Version)
}

func (suite *KeeperTestSuite) connOpenTry(chainID string, connectionID, clientID, counterpartyClientID, counterpartyConnID string) {
	app := suite.apps[chainID]
	counterparty := types.NewCounterparty(counterpartyClientID, counterpartyConnID, app.connKeeper.GetCommitmentPrefix())

	connKey := fmt.Sprintf("%s/%s", types.SubModuleName, types.ConnectionPath(counterpartyConnID))
	otherChainID := ChainIDGaia1
	if chainID == ChainIDGaia1 {
		otherChainID = ChainIDGaia2
	}
	proof, h := suite.queryProof(otherChainID, connKey)

	err := app.connKeeper.ConnOpenTry(app.ctx, connectionID, counterparty, clientID, types.GetCompatibleVersions(), proof, uint64(h), 0)
	suite.Nil(err)

	commitID := app.store.Commit()
	app.ctx.WithBlockHeight(commitID.Version)

	//check connection state
	conn, existed := app.connKeeper.GetConnection(app.ctx, connectionID)
	suite.True(existed)
	suite.Equal(types.TRYOPEN, conn.State)
}

func (suite *KeeperTestSuite) connOpenAck(chainID string, connectionID, counterpartyConnID string) {
	app := suite.apps[chainID]
	connKey := fmt.Sprintf("%s/%s", types.SubModuleName, types.ConnectionPath(counterpartyConnID))
	otherChainID := ChainIDGaia1
	if chainID == ChainIDGaia1 {
		otherChainID = ChainIDGaia2
	}
	proof, h := suite.queryProof(otherChainID, connKey)

	err := app.connKeeper.ConnOpenAck(app.ctx, connectionID, types.GetCompatibleVersions()[0], proof, uint64(h), 0)
	suite.Nil(err)

	commitID := app.store.Commit()
	app.ctx.WithBlockHeight(commitID.Version)

	//check connection state
	conn, existed := app.connKeeper.GetConnection(app.ctx, connectionID)
	suite.True(existed)
	suite.Equal(types.OPEN, conn.State)
}

func (suite *KeeperTestSuite) connOpenConfirm(chainID string, connectionID, counterpartyConnID string) {
	app := suite.apps[chainID]
	connKey := fmt.Sprintf("%s/%s", types.SubModuleName, types.ConnectionPath(counterpartyConnID))
	otherChainID := ChainIDGaia1
	if chainID == ChainIDGaia1 {
		otherChainID = ChainIDGaia2
	}
	proof, h := suite.queryProof(otherChainID, connKey)

	err := app.connKeeper.ConnOpenConfirm(app.ctx, connectionID, proof, uint64(h))
	suite.Nil(err)

	commitID := app.store.Commit()
	app.ctx.WithBlockHeight(commitID.Version)

	//check connection state
	conn, existed := app.connKeeper.GetConnection(app.ctx, connectionID)
	suite.True(existed)
	suite.Equal(types.OPEN, conn.State)
}

func (suite *KeeperTestSuite) TestHandshake() {
	//get gaia1 consensusState
	state1 := suite.getConsensusState(ChainIDGaia1)
	//create gaia1's client on gaia2
	suite.createClient(ChainIDGaia2, ClientToGaia2, clientType, state1)

	//get gaia2 consensusState
	state2 := suite.getConsensusState(ChainIDGaia2)
	// create gaia2's client on gaia1
	suite.createClient(ChainIDGaia1, ClientToGaia1, clientType, state2)

	//===========OpenInit on gaia2===========
	suite.connOpenInit(ChainIDGaia2, ConnectionToGaia1, ClientToGaia2, ClientToGaia1, ConnectionToGaia2)

	//===========OpenTry on gaia1===========
	// update gaia1 consensusState(should be UpdateClient)
	suite.updateClient(ChainIDGaia1, ClientToGaia1)
	// open-try on gaia1
	suite.connOpenTry(ChainIDGaia1, ConnectionToGaia2, ClientToGaia1, ClientToGaia2, ConnectionToGaia1)

	//===========ConnOpenAck on gaia2===========
	// update gaia2 consensusState(should be UpdateClient)
	suite.updateClient(ChainIDGaia2, ClientToGaia2)
	suite.connOpenAck(ChainIDGaia2, ConnectionToGaia1, ConnectionToGaia2)

	//===========ConnOpenConfirm on gaia1===========
	// update gaia1 consensusState(should be UpdateClient)
	suite.updateClient(ChainIDGaia1, ClientToGaia1)
	suite.connOpenConfirm(ChainIDGaia1, ConnectionToGaia2, ConnectionToGaia1)

}
