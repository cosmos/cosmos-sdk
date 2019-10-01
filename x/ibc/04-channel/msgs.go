package channel

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

const Route = "ibc"

type MsgOpenInit struct {
	PortID    string         `json:"port_id"`
	ChannelID string         `json:"channel_id"`
	Channel   Channel        `json:"channel"`
	Signer    sdk.AccAddress `json:"signer"`
}

var _ sdk.Msg = MsgOpenInit{}

func (msg MsgOpenInit) Route() string {
	return Route
}

func (msg MsgOpenInit) Type() string {
	return "open-init"
}

func (msg MsgOpenInit) ValidateBasic() sdk.Error {
	return nil // TODO
}

func (msg MsgOpenInit) GetSignBytes() []byte {
	return sdk.MustSortJSON(msgCdc.MustMarshalJSON(msg))
}

func (msg MsgOpenInit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

type MsgOpenTry struct {
	PortID    string             `json:"port_id"`
	ChannelID string             `json:"channel_id"`
	Channel   Channel            `json:"channel"`
	Proofs    []commitment.Proof `json:"proofs"`
	Height    uint64             `json:"height"`
	Signer    sdk.AccAddress     `json:"signer"`
}

var _ sdk.Msg = MsgOpenTry{}

func (msg MsgOpenTry) Route() string {
	return Route
}

func (msg MsgOpenTry) Type() string {
	return "open-try"
}

func (msg MsgOpenTry) ValidateBasic() sdk.Error {
	return nil // TODO
}

func (msg MsgOpenTry) GetSignBytes() []byte {
	return sdk.MustSortJSON(msgCdc.MustMarshalJSON(msg))
}

func (msg MsgOpenTry) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

type MsgOpenAck struct {
	PortID    string             `json:"port_id"`
	ChannelID string             `json:"channel_id"`
	Proofs    []commitment.Proof `json:"proofs"`
	Height    uint64             `json:"height"`
	Signer    sdk.AccAddress     `json:"signer"`
}

var _ sdk.Msg = MsgOpenAck{}

func (msg MsgOpenAck) Route() string {
	return Route
}

func (msg MsgOpenAck) Type() string {
	return "open-ack"
}

func (msg MsgOpenAck) ValidateBasic() sdk.Error {
	return nil // TODO
}

func (msg MsgOpenAck) GetSignBytes() []byte {
	return sdk.MustSortJSON(msgCdc.MustMarshalJSON(msg))
}

func (msg MsgOpenAck) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

type MsgOpenConfirm struct {
	PortID    string             `json:"port_id"`
	ChannelID string             `json:"channel_id"`
	Proofs    []commitment.Proof `json:"proofs"`
	Height    uint64             `json:"height"`
	Signer    sdk.AccAddress     `json:"signer"`
}

var _ sdk.Msg = MsgOpenConfirm{}

func (msg MsgOpenConfirm) Route() string {
	return Route
}

func (msg MsgOpenConfirm) Type() string {
	return "open-confirm"
}

func (msg MsgOpenConfirm) ValidateBasic() sdk.Error {
	return nil // TODO
}

func (msg MsgOpenConfirm) GetSignBytes() []byte {
	return sdk.MustSortJSON(msgCdc.MustMarshalJSON(msg))
}

func (msg MsgOpenConfirm) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

// PortID dependent on type
// ChannelID can be empty if batched & not first MsgPacket
// Height uint64 // height of the commitment root for the proofs
type MsgPacket struct {
	Packet    `json:"packet" yaml:"packet"`
	ChannelID string             `json:"channel_id,omitempty" yaml:"channel_id"`
	Proofs    []commitment.Proof `json:"proofs" yaml:"proofs"`
	Height    uint64             `json:"height" yaml:"height"`
	Signer    sdk.AccAddress     `json:"signer,omitempty" yaml:"signer"`
}

var _ sdk.Msg = MsgPacket{}

func (msg MsgPacket) ValidateBasic() sdk.Error {
	// Check PortID ChannelID len
	// Check packet != nil
	// Check proofs != nil
	// Signer can be empty
	return nil // TODO
}

func (msg MsgPacket) Route() string {
	return msg.ReceiverPort()
}

func (msg MsgPacket) GetSignBytes() []byte {
	return sdk.MustSortJSON(msgCdc.MustMarshalJSON(msg))
}

func (msg MsgPacket) GetSigners() []sdk.AccAddress {
	if msg.Signer.Empty() {
		return []sdk.AccAddress{}
	}
	return []sdk.AccAddress{msg.Signer}
}
