package keeper_test

import (
	"fmt"
	"testing"
	"time"

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
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

const (
	clientType = clientexported.Tendermint
	storeKey   = ibctypes.StoreKey
	chainID    = "gaia"

	testClientIDA     = "testclientida"
	testConnectionIDA = "connectionidatob"

	testClientIDB     = "testclientidb"
	testConnectionIDB = "connectionidbtoa"

	testClientID3     = "testclientidthree"
	testConnectionID3 = "connectionidthree"

	trustingPeriod time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod      time.Duration = time.Hour * 24 * 7 * 3
)

type KeeperTestSuite struct {
	suite.Suite

	cdc *codec.Codec

	// ChainA testing fields
	chainA *TestChain

	// ChainB testing fields
	chainB *TestChain
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.chainA = NewTestChain(testClientIDA)
	suite.chainB = NewTestChain(testClientIDB)

	suite.cdc = suite.chainA.App.Codec()
}

// nolint: unused
func (suite *KeeperTestSuite) queryProof(key []byte) (commitment.Proof, int64) {
	res := suite.chainA.App.Query(abci.RequestQuery{
		Path:   fmt.Sprintf("store/%s/key", storeKey),
		Height: suite.chainA.App.LastBlockHeight(),
		Data:   key,
		Prove:  true,
	})

	proof := commitment.Proof{
		Proof: res.Proof,
	}

	return proof, res.Height
}
func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestSetAndGetConnection() {
	_, existed := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetConnection(suite.chainA.GetContext(), testConnectionIDA)
	suite.Require().False(existed)

	counterparty := types.NewCounterparty(testClientIDA, testConnectionIDA, suite.chainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())
	expConn := types.NewConnectionEnd(exported.INIT, testClientIDA, counterparty, types.GetCompatibleVersions())
	suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainA.GetContext(), testConnectionIDA, expConn)
	conn, existed := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetConnection(suite.chainA.GetContext(), testConnectionIDA)
	suite.Require().True(existed)
	suite.Require().EqualValues(expConn, conn)
}

func (suite *KeeperTestSuite) TestSetAndGetClientConnectionPaths() {
	_, existed := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetClientConnectionPaths(suite.chainA.GetContext(), testClientIDA)
	suite.False(existed)

	suite.chainA.App.IBCKeeper.ConnectionKeeper.SetClientConnectionPaths(suite.chainA.GetContext(), testClientIDA, types.GetCompatibleVersions())
	paths, existed := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetClientConnectionPaths(suite.chainA.GetContext(), testClientIDA)
	suite.True(existed)
	suite.EqualValues(types.GetCompatibleVersions(), paths)
}

func (suite KeeperTestSuite) TestGetAllConnections() {
	// Connection (Counterparty): A(C) -> C(B) -> B(A)
	counterparty1 := types.NewCounterparty(testClientIDA, testConnectionIDA, suite.chainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())
	counterparty2 := types.NewCounterparty(testClientIDB, testConnectionIDB, suite.chainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())
	counterparty3 := types.NewCounterparty(testClientID3, testConnectionID3, suite.chainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())

	conn1 := types.NewConnectionEnd(exported.INIT, testClientIDA, counterparty3, types.GetCompatibleVersions())
	conn2 := types.NewConnectionEnd(exported.INIT, testClientIDB, counterparty1, types.GetCompatibleVersions())
	conn3 := types.NewConnectionEnd(exported.UNINITIALIZED, testClientID3, counterparty2, types.GetCompatibleVersions())

	expConnections := []types.ConnectionEnd{conn1, conn2, conn3}

	suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainA.GetContext(), testConnectionIDA, expConnections[0])
	suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainA.GetContext(), testConnectionIDB, expConnections[1])
	suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainA.GetContext(), testConnectionID3, expConnections[2])

	connections := suite.chainA.App.IBCKeeper.ConnectionKeeper.GetAllConnections(suite.chainA.GetContext())
	suite.Require().Len(connections, len(expConnections))
	suite.Require().ElementsMatch(expConnections, connections)
}

type TestChain struct {
	ClientID string
	App      *simapp.SimApp
	Header   ibctmtypes.Header
	Vals     *tmtypes.ValidatorSet
	Signers  []tmtypes.PrivValidator
}

func NewTestChain(clientID string) *TestChain {
	privVal := tmtypes.NewMockPV()
	validator := tmtypes.NewValidator(privVal.GetPubKey(), 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})
	signers := []tmtypes.PrivValidator{privVal}
	now := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)

	header := ibctmtypes.CreateTestHeader(clientID, 1, now, valSet, valSet, signers)

	return &TestChain{
		ClientID: clientID,
		App:      simapp.Setup(false),
		Header:   header,
		Vals:     valSet,
		Signers:  signers,
	}
}

// Creates simple context for testing purposes
func (chain *TestChain) GetContext() sdk.Context {
	return chain.App.BaseApp.NewContext(false, abci.Header{ChainID: chain.Header.ChainID, Height: chain.Header.Height})
}

// createClient will create a client for clientChain on targetChain
func (target *TestChain) CreateClient(client *TestChain) error {
	client.Header = nextHeader(client)
	// Commit and create a new block on appTarget to get a fresh CommitID
	client.App.Commit()
	client.App.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: client.Header.Height, Time: client.Header.Time}})

	// Create target ctx
	ctxTarget := target.GetContext()

	// create client
	clientState, err := ibctmtypes.Initialize(client.ClientID, trustingPeriod, ubdPeriod, client.Header)
	if err != nil {
		return err
	}
	_, err = target.App.IBCKeeper.ClientKeeper.CreateClient(ctxTarget, clientState, client.Header.ConsensusState())
	if err != nil {
		return err
	}
	return nil

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

func (target *TestChain) updateClient(client *TestChain) {
	// Create target ctx
	ctxTarget := target.GetContext()

	// if clientState does not already exist, return without updating
	_, found := target.App.IBCKeeper.ClientKeeper.GetClientState(
		ctxTarget, client.ClientID,
	)
	if !found {
		return
	}

	// always commit when updateClient and begin a new block
	client.App.Commit()
	commitID := client.App.LastCommitID()

	client.Header = nextHeader(client)

	client.App.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: client.Header.Height, Time: client.Header.Time}})

	consensusState := ibctmtypes.ConsensusState{
		Height:       uint64(client.Header.Height),
		Timestamp:    client.Header.Time,
		Root:         commitment.NewRoot(commitID.Hash),
		ValidatorSet: client.Vals,
	}

	target.App.IBCKeeper.ClientKeeper.SetClientConsensusState(
		ctxTarget, client.ClientID, uint64(client.Header.Height), consensusState,
	)
	target.App.IBCKeeper.ClientKeeper.SetClientState(
		ctxTarget, ibctmtypes.NewClientState(client.ClientID, trustingPeriod, ubdPeriod, client.Header),
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

func (chain *TestChain) createConnection(
	connID, counterpartyConnID, clientID, counterpartyClientID string,
	state exported.State,
) types.ConnectionEnd {
	counterparty := types.NewCounterparty(counterpartyClientID, counterpartyConnID, chain.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())
	connection := types.ConnectionEnd{
		State:        state,
		ClientID:     clientID,
		Counterparty: counterparty,
		Versions:     types.GetCompatibleVersions(),
	}
	ctx := chain.GetContext()
	chain.App.IBCKeeper.ConnectionKeeper.SetConnection(ctx, connID, connection)
	return connection
}

func (chain *TestChain) createChannel(
	portID, channelID, counterpartyPortID, counterpartyChannelID string,
	state channelexported.State, order channelexported.Order, connectionID string,
) channeltypes.Channel {
	counterparty := channeltypes.NewCounterparty(counterpartyPortID, counterpartyChannelID)
	channel := channeltypes.NewChannel(state, order, counterparty,
		[]string{connectionID}, "1.0",
	)
	ctx := chain.GetContext()
	chain.App.IBCKeeper.ChannelKeeper.SetChannel(ctx, portID, channelID, channel)
	return channel
}

func nextHeader(chain *TestChain) ibctmtypes.Header {
	return ibctmtypes.CreateTestHeader(chain.Header.ChainID, chain.Header.Height+1,
		chain.Header.Time.Add(time.Minute), chain.Vals, chain.Vals, chain.Signers)
}
