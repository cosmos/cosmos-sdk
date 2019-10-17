package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

const (
	// TODO: double check lenght. Eventually move to ICS24
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

// NewMsgChannelOpenInit creates a new MsgChannelCloseInit MsgChannelOpenInit
func NewMsgChannelOpenInit(
	portID, channelID string, channel Channel, signer sdk.AccAddress,
) MsgChannelOpenInit {
	return MsgChannelOpenInit{
		PortID:    portID,
		ChannelID: channelID,
		Channel:   channel,
		Signer:    signer,
	}
}

// Route implements sdk.Msg
func (msg MsgChannelOpenInit) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgChannelOpenInit) Type() string {
	return "channel_open_init"
}

// ValidateBasic implements sdk.Msg
func (msg MsgChannelOpenInit) ValidateBasic() sdk.Error {
	if strings.TrimSpace(msg.PortID) == "" {
		return ErrInvalidPortID(DefaultCodespace, "port ID can't be blank")
	}

	if len(msg.PortID) > lenPortID {
		return ErrInvalidPortID(DefaultCodespace, fmt.Sprintf("port ID length can't be > %d", lenPortID))
	}

	if strings.TrimSpace(msg.ChannelID) == "" {
		return ErrInvalidPortID(DefaultCodespace, "channel ID can't be blank")
	}

	if len(msg.ChannelID) > lenChanID {
		return ErrInvalidChannelID(DefaultCodespace, fmt.Sprintf("channel ID length can't be > %d", lenChanID))
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

// NewMsgChannelOpenTry creates a new MsgChannelOpenTry instance
func NewMsgChannelOpenTry(
	portID, channelID string, channel Channel, cpv string, proofInit ics23.Proof,
	proofHeight uint64, signer sdk.AccAddress,
) MsgChannelOpenTry {
	return MsgChannelOpenTry{
		PortID:              portID,
		ChannelID:           channelID,
		Channel:             channel,
		CounterpartyVersion: cpv,
		ProofInit:           proofInit,
		ProofHeight:         proofHeight,
		Signer:              signer,
	}
}

// Route implements sdk.Msg
func (msg MsgChannelOpenTry) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgChannelOpenTry) Type() string {
	return "channel_open_try"
}

// ValidateBasic implements sdk.Msg
func (msg MsgChannelOpenTry) ValidateBasic() sdk.Error {

	if strings.TrimSpace(msg.PortID) == "" {
		return ErrInvalidPortID(DefaultCodespace, "port ID can't be blank")
	}

	if len(msg.PortID) > lenPortID {
		return ErrInvalidPortID(DefaultCodespace, fmt.Sprintf("port ID length can't be > %d", lenPortID))
	}

	if strings.TrimSpace(msg.ChannelID) == "" {
		return ErrInvalidPortID(DefaultCodespace, "channel ID can't be blank")
	}

	if len(msg.ChannelID) > lenChanID {
		return ErrInvalidChannelID(DefaultCodespace, fmt.Sprintf("channel ID length can't be > %d", lenChanID))
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

// NewMsgChannelOpenAck creates a new MsgChannelOpenAck instance
func NewMsgChannelOpenAck(
	portID, channelID string, cpv string, proofTry ics23.Proof, proofHeight uint64,
	signer sdk.AccAddress,
) MsgChannelOpenAck {
	return MsgChannelOpenAck{
		PortID:              portID,
		ChannelID:           channelID,
		CounterpartyVersion: cpv,
		ProofTry:            proofTry,
		ProofHeight:         proofHeight,
		Signer:              signer,
	}
}

// Route implements sdk.Msg
func (msg MsgChannelOpenAck) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgChannelOpenAck) Type() string {
	return "channel_open_ack"
}

// ValidateBasic implements sdk.Msg
func (msg MsgChannelOpenAck) ValidateBasic() sdk.Error {
	if strings.TrimSpace(msg.PortID) == "" {
		return ErrInvalidPortID(DefaultCodespace, "port ID can't be blank")
	}

	if len(msg.PortID) > lenPortID {
		return ErrInvalidPortID(DefaultCodespace, fmt.Sprintf("port ID length can't be > %d", lenPortID))
	}

	if strings.TrimSpace(msg.ChannelID) == "" {
		return ErrInvalidPortID(DefaultCodespace, "channel ID can't be blank")
	}

	if len(msg.ChannelID) > lenChanID {
		return ErrInvalidChannelID(DefaultCodespace, fmt.Sprintf("channel ID length can't be > %d", lenChanID))
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

// NewMsgChannelOpenConfirm creates a new MsgChannelOpenConfirm instance
func NewMsgChannelOpenConfirm(
	portID, channelID string, proofAck ics23.Proof, proofHeight uint64,
	signer sdk.AccAddress,
) MsgChannelOpenConfirm {
	return MsgChannelOpenConfirm{
		PortID:      portID,
		ChannelID:   channelID,
		ProofAck:    proofAck,
		ProofHeight: proofHeight,
		Signer:      signer,
	}
}

// Route implements sdk.Msg
func (msg MsgChannelOpenConfirm) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgChannelOpenConfirm) Type() string {
	return "channel_open_confirm"
}

// ValidateBasic implements sdk.Msg
func (msg MsgChannelOpenConfirm) ValidateBasic() sdk.Error {
	if strings.TrimSpace(msg.PortID) == "" {
		return ErrInvalidPortID(DefaultCodespace, "port ID can't be blank")
	}

	if len(msg.PortID) > lenPortID {
		return ErrInvalidPortID(DefaultCodespace, fmt.Sprintf("port ID length can't be > %d", lenPortID))
	}

	if strings.TrimSpace(msg.ChannelID) == "" {
		return ErrInvalidPortID(DefaultCodespace, "channel ID can't be blank")
	}

	if len(msg.ChannelID) > lenChanID {
		return ErrInvalidChannelID(DefaultCodespace, fmt.Sprintf("channel ID length can't be > %d", lenChanID))
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

// NewMsgChannelCloseInit creates a new MsgChannelCloseInit instance
func NewMsgChannelCloseInit(portID string, channelID string, signer sdk.AccAddress) MsgChannelCloseInit {
	return MsgChannelCloseInit{
		PortID:    portID,
		ChannelID: channelID,
		Signer:    signer,
	}
}

// Route implements sdk.Msg
func (msg MsgChannelCloseInit) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgChannelCloseInit) Type() string {
	return "channel_close_init"
}

// ValidateBasic implements sdk.Msg
func (msg MsgChannelCloseInit) ValidateBasic() sdk.Error {
	if strings.TrimSpace(msg.PortID) == "" {
		return ErrInvalidPortID(DefaultCodespace, "port ID can't be blank")
	}

	if len(msg.PortID) > lenPortID {
		return ErrInvalidPortID(DefaultCodespace, fmt.Sprintf("port ID length can't be > %d", lenPortID))
	}

	if strings.TrimSpace(msg.ChannelID) == "" {
		return ErrInvalidPortID(DefaultCodespace, "channel ID can't be blank")
	}

	if len(msg.ChannelID) > lenChanID {
		return ErrInvalidChannelID(DefaultCodespace, fmt.Sprintf("channel ID length can't be > %d", lenChanID))
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

// NewMsgChannelCloseConfirm creates a new MsgChannelCloseConfirm instance
func NewMsgChannelCloseConfirm(
	portID, channelID string, proofInit ics23.Proof, proofHeight uint64,
	signer sdk.AccAddress,
) MsgChannelCloseConfirm {
	return MsgChannelCloseConfirm{
		PortID:      portID,
		ChannelID:   channelID,
		ProofInit:   proofInit,
		ProofHeight: proofHeight,
		Signer:      signer,
	}
}

// Route implements sdk.Msg
func (msg MsgChannelCloseConfirm) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgChannelCloseConfirm) Type() string {
	return "channel_close_confirm"
}

// ValidateBasic implements sdk.Msg
func (msg MsgChannelCloseConfirm) ValidateBasic() sdk.Error {
	if strings.TrimSpace(msg.PortID) == "" {
		return ErrInvalidPortID(DefaultCodespace, "port ID can't be blank")
	}

	if len(msg.PortID) > lenPortID {
		return ErrInvalidPortID(DefaultCodespace, fmt.Sprintf("port ID length can't be > %d", lenPortID))
	}

	if strings.TrimSpace(msg.ChannelID) == "" {
		return ErrInvalidPortID(DefaultCodespace, "channel ID can't be blank")
	}

	if len(msg.ChannelID) > lenChanID {
		return ErrInvalidChannelID(DefaultCodespace, fmt.Sprintf("channel ID length can't be > %d", lenChanID))
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

// NewMsgSendPacket creates a new MsgSendPacket instance
func NewMsgSendPacket(packet exported.PacketI, signer sdk.AccAddress) MsgSendPacket {
	return MsgSendPacket{
		Packet: packet,
		Signer: signer,
	}
}

// Route implements sdk.Msg
func (msg MsgSendPacket) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgSendPacket) Type() string {
	return "send_packet"
}

// ValidateBasic implements sdk.Msg
func (msg MsgSendPacket) ValidateBasic() sdk.Error {
	// TODO: move to Packet validation function
	if strings.TrimSpace(msg.Packet.SourcePort()) == "" {
		return ErrInvalidPortID(DefaultCodespace, "packet source port ID can't be blank")
	}

	if len(msg.Packet.SourcePort()) > lenPortID {
		return ErrInvalidPortID(DefaultCodespace, fmt.Sprintf("packet source port ID length can't be > %d", lenPortID))
	}

	if strings.TrimSpace(msg.Packet.DestPort()) == "" {
		return ErrInvalidPortID(DefaultCodespace, "packet destination port ID can't be blank")
	}

	if len(msg.Packet.DestPort()) > lenPortID {
		return ErrInvalidPortID(DefaultCodespace, fmt.Sprintf("packet destination port ID length can't be > %d", lenPortID))
	}

	if strings.TrimSpace(msg.Packet.SourceChannel()) == "" {
		return ErrInvalidPortID(DefaultCodespace, "packet source channel ID can't be blank")
	}

	if len(msg.Packet.SourceChannel()) > lenChanID {
		return ErrInvalidChannelID(DefaultCodespace, fmt.Sprintf("packet source channel ID length can't be > %d", lenChanID))
	}

	if strings.TrimSpace(msg.Packet.DestChannel()) == "" {
		return ErrInvalidPortID(DefaultCodespace, "packet destination channel ID can't be blank")
	}

	if len(msg.Packet.DestChannel()) > lenChanID {
		return ErrInvalidChannelID(DefaultCodespace, fmt.Sprintf("packet destination channel ID length can't be > %d", lenChanID))
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
