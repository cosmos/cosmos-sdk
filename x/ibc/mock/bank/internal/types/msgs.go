package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ics04 "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

const (
	TypeMsgTransfer           = "transfer"
	TypeMsgSendTransferPacket = "send-transfer-packet"
)

type MsgTransfer struct {
	SrcPort      string         `json:"src_port" yaml:"src_port"`
	SrcChannel   string         `json:"src_channel" yaml:"src_channel"`
	Denomination string         `json:"denomination" yaml:"denomination"`
	Amount       sdk.Int        `json:"amount" yaml:"amount"`
	Sender       sdk.AccAddress `json:"sender" yaml:"sender"`
	Receiver     string         `json:"receiver" yaml:"receiver"`
	Source       bool           `json:"source" yaml:"source"`
}
type MsgSendTransferPacket struct {
	Packet    ics04.PacketI  `json:"packet" yaml:"packet"`
	ChannelID string         `json:"channel_id" yaml:"channel_id"`
	Proofs    []ics23.Proof  `json:"proofs" yaml:"proofs"`
	Height    uint64         `json:"height" yaml:"height"`
	Signer    sdk.AccAddress `json:"signer" yaml:"signer"`
}

func NewMsgTransfer(srcPort, srcChannel string, denom string, amount sdk.Int, sender sdk.AccAddress, receiver string, source bool) MsgTransfer {
	return MsgTransfer{
		SrcPort:      srcPort,
		SrcChannel:   srcChannel,
		Denomination: denom,
		Amount:       amount,
		Sender:       sender,
		Receiver:     receiver,
		Source:       source,
	}
}

func (MsgTransfer) Route() string {
	return RouterKey
}

func (MsgTransfer) Type() string {
	return TypeMsgTransfer
}

func (msg MsgTransfer) ValidateBasic() sdk.Error {
	if !msg.Amount.IsPositive() {
		return sdk.NewError(sdk.CodespaceType(DefaultCodespace), CodeInvalidAmount, "invalid amount")
	}

	if msg.Sender.Empty() || len(msg.Receiver) == 0 {
		return sdk.NewError(sdk.CodespaceType(DefaultCodespace), CodeInvalidAddress, "invalid address")
	}

	return nil
}

func (msg MsgTransfer) GetSignBytes() []byte {
	return sdk.MustSortJSON(MouduleCdc.MustMarshalJSON(msg))
}

func (msg MsgTransfer) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

func NewMsgSendTransferPacket(packet ics04.PacketI, channelID string, proofs []ics23.Proof, height uint64, signer sdk.AccAddress) MsgSendTransferPacket {
	return MsgSendTransferPacket{
		Packet:    packet,
		ChannelID: channelID,
		Proofs:    proofs,
		Height:    height,
		Signer:    signer,
	}
}

func (MsgSendTransferPacket) Route() string {
	return RouterKey
}

func (MsgSendTransferPacket) Type() string {
	return TypeMsgSendTransferPacket
}

func (msg MsgSendTransferPacket) ValidateBasic() sdk.Error {
	if msg.Proofs == nil {
		return sdk.NewError(sdk.CodespaceType(DefaultCodespace), CodeProofMissing, "proof missing")
	}

	if msg.Signer.Empty() {
		return sdk.NewError(sdk.CodespaceType(DefaultCodespace), CodeInvalidAddress, "invalid signer")
	}

	return nil
}

func (msg MsgSendTransferPacket) GetSignBytes() []byte {
	return sdk.MustSortJSON(MouduleCdc.MustMarshalJSON(msg))
}

func (msg MsgSendTransferPacket) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
