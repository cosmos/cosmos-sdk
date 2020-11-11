package keeper_test

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

type OnRecvPacketTestCase  = struct {
	// The required subset of bank balances
	bankBefore []bank.Balance
	// The packet to process
	packet types.FungibleTokenPacketData
	// The expected changes in the bank
	bankChange []bank.Balance
	// Whether OnRecvPacket should pass or fail
	pass bool
}

type OwnedCoin struct {
	Address string
	Denom string
}
type Bank struct {
	balances map[OwnedCoin]sdk.Int
}

// Make an empty bank
func MakeBank() Bank {
	return Bank{ balances: make(map[OwnedCoin]sdk.Int) }
}

// Subtract other bank from this bank
func (bank *Bank) Sub(other *Bank) Bank {
	diff := MakeBank()
	for coin, amount := range bank.balances {
		otherAmount, exists := other.balances[coin]
		if exists {
			diff.balances[coin] = amount.Sub(otherAmount)
		} else {
			diff.balances[coin] = amount
		}
	}
	for coin, amount := range other.balances {
		if _, exists := bank.balances[coin]; !exists {
			diff.balances[coin] = amount.Neg()
		}
	}
	return diff
}

// Set specific bank balance
func (bank *Bank) SetBalance(address string, denom string, amount sdk.Int) {
	bank.balances[OwnedCoin{address, denom}] = amount
}

// String representation of all bank balances
func (bank *Bank) String() string {
	str := ""
	for coin, amount := range bank.balances {
		str += coin.Address + " : " + coin.Denom + " = " + amount.String() + "\n"
	}
	return str
}

// String representation of non-zero bank balances
func (bank *Bank) NonZeroString() string {
	str := ""
	for coin, amount := range bank.balances {
		if !amount.IsZero() {
			str += coin.Address + " : " + coin.Denom + " = " + amount.String() + "\n"
		}
	}
	return str
}

// Construct a bank out of the chain bank
func BankOfChain(chain *ibctesting.TestChain) Bank {
	bank := MakeBank()
	chain.App.BankKeeper.IterateAllBalances(chain.GetContext(), func(address sdk.AccAddress, coin sdk.Coin) (stop bool){
		bank.SetBalance(address.String(), coin.Denom, coin.Amount)
		return false
	})
	return bank
}

// Set balances of the chain bank for balances present in the bank
func (suite *KeeperTestSuite) SetChainBankBalances(chain *ibctesting.TestChain, bank *Bank) error {
	for coin, amount := range bank.balances {
		address, err := sdk.AccAddressFromBech32(coin.Address)
		if err != nil {
			return err
		}
		chain.App.BankKeeper.SetBalance(chain.GetContext(), address, sdk.NewCoin(coin.Denom, amount))
	}
	return nil
}


func (suite *KeeperTestSuite) CheckBankBalances(chain *ibctesting.TestChain, bankBefore *Bank, expectedBankChange *Bank) error {
	bankAfter := BankOfChain(chain)
	bankChange := bankAfter.Sub(bankBefore)
	diff := bankChange.Sub(expectedBankChange)
	NonZeroString := diff.NonZeroString()
	if len(NonZeroString) != 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, NonZeroString)
	}
	return nil
}

func (suite *KeeperTestSuite) TestModelBasedOnRecvPacket() {
	var (
		channelA, channelB ibctesting.TestChannel
	)

	suite.Run(fmt.Sprintf("Model based test for OnRecvPacket"), func() {
		suite.SetupTest() // reset
		_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, ibctesting.Tendermint)
		channelA, channelB = suite.coordinator.CreateTransferChannels(suite.chainA, suite.chainB, connA, connB, channeltypes.UNORDERED)

		seq := uint64(1)
		coin := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))
		data := types.NewFungibleTokenPacketData(coin.Denom, coin.Amount.Uint64(), suite.chainB.SenderAccount.GetAddress().String(), suite.chainA.SenderAccount.GetAddress().String())
		packet := channeltypes.NewPacket(data.GetBytes(), seq, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, clienttypes.NewHeight(0, 100), 0)
		err := suite.chainB.App.TransferKeeper.OnRecvPacket(suite.chainB.GetContext(), packet, data)
		suite.Require().NoError(err)

		emptyBank := MakeBank()
		err = suite.CheckBankBalances(suite.chainA, &emptyBank, &emptyBank)
		suite.Require().NoError(err)
	})
}
