package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// FungibleTokenPacketData defines a struct for the packet payload
// See FungibleTokenPacketData spec: https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#data-structures
type FungibleTokenPacketData struct {
	Amount   sdk.Coins `json:"amount" yaml:"amount"`     // the tokens to be transferred
	Sender   string    `json:"sender" yaml:"sender"`     // the sender address
	Receiver string    `json:"receiver" yaml:"receiver"` // the recipient address on the destination chain
}

// NewFungibleTokenPacketData contructs a new FungibleTokenPacketData instance
func NewFungibleTokenPacketData(
	amount sdk.Coins, sender, receiver string) FungibleTokenPacketData {
	return FungibleTokenPacketData{
		Amount:   amount,
		Sender:   sender,
		Receiver: receiver,
	}
}

// String returns a string representation of FungibleTokenPacketData
func (ftpd FungibleTokenPacketData) String() string {
	return fmt.Sprintf(`FungibleTokenPacketData:
	Amount:               %s
	Sender:               %s
	Receiver:             %s`,
		ftpd.Amount.String(),
		ftpd.Sender,
		ftpd.Receiver,
	)
}

// ValidateBasic is used for validating the token transfer
func (ftpd FungibleTokenPacketData) ValidateBasic() error {
	if !ftpd.Amount.IsAllPositive() {
		return sdkerrors.ErrInsufficientFunds
	}
	if !ftpd.Amount.IsValid() {
		return sdkerrors.ErrInvalidCoins
	}
	if ftpd.Sender == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender address")
	}
	if ftpd.Receiver == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing receiver address")
	}
	return nil
}

// GetBytes is a helper for serialising
func (ftpd FungibleTokenPacketData) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(ftpd))
}

// FungibleTokenPacketAcknowledgement contains a boolean success flag and an optional error msg
// error msg is empty string on success
// See spec for onAcknowledgePacket: https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#packet-relay
type FungibleTokenPacketAcknowledgement struct {
	Success bool   `json:"success" yaml:"success"`
	Error   string `json:"error" yaml:"error"`
}

// GetBytes is a helper for serialising
func (ack FungibleTokenPacketAcknowledgement) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(ack))
}
