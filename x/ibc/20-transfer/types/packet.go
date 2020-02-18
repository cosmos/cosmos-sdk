package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
)

var _ channelexported.PacketDataI = FungibleTokenPacketData{}

// FungibleTokenPacketData defines a struct for the packet payload
// See FungibleTokenPacketData spec: https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#data-structures
type FungibleTokenPacketData struct {
	Amount   sdk.Coins      `json:"amount" yaml:"amount"`     // the tokens to be transferred
	Sender   sdk.AccAddress `json:"sender" yaml:"sender"`     // the sender address
	Receiver sdk.AccAddress `json:"receiver" yaml:"receiver"` // the recipient address on the destination chain
	Source   bool           `json:"source" yaml:"source"`     // indicates if the sending chain is the source chain of the tokens to be transferred
	Timeout  uint64         `json:"timeout" yaml:"timeout"`
}

// NewFungibleTokenPacketData contructs a new FungibleTokenPacketData instance
func NewFungibleTokenPacketData(
	amount sdk.Coins, sender, receiver sdk.AccAddress,
	source bool, timeout uint64) FungibleTokenPacketData {
	return FungibleTokenPacketData{
		Amount:   amount,
		Sender:   sender,
		Receiver: receiver,
		Source:   source,
		Timeout:  timeout,
	}
}

// String returns a string representation of FungibleTokenPacketData
func (ftpd FungibleTokenPacketData) String() string {
	return fmt.Sprintf(`FungibleTokenPacketData:
	Amount:               %s
	Sender:               %s
	Receiver:             %s
	Source:               %v`,
		ftpd.Amount.String(),
		ftpd.Sender,
		ftpd.Receiver,
		ftpd.Source,
	)
}

// ValidateBasic implements channelexported.PacketDataI
func (ftpd FungibleTokenPacketData) ValidateBasic() error {
	if !ftpd.Amount.IsAllPositive() {
		return sdkerrors.ErrInsufficientFunds
	}
	if !ftpd.Amount.IsValid() {
		return sdkerrors.ErrInvalidCoins
	}
	if ftpd.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender address")
	}
	if ftpd.Receiver.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing receiver address")
	}
	if ftpd.Timeout == 0 {
		return sdkerrors.Wrap(ErrInvalidPacketTimeout, "timeout cannot be 0")
	}
	return nil
}

// GetBytes implements channelexported.PacketDataI
func (ftpd FungibleTokenPacketData) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(ftpd))
}

// GetTimeoutHeight implements channelexported.PacketDataI
func (ftpd FungibleTokenPacketData) GetTimeoutHeight() uint64 {
	return ftpd.Timeout
}

// Type implements channelexported.PacketDataI
func (ftpd FungibleTokenPacketData) Type() string {
	return "ics20/transfer"
}

var _ channelexported.PacketAcknowledgementI = AckDataTransfer{}

// AckDataTransfer is a no-op packet
// See spec for onAcknowledgePacket: https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#packet-relay
type AckDataTransfer struct{}

// GetBytes implements channelexported.PacketAcknowledgementI
func (ack AckDataTransfer) GetBytes() []byte {
	return []byte("fungible token transfer ack")
}
