package transfer_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

// define variables used for testing
var (
	testCoins, _          = sdk.ParseCoins("100atom")
	testPrefixedCoins1, _ = sdk.ParseCoins(fmt.Sprintf("100%satom", types.GetDenomPrefix(testPort1, testChannel1)))
	testPrefixedCoins2, _ = sdk.ParseCoins(fmt.Sprintf("100%satom", types.GetDenomPrefix(testPort2, testChannel2)))
)

type HandlerTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
}

func (suite *HandlerTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))
}

func (suite *HandlerTestSuite) TestHandleMsgTransfer() {
	_, _, connectionA, connectiobB, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)

	msg := types.NewMsgTransfer(testPort1, testChannel1, testPrefixedCoins2, testAddr1, testAddr2.String(), 110, 0)
	res, err := handler(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Nil(res, "%+v", res) // channel does not exist

	// Setup channel from A to B
	suite.chainA.CreateClient(suite.chainB)
	suite.chainA.createConnection(testConnection, testConnection, testClientIDB, testClientIDA, connectiontypes.OPEN)
	suite.chainA.createChannel(testPort1, testChannel1, testPort2, testChannel2, channeltypes.OPEN, channeltypes.ORDERED, testConnection)

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
	msg = types.NewMsgTransfer(testPort1, testChannel1, testPrefixedCoins2, testAddr1, testAddr2.String(), 110, 0)
	_ = suite.chainA.App.BankKeeper.SetBalances(ctx, testAddr1, testPrefixedCoins2)

	res, err = handler(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Nil(res, "%+v", res) // incorrect denom prefix

	msg = types.NewMsgTransfer(testPort1, testChannel1, testPrefixedCoins1, testAddr1, testAddr2.String(), 110, 0)
	suite.chainA.App.BankKeeper.SetSupply(ctx, banktypes.NewSupply(testPrefixedCoins1))
	_ = suite.chainA.App.BankKeeper.SetBalances(ctx, testAddr1, testPrefixedCoins1)

	res, err = handler(ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res, "%+v", res) // successfully executed
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
