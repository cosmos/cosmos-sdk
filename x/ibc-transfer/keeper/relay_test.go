package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

// test sending from chainA to chainB using both coins that orignate on this
// chain and that came from chainB
func (suite *KeeperTestSuite) TestSendTransfer() {
	var (
		amount             sdk.Coins
		channelA, channelB ibctesting.TestChannel
		err                error
	)

	testCases := []struct {
		msg           string
		malleate      func()
		isSourceChain bool
		expPass       bool
	}{
		{"successful transfer from source chain",
			func() {
				_, _, _, _, channelA, channelB = suite.coordinator.Setup(suite.chainA, suite.chainB)
				amount = ibctesting.NewTransferCoins(channelB, sdk.DefaultBondDenom, 100)
			}, true, true},
		{"successful transfer with coins from counterparty chain",
			func() {
				// send coins from chainA back to chainB
				_, _, _, _, channelA, channelB = suite.coordinator.Setup(suite.chainA, suite.chainB)
				amount = ibctesting.NewTransferCoins(channelA, sdk.DefaultBondDenom, 100)
			}, false, true},
		{"source channel not found",
			func() {
				// channel references wrong ID
				_, _, _, _, channelA, channelB = suite.coordinator.Setup(suite.chainA, suite.chainB)
				channelA.ID = ibctesting.InvalidID
				amount = ibctesting.NewTransferCoins(channelB, sdk.DefaultBondDenom, 100)
			}, true, false},
		{"next seq send not found",
			func() {
				_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
				channelA = connA.NextTestChannel()
				channelB = connB.NextTestChannel()
				// manually create channel so next seq send is never set
				suite.chainA.App.IBCKeeper.ChannelKeeper.SetChannel(
					suite.chainA.GetContext(),
					channelA.PortID, channelA.ID,
					channeltypes.NewChannel(channeltypes.OPEN, channeltypes.ORDERED, channeltypes.NewCounterparty(channelB.PortID, channelB.ID), []string{connA.ID}, ibctesting.ChannelVersion),
				)
				suite.chainA.CreateChannelCapability(channelA.PortID, channelA.ID)
				amount = ibctesting.NewTransferCoins(channelB, sdk.DefaultBondDenom, 100)
			}, true, false},

		// createOutgoingPacket tests
		// - source chain
		{"send coins failed",
			func() {
				_, _, _, _, channelA, channelB = suite.coordinator.Setup(suite.chainA, suite.chainB)
				amount = ibctesting.NewTransferCoins(channelB, "randomdenom", 100)
			}, true, false},
		// - receiving chain
		{"send from module account failed",
			func() {
				_, _, _, _, channelA, channelB = suite.coordinator.Setup(suite.chainA, suite.chainB)
				amount = ibctesting.NewTransferCoins(channelA, "randomdenom", 100)
			}, false, false},
		{"channel capability not found",
			func() {
				_, _, _, _, channelA, channelB = suite.coordinator.Setup(suite.chainA, suite.chainB)
				cap := suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)

				// Release channel capability
				suite.chainA.App.ScopedTransferKeeper.ReleaseCapability(suite.chainA.GetContext(), cap)
			}, true, false},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			if tc.isSourceChain {
				// use channelA for source
				err = suite.chainA.App.TransferKeeper.SendTransfer(
					suite.chainA.GetContext(), channelA.PortID, channelA.ID, amount,
					suite.chainA.SenderAccount.GetAddress(), suite.chainB.SenderAccount.GetAddress().String(), 110, 0,
				)
			} else {
				// send coins from chainB to chainA
				coinFromBToA := ibctesting.NewTransferCoins(channelA, sdk.DefaultBondDenom, 100)
				transferMsg := types.NewMsgTransfer(channelB.PortID, channelB.ID, coinFromBToA, suite.chainB.SenderAccount.GetAddress(), suite.chainA.SenderAccount.GetAddress().String(), 110, 0)
				err = suite.coordinator.SendMsgs(suite.chainB, suite.chainA, channelA.ClientID, []sdk.Msg{transferMsg})
				suite.Require().NoError(err) // message committed

				// TODO: retreive packet sequence from the resulting events in the commit above

				// receive coins on chainA from chainB
				fungibleTokenPacket := types.NewFungibleTokenPacketData(coinFromBToA, suite.chainB.SenderAccount.GetAddress().String(), suite.chainA.SenderAccount.GetAddress().String())
				packet := channeltypes.NewPacket(fungibleTokenPacket.GetBytes(), 1, channelB.PortID, channelB.ID, channelA.PortID, channelA.ID, 110, 0)

				// get proof of packet commitment from chainB
				packetKey := host.KeyPacketCommitment(packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
				proof, proofHeight := suite.chainB.QueryProof(packetKey)

				recvMsg := channeltypes.NewMsgRecvPacket(packet, proof, proofHeight, suite.chainA.SenderAccount.GetAddress())
				err = suite.coordinator.SendMsgs(suite.chainA, suite.chainB, channelB.ClientID, []sdk.Msg{recvMsg})
				suite.Require().NoError(err) // message committed

				// use channelB for source
				err = suite.chainA.App.TransferKeeper.SendTransfer(
					suite.chainA.GetContext(), channelA.PortID, channelA.ID, amount,
					suite.chainA.SenderAccount.GetAddress(), suite.chainB.SenderAccount.GetAddress().String(), 110, 0,
				)
			}

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

/*
func (suite *KeeperTestSuite) TestOnRecvPacket() {
	_, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)

	prefixCoins = types.GetPrefixedCoins(channelA.PortID, channelA.ID, sdk.NewInt64Coin("atom", 100))
	sourceCoins := types.GetPrefixedCoins(channelB.PortID, channelB.ID, sdk.NewInt64Coin("atom", 100))
	data := types.NewFungibleTokenPacketData(prefixCoins, suite.sender.String(), suite.receiver.String())

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		// {"success receive from source chain",
		// 	func() {}, true},
		{
			"empty amount",
			func() {
				data.Amount = nil
			},
			false,
		},
		{
			"invalid receiver address",
			func() {
				data.Amount = prefixCoins
				data.Receiver = "gaia1scqhwpgsmr6vmztaa7suurfl52my6nd2kmrudl"
			},
			false,
		},
		{
			"no dest prefix on coin denom",
			func() {
				data.Amount = testCoins
				data.Receiver = suite.receiver.String()
			},
			false,
		},
		// onRecvPacket
		// - source chain
		{
			"mint failed",
			func() {
				data.Amount = sourceCoins
				data.Amount[0].Amount = sdk.ZeroInt()
			},
			false,
		},
		// {
		// 	"success receive",
		// 	func() {
		// 		data.Amount = sourceCoins
		// 	},
		// 	true,
		// },
		// // - receiving chain
		// {"incorrect dest prefix on coin denom",
		// 	func() {
		// 		data.Amount = prefixCoins
		// 	}, false},
		// {"success receive from external chain",
		// 	func() {
		// 		data.Amount = prefixCoins
		// 		escrow := types.GetEscrowAddress(testPort2, testChannel2)
		// 		_, err := suite.chainA.App.BankKeeper.AddCoins(suite.chainA.GetContext(), escrow, testCoins)
		// 		suite.Require().NoError(err)
		// 	}, true},
	}

	packet := channeltypes.NewPacket(data.GetBytes(), 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, 100, 0)

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

// // TestOnAcknowledgementPacket tests that successful acknowledgement is a no-op
// // and failure acknowledment leads to refund
// func (suite *KeeperTestSuite) TestOnAcknowledgementPacket() {
// 	data := types.NewFungibleTokenPacketData(prefixCoins2, testAddr1.String(), testAddr2.String())

// 	successAck := types.FungibleTokenPacketAcknowledgement{
// 		Success: true,
// 	}
// 	failedAck := types.FungibleTokenPacketAcknowledgement{
// 		Success: false,
// 		Error:   "failed packet transfer",
// 	}

// 	testCases := []struct {
// 		msg      string
// 		ack      types.FungibleTokenPacketAcknowledgement
// 		malleate func()
// 		source   bool
// 		success  bool // success of ack
// 	}{
// 		{"success ack causes no-op", successAck,
// 			func() {}, true, true},
// 		{"successful refund from source chain", failedAck,
// 			func() {
// 				escrow := types.GetEscrowAddress(testPort1, testChannel1)
// 				_, err := suite.chainA.App.BankKeeper.AddCoins(suite.chainA.GetContext(), escrow, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(100))))
// 				suite.Require().NoError(err)
// 			}, true, false},
// 		{"successful refund from external chain", failedAck,
// 			func() {
// 				data.Amount = prefixCoins
// 			}, false, false},
// 	}

// 	packet := channeltypes.NewPacket(data.GetBytes(), 1, testPort1, testChannel1, testPort2, testChannel2, 100, 0)

// 	for i, tc := range testCases {
// 		tc := tc
// 		i := i
// 		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
// 			suite.SetupTest() // reset

// 			tc.malleate()

// 			var denom string
// 			if tc.source {
// 				prefix := types.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel())
// 				denom = prefixCoins2[0].Denom[len(prefix):]
// 			} else {
// 				denom = data.Amount[0].Denom
// 			}

// 			preCoin := suite.chainA.App.BankKeeper.GetBalance(suite.chainA.GetContext(), testAddr1, denom)

// 			err := suite.chainA.App.TransferKeeper.OnAcknowledgementPacket(suite.chainA.GetContext(), packet, data, tc.ack)
// 			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)

// 			postCoin := suite.chainA.App.BankKeeper.GetBalance(suite.chainA.GetContext(), testAddr1, denom)
// 			deltaAmount := postCoin.Amount.Sub(preCoin.Amount)

// 			if tc.success {
// 				suite.Require().Equal(sdk.ZeroInt(), deltaAmount, "successful ack changed balance")
// 			} else {
// 				suite.Require().Equal(prefixCoins2[0].Amount, deltaAmount, "failed ack did not trigger refund")
// 			}
// 		})
// 	}
// }

// // TestOnTimeoutPacket test private refundPacket function since it is a simple wrapper over it
// func (suite *KeeperTestSuite) TestOnTimeoutPacket() {
// 	data := types.NewFungibleTokenPacketData(prefixCoins2, testAddr1.String(), testAddr2.String())
// 	testCoins2 := sdk.NewCoins(sdk.NewCoin("bank/firstchannel/atom", sdk.NewInt(100)))

// 	testCases := []struct {
// 		msg      string
// 		malleate func()
// 		source   bool
// 		expPass  bool
// 	}{
// 		{"successful timeout from source chain",
// 			func() {
// 				escrow := types.GetEscrowAddress(testPort1, testChannel1)
// 				_, err := suite.chainA.App.BankKeeper.AddCoins(suite.chainA.GetContext(), escrow, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(100))))
// 				suite.Require().NoError(err)
// 			}, true, true},
// 		{"successful timeout from external chain",
// 			func() {
// 				data.Amount = testCoins2
// 			}, false, true},
// 		{"no source prefix on coin denom",
// 			func() {
// 				data.Amount = prefixCoins2
// 			}, false, false},
// 		{"unescrow failed",
// 			func() {
// 			}, true, false},
// 		{"mint failed",
// 			func() {
// 				data.Amount[0].Denom = prefixCoins2[0].Denom
// 				data.Amount[0].Amount = sdk.ZeroInt()
// 			}, true, false},
// 	}

// 	packet := channeltypes.NewPacket(data.GetBytes(), 1, testPort1, testChannel1, testPort2, testChannel2, 100, 0)

// 	for i, tc := range testCases {
// 		tc := tc
// 		i := i
// 		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
// 			suite.SetupTest() // reset

// 			tc.malleate()

// 			var denom string
// 			if tc.source {
// 				prefix := types.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel())
// 				denom = prefixCoins2[0].Denom[len(prefix):]
// 			} else {
// 				denom = data.Amount[0].Denom
// 			}

// 			preCoin := suite.chainA.App.BankKeeper.GetBalance(suite.chainA.GetContext(), testAddr1, denom)

// 			err := suite.chainA.App.TransferKeeper.OnTimeoutPacket(suite.chainA.GetContext(), packet, data)

// 			postCoin := suite.chainA.App.BankKeeper.GetBalance(suite.chainA.GetContext(), testAddr1, denom)
// 			deltaAmount := postCoin.Amount.Sub(preCoin.Amount)

// 			if tc.expPass {
// 				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
// 				suite.Require().Equal(prefixCoins2[0].Amount.Int64(), deltaAmount.Int64(), "successful timeout did not trigger refund")
// 			} else {
// 				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
// 			}
// 		})
// 	}
// }
*/
