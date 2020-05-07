package types

import (
	"encoding/base64"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

var _ sdk.Msg = MsgChannelOpenInit{}

// NewMsgChannelOpenInit creates a new MsgChannelCloseInit MsgChannelOpenInit
func NewMsgChannelOpenInit(
	portID, channelID string, version string, channelOrder ibctypes.Order, connectionHops []string,
	counterpartyPortID, counterpartyChannelID string, signer sdk.AccAddress,
) MsgChannelOpenInit {
	counterparty := NewCounterparty(counterpartyPortID, counterpartyChannelID)
	channel := NewChannel(ibctypes.INIT, channelOrder, counterparty, connectionHops, version)
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
func (msg MsgChannelOpenInit) ValidateBasic() error {
	if err := host.DefaultPortIdentifierValidator(msg.PortID); err != nil {
		return sdkerrors.Wrap(err, "invalid port ID")
	}
	if err := host.DefaultChannelIdentifierValidator(msg.ChannelID); err != nil {
		return sdkerrors.Wrap(err, "invalid channel ID")
	}
	// Signer can be empty
	return msg.Channel.ValidateBasic()
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

// NewMsgChannelOpenTry creates a new MsgChannelOpenTry instance
func NewMsgChannelOpenTry(
	portID, channelID, version string, channelOrder ibctypes.Order, connectionHops []string,
	counterpartyPortID, counterpartyChannelID, counterpartyVersion string,
	proofInit commitmenttypes.MerkleProof, proofHeight uint64, signer sdk.AccAddress,
) MsgChannelOpenTry {
	counterparty := NewCounterparty(counterpartyPortID, counterpartyChannelID)
	channel := NewChannel(ibctypes.INIT, channelOrder, counterparty, connectionHops, version)
	return MsgChannelOpenTry{
		PortID:              portID,
		ChannelID:           channelID,
		Channel:             channel,
		CounterpartyVersion: counterpartyVersion,
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
func (msg MsgChannelOpenTry) ValidateBasic() error {
	if err := host.DefaultPortIdentifierValidator(msg.PortID); err != nil {
		return sdkerrors.Wrap(err, "invalid port ID")
	}
	if err := host.DefaultChannelIdentifierValidator(msg.ChannelID); err != nil {
		return sdkerrors.Wrap(err, "invalid channel ID")
	}
	if strings.TrimSpace(msg.CounterpartyVersion) == "" {
		return sdkerrors.Wrap(ErrInvalidCounterparty, "counterparty version cannot be blank")
	}
	if msg.ProofInit.IsEmpty() {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if err := msg.ProofInit.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof init cannot be nil")
	}
	if msg.ProofHeight == 0 {
		return sdkerrors.Wrap(ibctypes.ErrInvalidHeight, "proof height must be > 0")
	}
	// Signer can be empty
	return msg.Channel.ValidateBasic()
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

// NewMsgChannelOpenAck creates a new MsgChannelOpenAck instance
func NewMsgChannelOpenAck(
	portID, channelID string, cpv string, proofTry commitmenttypes.MerkleProof, proofHeight uint64,
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
func (msg MsgChannelOpenAck) ValidateBasic() error {
	if err := host.DefaultPortIdentifierValidator(msg.PortID); err != nil {
		return sdkerrors.Wrap(err, "invalid port ID")
	}
	if err := host.DefaultChannelIdentifierValidator(msg.ChannelID); err != nil {
		return sdkerrors.Wrap(err, "invalid channel ID")
	}
	if strings.TrimSpace(msg.CounterpartyVersion) == "" {
		return sdkerrors.Wrap(ErrInvalidCounterparty, "counterparty version cannot be blank")
	}
	if msg.ProofTry.IsEmpty() {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if err := msg.ProofTry.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof try cannot be nil")
	}
	if msg.ProofHeight == 0 {
		return sdkerrors.Wrap(ibctypes.ErrInvalidHeight, "proof height must be > 0")
	}
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

// NewMsgChannelOpenConfirm creates a new MsgChannelOpenConfirm instance
func NewMsgChannelOpenConfirm(
	portID, channelID string, proofAck commitmenttypes.MerkleProof, proofHeight uint64,
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
func (msg MsgChannelOpenConfirm) ValidateBasic() error {
	if err := host.DefaultPortIdentifierValidator(msg.PortID); err != nil {
		return sdkerrors.Wrap(err, "invalid port ID")
	}
	if err := host.DefaultChannelIdentifierValidator(msg.ChannelID); err != nil {
		return sdkerrors.Wrap(err, "invalid channel ID")
	}
	if msg.ProofAck.IsEmpty() {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if err := msg.ProofAck.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof ack cannot be nil")
	}
	if msg.ProofHeight == 0 {
		return sdkerrors.Wrap(ibctypes.ErrInvalidHeight, "proof height must be > 0")
	}
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

// NewMsgChannelCloseInit creates a new MsgChannelCloseInit instance
func NewMsgChannelCloseInit(
	portID string, channelID string, signer sdk.AccAddress,
) MsgChannelCloseInit {
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
func (msg MsgChannelCloseInit) ValidateBasic() error {
	if err := host.DefaultPortIdentifierValidator(msg.PortID); err != nil {
		return sdkerrors.Wrap(err, "invalid port ID")
	}
	if err := host.DefaultChannelIdentifierValidator(msg.ChannelID); err != nil {
		return sdkerrors.Wrap(err, "invalid channel ID")
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

// NewMsgChannelCloseConfirm creates a new MsgChannelCloseConfirm instance
func NewMsgChannelCloseConfirm(
	portID, channelID string, proofInit commitmenttypes.MerkleProof, proofHeight uint64,
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
func (msg MsgChannelCloseConfirm) ValidateBasic() error {
	if err := host.DefaultPortIdentifierValidator(msg.PortID); err != nil {
		return sdkerrors.Wrap(err, "invalid port ID")
	}
	if err := host.DefaultChannelIdentifierValidator(msg.ChannelID); err != nil {
		return sdkerrors.Wrap(err, "invalid channel ID")
	}
	if msg.ProofInit.IsEmpty() {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if err := msg.ProofInit.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof init cannot be nil")
	}
	if msg.ProofHeight == 0 {
		return sdkerrors.Wrap(ibctypes.ErrInvalidHeight, "proof height must be > 0")
	}
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

var _ sdk.Msg = MsgPacket{}

// NewMsgPacket constructs new MsgPacket
func NewMsgPacket(
	packet Packet, proof commitmenttypes.MerkleProof, proofHeight uint64,
	signer sdk.AccAddress,
) MsgPacket {
	return MsgPacket{
		Packet:      packet,
		Proof:       proof,
		ProofHeight: proofHeight,
		Signer:      signer,
	}
}

// Route implements sdk.Msg
func (msg MsgPacket) Route() string {
	return ibctypes.RouterKey
}

// ValidateBasic implements sdk.Msg
func (msg MsgPacket) ValidateBasic() error {
	if msg.Proof.IsEmpty() {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if err := msg.Proof.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof ack cannot be nil")
	}
	if msg.ProofHeight == 0 {
		return sdkerrors.Wrap(ibctypes.ErrInvalidHeight, "proof height must be > 0")
	}
	if msg.Signer.Empty() {
		return sdkerrors.ErrInvalidAddress
	}

	return msg.Packet.ValidateBasic()
}

// GetSignBytes implements sdk.Msg
func (msg MsgPacket) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetDataSignBytes returns the base64-encoded bytes used for the
// data field when signing the packet.
func (msg MsgPacket) GetDataSignBytes() []byte {
	s := "\"" + base64.StdEncoding.EncodeToString(msg.Packet.Data) + "\""
	return []byte(s)
}

// GetSigners implements sdk.Msg
func (msg MsgPacket) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

// Type implements sdk.Msg
func (msg MsgPacket) Type() string {
	return "ics04/opaque"
}

var _ sdk.Msg = MsgTimeout{}

// NewMsgTimeout constructs new MsgTimeout
func NewMsgTimeout(
	packet Packet, nextSequenceRecv uint64, proof commitmenttypes.MerkleProof,
	proofHeight uint64, signer sdk.AccAddress,
) MsgTimeout {
	return MsgTimeout{
		Packet:           packet,
		NextSequenceRecv: nextSequenceRecv,
		Proof:            proof,
		ProofHeight:      proofHeight,
		Signer:           signer,
	}
}

// Route implements sdk.Msg
func (msg MsgTimeout) Route() string {
	return ibctypes.RouterKey
}

// ValidateBasic implements sdk.Msg
func (msg MsgTimeout) ValidateBasic() error {
	if msg.Proof.IsEmpty() {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if err := msg.Proof.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof ack cannot be nil")
	}
	if msg.ProofHeight == 0 {
		return sdkerrors.Wrap(ibctypes.ErrInvalidHeight, "proof height must be > 0")
	}
	if msg.Signer.Empty() {
		return sdkerrors.ErrInvalidAddress
	}

	return msg.Packet.ValidateBasic()
}

// GetSignBytes implements sdk.Msg
func (msg MsgTimeout) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgTimeout) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

// Type implements sdk.Msg
func (msg MsgTimeout) Type() string {
	return "ics04/timeout"
}

var _ sdk.Msg = MsgAcknowledgement{}

// NewMsgAcknowledgement constructs a new MsgAcknowledgement
func NewMsgAcknowledgement(
	packet Packet, ack []byte, proof commitmenttypes.MerkleProof, proofHeight uint64, signer sdk.AccAddress) MsgAcknowledgement {
	return MsgAcknowledgement{
		Packet:          packet,
		Acknowledgement: ack,
		Proof:           proof,
		ProofHeight:     proofHeight,
		Signer:          signer,
	}
}

// Route implements sdk.Msg
func (msg MsgAcknowledgement) Route() string {
	return ibctypes.RouterKey
}

// ValidateBasic implements sdk.Msg
func (msg MsgAcknowledgement) ValidateBasic() error {
	if msg.Proof.IsEmpty() {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if err := msg.Proof.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof ack cannot be nil")
	}
	if len(msg.Acknowledgement) > 100 {
		return sdkerrors.Wrap(ErrAcknowledgementTooLong, "acknowledgement cannot exceed 100 bytes")
	}
	if msg.ProofHeight == 0 {
		return sdkerrors.Wrap(ibctypes.ErrInvalidHeight, "proof height must be > 0")
	}
	if msg.Signer.Empty() {
		return sdkerrors.ErrInvalidAddress
	}

	return msg.Packet.ValidateBasic()
}

// GetSignBytes implements sdk.Msg
func (msg MsgAcknowledgement) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgAcknowledgement) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

// Type implements sdk.Msg
func (msg MsgAcknowledgement) Type() string {
	return "ics04/opaque"
}
