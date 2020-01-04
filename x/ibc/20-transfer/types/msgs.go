package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

type MsgTransfer struct {
	SourcePort    string         `json:"source_port" yaml:"source_port"`       // the port on which the packet will be sent
	SourceChannel string         `json:"source_channel" yaml:"source_channel"` // the channel by which the packet will be sent
	Amount        sdk.Coins      `json:"amount" yaml:"amount"`                 // the tokens to be transferred
	Sender        sdk.AccAddress `json:"sender" yaml:"sender"`                 // the sender address
	Receiver      sdk.AccAddress `json:"receiver" yaml:"receiver"`             // the recipient address on the destination chain
	Source        bool           `json:"source" yaml:"source"`                 // indicates if the sending chain is the source chain of the tokens to be transferred
}

// NewMsgTransfer creates a new MsgTransfer instance
func NewMsgTransfer(
	sourcePort, sourceChannel string, amount sdk.Coins, sender, receiver sdk.AccAddress, source bool,
) MsgTransfer {
	return MsgTransfer{
		SourcePort:    sourcePort,
		SourceChannel: sourceChannel,
		Amount:        amount,
		Sender:        sender,
		Receiver:      receiver,
		Source:        source,
	}
}

// Route implements sdk.Msg
func (MsgTransfer) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (MsgTransfer) Type() string {
	return "transfer"
}

// ValidateBasic implements sdk.Msg
func (msg MsgTransfer) ValidateBasic() error {
	if err := host.DefaultPortIdentifierValidator(msg.SourcePort); err != nil {
		return sdkerrors.Wrap(err, "invalid source port ID")
	}
	if err := host.DefaultChannelIdentifierValidator(msg.SourceChannel); err != nil {
		return sdkerrors.Wrap(err, "invalid source channel ID")
	}
	if !msg.Amount.IsAllPositive() {
		return sdkerrors.ErrInsufficientFunds
	}
	if !msg.Amount.IsValid() {
		return sdkerrors.ErrInvalidCoins
	}
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender address")
	}
	if msg.Receiver.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing recipient address")
	}
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgTransfer) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgTransfer) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

type MsgRecvPacket struct {
	Packet channelexported.PacketI `json:"packet" yaml:"packet"`
	Proofs []commitment.Proof      `json:"proofs" yaml:"proofs"`
	Height uint64                  `json:"height" yaml:"height"`
	Signer sdk.AccAddress          `json:"signer" yaml:"signer"`
}

// NewMsgRecvPacket creates a new MsgRecvPacket instance
func NewMsgRecvPacket(packet channelexported.PacketI, proofs []commitment.Proof, height uint64, signer sdk.AccAddress) MsgRecvPacket {
	return MsgRecvPacket{
		Packet: packet,
		Proofs: proofs,
		Height: height,
		Signer: signer,
	}
}

// Route implements sdk.Msg
func (MsgRecvPacket) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (MsgRecvPacket) Type() string {
	return "recv_packet"
}

// ValidateBasic implements sdk.Msg
func (msg MsgRecvPacket) ValidateBasic() error {
	if msg.Height == 0 {
		return sdkerrors.Wrap(ibctypes.ErrInvalidHeight, "height must be > 0")
	}
	if msg.Proofs == nil || len(msg.Proofs) == 0 {
		return sdkerrors.Wrap(commitment.ErrInvalidProof, "missing proof")
	}
	for _, proof := range msg.Proofs {
		if err := proof.ValidateBasic(); err != nil {
			return err
		}
	}
	if msg.Signer.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender address")
	}
	return msg.Packet.ValidateBasic()
}

// GetSignBytes implements sdk.Msg
func (msg MsgRecvPacket) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgRecvPacket) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
