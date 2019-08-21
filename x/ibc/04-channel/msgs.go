package channel

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

const Route = "ibc"

type MsgOpenInit struct {
	ConnectionID       string
	ChannelID          string
	Channel            Channel
	CounterpartyClient string
	NextTimeout        uint64
	Signer             sdk.AccAddress
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
	return nil // TODO
}

func (msg MsgOpenInit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

type MsgOpenTry struct {
	ConnectionID       string
	ChannelID          string
	Channel            Channel
	CounterpartyClient string
	Timeout            uint64
	NextTimeout        uint64
	Proofs             []commitment.Proof
	Signer             sdk.AccAddress
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
	return nil // TODO
}

func (msg MsgOpenTry) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

type MsgOpenAck struct {
	ConnectionID string
	ChannelID    string
	Timeout      uint64
	NextTimeout  uint64
	Proofs       []commitment.Proof
	Signer       sdk.AccAddress
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
	return nil // TODO
}

func (msg MsgOpenAck) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

type MsgOpenConfirm struct {
	ConnectionID string
	ChannelID    string
	Timeout      uint64
	Proofs       []commitment.Proof
	Signer       sdk.AccAddress
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
	return nil // TODO
}

func (msg MsgOpenConfirm) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

type MsgPacket struct {
	Packet `json:"packet" yaml:"packet"`
	// PortID, ChannelID can be empty if batched & not first MsgPacket
	PortID    string `json:"port_id,omitempty" yaml:"port_id"`
	ChannelID string `json:"channel_id,omitempty" yaml:"channel_id"`
	// Height uint64 // height of the commitment root for the proofs
	Proofs []commitment.Proof `json:"proofs" yaml:"proofs"`
	Signer sdk.AccAddress     `json:"signer,omitempty" yaml:"signer"`
}

var _ sdk.Msg = MsgPacket{}

func (msg MsgPacket) ValidateBasic() sdk.Error {
	// Check PortID ChannelID len
	// Check packet != nil
	// Check proofs != nil
	// Signer can be empty
	return nil // TODO
}

func (msg MsgPacket) GetSignBytes() []byte {
	return msgCdc.MustMarshalJSON(msg) // TODO: Sort 
}

func (msg MsgPacket) GetSigners() []sdk.AccAddress {
	if msg.Signer.Empty() {
		return []sdk.AccAddress{}
	}
	return []sdk.AccAddress{msg.Signer}
}
