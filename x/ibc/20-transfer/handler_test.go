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
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	transfer "github.com/cosmos/cosmos-sdk/x/ibc/20-transfer"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// define constants used for testing
const (
	testChainID    = "test-chain-id"
	testClient     = "test-client"
	testClientType = clientexported.Tendermint

	testConnection = "testconnection"
	testPort1      = "bank"
	testPort2      = "testportid"
	testChannel1   = "firstchannel"
	testChannel2   = "secondchannel"

	testChannelOrder   = channelexported.UNORDERED
	testChannelVersion = "1.0"

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

	cdc    *codec.Codec
	ctx    sdk.Context
	app    *simapp.SimApp
	valSet *tmtypes.ValidatorSet
}

func (suite *HandlerTestSuite) SetupTest() {
	isCheckTx := false
	app := simapp.Setup(isCheckTx)

	suite.cdc = app.Codec()
	suite.ctx = app.BaseApp.NewContext(isCheckTx, abci.Header{})
	suite.app = app

	privVal := tmtypes.NewMockPV()

	validator := tmtypes.NewValidator(privVal.GetPubKey(), 1)
	suite.valSet = tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})

	suite.createClient()
	suite.createConnection(connectionexported.OPEN)
}

func (suite *HandlerTestSuite) createClient() {
	suite.app.Commit()
	commitID := suite.app.LastCommitID()

	suite.app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: suite.app.LastBlockHeight() + 1}})
	suite.ctx = suite.app.BaseApp.NewContext(false, abci.Header{})

	consensusState := ibctmtypes.ConsensusState{
		Root:         commitment.NewRoot(commitID.Hash),
		ValidatorSet: suite.valSet,
	}

	clientState, err := ibctmtypes.Initialize(testClient, testClient, consensusState, trustingPeriod, ubdPeriod)
	suite.NoError(err)
	_, err = suite.app.IBCKeeper.ClientKeeper.CreateClient(suite.ctx, clientState, consensusState)
	suite.NoError(err)
}

func (suite *HandlerTestSuite) updateClient() {
	// always commit and begin a new block on updateClient
	suite.app.Commit()
	commitID := suite.app.LastCommitID()

	suite.app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: suite.app.LastBlockHeight() + 1}})
	suite.ctx = suite.app.BaseApp.NewContext(false, abci.Header{})

	state := ibctmtypes.ConsensusState{
		Root: commitment.NewRoot(commitID.Hash),
	}

	suite.app.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.ctx, testClient, 1, state)
}

func (suite *HandlerTestSuite) createConnection(state connectionexported.State) {
	connection := connection.ConnectionEnd{
		State:    state,
		ClientID: testClient,
		Counterparty: connection.Counterparty{
			ClientID:     testClient,
			ConnectionID: testConnection,
			Prefix:       suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(),
		},
		Versions: connection.GetCompatibleVersions(),
	}

	suite.app.IBCKeeper.ConnectionKeeper.SetConnection(suite.ctx, testConnection, connection)
}

func (suite *HandlerTestSuite) createChannel(
	portID, channnelID, connnnectionID, counterpartyPortID, counterpartyChannelID string, state channelexported.State) {
	ch := channel.Channel{
		State:    state,
		Ordering: testChannelOrder,
		Counterparty: channel.Counterparty{
			PortID:    counterpartyPortID,
			ChannelID: counterpartyChannelID,
		},
		ConnectionHops: []string{connnnectionID},
		Version:        testChannelVersion,
	}

	suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, portID, channnelID, ch)
}

func (suite *HandlerTestSuite) queryProof(key []byte) (proof commitment.Proof, height int64) {
	res := suite.app.Query(abci.RequestQuery{
		Path:  fmt.Sprintf("store/%s/key", ibctypes.StoreKey),
		Data:  key,
		Prove: true,
	})

	height = res.Height
	proof = commitment.Proof{
		Proof: res.Proof,
	}

	return
}

func (suite *HandlerTestSuite) TestHandleMsgTransfer() {
	source := true

	handler := transfer.NewHandler(suite.app.TransferKeeper)

	msg := transfer.NewMsgTransfer(testPort1, testChannel1, 10, testCoins, testAddr1, testAddr2, source)
	res, err := handler(suite.ctx, msg)
	suite.Require().Error(err)
	suite.Require().Nil(res, "%+v", res) // channel does not exist

	suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, channelexported.OPEN)
	res, err = handler(suite.ctx, msg)
	suite.Require().Error(err)
	suite.Require().Nil(res, "%+v", res) // next send sequence not found

	nextSeqSend := uint64(1)
	suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, testPort1, testChannel1, nextSeqSend)
	res, err = handler(suite.ctx, msg)
	suite.Require().Error(err)
	suite.Require().Nil(res, "%+v", res) // sender has insufficient coins

	_ = suite.app.BankKeeper.SetBalances(suite.ctx, testAddr1, testCoins)
	res, err = handler(suite.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res, "%+v", res) // successfully executed

	// test when the source is false
	source = false

	msg = transfer.NewMsgTransfer(testPort1, testChannel1, 10, testPrefixedCoins2, testAddr1, testAddr2, source)
	_ = suite.app.BankKeeper.SetBalances(suite.ctx, testAddr1, testPrefixedCoins2)

	res, err = handler(suite.ctx, msg)
	suite.Require().Error(err)
	suite.Require().Nil(res, "%+v", res) // incorrect denom prefix

	msg = transfer.NewMsgTransfer(testPort1, testChannel1, 10, testPrefixedCoins1, testAddr1, testAddr2, source)
	suite.app.SupplyKeeper.SetSupply(suite.ctx, supply.NewSupply(testPrefixedCoins1))
	_ = suite.app.BankKeeper.SetBalances(suite.ctx, testAddr1, testPrefixedCoins1)

	res, err = handler(suite.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res, "%+v", res) // successfully executed
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
