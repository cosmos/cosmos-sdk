package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

func (suite *KeeperTestSuite) TestSendTransfer() {
	testCases := []struct {
		msg           string
		amount        sdk.Coins
		isSourceChain bool
		malleate      func()
		expPass       bool
	}{
		{"sucess transfer from source chain", testCoins,
			true, func() {
				suite.app.BankKeeper.AddCoins(suite.ctx, testAddr1, testCoins)
				suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, channelexported.OPEN)
				suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, testPort1, testChannel1, 1)
			}, true},
		{"sucess transfer from source chain with denom prefix", destCoins2,
			true, func() {
				suite.app.BankKeeper.AddCoins(suite.ctx, testAddr1, testCoins)
				suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, channelexported.OPEN)
				suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, testPort1, testChannel1, 1)
			}, true},
		{"sucess transfer from external chain", testCoins,
			false, func() {
				suite.app.SupplyKeeper.SetSupply(suite.ctx, supply.NewSupply(destCoins))
				suite.app.BankKeeper.AddCoins(suite.ctx, testAddr1, destCoins)
				suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, channelexported.OPEN)
				suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, testPort1, testChannel1, 1)
			}, true},
		{"source channel not found", testCoins,
			true, func() {}, false},
		{"next seq send not found", testCoins,
			true, func() {
				suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, channelexported.OPEN)
			}, false},
		// createOutgoingPacket tests
		// - source chain
		{"send coins failed", testCoins,
			true, func() {
				suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, channelexported.OPEN)
				suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, testPort1, testChannel1, 1)
			}, false},
		// - receiving chain
		{"send from module account dailed", testCoins,
			false, func() {
				suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, channelexported.OPEN)
				suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, testPort1, testChannel1, 1)
			}, false},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			err := suite.app.TransferKeeper.SendTransfer(
				suite.ctx, testPort1, testChannel1, 100, tc.amount, testAddr1, testAddr2, tc.isSourceChain,
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
	var packet channeltypes.Packet
	var data types.FungibleTokenPacketData
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset
			tc.malleate()

			err := suite.app.TransferKeeper.ReceiveTransfer(suite.ctx, packet, data)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}
