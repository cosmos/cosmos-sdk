package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

// test sending from chainA to chainB using both coin that orignate on
// chainA and coin that orignate on chainB
func (suite *KeeperTestSuite) TestSendTransfer() {
	var (
		amount             sdk.Coin
		channelA, channelB ibctesting.TestChannel
		err                error
	)

	testCases := []struct {
		msg            string
		malleate       func()
		sendFromSource bool
		expPass        bool
	}{
		{"successful transfer from source chain",
			func() {
				_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
				channelA, channelB = suite.coordinator.CreateTransferChannels(suite.chainA, suite.chainB, connA, connB, channeltypes.UNORDERED)
				amount = sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))
			}, true, true},
		{"successful transfer with coin from counterparty chain",
			func() {
				// send coin from chainA back to chainB
				_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
				channelA, channelB = suite.coordinator.CreateTransferChannels(suite.chainA, suite.chainB, connA, connB, channeltypes.UNORDERED)
				amount = types.GetTransferCoin(channelA.PortID, channelA.ID, sdk.DefaultBondDenom, 100)
			}, false, true},
		{"source channel not found",
			func() {
				// channel references wrong ID
				_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
				channelA, channelB = suite.coordinator.CreateTransferChannels(suite.chainA, suite.chainB, connA, connB, channeltypes.UNORDERED)
				channelA.ID = ibctesting.InvalidID
				amount = sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))
			}, true, false},
		{"next seq send not found",
			func() {
				_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
				channelA = connA.NextTestChannel(ibctesting.TransferPort)
				channelB = connB.NextTestChannel(ibctesting.TransferPort)
				// manually create channel so next seq send is never set
				suite.chainA.App.IBCKeeper.ChannelKeeper.SetChannel(
					suite.chainA.GetContext(),
					channelA.PortID, channelA.ID,
					channeltypes.NewChannel(channeltypes.OPEN, channeltypes.ORDERED, channeltypes.NewCounterparty(channelB.PortID, channelB.ID), []string{connA.ID}, ibctesting.DefaultChannelVersion),
				)
				suite.chainA.CreateChannelCapability(channelA.PortID, channelA.ID)
				amount = sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))
			}, true, false},

		// createOutgoingPacket tests
		// - source chain
		{"send coin failed",
			func() {
				_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
				channelA, channelB = suite.coordinator.CreateTransferChannels(suite.chainA, suite.chainB, connA, connB, channeltypes.UNORDERED)
				amount = sdk.NewCoin("randomdenom", sdk.NewInt(100))
			}, true, false},
		// - receiving chain
		{"send from module account failed",
			func() {
				_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
				channelA, channelB = suite.coordinator.CreateTransferChannels(suite.chainA, suite.chainB, connA, connB, channeltypes.UNORDERED)
				amount = types.GetTransferCoin(channelA.PortID, channelA.ID, " randomdenom", 100)
			}, false, false},
		{"channel capability not found",
			func() {
				_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
				channelA, channelB = suite.coordinator.CreateTransferChannels(suite.chainA, suite.chainB, connA, connB, channeltypes.UNORDERED)
				cap := suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)

				// Release channel capability
				suite.chainA.App.ScopedTransferKeeper.ReleaseCapability(suite.chainA.GetContext(), cap)
				amount = sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))
			}, true, false},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			if !tc.sendFromSource {
				// send coin from chainB to chainA
				coinFromBToA := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))
				transferMsg := types.NewMsgTransfer(channelB.PortID, channelB.ID, coinFromBToA, suite.chainB.SenderAccount.GetAddress(), suite.chainA.SenderAccount.GetAddress().String(), clienttypes.NewHeight(0, 110), 0)
				err = suite.coordinator.SendMsg(suite.chainB, suite.chainA, channelA.ClientID, transferMsg)
				suite.Require().NoError(err) // message committed

				// receive coin on chainA from chainB
				fungibleTokenPacket := types.NewFungibleTokenPacketData(coinFromBToA.Denom, coinFromBToA.Amount.Uint64(), suite.chainB.SenderAccount.GetAddress().String(), suite.chainA.SenderAccount.GetAddress().String())
				packet := channeltypes.NewPacket(fungibleTokenPacket.GetBytes(), 1, channelB.PortID, channelB.ID, channelA.PortID, channelA.ID, clienttypes.NewHeight(0, 110), 0)

				// get proof of packet commitment from chainB
				packetKey := host.KeyPacketCommitment(packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
				proof, proofHeight := suite.chainB.QueryProof(packetKey)

				recvMsg := channeltypes.NewMsgRecvPacket(packet, proof, proofHeight, suite.chainA.SenderAccount.GetAddress())
				err = suite.coordinator.SendMsg(suite.chainA, suite.chainB, channelB.ClientID, recvMsg)
				suite.Require().NoError(err) // message committed
			}

			err = suite.chainA.App.TransferKeeper.SendTransfer(
				suite.chainA.GetContext(), channelA.PortID, channelA.ID, amount,
				suite.chainA.SenderAccount.GetAddress(), suite.chainB.SenderAccount.GetAddress().String(), clienttypes.NewHeight(0, 110), 0,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// test receiving coin on chainB with coin that orignate on chainA and
// coin that orignated on chainB (source). The bulk of the testing occurs
// in the test case for loop since setup is intensive for all cases. The
// malleate function allows for testing invalid cases.
func (suite *KeeperTestSuite) TestOnRecvPacket() {
	var (
		channelA, channelB ibctesting.TestChannel
		trace              types.DenomTrace
		amount             sdk.Int
		receiver           string
	)

	testCases := []struct {
		msg          string
		malleate     func()
		recvIsSource bool // the receiving chain is the source of the coin originally
		expPass      bool
	}{
		{"success receive on source chain", func() {}, true, true},
		{"success receive with coin from another chain as source", func() {}, false, true},
		{"empty coin", func() {
			trace = types.DenomTrace{}
			amount = sdk.ZeroInt()
		}, true, false},
		{"invalid receiver address", func() {
			receiver = "gaia1scqhwpgsmr6vmztaa7suurfl52my6nd2kmrudl"
		}, true, false},

		// onRecvPacket
		// - coin from chain chainA
		{"failure: mint zero coin", func() {
			amount = sdk.ZeroInt()
		}, false, false},

		// - coin being sent back to original chain (chainB)
		{"tries to unescrow more tokens than allowed", func() {
			amount = sdk.NewInt(1000000)
		}, true, false},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			clientA, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			channelA, channelB = suite.coordinator.CreateTransferChannels(suite.chainA, suite.chainB, connA, connB, channeltypes.UNORDERED)
			receiver = suite.chainB.SenderAccount.GetAddress().String() // must be explicitly changed in malleate

			amount = sdk.NewInt(100) // must be explicitly changed in malleate
			seq := uint64(1)

			if tc.recvIsSource {
				// send coin from chainB to chainA, receive them, acknowledge them, and send back to chainB
				coinFromBToA := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))
				transferMsg := types.NewMsgTransfer(channelB.PortID, channelB.ID, coinFromBToA, suite.chainB.SenderAccount.GetAddress(), suite.chainA.SenderAccount.GetAddress().String(), clienttypes.NewHeight(0, 110), 0)
				err := suite.coordinator.SendMsg(suite.chainB, suite.chainA, channelA.ClientID, transferMsg)
				suite.Require().NoError(err) // message committed

				// relay send packet
				fungibleTokenPacket := types.NewFungibleTokenPacketData(coinFromBToA.Denom, coinFromBToA.Amount.Uint64(), suite.chainB.SenderAccount.GetAddress().String(), suite.chainA.SenderAccount.GetAddress().String())
				packet := channeltypes.NewPacket(fungibleTokenPacket.GetBytes(), 1, channelB.PortID, channelB.ID, channelA.PortID, channelA.ID, clienttypes.NewHeight(0, 110), 0)
				ack := channeltypes.NewResultAcknowledgement([]byte{byte(1)})
				err = suite.coordinator.RelayPacket(suite.chainB, suite.chainA, clientB, clientA, packet, ack.GetBytes())
				suite.Require().NoError(err) // relay committed

				seq++

				// NOTE: trace must be explicitly changed in malleate to test invalid cases
				trace = types.ParseDenomTrace(types.GetPrefixedDenom(channelA.PortID, channelA.ID, sdk.DefaultBondDenom))
			} else {
				trace = types.ParseDenomTrace(sdk.DefaultBondDenom)
			}

			// send coin from chainA to chainB
			transferMsg := types.NewMsgTransfer(channelA.PortID, channelA.ID, sdk.NewCoin(trace.IBCDenom(), amount), suite.chainA.SenderAccount.GetAddress(), receiver, clienttypes.NewHeight(0, 110), 0)
			err := suite.coordinator.SendMsg(suite.chainA, suite.chainB, channelB.ClientID, transferMsg)
			suite.Require().NoError(err) // message committed

			tc.malleate()

			data := types.NewFungibleTokenPacketData(trace.GetFullDenomPath(), amount.Uint64(), suite.chainA.SenderAccount.GetAddress().String(), receiver)
			packet := channeltypes.NewPacket(data.GetBytes(), seq, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, clienttypes.NewHeight(0, 100), 0)

			err = suite.chainB.App.TransferKeeper.OnRecvPacket(suite.chainB.GetContext(), packet, data)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

/*
// TestOnAcknowledgementPacket tests that successful acknowledgement is a no-op
// and failure acknowledment leads to refund when attempting to send from chainA
// to chainB.
func (suite *KeeperTestSuite) TestOnAcknowledgementPacket() {
	var (
		successAck = channeltypes.NewResultAcknowledgement([]byte{byte(1)})
		failedAck  = channeltypes.NewErrorAcknowledgement("failed packet transfer")

		channelA, channelB ibctesting.TestChannel
		coin               sdk.Coin
	)

	testCases := []struct {
		msg      string
		ack      channeltypes.Acknowledgement
		malleate func()
		source   bool
		success  bool // success of ack
	}{
		{"success ack causes no-op", successAck, func() {
			coin = types.GetPrefixedCoin(channelB.PortID, channelB.ID, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)))
		}, true, true},
		{"successful refund from source chain", failedAck,
			func() {
				escrow := types.GetEscrowAddress(channelA.PortID, channelA.ID)
				_, err := suite.chainA.App.BankKeeper.AddCoins(suite.chainA.GetContext(), escrow, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))))
				suite.Require().NoError(err)

				coin = types.GetPrefixedCoin(channelB.PortID, channelB.ID, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)))
			}, true, false},
		{"successful refund from external chain", failedAck,
			func() {
				coin = types.GetPrefixedCoin(channelA.PortID, channelA.ID, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)))
			}, false, false},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset
			_, _, _, _, channelA, channelB = suite.coordinator.Setup(suite.chainA, suite.chainB, channeltypes.UNORDERED)

			tc.malleate()

			data := types.NewFungibleTokenPacketData(coin.Denom, coin.Amount.Uint64(), suite.chainA.SenderAccount.GetAddress().String(), suite.chainB.SenderAccount.GetAddress().String())
			packet := channeltypes.NewPacket(data.GetBytes(), 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, clienttypes.NewHeight(0, 100), 0)

			var denom string
			if tc.source {
				// peel off ibc denom if returning back to original chain
				denom = sdk.DefaultBondDenom
			} else {
				denom = coin.Denom
			}

			preCoin := suite.chainA.App.BankKeeper.GetBalance(suite.chainA.GetContext(), suite.chainA.SenderAccount.GetAddress(), denom)

			err := suite.chainA.App.TransferKeeper.OnAcknowledgementPacket(suite.chainA.GetContext(), packet, data, tc.ack)
			suite.Require().NoError(err)

			postCoin := suite.chainA.App.BankKeeper.GetBalance(suite.chainA.GetContext(), suite.chainA.SenderAccount.GetAddress(), denom)
			deltaAmount := postCoin.Amount.Sub(preCoin.Amount)

			if tc.success {
				suite.Require().Equal(int64(0), deltaAmount.Int64(), "successful ack changed balance")
			} else {
				suite.Require().Equal(coin.Amount, deltaAmount, "failed ack did not trigger refund")
			}
		})
	}
}

// TestOnTimeoutPacket test private refundPacket function since it is a simple
// wrapper over it. The actual timeout does not matter since IBC core logic
// is not being tested. The test is timing out a send from chainA to chainB
// so the refunds are occurring on chainA.
func (suite *KeeperTestSuite) TestOnTimeoutPacket() {
	var (
		channelA, channelB ibctesting.TestChannel
		coin               sdk.Coin
	)

	testCases := []struct {
		msg      string
		malleate func()
		source   bool
		expPass  bool
	}{
		{"successful timeout from source chain",
			func() {
				escrow := types.GetEscrowAddress(channelA.PortID, channelA.ID)
				_, err := suite.chainA.App.BankKeeper.AddCoins(suite.chainA.GetContext(), escrow, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))))
				suite.Require().NoError(err)

				coin = types.GetPrefixedCoin(channelB.PortID, channelB.ID, sdk.DefaultBondDenom, 100)
			}, true, true},
		{"successful timeout from external chain",
			func() {
				coin = types.GetPrefixedCoin(channelA.PortID, channelA.ID, sdk.DefaultBondDenom, 100)
			}, false, true},
		{"no source prefix on coin denom",
			func() {
				coin = sdk.NewCoin("bitcoin", sdk.NewInt(100))
			}, false, false},
		{"unescrow failed",
			func() {
				coin = types.GetPrefixedCoin(channelB.PortID, channelB.ID, sdk.DefaultBondDenom, 100)
			}, true, false},
		{"mint failed",
			func() {
				coin = types.GetPrefixedCoin(channelA.PortID, channelA.ID, sdk.DefaultBondDenom, 0)
			}, true, false},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset
			_, _, connA, connB := suite.coordinator.SetupClientsConnections(suite.chainA, suite.chainB, exported.Tendermint)
			channelA, channelB = suite.coordinator.CreateTransferChannels(suite.chainA, suite.chainB, connA, connB, channeltypes.UNORDERED)

			tc.malleate()

			data := types.NewFungibleTokenPacketData(coin.Denom, coin.Amount.Uint64(), suite.chainA.SenderAccount.GetAddress().String(), suite.chainB.SenderAccount.GetAddress().String())
			packet := channeltypes.NewPacket(data.GetBytes(), 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, clienttypes.NewHeight(0, 100), 0)

			var denom string
			if tc.source {
				denom = sdk.DefaultBondDenom
			} else {
				denom = coin.Denom
			}

			preCoin := suite.chainA.App.BankKeeper.GetBalance(suite.chainA.GetContext(), suite.chainA.SenderAccount.GetAddress(), denom)

			err := suite.chainA.App.TransferKeeper.OnTimeoutPacket(suite.chainA.GetContext(), packet, data)

			postCoin := suite.chainA.App.BankKeeper.GetBalance(suite.chainA.GetContext(), suite.chainA.SenderAccount.GetAddress(), denom)
			deltaAmount := postCoin.Amount.Sub(preCoin.Amount)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(coin.Amount.Int64(), deltaAmount.Int64(), "successful timeout did not trigger refund")
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
*/
