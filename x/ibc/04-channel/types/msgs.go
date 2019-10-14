package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

var _ sdk.Msg = MsgChanOpenInit{}

type MsgChanOpenInit struct {
	PortID    string         `json:"port_id"`
	ChannelID string         `json:"channel_id"`
	Channel   Channel        `json:"channel"`
	Signer    sdk.AccAddress `json:"signer"`
}

// Route implements sdk.Msg
func (msg MsgChanOpenInit) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgChanOpenInit) Type() string {
	return "chan_open_init"
}

// ValidateBasic implements sdk.Msg
func (msg MsgChanOpenInit) ValidateBasic() sdk.Error {
	// TODO:
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgChanOpenInit) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgChanOpenInit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

var _ sdk.Msg = MsgChanOpenTry{}

type MsgChanOpenTry struct {
	PortID    string         `json:"port_id"`
	ChannelID string         `json:"channel_id"`
	Channel   Channel        `json:"channel"`
	Proofs    []ics23.Proof  `json:"proofs"`
	Height    uint64         `json:"height"`
	Signer    sdk.AccAddress `json:"signer"`
}

// Route implements sdk.Msg
func (msg MsgChanOpenTry) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgChanOpenTry) Type() string {
	return "chan_open_try"
}

// ValidateBasic implements sdk.Msg
func (msg MsgChanOpenTry) ValidateBasic() sdk.Error {
	// TODO:
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgChanOpenTry) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgChanOpenTry) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

var _ sdk.Msg = MsgChanOpenAck{}

type MsgChanOpenAck struct {
	PortID    string         `json:"port_id"`
	ChannelID string         `json:"channel_id"`
	Proofs    []ics23.Proof  `json:"proofs"`
	Height    uint64         `json:"height"`
	Signer    sdk.AccAddress `json:"signer"`
}

// Route implements sdk.Msg
func (msg MsgChanOpenAck) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgChanOpenAck) Type() string {
	return "chan_open_ack"
}

// ValidateBasic implements sdk.Msg
func (msg MsgChanOpenAck) ValidateBasic() sdk.Error {
	// TODO:
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgChanOpenAck) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgChanOpenAck) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

var _ sdk.Msg = MsgChanOpenConfirm{}

type MsgChanOpenConfirm struct {
	PortID    string         `json:"port_id"`
	ChannelID string         `json:"channel_id"`
	Proofs    []ics23.Proof  `json:"proofs"`
	Height    uint64         `json:"height"`
	Signer    sdk.AccAddress `json:"signer"`
}

// Route implements sdk.Msg
func (msg MsgChanOpenConfirm) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgChanOpenConfirm) Type() string {
	return "chan_open_confirm"
}

// ValidateBasic implements sdk.Msg
func (msg MsgChanOpenConfirm) ValidateBasic() sdk.Error {
	// TODO:
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgChanOpenConfirm) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgChanOpenConfirm) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

var _ sdk.Msg = MsgPacket{}

// MsgPacket PortID dependent on type
// ChannelID can be empty if batched & not first MsgPacket
// Height uint64 // height of the commitment root for the proofs
type MsgPacket struct {
	Packet    exported.PacketI `json:"packet" yaml:"packet"`
	ChannelID string           `json:"channel_id,omitempty" yaml:"channel_id"`
	Proofs    []ics23.Proof    `json:"proofs" yaml:"proofs"`
	Height    uint64           `json:"height" yaml:"height"`
	Signer    sdk.AccAddress   `json:"signer,omitempty" yaml:"signer"`
}

// Route implements sdk.Msg
func (msg MsgPacket) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgPacket) Type() string {
	return "packet"
}

// ValidateBasic implements sdk.Msg
func (msg MsgPacket) ValidateBasic() sdk.Error {
	// TODO:
	// Check PortID ChannelID len
	// Check packet != nil
	// Check proofs != nil
	// Signer can be empty
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgPacket) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgPacket) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
