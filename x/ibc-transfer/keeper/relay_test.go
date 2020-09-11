package keeper_test

// TODO: these tests are blocked on multiple bugs in the existing ibc-transfer
// code
// - https://github.com/cosmos/cosmos-sdk/issues/6649
// - https://github.com/cosmos/cosmos-sdk/issues/6827

/*
import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

// test sending from chainA to chainB using both coins that orignate on
// chainA and coins that orignate on chainB
func (suite *KeeperTestSuite) TestSendTransfer() {
	var (
		amount             sdk.Coins
		channelA, channelB ibctesting.TestChannel
		err                error
	)

	testCases := []struct {
		msg      string
		malleate func()
		source   bool
		expPass  bool
	}{
		{"successful transfer from source chain",
			func() {
				_, _, _, _, channelA, channelB = suite.coordinator.Setup(suite.chainA, suite.chainB)
				amount = types.GetTransferCoin(channelB, sdk.DefaultBondDenom, 100)
			}, true, true},
		{"successful transfer with coins from counterparty chain",
			func() {
				// send coins from chainA back to chainB
				_, _, _, _, channelA, channelB = suite.coordinator.Setup(suite.chainA, suite.chainB)
				amount = types.GetTransferCoin(channelA, sdk.DefaultBondDenom, 100)
			}, false, true},
		{"source channel not found",
			func() {
				// channel references wrong ID
				_, _, _, _, channelA, channelB = suite.coordinator.Setup(suite.chainA, suite.chainB)
				channelA.ID = ibctesting.InvalidID
				amount = types.GetTransferCoin(channelB, sdk.DefaultBondDenom, 100)
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
				amount = types.GetTransferCoin(channelB, sdk.DefaultBondDenom, 100)
			}, true, false},

		// createOutgoingPacket tests
		// - source chain
		{"send coins failed",
			func() {
				_, _, _, _, channelA, channelB = suite.coordinator.Setup(suite.chainA, suite.chainB)
				amount = types.GetTransferCoin(channelB, "randomdenom", 100)
			}, true, false},
		// - receiving chain
		{"send from module account failed",
			func() {
				_, _, _, _, channelA, channelB = suite.coordinator.Setup(suite.chainA, suite.chainB)
				amount = types.GetTransferCoin(channelA, "randomdenom", 100)
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

			if !tc.source {
				// send coins from chainB to chainA
				coinFromBToA := types.GetTransferCoin(channelA, sdk.DefaultBondDenom, 100)
				transferMsg := types.NewMsgTransfer(channelB.PortID, channelB.ID, coinFromBToA, suite.chainB.SenderAccount.GetAddress(), suite.chainA.SenderAccount.GetAddress().String(), 110, 0)
				err = suite.coordinator.SendMsgs(suite.chainB, suite.chainA, channelA.ClientID, transferMsg)
				suite.Require().NoError(err) // message committed

				// receive coins on chainA from chainB
				fungibleTokenPacket := types.NewFungibleTokenPacketData(coinFromBToA, suite.chainB.SenderAccount.GetAddress().String(), suite.chainA.SenderAccount.GetAddress().String())
				packet := channeltypes.NewPacket(fungibleTokenPacket.GetBytes(), 1, channelB.PortID, channelB.ID, channelA.PortID, channelA.ID, 110, 0)

				// get proof of packet commitment from chainB
				packetKey := host.KeyPacketCommitment(packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
				proof, proofHeight := suite.chainB.QueryProof(packetKey)

				recvMsg := channeltypes.NewMsgRecvPacket(packet, proof, proofHeight, suite.chainA.SenderAccount.GetAddress())
				err = suite.coordinator.SendMsgs(suite.chainA, suite.chainB, channelB.ClientID, recvMsg)
				suite.Require().NoError(err) // message committed
			}

			err = suite.chainA.App.TransferKeeper.SendTransfer(
				suite.chainA.GetContext(), channelA.PortID, channelA.ID, amount,
				suite.chainA.SenderAccount.GetAddress(), suite.chainB.SenderAccount.GetAddress().String(), 110, 0,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// test receiving coins on chainB with coins that orignate on chainA and
// coins that orignated on chainB. Coins from source (chainA) have channelB
// as the denom prefix. The bulk of the testing occurs in the test case
// for loop since setup is intensive for all cases. The malleate function
// allows for testing invalid cases.
func (suite *KeeperTestSuite) TestOnRecvPacket() {
	var (
		channelA, channelB ibctesting.TestChannel
		coins              sdk.Coins
		receiver           string
	)

	testCases := []struct {
		msg      string
		malleate func()
		source   bool
		expPass  bool
	}{
		{"success receive from source chain", func() {}, true, true},
		{"success receive with coins orignated on this chain", func() {}, false, true},
		{"empty amount", func() {
			coins = nil
		}, true, false},
		{"invalid receiver address", func() {
			receiver = "gaia1scqhwpgsmr6vmztaa7suurfl52my6nd2kmrudl"
		}, true, false},
		{"no dest prefix on coin denom", func() {
			coins = sdk.NewCoins(sdk.NewInt64Coin("bitcoin", 100))
		}, false, false},

		// onRecvPacket
		// - coins from source chain (chainA)
		{"failure: mint zero coins", func() {
			coins = types.GetTransferCoin(channelB, sdk.DefaultBondDenom, 0)
		}, true, false},

		// - coins being sent back to original chain (chainB)
		{"tries to unescrow more tokens than allowed", func() {
			coins = types.GetTransferCoin(channelA, sdk.DefaultBondDenom, 1000000)
		}, false, false},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset
			_, _, _, _, channelA, channelB = suite.coordinator.Setup(suite.chainA, suite.chainB)
			receiver = suite.chainB.SenderAccount.GetAddress().String() // must be explicitly changed in malleate

			seq := uint64(1)

			if !tc.source {
				// send coins from chainB to chainA, receive them, acknowledge them, and send back to chainB
				coinFromBToA := types.GetTransferCoin(channelA, sdk.DefaultBondDenom, 100)
				transferMsg := types.NewMsgTransfer(channelB.PortID, channelB.ID, coinFromBToA, suite.chainB.SenderAccount.GetAddress(), suite.chainA.SenderAccount.GetAddress().String(), 110, 0)
				err := suite.coordinator.SendMsgs(suite.chainB, suite.chainA, channelA.ClientID, transferMsg)
				suite.Require().NoError(err) // message committed

				// relay send packet
				fungibleTokenPacket := types.NewFungibleTokenPacketData(coinFromBToA, suite.chainB.SenderAccount.GetAddress().String(), suite.chainA.SenderAccount.GetAddress().String())
				packet := channeltypes.NewPacket(fungibleTokenPacket.GetBytes(), 1, channelB.PortID, channelB.ID, channelA.PortID, channelA.ID, 110, 0)
				ack := channeltypes.NewResultAcknowledgement([]byte{byte(1)})
				err = suite.coordinator.RelayPacket(suite.chainB, suite.chainA, clientB, clientA, packet, ack.GetBytes())
				suite.Require().NoError(err) // relay committed

				seq++
				// NOTE: coins must be explicitly changed in malleate to test invalid cases
				coins = types.GetTransferCoin(channelA, sdk.DefaultBondDenom, 100)
			} else {
				coins = types.GetTransferCoin(channelB, sdk.DefaultBondDenom, 100)
			}

			// send coins from chainA to chainB
			transferMsg := types.NewMsgTransfer(channelA.PortID, channelA.ID, coins, suite.chainA.SenderAccount.GetAddress(), receiver, 110, 0)
			err := suite.coordinator.SendMsgs(suite.chainA, suite.chainB, channelB.ClientID, transferMsg)
			suite.Require().NoError(err) // message committed

			tc.malleate()

			data := types.NewFungibleTokenPacketData(coins, suite.chainA.SenderAccount.GetAddress().String(), receiver)
			packet := channeltypes.NewPacket(data.GetBytes(), seq, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, 100, 0)

			err = suite.chainB.App.TransferKeeper.OnRecvPacket(suite.chainB.GetContext(), packet, data)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestOnAcknowledgementPacket tests that successful acknowledgement is a no-op
// and failure acknowledment leads to refund when attempting to send from chainA
// to chainB.
func (suite *KeeperTestSuite) TestOnAcknowledgementPacket() {
	var (
		successAck = channeltypes.NewResultAcknowledgement([]byte{byte(1)})
		failedAck = channeltypes.NewErrorAcknowledgement("failed packet transfer")

		channelA, channelB ibctesting.TestChannel
		coins              sdk.Coins
	)

	testCases := []struct {
		msg      string
		ack      channeltypes.Acknowledgement
		malleate func()
		source   bool
		success  bool // success of ack
	}{
		{"success ack causes no-op", successAck, func() {
			coins = types.GetTransferCoin(channelB, sdk.DefaultBondDenom, 100)
		}, true, true},
		{"successful refund from source chain", failedAck,
			func() {
				escrow := types.GetEscrowAddress(channelA.PortID, channelA.ID)
				_, err := suite.chainA.App.BankKeeper.AddCoins(suite.chainA.GetContext(), escrow, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))))
				suite.Require().NoError(err)

				coins = types.GetTransferCoin(channelB, sdk.DefaultBondDenom, 100)
			}, true, false},
		{"successful refund from external chain", failedAck,
			func() {
				coins = types.GetTransferCoin(channelA, sdk.DefaultBondDenom, 100)
			}, false, false},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset
			_, _, _, _, channelA, channelB = suite.coordinator.Setup(suite.chainA, suite.chainB)

			tc.malleate()

			data := types.NewFungibleTokenPacketData(coins, suite.chainA.SenderAccount.GetAddress().String(), suite.chainB.SenderAccount.GetAddress().String())
			packet := channeltypes.NewPacket(data.GetBytes(), 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, 100, 0)

			var denom string
			if tc.source {
				// peel off ibc denom if returning back to original chain
				denom = sdk.DefaultBondDenom
			} else {
				denom = coins[0].Denom
			}

			preCoin := suite.chainA.App.BankKeeper.GetBalance(suite.chainA.GetContext(), suite.chainA.SenderAccount.GetAddress(), denom)

			err := suite.chainA.App.TransferKeeper.OnAcknowledgementPacket(suite.chainA.GetContext(), packet, data, tc.ack)
			suite.Require().NoError(err)

			postCoin := suite.chainA.App.BankKeeper.GetBalance(suite.chainA.GetContext(), suite.chainA.SenderAccount.GetAddress(), denom)
			deltaAmount := postCoin.Amount.Sub(preCoin.Amount)

			if tc.success {
				suite.Require().Equal(int64(0), deltaAmount.Int64(), "successful ack changed balance")
			} else {
				suite.Require().Equal(coins[0].Amount, deltaAmount, "failed ack did not trigger refund")
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
		coins              sdk.Coins
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

				coins = types.GetTransferCoin(channelB, sdk.DefaultBondDenom, 100)
			}, true, true},
		{"successful timeout from external chain",
			func() {
				coins = types.GetTransferCoin(channelA, sdk.DefaultBondDenom, 100)
			}, false, true},
		{"no source prefix on coin denom",
			func() {
				coins = sdk.NewCoins(sdk.NewCoin("bitcoin", sdk.NewInt(100)))
			}, false, false},
		{"unescrow failed",
			func() {
				coins = types.GetTransferCoin(channelB, sdk.DefaultBondDenom, 100)
			}, true, false},
		{"mint failed",
			func() {
				coins = types.GetTransferCoin(channelA, sdk.DefaultBondDenom, 0)
			}, true, false},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset
			_, _, _, _, channelA, channelB = suite.coordinator.Setup(suite.chainA, suite.chainB)

			tc.malleate()

			data := types.NewFungibleTokenPacketData(coins, suite.chainA.SenderAccount.GetAddress().String(), suite.chainB.SenderAccount.GetAddress().String())
			packet := channeltypes.NewPacket(data.GetBytes(), 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, 100, 0)

			var denom string
			if tc.source {
				denom = sdk.DefaultBondDenom
			} else {
				denom = coins[0].Denom
			}

			preCoin := suite.chainA.App.BankKeeper.GetBalance(suite.chainA.GetContext(), suite.chainA.SenderAccount.GetAddress(), denom)

			err := suite.chainA.App.TransferKeeper.OnTimeoutPacket(suite.chainA.GetContext(), packet, data)

			postCoin := suite.chainA.App.BankKeeper.GetBalance(suite.chainA.GetContext(), suite.chainA.SenderAccount.GetAddress(), denom)
			deltaAmount := postCoin.Amount.Sub(preCoin.Amount)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(coins[0].Amount.Int64(), deltaAmount.Int64(), "successful timeout did not trigger refund")
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
*/
