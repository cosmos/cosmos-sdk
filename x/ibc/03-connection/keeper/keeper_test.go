package keeper_test

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

const (
	clientType = clientexported.Tendermint
	storeKey   = ibctypes.StoreKey
	chainID    = "gaia"
	testHeight = 10

	testClientID1     = "testclientidone"
	testConnectionID1 = "connectionidone"

	testClientID2     = "testclientidtwo"
	testConnectionID2 = "connectionidtwo"

	testClientID3     = "testclientidthree"
	testConnectionID3 = "connectionidthree"
)

type KeeperTestSuite struct {
	suite.Suite

	cdc            *codec.Codec
	ctx            sdk.Context
	app            *simapp.SimApp
	valSet         *tmtypes.ValidatorSet
	lastValSet     *tmtypes.ValidatorSet
	consensusState clientexported.ConsensusState
	header         tendermint.Header
}

func (suite *KeeperTestSuite) SetupTest() {
	isCheckTx := false
	app := simapp.Setup(isCheckTx)

	suite.cdc = app.Codec()
	suite.ctx = app.BaseApp.NewContext(isCheckTx, abci.Header{ChainID: chainID, Height: 1})
	suite.app = app

	var signers []tmtypes.PrivValidator

	var validators staking.Validators
	for i := 1; i < 11; i++ {
		privVal := tmtypes.NewMockPV()
		pk := privVal.GetPubKey()
		val := staking.NewValidator(sdk.ValAddress(pk.Address()), pk, staking.Description{})
		val.Status = sdk.Bonded
		val.Tokens = sdk.NewInt(rand.Int63())
		validators = append(validators, val)
		signers = append(signers, privVal)

		suite.lastValSet = suite.valSet
		suite.valSet = tmtypes.NewValidatorSet(validators.ToTmValidators())
		app.StakingKeeper.SetHistoricalInfo(suite.ctx, int64(i), staking.NewHistoricalInfo(suite.ctx.BlockHeader(), validators))
	}
	suite.header = tendermint.CreateTestHeader(chainID, testHeight, suite.lastValSet, suite.valSet, signers)

	suite.consensusState = tendermint.ConsensusState{
		Root:             commitment.NewRoot(suite.header.AppHash),
		ValidatorSetHash: suite.valSet.Hash(),
	}
}

// nolint: unused
func (suite *KeeperTestSuite) queryProof(key []byte) (commitment.Proof, int64) {
	res := suite.app.Query(abci.RequestQuery{
		Path:   fmt.Sprintf("store/%s/key", storeKey),
		Height: suite.app.LastBlockHeight(),
		Data:   key,
		Prove:  true,
	})

	proof := commitment.Proof{
		Proof: res.Proof,
	}

	return proof, res.Height
}

func (suite *KeeperTestSuite) createClient(clientID string) {
	suite.app.Commit()
	commitID := suite.app.LastCommitID()

	suite.app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: suite.app.LastBlockHeight() + 1}})
	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)

	consensusState := tendermint.ConsensusState{
		Root:             commitment.NewRoot(commitID.Hash),
		ValidatorSetHash: suite.valSet.Hash(),
	}

	_, err := suite.app.IBCKeeper.ClientKeeper.CreateClient(suite.ctx, clientID, clientType, consensusState)
	suite.Require().NoError(err)

	// _, _, err := simapp.SignCheckDeliver(
	// 	suite.T(),
	// 	suite.cdc,
	// 	suite.app.BaseApp,
	// 	suite.ctx.BlockHeader(),
	// 	[]sdk.Msg{clienttypes.NewMsgCreateClient(clientID, clientexported.ClientTypeTendermint, consState, accountAddress)},
	// 	[]uint64{baseAccount.GetAccountNumber()},
	// 	[]uint64{baseAccount.GetSequence()},
	// 	true, true, accountPrivKey,
	// )
}

func (suite *KeeperTestSuite) updateClient(clientID string) {
	// always commit when updateClient and begin a new block
	suite.app.Commit()
	commitID := suite.app.LastCommitID()

	suite.app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: suite.app.LastBlockHeight() + 1}})
	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)

	consensusState := tendermint.ConsensusState{
		Root:             commitment.NewRoot(commitID.Hash),
		ValidatorSetHash: suite.valSet.Hash(),
	}

	suite.app.IBCKeeper.ClientKeeper.SetConsensusState(
		suite.ctx, clientID, uint64(suite.app.LastBlockHeight()), consensusState,
	)

	// _, _, err := simapp.SignCheckDeliver(
	// 	suite.T(),
	// 	suite.cdc,
	// 	suite.app.BaseApp,
	// 	suite.ctx.BlockHeader(),
	// 	[]sdk.Msg{clienttypes.NewMsgUpdateClient(clientID, suite.header, accountAddress)},
	// 	[]uint64{baseAccount.GetAccountNumber()},
	// 	[]uint64{baseAccount.GetSequence()},
	// 	true, true, accountPrivKey,
	// )
	// suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) createConnection(
	connID, counterpartyConnID, clientID, counterpartyClientID string,
	state exported.State,
) types.ConnectionEnd {
	counterparty := types.NewCounterparty(counterpartyClientID, counterpartyConnID, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())
	connection := types.ConnectionEnd{
		State:        state,
		ClientID:     clientID,
		Counterparty: counterparty,
		Versions:     types.GetCompatibleVersions(),
	}
	suite.app.IBCKeeper.ConnectionKeeper.SetConnection(suite.ctx, connID, connection)
	return connection
}

func (suite *KeeperTestSuite) createChannel(
	portID, channelID, counterpartyPortID, counterpartyChannelID string,
	state channelexported.State, order channelexported.Order, connectionID string,
) channeltypes.Channel {
	counterparty := channeltypes.NewCounterparty(counterpartyPortID, counterpartyChannelID)
	channel := channeltypes.NewChannel(state, order, counterparty,
		[]string{connectionID}, "1.0",
	)
	suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, portID, channelID, channel)
	return channel
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestSetAndGetConnection() {
	_, existed := suite.app.IBCKeeper.ConnectionKeeper.GetConnection(suite.ctx, testConnectionID1)
	suite.Require().False(existed)

	counterparty := types.NewCounterparty(testClientID1, testConnectionID1, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())
	expConn := types.NewConnectionEnd(exported.INIT, testClientID1, counterparty, types.GetCompatibleVersions())
	suite.app.IBCKeeper.ConnectionKeeper.SetConnection(suite.ctx, testConnectionID1, expConn)
	conn, existed := suite.app.IBCKeeper.ConnectionKeeper.GetConnection(suite.ctx, testConnectionID1)
	suite.Require().True(existed)
	suite.Require().EqualValues(expConn, conn)
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

	conn1 := types.NewConnectionEnd(exported.INIT, testClientID1, counterparty3, types.GetCompatibleVersions())
	conn2 := types.NewConnectionEnd(exported.INIT, testClientID2, counterparty1, types.GetCompatibleVersions())
	conn3 := types.NewConnectionEnd(exported.UNINITIALIZED, testClientID3, counterparty2, types.GetCompatibleVersions())

	expConnections := []types.ConnectionEnd{conn1, conn2, conn3}

	suite.app.IBCKeeper.ConnectionKeeper.SetConnection(suite.ctx, testConnectionID1, expConnections[0])
	suite.app.IBCKeeper.ConnectionKeeper.SetConnection(suite.ctx, testConnectionID2, expConnections[1])
	suite.app.IBCKeeper.ConnectionKeeper.SetConnection(suite.ctx, testConnectionID3, expConnections[2])

	connections := suite.app.IBCKeeper.ConnectionKeeper.GetAllConnections(suite.ctx)
	suite.Require().Len(connections, len(expConnections))
	suite.Require().ElementsMatch(expConnections, connections)
}

// Mocked types
// TODO: fix tests and replace for real proofs

var (
	_ commitment.ProofI = validProof{}
	_ commitment.ProofI = invalidProof{}
)

type (
	validProof   struct{}
	invalidProof struct{}
)

func (validProof) GetCommitmentType() commitment.Type {
	return commitment.Merkle
}

func (validProof) VerifyMembership(
	root commitment.RootI, path commitment.PathI, value []byte) error {
	return nil
}

func (validProof) VerifyNonMembership(root commitment.RootI, path commitment.PathI) error {
	return nil
}

func (validProof) ValidateBasic() error {
	return nil
}

func (validProof) IsEmpty() bool {
	return false
}

func (invalidProof) GetCommitmentType() commitment.Type {
	return commitment.Merkle
}

func (invalidProof) VerifyMembership(
	root commitment.RootI, path commitment.PathI, value []byte) error {
	return errors.New("proof failed")
}

func (invalidProof) VerifyNonMembership(root commitment.RootI, path commitment.PathI) error {
	return errors.New("proof failed")
}

func (invalidProof) ValidateBasic() error {
	return errors.New("invalid proof")
}

func (invalidProof) IsEmpty() bool {
	return true
}
