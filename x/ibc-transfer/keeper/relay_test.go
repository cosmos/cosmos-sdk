package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

func (suite *KeeperTestSuite) TestSendTransfer() {
	testCoins2 := sdk.NewCoins(sdk.NewCoin("testportid/secondchannel/atom", sdk.NewInt(100)))
	capName := host.ChannelCapabilityPath(testPort1, testChannel1)

	testCases := []struct {
		msg           string
		amount        sdk.Coins
		malleate      func()
		isSourceChain bool
		expPass       bool
	}{
		{"successful transfer from source chain", testCoins2,
			func() {
				suite.chainA.App.BankKeeper.AddCoins(suite.chainA.GetContext(), testAddr1, testCoins)
				suite.chainA.CreateClient(suite.chainB)
				suite.chainA.createConnection(testConnection, testConnection, testClientIDB, testClientIDA, connectiontypes.OPEN)
				suite.chainA.createChannel(testPort1, testChannel1, testPort2, testChannel2, channeltypes.OPEN, channeltypes.ORDERED, testConnection)
				suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.chainA.GetContext(), testPort1, testChannel1, 1)
			}, true, true},
		{"successful transfer from external chain", prefixCoins,
			func() {
				suite.chainA.App.BankKeeper.SetSupply(suite.chainA.GetContext(), banktypes.NewSupply(prefixCoins))
				_, err := suite.chainA.App.BankKeeper.AddCoins(suite.chainA.GetContext(), testAddr1, prefixCoins)
				suite.Require().NoError(err)
				suite.chainA.CreateClient(suite.chainB)
				suite.chainA.createConnection(testConnection, testConnection, testClientIDB, testClientIDA, connectiontypes.OPEN)
				suite.chainA.createChannel(testPort1, testChannel1, testPort2, testChannel2, channeltypes.OPEN, channeltypes.ORDERED, testConnection)
				suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.chainA.GetContext(), testPort1, testChannel1, 1)
			}, false, true},
		{"source channel not found", testCoins,
			func() {}, true, false},
		{"next seq send not found", testCoins,
			func() {
				suite.chainA.CreateClient(suite.chainB)
				suite.chainA.createConnection(testConnection, testConnection, testClientIDB, testClientIDA, connectiontypes.OPEN)
				suite.chainA.createChannel(testPort1, testChannel1, testPort2, testChannel2, channeltypes.OPEN, channeltypes.ORDERED, testConnection)
			}, true, false},
		// createOutgoingPacket tests
		// - source chain
		{"send coins failed", testCoins,
			func() {
				suite.chainA.CreateClient(suite.chainB)
				suite.chainA.createConnection(testConnection, testConnection, testClientIDB, testClientIDA, connectiontypes.OPEN)
				suite.chainA.createChannel(testPort1, testChannel1, testPort2, testChannel2, channeltypes.OPEN, channeltypes.ORDERED, testConnection)
				suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.chainA.GetContext(), testPort1, testChannel1, 1)
			}, true, false},
		// - receiving chain
		{"send from module account failed", testCoins,
			func() {
				suite.chainA.CreateClient(suite.chainB)
				suite.chainA.createConnection(testConnection, testConnection, testClientIDB, testClientIDA, connectiontypes.OPEN)
				suite.chainA.createChannel(testPort1, testChannel1, testPort2, testChannel2, channeltypes.OPEN, channeltypes.ORDERED, testConnection)
				suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.chainA.GetContext(), testPort1, testChannel1, 1)
			}, false, false},
		{"channel capability not found", testCoins,
			func() {
				suite.chainA.App.BankKeeper.AddCoins(suite.chainA.GetContext(), testAddr1, testCoins)
				suite.chainA.CreateClient(suite.chainB)
				suite.chainA.createConnection(testConnection, testConnection, testClientIDB, testClientIDA, connectiontypes.OPEN)
				suite.chainA.createChannel(testPort1, testChannel1, testPort2, testChannel2, channeltypes.OPEN, channeltypes.ORDERED, testConnection)
				suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.chainA.GetContext(), testPort1, testChannel1, 1)
				// Release channel capability
				cap, _ := suite.chainA.App.ScopedTransferKeeper.GetCapability(suite.chainA.GetContext(), capName)
				suite.chainA.App.ScopedTransferKeeper.ReleaseCapability(suite.chainA.GetContext(), cap)
			}, true, false},
	}

	for i, tc := range testCases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			// create channel capability from ibc scoped keeper and claim with transfer scoped keeper
			cap, err := suite.chainA.App.ScopedIBCKeeper.NewCapability(suite.chainA.GetContext(), capName)
			suite.Require().Nil(err, "could not create capability")
			err = suite.chainA.App.ScopedTransferKeeper.ClaimCapability(suite.chainA.GetContext(), cap, capName)
			suite.Require().Nil(err, "transfer module could not claim capability")

			tc.malleate()

			err = suite.chainA.App.TransferKeeper.SendTransfer(
				suite.chainA.GetContext(), testPort1, testChannel1, tc.amount, testAddr1, testAddr2.String(), 3, 110, 0,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestOnRecvPacket() {
	data := types.NewFungibleTokenPacketData(prefixCoins2, testAddr1.String(), testAddr2.String())

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"success receive from source chain",
			func() {}, true},
		// onRecvPacket
		// - source chain
		{"no dest prefix on coin denom",
			func() {
				data.Amount = testCoins
			}, false},
		{"mint failed",
			func() {
				data.Amount = prefixCoins2
				data.Amount[0].Amount = sdk.ZeroInt()
			}, false},
		// - receiving chain
		{"incorrect dest prefix on coin denom",
			func() {
				data.Amount = prefixCoins
			}, false},
		{"success receive from external chain",
			func() {
				data.Amount = prefixCoins
				escrow := types.GetEscrowAddress(testPort2, testChannel2)
				_, err := suite.chainA.App.BankKeeper.AddCoins(suite.chainA.GetContext(), escrow, testCoins)
				suite.Require().NoError(err)
			}, true},
	}

	packet := channeltypes.NewPacket(data.GetBytes(), 1, testPort1, testChannel1, testPort2, testChannel2, 3, 100, 0)

	for i, tc := range testCases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset
			tc.malleate()

			err := suite.chainA.App.TransferKeeper.OnRecvPacket(suite.chainA.GetContext(), packet, data)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

// TestOnAcknowledgementPacket tests that successful acknowledgement is a no-op
// and failure acknowledment leads to refund
func (suite *KeeperTestSuite) TestOnAcknowledgementPacket() {
	data := types.NewFungibleTokenPacketData(prefixCoins2, testAddr1.String(), testAddr2.String())

	successAck := types.FungibleTokenPacketAcknowledgement{
		Success: true,
	}
	failedAck := types.FungibleTokenPacketAcknowledgement{
		Success: false,
		Error:   "failed packet transfer",
	}

	testCases := []struct {
		msg      string
		ack      types.FungibleTokenPacketAcknowledgement
		malleate func()
		source   bool
		success  bool // success of ack
	}{
		{"success ack causes no-op", successAck,
			func() {}, true, true},
		{"successful refund from source chain", failedAck,
			func() {
				escrow := types.GetEscrowAddress(testPort1, testChannel1)
				_, err := suite.chainA.App.BankKeeper.AddCoins(suite.chainA.GetContext(), escrow, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(100))))
				suite.Require().NoError(err)
			}, true, false},
		{"successful refund from external chain", failedAck,
			func() {
				data.Amount = prefixCoins
			}, false, false},
	}

	packet := channeltypes.NewPacket(data.GetBytes(), 1, testPort1, testChannel1, testPort2, testChannel2, 3, 100, 0)

	for i, tc := range testCases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			var denom string
			if tc.source {
				prefix := types.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel())
				denom = prefixCoins2[0].Denom[len(prefix):]
			} else {
				denom = data.Amount[0].Denom
			}

			preCoin := suite.chainA.App.BankKeeper.GetBalance(suite.chainA.GetContext(), testAddr1, denom)

			err := suite.chainA.App.TransferKeeper.OnAcknowledgementPacket(suite.chainA.GetContext(), packet, data, tc.ack)
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)

			postCoin := suite.chainA.App.BankKeeper.GetBalance(suite.chainA.GetContext(), testAddr1, denom)
			deltaAmount := postCoin.Amount.Sub(preCoin.Amount)

			if tc.success {
				suite.Require().Equal(sdk.ZeroInt(), deltaAmount, "successful ack changed balance")
			} else {
				suite.Require().Equal(prefixCoins2[0].Amount, deltaAmount, "failed ack did not trigger refund")
			}
		})
	}
}

// TestOnTimeoutPacket test private refundPacket function since it is a simple wrapper over it
func (suite *KeeperTestSuite) TestOnTimeoutPacket() {
	data := types.NewFungibleTokenPacketData(prefixCoins2, testAddr1.String(), testAddr2.String())
	testCoins2 := sdk.NewCoins(sdk.NewCoin("bank/firstchannel/atom", sdk.NewInt(100)))

	testCases := []struct {
		msg      string
		malleate func()
		source   bool
		expPass  bool
	}{
		{"successful timeout from source chain",
			func() {
				escrow := types.GetEscrowAddress(testPort1, testChannel1)
				_, err := suite.chainA.App.BankKeeper.AddCoins(suite.chainA.GetContext(), escrow, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(100))))
				suite.Require().NoError(err)
			}, true, true},
		{"successful timeout from external chain",
			func() {
				data.Amount = testCoins2
			}, false, true},
		{"no source prefix on coin denom",
			func() {
				data.Amount = prefixCoins2
			}, false, false},
		{"unescrow failed",
			func() {
			}, true, false},
		{"mint failed",
			func() {
				data.Amount[0].Denom = prefixCoins2[0].Denom
				data.Amount[0].Amount = sdk.ZeroInt()
			}, true, false},
	}

	packet := channeltypes.NewPacket(data.GetBytes(), 1, testPort1, testChannel1, testPort2, testChannel2, 3, 100, 0)

	for i, tc := range testCases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			var denom string
			if tc.source {
				prefix := types.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel())
				denom = prefixCoins2[0].Denom[len(prefix):]
			} else {
				denom = data.Amount[0].Denom
			}

			preCoin := suite.chainA.App.BankKeeper.GetBalance(suite.chainA.GetContext(), testAddr1, denom)

			err := suite.chainA.App.TransferKeeper.OnTimeoutPacket(suite.chainA.GetContext(), packet, data)

			postCoin := suite.chainA.App.BankKeeper.GetBalance(suite.chainA.GetContext(), testAddr1, denom)
			deltaAmount := postCoin.Amount.Sub(preCoin.Amount)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
				suite.Require().Equal(prefixCoins2[0].Amount.Int64(), deltaAmount.Int64(), "successful timeout did not trigger refund")
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}
