package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	lite "github.com/tendermint/tendermint/lite2"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	storeKey = host.StoreKey

	testClientIDA     = "testclientida" // chainid for chainA also chainB's clientID for A's liteclient
	testConnectionIDA = "connectionidatob"

	testClientIDB     = "testclientidb" // chainid for chainB also chainA's clientID for B's liteclient
	testConnectionIDB = "connectionidbtoa"

	testClientID3     = "testclientidthree"
	testConnectionID3 = "connectionidthree"

	trustingPeriod time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod      time.Duration = time.Hour * 24 * 7 * 3
	maxClockDrift  time.Duration = time.Second * 10

	nextTimestamp = 10 // increment used for the next header's timestamp
)

var (
	timestamp = time.Now() // starting timestamp for the client test chain
)

type KeeperTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain

	// TODO: delete
	cdc *codec.Codec

	// ChainA testing fields
	oldchainA *TestChain

	// ChainB testing fields
	oldchainB *TestChain
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))

	// TODO: delete
	suite.oldchainA = NewTestChain(testClientIDA)
	suite.oldchainB = NewTestChain(testClientIDB)

	suite.cdc = suite.oldchainA.App.Codec()
}

// nolint: unused
func queryProof(chain *TestChain, key []byte) ([]byte, uint64) {
	res := chain.App.Query(abci.RequestQuery{
		Path:   fmt.Sprintf("store/%s/key", storeKey),
		Height: chain.App.LastBlockHeight(),
		Data:   key,
		Prove:  true,
	})

	merkleProof := commitmenttypes.MerkleProof{
		Proof: res.Proof,
	}

	proof, _ := chain.App.AppCodec().MarshalBinaryBare(&merkleProof)

	return proof, uint64(res.Height)
}
func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestSetAndGetConnection() {
	_, existed := suite.oldchainA.App.IBCKeeper.ConnectionKeeper.GetConnection(suite.oldchainA.GetContext(), testConnectionIDA)
	suite.Require().False(existed)

	counterparty := types.NewCounterparty(testClientIDA, testConnectionIDA, commitmenttypes.NewMerklePrefix(suite.oldchainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes()))
	expConn := types.NewConnectionEnd(types.INIT, testConnectionIDB, testClientIDB, counterparty, types.GetCompatibleVersions())
	suite.oldchainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.oldchainA.GetContext(), testConnectionIDA, expConn)
	conn, existed := suite.oldchainA.App.IBCKeeper.ConnectionKeeper.GetConnection(suite.oldchainA.GetContext(), testConnectionIDA)
	suite.Require().True(existed)
	suite.Require().EqualValues(expConn, conn)
}

func (suite *KeeperTestSuite) TestSetAndGetClientConnectionPaths() {
	_, existed := suite.oldchainA.App.IBCKeeper.ConnectionKeeper.GetClientConnectionPaths(suite.oldchainA.GetContext(), testClientIDA)
	suite.False(existed)

	suite.oldchainA.App.IBCKeeper.ConnectionKeeper.SetClientConnectionPaths(suite.oldchainA.GetContext(), testClientIDB, types.GetCompatibleVersions())
	paths, existed := suite.oldchainA.App.IBCKeeper.ConnectionKeeper.GetClientConnectionPaths(suite.oldchainA.GetContext(), testClientIDB)
	suite.True(existed)
	suite.EqualValues(types.GetCompatibleVersions(), paths)
}

func (suite KeeperTestSuite) TestGetAllConnections() {
	// Connection (Counterparty): A(C) -> C(B) -> B(A)
	counterparty1 := types.NewCounterparty(testClientIDA, testConnectionIDA, commitmenttypes.NewMerklePrefix(suite.oldchainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes()))
	counterparty2 := types.NewCounterparty(testClientIDB, testConnectionIDB, commitmenttypes.NewMerklePrefix(suite.oldchainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes()))
	counterparty3 := types.NewCounterparty(testClientID3, testConnectionID3, commitmenttypes.NewMerklePrefix(suite.oldchainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes()))

	conn1 := types.NewConnectionEnd(types.INIT, testConnectionIDA, testClientIDA, counterparty3, types.GetCompatibleVersions())
	conn2 := types.NewConnectionEnd(types.INIT, testConnectionIDB, testClientIDB, counterparty1, types.GetCompatibleVersions())
	conn3 := types.NewConnectionEnd(types.UNINITIALIZED, testConnectionID3, testClientID3, counterparty2, types.GetCompatibleVersions())

	expConnections := []types.ConnectionEnd{conn1, conn2, conn3}

	for i := range expConnections {
		suite.oldchainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.oldchainA.GetContext(), expConnections[i].ID, expConnections[i])
	}

	connections := suite.oldchainA.App.IBCKeeper.ConnectionKeeper.GetAllConnections(suite.oldchainA.GetContext())
	suite.Require().Len(connections, len(expConnections))
	suite.Require().Equal(expConnections, connections)
}

func (suite KeeperTestSuite) TestGetAllClientConnectionPaths() {
	clients := []clientexported.ClientState{
		ibctmtypes.NewClientState(testClientIDA, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, ibctmtypes.Header{}, commitmenttypes.GetSDKSpecs()),
		ibctmtypes.NewClientState(testClientIDB, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, ibctmtypes.Header{}, commitmenttypes.GetSDKSpecs()),
		ibctmtypes.NewClientState(testClientID3, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, ibctmtypes.Header{}, commitmenttypes.GetSDKSpecs()),
	}

	for i := range clients {
		suite.oldchainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.oldchainA.GetContext(), clients[i])
	}

	expPaths := []types.ConnectionPaths{
		types.NewConnectionPaths(testClientIDA, []string{host.ConnectionPath(testConnectionIDA)}),
		types.NewConnectionPaths(testClientIDB, []string{host.ConnectionPath(testConnectionIDB), host.ConnectionPath(testConnectionID3)}),
	}

	for i := range expPaths {
		suite.oldchainA.App.IBCKeeper.ConnectionKeeper.SetClientConnectionPaths(suite.oldchainA.GetContext(), expPaths[i].ClientID, expPaths[i].Paths)
	}

	connPaths := suite.oldchainA.App.IBCKeeper.ConnectionKeeper.GetAllClientConnectionPaths(suite.oldchainA.GetContext())
	suite.Require().Len(connPaths, 2)
	suite.Require().Equal(connPaths, expPaths)
}

// TestGetTimestampAtHeight verifies if the clients on each chain return the correct timestamp
// for the other chain.
func (suite *KeeperTestSuite) TestGetTimestampAtHeight() {
	cases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"verification success", func() {
			suite.oldchainA.CreateClient(suite.oldchainB)
		}, true},
		{"client state not found", func() {}, false},
	}

	for i, tc := range cases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			// create and store a connection to chainB on chainA
			connection := suite.oldchainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.OPEN)

			actualTimestamp, err := suite.oldchainA.App.IBCKeeper.ConnectionKeeper.GetTimestampAtHeight(
				suite.oldchainA.GetContext(), connection, uint64(suite.oldchainB.Header.Height),
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
				suite.Require().EqualValues(uint64(suite.oldchainB.Header.Time.UnixNano()), actualTimestamp)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

// TestChain is a testing struct that wraps a simapp with the latest Header, Vals and Signers
// It also contains a field called ClientID. This is the clientID that *other* chains use
// to refer to this TestChain. For simplicity's sake it is also the chainID on the TestChain Header
type TestChain struct {
	ClientID string
	App      *simapp.SimApp
	Header   ibctmtypes.Header
	Vals     *tmtypes.ValidatorSet
	Signers  []tmtypes.PrivValidator
}

func NewTestChain(clientID string) *TestChain {
	privVal := tmtypes.NewMockPV()

	pubKey, err := privVal.GetPubKey()
	if err != nil {
		panic(err)
	}

	validator := tmtypes.NewValidator(pubKey, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})
	signers := []tmtypes.PrivValidator{privVal}

	header := ibctmtypes.CreateTestHeader(clientID, 1, timestamp, valSet, signers)

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
	return chain.App.BaseApp.NewContext(false, abci.Header{ChainID: chain.Header.SignedHeader.Header.ChainID, Height: chain.Header.SignedHeader.Header.Height})
}

// createClient will create a client for clientChain on targetChain
func (chain *TestChain) CreateClient(client *TestChain) error {
	client.Header = nextHeader(client)
	// Commit and create a new block on appTarget to get a fresh CommitID
	client.App.Commit()
	commitID := client.App.LastCommitID()
	client.App.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: client.Header.SignedHeader.Header.Height, Time: client.Header.Time}})

	// Set HistoricalInfo on client chain after Commit
	ctxClient := client.GetContext()
	validator := stakingtypes.NewValidator(
		sdk.ValAddress(client.Vals.Validators[0].Address), client.Vals.Validators[0].PubKey, stakingtypes.Description{},
	)
	validator.Status = sdk.Bonded
	validator.Tokens = sdk.NewInt(1000000) // get one voting power
	validators := []stakingtypes.Validator{validator}
	histInfo := stakingtypes.HistoricalInfo{
		Header: abci.Header{
			Time:    client.Header.Time,
			AppHash: commitID.Hash,
		},
		Valset: validators,
	}
	client.App.StakingKeeper.SetHistoricalInfo(ctxClient, client.Header.SignedHeader.Header.Height, histInfo)

	// also set staking params
	stakingParams := stakingtypes.DefaultParams()
	stakingParams.HistoricalEntries = 10
	client.App.StakingKeeper.SetParams(ctxClient, stakingParams)

	// Create target ctx
	ctxTarget := chain.GetContext()

	// create client
	clientState, err := ibctmtypes.Initialize(client.ClientID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, client.Header, commitmenttypes.GetSDKSpecs())
	if err != nil {
		return err
	}
	_, err = chain.App.IBCKeeper.ClientKeeper.CreateClient(ctxTarget, clientState, client.Header.ConsensusState())
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

func (chain *TestChain) updateClient(client *TestChain) {
	// Create chain ctx
	ctxTarget := chain.GetContext()

	// if clientState does not already exist, return without updating
	_, found := chain.App.IBCKeeper.ClientKeeper.GetClientState(
		ctxTarget, client.ClientID,
	)
	if !found {
		return
	}

	// always commit when updateClient and begin a new block
	client.App.Commit()
	commitID := client.App.LastCommitID()
	client.Header = nextHeader(client)

	/*
		err := chain.App.IBCKeeper.ClientKeeper.UpdateClient(ctxTarget, client.ClientID, client.Header)
		if err != nil {
			panic(err)
		}
	*/

	client.App.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: client.Header.SignedHeader.Header.Height, Time: client.Header.Time}})

	// Set HistoricalInfo on client chain after Commit
	ctxClient := client.GetContext()
	validator := stakingtypes.NewValidator(
		sdk.ValAddress(client.Vals.Validators[0].Address), client.Vals.Validators[0].PubKey, stakingtypes.Description{},
	)
	validator.Status = sdk.Bonded
	validator.Tokens = sdk.NewInt(1000000)
	validators := []stakingtypes.Validator{validator}
	histInfo := stakingtypes.HistoricalInfo{
		Header: abci.Header{
			Time:    client.Header.Time,
			AppHash: commitID.Hash,
		},
		Valset: validators,
	}
	client.App.StakingKeeper.SetHistoricalInfo(ctxClient, client.Header.SignedHeader.Header.Height, histInfo)

	consensusState := ibctmtypes.ConsensusState{
		Height:       client.Header.GetHeight(),
		Timestamp:    client.Header.Time,
		Root:         commitmenttypes.NewMerkleRoot(commitID.Hash),
		ValidatorSet: client.Vals,
	}

	chain.App.IBCKeeper.ClientKeeper.SetClientConsensusState(
		ctxTarget, client.ClientID, client.Header.GetHeight(), consensusState,
	)
	chain.App.IBCKeeper.ClientKeeper.SetClientState(
		ctxTarget, ibctmtypes.NewClientState(client.ClientID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, client.Header, commitmenttypes.GetSDKSpecs()),
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
	state types.State,
) types.ConnectionEnd {
	counterparty := types.NewCounterparty(counterpartyClientID, counterpartyConnID, commitmenttypes.NewMerklePrefix(chain.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes()))
	connection := types.ConnectionEnd{
		State:        state,
		ID:           connID,
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
	state channeltypes.State, order channeltypes.Order, connectionID string,
) channeltypes.Channel {
	counterparty := channeltypes.NewCounterparty(counterpartyPortID, counterpartyChannelID)
	channel := channeltypes.NewChannel(state, order, counterparty, []string{connectionID}, "1.0")
	ctx := chain.GetContext()
	chain.App.IBCKeeper.ChannelKeeper.SetChannel(ctx, portID, channelID, channel)
	return channel
}

func nextHeader(chain *TestChain) ibctmtypes.Header {
	return ibctmtypes.CreateTestHeader(
		chain.Header.SignedHeader.Header.ChainID,
		chain.Header.SignedHeader.Header.Height+1,
		chain.Header.Time.Add(nextTimestamp), chain.Vals, chain.Signers,
	)
}

func prefixedClientKey(clientID string, key []byte) []byte {
	return append([]byte("clients/"+clientID+"/"), key...)
}
