package paymentchannels

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PaymentChannel is a one way payment channel
type PaymentChannel struct {
	ChannelID       uint          `json:"channelId"`
	Participants    []sdk.Account `json:"participants"`
	Pot             sdk.Coins     `json:"pot"`
	LastSettlement  Settlement    `json:"lastState"`
	ChallengePeriod uint          `json:"challenge_period"`
}

// String provides a human-readable representation of a coin
func (channel PapymentChannel) String() string {
	return fmt.Sprintf("%v %v", channel.Participants, channel.Amount)
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
		return false
	}

	// check that coins aren't created or destroyed
	totalCoins := sdk.Coins{}

	for _, accountCoins := range settlement.Distribution {
		totalCoins.plus(accountCoins)
	}

	if !totalCoins.IsEqual(channel.Pot) {
		return false
	}
	
	for i, participant := range channel.Participants {
		

	if !sig.PubKey.VerifyBytes(msg.GetSignBytes(), sig.Signature) {
					return ctx,
						sdk.ErrUnauthorized("").Result(),
						true
				}
}

// Settlement is a settlement of a PaymentChannel
type Settlement struct {
	ChannelID    uint               `json:"channelId"`
	Sequence     uint               `json:"sequence"`
	Distribution []sdk.Coins        `json:"disribution"`
	Signatures   []crypto.Signature `json:"signatures"`
}

func (settlement Settlement) String() string {
	return fmt.Sprintf("%v %v %v", channel.Participants, channel.Amount, channel.Distribution)
}
