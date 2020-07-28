package transfer_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfer "github.com/cosmos/cosmos-sdk/x/ibc-transfer"
	"github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
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

// constructs a send from chainA to chainB on the established channel/connection
// and sends the coins back from chainB to chainA.
// FIX: this test currently passes because source is incorrectly determined
// by the ibc-transfer module, so what actually occurs is chainA and chainB
// send coins to each other, but no coins are ever sent back. This can be
// fixed by receving and acknowledeging the send on the counterparty chain.
// https://github.com/cosmos/cosmos-sdk/issues/6827
func (suite *HandlerTestSuite) TestHandleMsgTransfer() {
	clientA, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)
	handlerA := ibctransfer.NewHandler(suite.chainA.App.TransferKeeper)

	coinToSendToB := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))

	// send from chainA to chainB
	msg := types.NewMsgTransfer(channelA.PortID, channelA.ID, coinToSendToB, suite.chainA.SenderAccount.GetAddress(), suite.chainB.SenderAccount.GetAddress().String(), true, 110, 0)

	res, err := handlerA(suite.chainA.GetContext(), msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res, "%+v", res) // successfully executed

	err = suite.coordinator.SendMsgs(suite.chainA, suite.chainB, clientB, msg)
	suite.Require().NoError(err) // message committed

	handlerB := ibctransfer.NewHandler(suite.chainB.App.TransferKeeper)

	coinToSendBackToA := sdk.NewCoin(fmt.Sprintf("%s/%s/%s", channelB.PortID, channelB.ID, sdk.DefaultBondDenom), sdk.NewInt(100))

	// send from chainB back to chainA
	msg = types.NewMsgTransfer(channelA.PortID, channelA.ID, coinToSendBackToA, suite.chainB.SenderAccount.GetAddress(), suite.chainA.SenderAccount.GetAddress().String(), false, 110, 0)

	res, err = handlerB(suite.chainB.GetContext(), msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res, "%+v", res) // successfully executed

	err = suite.coordinator.SendMsgs(suite.chainB, suite.chainA, clientA, msg)
	suite.Require().NoError(err) // message committed
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
