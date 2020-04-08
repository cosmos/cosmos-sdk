package transfer_test

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
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	transfer "github.com/cosmos/cosmos-sdk/x/ibc/20-transfer"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// define constants used for testing
const (
	testClientIDA = "testclientida"
	testClientIDB = "testclientidb"

	testConnection = "testconnection"
	testPort1      = "bank"
	testPort2      = "testportid"
	testChannel1   = "firstchannel"
	testChannel2   = "secondchannel"

	trustingPeriod time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod      time.Duration = time.Hour * 24 * 7 * 3
)

// define variables used for testing
var (
	testAddr1 = sdk.AccAddress([]byte("testaddr1"))
	testAddr2 = sdk.AccAddress([]byte("testaddr2"))

	testCoins, _          = sdk.ParseCoins("100atom")
	testPrefixedCoins1, _ = sdk.ParseCoins(fmt.Sprintf("100%satom", types.GetDenomPrefix(testPort1, testChannel1)))
	testPrefixedCoins2, _ = sdk.ParseCoins(fmt.Sprintf("100%satom", types.GetDenomPrefix(testPort2, testChannel2)))
)

type HandlerTestSuite struct {
	suite.Suite

	cdc *codec.Codec

	chainA *TestChain
	chainB *TestChain
}

func (suite *HandlerTestSuite) SetupTest() {
	suite.chainA = NewTestChain(testClientIDA)
	suite.chainB = NewTestChain(testClientIDB)

	suite.cdc = suite.chainA.App.Codec()
}

func (suite *HandlerTestSuite) TestHandleMsgTransfer() {
	handler := transfer.NewHandler(suite.chainA.App.TransferKeeper)

	// create channel capability from ibc scoped keeper and claim with transfer scoped keeper
	capName := ibctypes.ChannelCapabilityPath(testPort1, testChannel1)
	cap, err := suite.chainA.App.ScopedIBCKeeper.NewCapability(suite.chainA.GetContext(), capName)
	suite.Require().Nil(err, "could not create capability")
	err = suite.chainA.App.ScopedTransferKeeper.ClaimCapability(suite.chainA.GetContext(), cap, capName)
	suite.Require().Nil(err, "transfer module could not claim capability")

	ctx := suite.chainA.GetContext()
	msg := transfer.NewMsgTransfer(testPort1, testChannel1, 10, testPrefixedCoins2, testAddr1, testAddr2)
	res, err := handler(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Nil(res, "%+v", res) // channel does not exist

	// Setup channel from A to B
	suite.chainA.CreateClient(suite.chainB)
	suite.chainA.createConnection(testConnection, testConnection, testClientIDB, testClientIDA, connectionexported.OPEN)
	suite.chainA.createChannel(testPort1, testChannel1, testPort2, testChannel2, channelexported.OPEN, channelexported.ORDERED, testConnection)

	res, err = handler(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Nil(res, "%+v", res) // next send sequence not found

	nextSeqSend := uint64(1)
	suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(ctx, testPort1, testChannel1, nextSeqSend)
	res, err = handler(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Nil(res, "%+v", res) // sender has insufficient coins

	_ = suite.chainA.App.BankKeeper.SetBalances(ctx, testAddr1, testCoins)
	res, err = handler(ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res, "%+v", res) // successfully executed

	// test when the source is false
	msg = transfer.NewMsgTransfer(testPort1, testChannel1, 10, testPrefixedCoins2, testAddr1, testAddr2)
	_ = suite.chainA.App.BankKeeper.SetBalances(ctx, testAddr1, testPrefixedCoins2)

	res, err = handler(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Nil(res, "%+v", res) // incorrect denom prefix

	msg = transfer.NewMsgTransfer(testPort1, testChannel1, 10, testPrefixedCoins1, testAddr1, testAddr2)
	suite.chainA.App.SupplyKeeper.SetSupply(ctx, supply.NewSupply(testPrefixedCoins1))
	_ = suite.chainA.App.BankKeeper.SetBalances(ctx, testAddr1, testPrefixedCoins1)

	res, err = handler(ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res, "%+v", res) // successfully executed
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
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

	header := ibctmtypes.CreateTestHeader(clientID, 1, now, valSet, signers)

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
func (chain *TestChain) CreateClient(client *TestChain) error {
	client.Header = nextHeader(client)
	// Commit and create a new block on appTarget to get a fresh CommitID
	client.App.Commit()
	commitID := client.App.LastCommitID()
	client.App.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: client.Header.Height, Time: client.Header.Time}})

	// Set HistoricalInfo on client chain after Commit
	ctxClient := client.GetContext()
	validator := staking.NewValidator(
		sdk.ValAddress(client.Vals.Validators[0].Address), client.Vals.Validators[0].PubKey, staking.Description{},
	)
	validator.Status = sdk.Bonded
	validator.Tokens = sdk.NewInt(1000000) // get one voting power
	validators := []staking.Validator{validator}
	histInfo := staking.HistoricalInfo{
		Header: abci.Header{
			AppHash: commitID.Hash,
		},
		Valset: validators,
	}
	client.App.StakingKeeper.SetHistoricalInfo(ctxClient, client.Header.Height, histInfo)

	// Create target ctx
	ctxTarget := chain.GetContext()

	// create client
	clientState, err := ibctmtypes.Initialize(client.ClientID, trustingPeriod, ubdPeriod, client.Header)
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
	// 	ctx.BlockHeader(),
	// 	[]sdk.Msg{clienttypes.NewMsgCreateClient(clientID, clientexported.ClientTypeTendermint, consState, accountAddress)},
	// 	[]uint64{baseAccount.GetAccountNumber()},
	// 	[]uint64{baseAccount.GetSequence()},
	// 	true, true, accountPrivKey,
	// )
}

// nolint: unused
func (chain *TestChain) updateClient(client *TestChain) {
	// Create target ctx
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
	client.App.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: client.Header.Height, Time: client.Header.Time}})

	// Set HistoricalInfo on client chain after Commit
	ctxClient := client.GetContext()
	validator := staking.NewValidator(
		sdk.ValAddress(client.Vals.Validators[0].Address), client.Vals.Validators[0].PubKey, staking.Description{},
	)
	validator.Status = sdk.Bonded
	validator.Tokens = sdk.NewInt(1000000)
	validators := []staking.Validator{validator}
	histInfo := staking.HistoricalInfo{
		Header: abci.Header{
			AppHash: commitID.Hash,
		},
		Valset: validators,
	}
	client.App.StakingKeeper.SetHistoricalInfo(ctxClient, client.Header.Height, histInfo)

	consensusState := ibctmtypes.ConsensusState{
		Height:       uint64(client.Header.Height),
		Timestamp:    client.Header.Time,
		Root:         commitmenttypes.NewMerkleRoot(commitID.Hash),
		ValidatorSet: client.Vals,
	}

	chain.App.IBCKeeper.ClientKeeper.SetClientConsensusState(
		ctxTarget, client.ClientID, uint64(client.Header.Height), consensusState,
	)
	chain.App.IBCKeeper.ClientKeeper.SetClientState(
		ctxTarget, ibctmtypes.NewClientState(client.ClientID, trustingPeriod, ubdPeriod, client.Header),
	)

	// _, _, err := simapp.SignCheckDeliver(
	// 	suite.T(),
	// 	suite.cdc,
	// 	suite.app.BaseApp,
	// 	ctx.BlockHeader(),
	// 	[]sdk.Msg{clienttypes.NewMsgUpdateClient(clientID, suite.header, accountAddress)},
	// 	[]uint64{baseAccount.GetAccountNumber()},
	// 	[]uint64{baseAccount.GetSequence()},
	// 	true, true, accountPrivKey,
	// )
	// suite.Require().NoError(err)
}

func (chain *TestChain) createConnection(
	connID, counterpartyConnID, clientID, counterpartyClientID string,
	state connectionexported.State,
) connectiontypes.ConnectionEnd {
	counterparty := connectiontypes.NewCounterparty(counterpartyClientID, counterpartyConnID, chain.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())
	connection := connectiontypes.ConnectionEnd{
		State:        state,
		ClientID:     clientID,
		Counterparty: counterparty,
		Versions:     connectiontypes.GetCompatibleVersions(),
	}
	ctx := chain.GetContext()
	chain.App.IBCKeeper.ConnectionKeeper.SetConnection(ctx, connID, connection)
	return connection
}

// nolint: unused
func (chain *TestChain) createChannel(
	portID, channelID, counterpartyPortID, counterpartyChannelID string,
	state channelexported.State, order exported.Order, connectionID string,
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
		chain.Header.Time.Add(time.Minute), chain.Vals, chain.Signers)
}
