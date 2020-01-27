package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
)

var _ channelexported.PacketDataI = PacketDataTransfer{}

// PacketDataTransfer defines a struct for the packet payload
// See FungibleTokenPacketData spec: https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#data-structures
type PacketDataTransfer struct {
	Amount   sdk.Coins      `json:"amount" yaml:"amount"`     // the tokens to be transferred
	Sender   sdk.AccAddress `json:"sender" yaml:"sender"`     // the sender address
	Receiver sdk.AccAddress `json:"receiver" yaml:"receiver"` // the recipient address on the destination chain
	Source   bool           `json:"source" yaml:"source"`     // indicates if the sending chain is the source chain of the tokens to be transferred
	Timeout  uint64         `json:"timeout" yaml:"timeout"`
}

// NewPacketDataTransfer contructs a new PacketDataTransfer
func NewPacketDataTransfer(amount sdk.Coins, sender, receiver sdk.AccAddress, source bool, timeout uint64) PacketDataTransfer {
	return PacketDataTransfer{
		Amount:   amount,
		Sender:   sender,
		Receiver: receiver,
		Source:   source,
		Timeout:  timeout,
	}
}

// String returns a string representation of PacketDataTransfer
func (pd PacketDataTransfer) String() string {
	return fmt.Sprintf(`PacketDataTransfer:
	Amount:               %s
	Sender:               %s
	Receiver:             %s
	Source:               %v`,
		pd.Amount.String(),
		pd.Sender,
		pd.Receiver,
		pd.Source,
	)
}

// ValidateBasic implements channelexported.PacketDataI
func (pd PacketDataTransfer) ValidateBasic() error {
	if !pd.Amount.IsAllPositive() {
		return sdkerrors.ErrInsufficientFunds
	}
	if !pd.Amount.IsValid() {
		return sdkerrors.ErrInvalidCoins
	}
	if pd.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender address")
	}
	if pd.Receiver.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing receiver address")
	}
	return nil
}

// GetBytes implements channelexported.PacketDataI
func (pd PacketDataTransfer) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(pd))
}

// GetTimeoutHeight implements channelexported.PacketDataI
func (pd PacketDataTransfer) GetTimeoutHeight() uint64 {
	return pd.Timeout
}

// Type implements channelexported.PacketDataI
func (pd PacketDataTransfer) Type() string {
	return "ics20/transfer"
}

var _ channelexported.PacketDataI = AckDataTransfer{}

// AckDataTransfer is a no-op packet
// See spec for onAcknowledgePacket: https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#packet-relay
type AckDataTransfer struct{}

// ValidateBasic implements channelexported.PacketDataI
func (ack AckDataTransfer) ValidateBasic() error {
	return nil
}

// GetBytes implements channelexported.PacketDataI
func (ack AckDataTransfer) GetBytes() []byte {
	return []byte("ok")
}

// GetTimeoutHeight implements channelexported.PacketDataI
func (ack AckDataTransfer) GetTimeoutHeight() uint64 {
	return 0
}

// Type implements channelexported.PacketDataI
func (ack AckDataTransfer) Type() string {
	return "ics20/transfer/ack"
}
