package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
)

func (suite *KeeperTestSuite) TestSendTransfer() {
	testCases := []struct {
		msg           string
		sourcePort    string
		sourceChannel string
		amount        sdk.Coins
		sender        sdk.AccAddress
		receiver      sdk.AccAddress
		isSourceChain bool
		malleate      func()
		expPass       bool
	}{
		// {"sucess transfer from source chain", testPort1, testChannel1, testCoins, testAddr1, testAddr2,
		// 	true, func() {}, true},
		// {"sucess transfer from external chain", testPort1, testChannel1, testCoins, testAddr1, testAddr2,
		// 	true, func() {}, true},
		{"source channel not found", testPort1, testChannel1, testCoins, testAddr1, testAddr2,
			true, func() {}, false},
		{"next seq send not found", testPort1, testChannel1, testCoins, testAddr1, testAddr2,
			true, func() {
				suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, channelexported.OPEN)
			}, false},
		// createOutgoingPacket tests
		// - source chain
		{"no prefix on transfer amount", testPort1, testChannel1, testCoins, testAddr1, testAddr2,
			true, func() {
				suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, channelexported.OPEN)
				suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, testPort1, testChannel1, 1)
			}, false},
		{"send coins failed", testPort1, testChannel1, testCoins, testAddr1, testAddr2,
			true, func() {
				suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, channelexported.OPEN)
				suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, testPort1, testChannel1, 1)
			}, true},
		// - receiving chain
		// {"no prefix on transfer amount", testPort1, testChannel1, testCoins, testAddr1, testAddr2,
		// 	false, func() {}, false},
		// {"send from module account dailed", testPort1, testChannel1, testCoins, testAddr1, testAddr2,
		// 	false, func() {}, false},
		// {"tokens burn failed", testPort1, testChannel1, testCoins, testAddr1, testAddr2,
		// 	false, func() {}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			err := suite.app.TransferKeeper.SendTransfer(
				suite.ctx, tc.sourcePort, tc.sourceChannel, 100, tc.amount, tc.sender, tc.receiver, tc.isSourceChain,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestReceiveTransfer() {
	// test the situation where the source is true
	source := true
	packetTimeout := uint64(100)

	packetData := types.NewFungibleTokenPacketData(testPrefixedCoins1, testAddr1, testAddr2, source, packetTimeout)
	err := suite.app.TransferKeeper.ReceiveTransfer(suite.ctx, testPort1, testChannel1, testPort2, testChannel2, packetData)
	suite.Error(err) // incorrect denom prefix

	packetData.Amount = testPrefixedCoins2
	err = suite.app.TransferKeeper.ReceiveTransfer(suite.ctx, testPort1, testChannel1, testPort2, testChannel2, packetData)
	suite.NoError(err) // successfully executed

	totalSupply := suite.app.SupplyKeeper.GetSupply(suite.ctx)
	suite.Equal(testPrefixedCoins2, totalSupply.GetTotal()) // supply should be inflated

	receiverCoins := suite.app.BankKeeper.GetAllBalances(suite.ctx, packetData.Receiver)
	suite.Equal(testPrefixedCoins2, receiverCoins)

	// test the situation where the source is false
	packetData.Source = false

	packetData.Amount = testPrefixedCoins2
	err = suite.app.TransferKeeper.ReceiveTransfer(suite.ctx, testPort1, testChannel1, testPort2, testChannel2, packetData)
	suite.Error(err) // incorrect denom prefix

	packetData.Amount = testPrefixedCoins1
	err = suite.app.TransferKeeper.ReceiveTransfer(suite.ctx, testPort1, testChannel1, testPort2, testChannel2, packetData)
	suite.Error(err) // insufficient coins in the corresponding escrow account

	escrowAddress := types.GetEscrowAddress(testPort2, testChannel2)
	_ = suite.app.BankKeeper.SetBalances(suite.ctx, escrowAddress, testCoins)
	_ = suite.app.BankKeeper.SetBalances(suite.ctx, packetData.Receiver, sdk.Coins{})
	err = suite.app.TransferKeeper.ReceiveTransfer(suite.ctx, testPort1, testChannel1, testPort2, testChannel2, packetData)
	suite.NoError(err) // successfully executed

	escrowCoins := suite.app.BankKeeper.GetAllBalances(suite.ctx, escrowAddress)
	suite.Equal(sdk.Coins(nil), escrowCoins)

	receiverCoins = suite.app.BankKeeper.GetAllBalances(suite.ctx, packetData.Receiver)
	suite.Equal(testCoins, receiverCoins)
}
