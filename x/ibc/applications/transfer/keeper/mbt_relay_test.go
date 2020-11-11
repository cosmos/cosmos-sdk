package keeper_test

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
	"strings"
)

type OnRecvPacketTestCase = struct {
	description string
	// The required subset of bank balances
	bankBefore []Balance
	// The packet to process
	packet types.FungibleTokenPacketData
	// The expected changes in the bank
	bankChange []Balance
	// Whether OnRecvPacket should pass or fail
	pass bool
}

type OwnedCoin struct {
	Address string
	Denom string
}

type Balance struct {
	Address string
	Denom string
	Amount sdk.Int
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

// Set several balances at once
func (bank *Bank) SetBalances(balances []Balance) {
	for _, balance := range balances {
		bank.balances[OwnedCoin{balance.Address, balance.Denom}] = balance.Amount
	}
}

// Set several balances at once
func BankFromBalances(balances []Balance) Bank{
	bank := MakeBank()
	for _, balance := range balances {
		bank.balances[OwnedCoin{balance.Address, balance.Denom}] = balance.Amount
	}
	return bank
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
		fullDenom := coin.Denom
		if strings.HasPrefix(coin.Denom, "ibc/") {
			fullDenom, _ = chain.App.TransferKeeper.DenomPathFromHash(chain.GetContext(), coin.Denom)
		}
		bank.SetBalance(address.String(), fullDenom, coin.Amount)
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
		err = chain.App.BankKeeper.SetBalance(chain.GetContext(), address, sdk.NewCoin(coin.Denom, amount))
		if err != nil {
			return err
		}
	}
	return nil
}

// Check that the state of the bank is the bankBefore + expectedBankChange
func (suite *KeeperTestSuite) CheckBankBalances(chain *ibctesting.TestChain, bankBefore *Bank, expectedBankChange *Bank) error {
	bankAfter := BankOfChain(chain)
	bankChange := bankAfter.Sub(bankBefore)
	diff := bankChange.Sub(expectedBankChange)
	NonZeroString := diff.NonZeroString()
	if len(NonZeroString) != 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "Unexpected changes in the bank: \n" + NonZeroString)
	}
	return nil
}


func StaticOnRecvPacketTestCases() []OnRecvPacketTestCase {
	return []OnRecvPacketTestCase {
		{
			description: "failure zero amount",
			bankBefore: []Balance{},
			packet: types.FungibleTokenPacketData {"stake", 0, "cosmos1dpv8nhpl26lfpcc0f9wyseyhn0l8sv894j8tw5","cosmos1dpv8nhpl26lfpcc0f9wyseyhn0l8sv894j8tw5"},
			bankChange: []Balance{},
			pass: false,
		},
		{
			description: "failure empty denomination",
			bankBefore: []Balance{},
			packet: types.FungibleTokenPacketData {"", 1, "cosmos1dpv8nhpl26lfpcc0f9wyseyhn0l8sv894j8tw5","cosmos1dpv8nhpl26lfpcc0f9wyseyhn0l8sv894j8tw5"},
			bankChange: []Balance{},
			pass: false,
		},
		{
			description: "success expected change",
			bankBefore: []Balance{},
			packet: types.FungibleTokenPacketData {"a", 1, "cosmos1dpv8nhpl26lfpcc0f9wyseyhn0l8sv894j8tw5","cosmos1dpv8nhpl26lfpcc0f9wyseyhn0l8sv894j8tw5"},
			bankChange: []Balance{{"cosmos1dpv8nhpl26lfpcc0f9wyseyhn0l8sv894j8tw5", "transfer/testchain1-conn0-chan0/a", sdk.NewInt(1)}},
			pass: true,
		},
	}
}


func (suite *KeeperTestSuite) TestModelBasedStaticOnRecvPacket() {

	var (
		channelA, channelB ibctesting.TestChannel
	)

	for _, tc := range StaticOnRecvPacketTestCases() {
		suite.Run(fmt.Sprintf("Case %s", tc.description), func() {
			suite.SetupTest() // reset
			_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, ibctesting.Tendermint)
			channelA, channelB = suite.coordinator.CreateTransferChannels(suite.chainA, suite.chainB, connA, connB, channeltypes.UNORDERED)

			seq := uint64(1)
			packet := channeltypes.NewPacket(tc.packet.GetBytes(), seq, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, clienttypes.NewHeight(0, 100), 0)

			bankBefore := BankFromBalances(tc.bankBefore)
			if suite.SetChainBankBalances(suite.chainB, &bankBefore) != nil {
				suite.Require().False(tc.pass)
				return
			}
			bankBefore = BankOfChain(suite.chainB)
			if suite.chainB.App.TransferKeeper.OnRecvPacket(suite.chainB.GetContext(), packet, tc.packet) != nil {
				suite.Require().False(tc.pass)
				return
			}
			expectedChange := BankFromBalances(tc.bankChange)
			if suite.CheckBankBalances(suite.chainB, &bankBefore, &expectedChange) != nil {
				suite.Require().False(tc.pass)
				return
			}
			suite.Require().True(tc.pass)
		})
	}
}
