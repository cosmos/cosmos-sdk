package keeper_test

import (
	"encoding/json"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
	"io/ioutil"
	"strings"
)


type TlaBalance struct {
	Address []string  `json:"address"`
	Denom []string    `json:"denom"`
	Amount int64      `json:"amount"`
}

type TlaFungibleTokenPacketData struct {
	Sender string     `json:"sender"`
	Receiver string   `json:"receiver"`
	Amount int        `json:"amount"`
	Denom []string    `json:"denom"`
}

type TlaFungibleTokenPacket struct {
	SourceChannel string `json:"sourceChannel"`
	SourcePort string    `json:"sourcePort"`
	DestChannel string   `json:"destChannel"`
	DestPort string      `json:"destPort"`
	Data TlaFungibleTokenPacketData `json:"data"`
}

type TlaOnRecvPacketTestCase = struct {
	// The required subset of bank balances
	BankBefore []TlaBalance        `json:"bankBefore"`
	// The packet to process
	Packet TlaFungibleTokenPacket  `json:"packet"`
	// The expected changes in the bank
	BankAfter []TlaBalance        `json:"bankAfter"`
	// Whether OnRecvPacket should fail or not
	Error bool                     `json:"error"`
}

type FungibleTokenPacket struct {
	SourceChannel string
	SourcePort string
	DestChannel string
	DestPort string
	Data types.FungibleTokenPacketData
}

type OnRecvPacketTestCase = struct {
	description string
	// The required subset of bank balances
	bankBefore []Balance
	// The packet to process
	packet FungibleTokenPacket
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

func AddressFromTla(addr []string) string {
	return strings.Join(addr, "/")
}

func DenomFromTla(denom []string) string {
	if len(denom) != 3 {
		panic("failed to convert from TLA+ denom")
	}
	s := ""
	if len(denom[0]) == 0 && len(denom[1]) == 0 {
		// native denomination
		s = denom[2]
	} else  {
		s = strings.Join(denom, "/")
	}
	return s
}

func BalanceFromTla(balance TlaBalance) Balance {
	return Balance{
		Address: AddressFromTla(balance.Address),
		Denom:   DenomFromTla(balance.Denom),
		Amount:  sdk.NewInt(balance.Amount),
	}
}

func BalancesFromTla(tla []TlaBalance) []Balance {
	balances := make([]Balance,0)
	for _, b := range tla {
		balances = append(balances, BalanceFromTla(b))
	}
	return balances
}

func FungibleTokenPacketFromTla(packet TlaFungibleTokenPacket) FungibleTokenPacket {
	return FungibleTokenPacket{
		SourceChannel: packet.SourceChannel,
		SourcePort:    packet.SourcePort,
		DestChannel:   packet.DestChannel,
		DestPort:      packet.DestPort,
		Data:          types.NewFungibleTokenPacketData(
			DenomFromTla(packet.Data.Denom),
			uint64(packet.Data.Amount),
			packet.Data.Sender,
			packet.Data.Receiver),
	}
}

func OnRecvPacketTestCaseFromTla(tc TlaOnRecvPacketTestCase) OnRecvPacketTestCase {
	return OnRecvPacketTestCase{
		description: "auto-generated",
		bankBefore:  BalancesFromTla(tc.BankBefore),
		packet:      FungibleTokenPacketFromTla(tc.Packet),
		bankChange:  BalancesFromTla(tc.BankAfter),
		pass:        !tc.Error,
	}
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
		//{
		//	description: "failure zero amount",
		//	bankBefore: []Balance{},
		//	data: types.FungibleTokenPacketData {"stake", 0, "cosmos1dpv8nhpl26lfpcc0f9wyseyhn0l8sv894j8tw5","cosmos1dpv8nhpl26lfpcc0f9wyseyhn0l8sv894j8tw5"},
		//	bankChange: []Balance{},
		//	pass: false,
		//},
		//{
		//	description: "failure empty denomination",
		//	bankBefore: []Balance{},
		//	data: types.FungibleTokenPacketData {"", 1, "cosmos1dpv8nhpl26lfpcc0f9wyseyhn0l8sv894j8tw5","cosmos1dpv8nhpl26lfpcc0f9wyseyhn0l8sv894j8tw5"},
		//	bankChange: []Balance{},
		//	pass: false,
		//},
		//{
		//	description: "success expected change",
		//	bankBefore: []Balance{},
		//	data: types.FungibleTokenPacketData {"a", 1, "cosmos1dpv8nhpl26lfpcc0f9wyseyhn0l8sv894j8tw5","cosmos1dpv8nhpl26lfpcc0f9wyseyhn0l8sv894j8tw5"},
		//	bankChange: []Balance{{"cosmos1dpv8nhpl26lfpcc0f9wyseyhn0l8sv894j8tw5", "transfer/testchain1-conn0-chan0/a", sdk.NewInt(1)}},
		//	pass: true,
		//},
	}
}

func (suite *KeeperTestSuite) TestModelBasedStaticOnRecvPacket() {
	var testCases = []TlaOnRecvPacketTestCase{}

	jsonBlob, err := ioutil.ReadFile("recv-test.json")
	if err != nil {
		panic(fmt.Errorf("Failed to read JSON test fixture: %w", err))
	}

	err = json.Unmarshal([]byte(jsonBlob), &testCases)
	if err != nil {
		panic(fmt.Errorf("Failed to parse JSON test fixture: %w", err))
	}

	var (
		channelA, channelB ibctesting.TestChannel
	)

	for _, tc := range StaticOnRecvPacketTestCases() {
		suite.Run(fmt.Sprintf("Case %s", tc.description), func() {
			suite.SetupTest() // reset
			_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, ibctesting.Tendermint)
			channelA, channelB = suite.coordinator.CreateTransferChannels(suite.chainA, suite.chainB, connA, connB, channeltypes.UNORDERED)

			seq := uint64(1)
			packet := channeltypes.NewPacket(tc.packet.Data.GetBytes(), seq, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, clienttypes.NewHeight(0, 100), 0)

			bankBefore := BankFromBalances(tc.bankBefore)
			if suite.SetChainBankBalances(suite.chainB, &bankBefore) != nil {
				panic("failed to set chain balances")
			}
			bankBefore = BankOfChain(suite.chainB)
			if suite.chainB.App.TransferKeeper.OnRecvPacket(suite.chainB.GetContext(), packet, tc.packet.Data) != nil {
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
