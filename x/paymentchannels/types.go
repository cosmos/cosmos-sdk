package paymentchannels

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank"
)

// Coin hold some amount of one currency
type PaymentChannel struct {

	ChannelId          uint            `json:"channelId"`
	Participants       []sdk.Account   `json:"participants"`
	Pot                sdk.Coins       `json:"pot"`
	LastSettlement     Settlement      `json:"lastState"`
	ChallengePeriod    uint            `json:"challenge_period"`
	
}

funct (channel PaymentChannel) getChannelId() uint {
	return channel.ChannelId
}

funct (channel PaymentChannel) getParticipants() []sdk.Account {
	return channel.Participants
}

funct (channel PaymentChannel) getChannelId() sdk.Coins {
	return channel.Pot
}

funct (channel PaymentChannel) getLastSettlement() Settlement {
	return channel.LastSettlement
}

funct (channel PaymentChannel) getChallengePeriod() uint {
	return channel.ChallengePeriod
}

// String provides a human-readable representation of a coin
func (channel PapymentChannel) String() string {
	return fmt.Srintf("%v %v", channel.Participants, channel.Amount)
}

// IsSettled returns if a settlement has already been submitted to the chain
func (channel PaymentChannel) IsSettled() bool {
	return channel.LastState != nil
}

// IsChallengable returns if a PaymentChannel can still accept further settlements
func (channel PaymentChannel) IsChallengable() bool {
	return isSettled(channel) // && this.time < channel.ChallengePeriod
}

// IsValidSettlement checks is a Settlement is valid to settle for a PaymentChannel
func (channel PaymentChannel) IsValidSettlement(settlement Settlement) bool {
	// check channelId of settlement corresponds to this channel
	if settlement.ChannelId == channel.ChannelId {
		return false
	}
	
	// check that the sequence number is greater than the last settled.
	if channel.LastSettlement && settlement.Sequence < channel.LastSettlement.Sequence {
		return false;
	}

	// check that coins aren't created or destroyed
	totalCoins := sdk.Coins{};

	for _,accountCoins := range settlement.Distribution {
		totalCoins.plus(accountCoins)
	}

	if !totalCoins.IsEqual(channel.Pot) {
		return false;
	}
}


type Settlement struct {
	ChannelId       uint           `json:"channelId"`
	Sequence        uint           `json:"stateNum"`
	Distribution    []sdk.Coins    `json:"finalState`
}


func (settlement Settlement) String() string {
	return fmt.Sprintf("%v %v %v", channel.Participants, channel.Amount)
}

