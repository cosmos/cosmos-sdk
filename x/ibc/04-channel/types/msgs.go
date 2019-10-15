package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

const (
	lenPortID = 64
	lenChanID = 64
)

var _ sdk.Msg = MsgChannelOpenInit{}

type MsgChannelOpenInit struct {
	PortID    string         `json:"port_id"`
	ChannelID string         `json:"channel_id"`
	Channel   Channel        `json:"channel"`
	Signer    sdk.AccAddress `json:"signer"`
}

func NewMsgChannelOpenInit(portID string, channelID string, channel Channel, signer sdk.AccAddress) MsgChannelOpenInit {
	return MsgChannelOpenInit{portID, channelID, channel, signer}
}

// Route implements sdk.Msg
func (msg MsgChannelOpenInit) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgChannelOpenInit) Type() string {
	return EventTypeChannelOpenInit
}

// ValidateBasic implements sdk.Msg
func (msg MsgChannelOpenInit) ValidateBasic() sdk.Error {
	// Check PortID
	if len(msg.PortID) > lenPortID {
		return ErrLenPortID(DefaultCodespace)
	}

	// Check ChanID
	if len(msg.ChannelID) > lenChanID {
		return ErrLenChanID(DefaultCodespace)
	}

	// Signer can be empty
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgChannelOpenInit) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgChannelOpenInit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

var _ sdk.Msg = MsgChannelOpenTry{}

type MsgChannelOpenTry struct {
	PortID              string         `json:"port_id"`
	ChannelID           string         `json:"channel_id"`
	Channel             Channel        `json:"channel"`
	CounterpartyVersion string         `json:"counterparty_version"`
	ProofInit           ics23.Proof    `json:"proof_init"`
	ProofHeight         uint64         `json:"proof_height"`
	Signer              sdk.AccAddress `json:"signer"`
}

func NewMsgChannelOpenTry(portID string, channelID string, channel Channel, cpv string, proofInit ics23.Proof, proofHeight uint64, signer sdk.AccAddress) MsgChannelOpenTry {
	return MsgChannelOpenTry{portID, channelID, channel, cpv, proofInit, proofHeight, signer}
}

// Route implements sdk.Msg
func (msg MsgChannelOpenTry) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgChannelOpenTry) Type() string {
	return EventTypeChannelOpenTry
}

// ValidateBasic implements sdk.Msg
func (msg MsgChannelOpenTry) ValidateBasic() sdk.Error {
	// Check PortID
	if len(msg.PortID) > lenPortID {
		return ErrLenPortID(DefaultCodespace)
	}

	// Check ChanID
	if len(msg.ChannelID) > lenChanID {
		return ErrLenChanID(DefaultCodespace)
	}

	// Check proofs != nil
	// Check channel != nil
	// Signer can be empty
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgChannelOpenTry) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgChannelOpenTry) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

var _ sdk.Msg = MsgChannelOpenAck{}

type MsgChannelOpenAck struct {
	PortID              string         `json:"port_id"`
	ChannelID           string         `json:"channel_id"`
	CounterpartyVersion string         `json:"counterparty_version"`
	ProofTry            ics23.Proof    `json:"proof_try"`
	ProofHeight         uint64         `json:"proof_height"`
	Signer              sdk.AccAddress `json:"signer"`
}

func NewMsgChannelOpenAck(port string, channelID string, cpv string, proofTry ics23.Proof, proofHeight uint64, signer sdk.AccAddress) MsgChannelOpenAck {
	return MsgChannelOpenAck{port, channelID, cpv, proofTry, proofHeight, signer}
}

// Route implements sdk.Msg
func (msg MsgChannelOpenAck) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgChannelOpenAck) Type() string {
	return EventTypeChannelOpenAck
}

// ValidateBasic implements sdk.Msg
func (msg MsgChannelOpenAck) ValidateBasic() sdk.Error {
	// Check PortID
	if len(msg.PortID) > lenPortID {
		return ErrLenPortID(DefaultCodespace)
	}

	// Check ChanID
	if len(msg.ChannelID) > lenChanID {
		return ErrLenChanID(DefaultCodespace)
	}

	// Check proofs != nil
	// Signer can be empty
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgChannelOpenAck) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgChannelOpenAck) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

var _ sdk.Msg = MsgChannelOpenConfirm{}

type MsgChannelOpenConfirm struct {
	PortID      string         `json:"port_id"`
	ChannelID   string         `json:"channel_id"`
	ProofAck    ics23.Proof    `json:"proof_ack"`
	ProofHeight uint64         `json:"proof_height"`
	Signer      sdk.AccAddress `json:"signer"`
}

func NewMsgChannelOpenConfirm(portID string, channelID string, proofAck ics23.Proof, proofHeight uint64, signer sdk.AccAddress) MsgChannelOpenConfirm {
	return MsgChannelOpenConfirm{portID, channelID, proofAck, proofHeight, signer}
}

// Route implements sdk.Msg
func (msg MsgChannelOpenConfirm) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgChannelOpenConfirm) Type() string {
	return EventTypeChannelOpenConfirm
}

// ValidateBasic implements sdk.Msg
func (msg MsgChannelOpenConfirm) ValidateBasic() sdk.Error {
	// Check PortID
	if len(msg.PortID) > lenPortID {
		return ErrLenPortID(DefaultCodespace)
	}

	// Check ChanID
	if len(msg.ChannelID) > lenChanID {
		return ErrLenChanID(DefaultCodespace)
	}

	// Check proofs != nil
	// Signer can be empty
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgChannelOpenConfirm) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgChannelOpenConfirm) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

var _ sdk.Msg = MsgChannelCloseInit{}

type MsgChannelCloseInit struct {
	PortID    string         `json:"port_id"`
	ChannelID string         `json:"channel_id"`
	Signer    sdk.AccAddress `json:"signer"`
}

func NewMsgChannelCloseInit(portID string, channelID string, signer sdk.AccAddress) MsgChannelCloseInit {
	return MsgChannelCloseInit{portID, channelID, signer}
}

// Route implements sdk.Msg
func (msg MsgChannelCloseInit) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgChannelCloseInit) Type() string {
	return EventTypeChannelCloseInit
}

// ValidateBasic implements sdk.Msg
func (msg MsgChannelCloseInit) ValidateBasic() sdk.Error {
	// Check PortID
	if len(msg.PortID) > lenPortID {
		return ErrLenPortID(DefaultCodespace)
	}

	// Check ChanID
	if len(msg.ChannelID) > lenChanID {
		return ErrLenChanID(DefaultCodespace)
	}

	// Signer can be empty
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgChannelCloseInit) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgChannelCloseInit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

var _ sdk.Msg = MsgChannelCloseConfirm{}

type MsgChannelCloseConfirm struct {
	PortID      string         `json:"port_id"`
	ChannelID   string         `json:"channel_id"`
	ProofInit   ics23.Proof    `json:"proof_init"`
	ProofHeight uint64         `json:"proof_height"`
	Signer      sdk.AccAddress `json:"signer"`
}

func NewMsgChannelCloseConfirm(portID string, channelID string, proofInit ics23.Proof, proofHeight uint64, signer sdk.AccAddress) MsgChannelCloseConfirm {
	return MsgChannelCloseConfirm{portID, channelID, proofInit, proofHeight, signer}
}

// Route implements sdk.Msg
func (msg MsgChannelCloseConfirm) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgChannelCloseConfirm) Type() string {
	return EventTypeChannelCloseConfirm
}

// ValidateBasic implements sdk.Msg
func (msg MsgChannelCloseConfirm) ValidateBasic() sdk.Error {
	// Check PortID
	if len(msg.PortID) > lenPortID {
		return ErrLenPortID(DefaultCodespace)
	}

	// Check ChanID
	if len(msg.ChannelID) > lenChanID {
		return ErrLenChanID(DefaultCodespace)
	}

	// Check proofs != nil
	// Signer can be empty
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgChannelCloseConfirm) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgChannelCloseConfirm) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

var _ sdk.Msg = MsgSendPacket{}

// MsgSendPacket PortID dependent on type
// ChannelID can be empty if batched & not first MsgSendPacket
// Height uint64 // height of the commitment root for the proofs
type MsgSendPacket struct {
	Packet    exported.PacketI `json:"packet" yaml:"packet"`
	ChannelID string           `json:"channel_id" yaml:"channel_id"`
	Proofs    []ics23.Proof    `json:"proofs" yaml:"proofs"`
	Height    uint64           `json:"height" yaml:"height"`
	Signer    sdk.AccAddress   `json:"signer" yaml:"signer"`
}

func NewMsgSendPacket(packet exported.PacketI, channelID string, proofs []ics23.Proof, height uint64, signer sdk.AccAddress) MsgSendPacket {
	return MsgSendPacket{packet, channelID, proofs, height, signer}
}

// Route implements sdk.Msg
func (msg MsgSendPacket) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgSendPacket) Type() string {
	return EventTypeSendPacket
}

// ValidateBasic implements sdk.Msg
func (msg MsgSendPacket) ValidateBasic() sdk.Error {
	// Check ChanID
	if len(msg.ChannelID) > lenChanID {
		return ErrLenChanID(DefaultCodespace)
	}

	// Check proofs != nil
	// Check packet != nil
	// Signer can be empty
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgSendPacket) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgSendPacket) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
