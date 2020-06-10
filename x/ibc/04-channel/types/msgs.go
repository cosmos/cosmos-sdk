package types

import (
	"encoding/base64"
	"fmt"
	"strings"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

var _ sdk.Msg = &MsgChannelOpenInit{}

// NewMsgChannelOpenInit creates a new MsgChannelCloseInit MsgChannelOpenInit
func NewMsgChannelOpenInit(
	portID, channelID string, version string, channelOrder Order, connectionHops []string,
	counterpartyPortID, counterpartyChannelID string, signer sdk.AccAddress,
) *MsgChannelOpenInit {
	counterparty := NewCounterparty(counterpartyPortID, counterpartyChannelID)
	channel := NewChannel(INIT, channelOrder, counterparty, connectionHops, version)
	return &MsgChannelOpenInit{
		PortID:    portID,
		ChannelID: channelID,
		Channel:   channel,
		Signer:    signer,
	}
}

// Route implements sdk.Msg
func (msg MsgChannelOpenInit) Route() string {
	return host.RouterKey
}

// Type implements sdk.Msg
func (msg MsgChannelOpenInit) Type() string {
	return "channel_open_init"
}

// ValidateBasic implements sdk.Msg
func (msg MsgChannelOpenInit) ValidateBasic() error {
	if err := host.PortIdentifierValidator(msg.PortID); err != nil {
		return sdkerrors.Wrap(err, "invalid port ID")
	}
	if err := host.ChannelIdentifierValidator(msg.ChannelID); err != nil {
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

var (
	_ sdk.Msg                          = &MsgChannelOpenTry{}
	_ cdctypes.UnpackInterfacesMessage = &MsgChannelOpenTry{}
)

// NewMsgChannelOpenTry creates a new MsgChannelOpenTry instance
func NewMsgChannelOpenTry(
	portID, channelID, version string, channelOrder Order, connectionHops []string,
	counterpartyPortID, counterpartyChannelID, counterpartyVersion string,
	proofInit commitmentexported.Proof, proofHeight uint64, signer sdk.AccAddress,
) (*MsgChannelOpenTry, error) {
	counterparty := NewCounterparty(counterpartyPortID, counterpartyChannelID)
	channel := NewChannel(INIT, channelOrder, counterparty, connectionHops, version)

	proofInitAny, err := proofInit.PackAny()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid proof init")
	}

	return &MsgChannelOpenTry{
		PortID:              portID,
		ChannelID:           channelID,
		Channel:             channel,
		CounterpartyVersion: counterpartyVersion,
		ProofInit:           *proofInitAny,
		ProofHeight:         proofHeight,
		Signer:              signer,
	}, nil
}

// Route implements sdk.Msg
func (msg MsgChannelOpenTry) Route() string {
	return host.RouterKey
}

// Type implements sdk.Msg
func (msg MsgChannelOpenTry) Type() string {
	return "channel_open_try"
}

// ValidateBasic implements sdk.Msg
func (msg MsgChannelOpenTry) ValidateBasic() error {
	if err := host.PortIdentifierValidator(msg.PortID); err != nil {
		return sdkerrors.Wrap(err, "invalid port ID")
	}
	if err := host.ChannelIdentifierValidator(msg.ChannelID); err != nil {
		return sdkerrors.Wrap(err, "invalid channel ID")
	}
	if strings.TrimSpace(msg.CounterpartyVersion) == "" {
		return sdkerrors.Wrap(ErrInvalidCounterparty, "counterparty version cannot be blank")
	}
	proofInit := msg.GetProofInit()
	if proofInit == nil || proofInit.IsEmpty() {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if err := proofInit.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof init failed basic validation")
	}
	if msg.ProofHeight == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "proof height must be > 0")
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

// GetProofInit returns the cached value from ProofInit. It returns nil if the value
// is not cached or if the proof doesn't cast to a commitment Proof.
func (msg MsgChannelOpenTry) GetProofInit() commitmentexported.Proof {
	proof, _ := commitmenttypes.UnpackAnyProof(&msg.ProofInit)
	return proof
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgChannelOpenTry) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	var proofInit commitmentexported.Proof
	err := unpacker.UnpackAny(&msg.ProofInit, &proofInit)
	if err != nil {
		return fmt.Errorf("proof init unpack failed: %w", err)
	}

	return nil
}

var (
	_ sdk.Msg                          = &MsgChannelOpenAck{}
	_ cdctypes.UnpackInterfacesMessage = &MsgChannelOpenAck{}
)

// NewMsgChannelOpenAck creates a new MsgChannelOpenAck instance
func NewMsgChannelOpenAck(
	portID, channelID string, cpv string, proofTry commitmentexported.Proof, proofHeight uint64,
	signer sdk.AccAddress,
) (*MsgChannelOpenAck, error) {
	proofTryAny, err := proofTry.PackAny()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid proof try")
	}

	return &MsgChannelOpenAck{
		PortID:              portID,
		ChannelID:           channelID,
		CounterpartyVersion: cpv,
		ProofTry:            *proofTryAny,
		ProofHeight:         proofHeight,
		Signer:              signer,
	}, nil
}

// Route implements sdk.Msg
func (msg MsgChannelOpenAck) Route() string {
	return host.RouterKey
}

// Type implements sdk.Msg
func (msg MsgChannelOpenAck) Type() string {
	return "channel_open_ack"
}

// ValidateBasic implements sdk.Msg
func (msg MsgChannelOpenAck) ValidateBasic() error {
	if err := host.PortIdentifierValidator(msg.PortID); err != nil {
		return sdkerrors.Wrap(err, "invalid port ID")
	}
	if err := host.ChannelIdentifierValidator(msg.ChannelID); err != nil {
		return sdkerrors.Wrap(err, "invalid channel ID")
	}
	if strings.TrimSpace(msg.CounterpartyVersion) == "" {
		return sdkerrors.Wrap(ErrInvalidCounterparty, "counterparty version cannot be blank")
	}
	proofTry := msg.GetProofTry()
	if proofTry == nil || proofTry.IsEmpty() {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if err := proofTry.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof try failed basic validation")
	}
	if msg.ProofHeight == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "proof height must be > 0")
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

// GetProofTry returns the cached value from ProofTry. It returns nil if the value
// is not cached or if the proof doesn't cast to a commitment Proof.
func (msg MsgChannelOpenAck) GetProofTry() commitmentexported.Proof {
	proof, _ := commitmenttypes.UnpackAnyProof(&msg.ProofTry)
	return proof
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgChannelOpenAck) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	var proofTry commitmentexported.Proof
	err := unpacker.UnpackAny(&msg.ProofTry, &proofTry)
	if err != nil {
		return fmt.Errorf("proof try unpack failed: %w", err)
	}

	return nil
}

var (
	_ sdk.Msg                          = &MsgChannelOpenConfirm{}
	_ cdctypes.UnpackInterfacesMessage = &MsgChannelOpenConfirm{}
)

// NewMsgChannelOpenConfirm creates a new MsgChannelOpenConfirm instance
func NewMsgChannelOpenConfirm(
	portID, channelID string, proofAck commitmentexported.Proof, proofHeight uint64,
	signer sdk.AccAddress,
) (*MsgChannelOpenConfirm, error) {
	proofAckAny, err := proofAck.PackAny()
	if err != nil {
		returnnil, sdkerrors.Wrap(err, "invalid proof ack")
	}
	return &MsgChannelOpenConfirm{
		PortID:      portID,
		ChannelID:   channelID,
		ProofAck:    *proofAckAny,
		ProofHeight: proofHeight,
		Signer:      signer,
	}, nil
}

// Route implements sdk.Msg
func (msg MsgChannelOpenConfirm) Route() string {
	return host.RouterKey
}

// Type implements sdk.Msg
func (msg MsgChannelOpenConfirm) Type() string {
	return "channel_open_confirm"
}

// ValidateBasic implements sdk.Msg
func (msg MsgChannelOpenConfirm) ValidateBasic() error {
	if err := host.PortIdentifierValidator(msg.PortID); err != nil {
		return sdkerrors.Wrap(err, "invalid port ID")
	}
	if err := host.ChannelIdentifierValidator(msg.ChannelID); err != nil {
		return sdkerrors.Wrap(err, "invalid channel ID")
	}
	proofAck := msg.GetProofAck()
	if proofAck == nil || proofAck.IsEmpty() {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if err := proofAck.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof ack failed basic validation")
	}
	if msg.ProofHeight == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "proof height must be > 0")
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

// GetProofAck returns the cached value from ProofAck. It returns nil if the value
// is not cached or if the proof doesn't cast to a commitment Proof.
func (msg MsgChannelOpenConfirm) GetProofAck() commitmentexported.Proof {
	proof, _ := commitmenttypes.UnpackAnyProof(&msg.ProofAck)
	return proof
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgChannelOpenConfirm) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	var proofAck commitmentexported.Proof
	err := unpacker.UnpackAny(&msg.ProofAck, &proofAck)
	if err != nil {
		return fmt.Errorf("proof ack unpack failed: %w", err)
	}

	return nil
}

var _ sdk.Msg = &MsgChannelCloseInit{}

// NewMsgChannelCloseInit creates a new MsgChannelCloseInit instance
func NewMsgChannelCloseInit(
	portID string, channelID string, signer sdk.AccAddress,
) *MsgChannelCloseInit {
	return &MsgChannelCloseInit{
		PortID:    portID,
		ChannelID: channelID,
		Signer:    signer,
	}
}

// Route implements sdk.Msg
func (msg MsgChannelCloseInit) Route() string {
	return host.RouterKey
}

// Type implements sdk.Msg
func (msg MsgChannelCloseInit) Type() string {
	return "channel_close_init"
}

// ValidateBasic implements sdk.Msg
func (msg MsgChannelCloseInit) ValidateBasic() error {
	if err := host.PortIdentifierValidator(msg.PortID); err != nil {
		return sdkerrors.Wrap(err, "invalid port ID")
	}
	if err := host.ChannelIdentifierValidator(msg.ChannelID); err != nil {
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

var (
	_ sdk.Msg                          = &MsgChannelCloseConfirm{}
	_ cdctypes.UnpackInterfacesMessage = &MsgChannelCloseConfirm{}
)

// NewMsgChannelCloseConfirm creates a new MsgChannelCloseConfirm instance
func NewMsgChannelCloseConfirm(
	portID, channelID string, proofInit commitmentexported.Proof, proofHeight uint64,
	signer sdk.AccAddress,
) (*MsgChannelCloseConfirm, error) {
	proofInitAny, err := proofInit.PackAny()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid proof init")
	}

	return &MsgChannelCloseConfirm{
		PortID:      portID,
		ChannelID:   channelID,
		ProofInit:   *proofInitAny,
		ProofHeight: proofHeight,
		Signer:      signer,
	}, nil
}

// Route implements sdk.Msg
func (msg MsgChannelCloseConfirm) Route() string {
	return host.RouterKey
}

// Type implements sdk.Msg
func (msg MsgChannelCloseConfirm) Type() string {
	return "channel_close_confirm"
}

// ValidateBasic implements sdk.Msg
func (msg MsgChannelCloseConfirm) ValidateBasic() error {
	if err := host.PortIdentifierValidator(msg.PortID); err != nil {
		return sdkerrors.Wrap(err, "invalid port ID")
	}
	if err := host.ChannelIdentifierValidator(msg.ChannelID); err != nil {
		return sdkerrors.Wrap(err, "invalid channel ID")
	}
	proofInit := msg.GetProofInit()
	if proofInit == nil || proofInit.IsEmpty() {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if err := proofInit.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof init failed basic validation")
	}
	if msg.ProofHeight == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "proof height must be > 0")
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

// GetProofInit returns the cached value from ProofInit. It returns nil if the value
// is not cached or if the proof doesn't cast to a commitment Proof.
func (msg MsgChannelCloseConfirm) GetProofInit() commitmentexported.Proof {
	proof, _ := commitmenttypes.UnpackAnyProof(&msg.ProofInit)
	return proof
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgChannelCloseConfirm) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	var proofInit commitmentexported.Proof
	err := unpacker.UnpackAny(&msg.ProofInit, &proofInit)
	if err != nil {
		return fmt.Errorf("proof init unpack failed: %w", err)
	}

	return nil
}

var (
	_ sdk.Msg                          = &MsgPacket{}
	_ cdctypes.UnpackInterfacesMessage = &MsgPacket{}
)

// NewMsgPacket constructs new MsgPacket
func NewMsgPacket(
	packet Packet, proof commitmentexported.Proof, proofHeight uint64,
	signer sdk.AccAddress,
) *(MsgPacket, error) {
	proofAny, err := proof.PackAny()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid packet proof")
	}

	return &MsgPacket{
		Packet:      packet,
		Proof:       *proofAny,
		ProofHeight: proofHeight,
		Signer:      signer,
	}, nil
}

// Route implements sdk.Msg
func (msg MsgPacket) Route() string {
	return host.RouterKey
}

// ValidateBasic implements sdk.Msg
func (msg MsgPacket) ValidateBasic() error {
	proof := msg.GetProof()
	if proof == nil || proof.IsEmpty() {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty packet proof")
	}
	if err := proof.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "packet proof failed basic validation")
	}
	if msg.ProofHeight == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "proof height must be > 0")
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

// GetProof returns the cached value from Proof. It returns nil if the value
// is not cached or if the proof doesn't cast to a commitment Proof.
func (msg MsgPacket) GetProof() commitmentexported.Proof {
	proof, _ := commitmenttypes.UnpackAnyProof(&msg.Proof)
	return proof
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgPacket) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	var proof commitmentexported.Proof
	return unpacker.UnpackAny(&msg.Proof, &proof)
}

var (
	_ sdk.Msg                          = &MsgTimeout{}
	_ cdctypes.UnpackInterfacesMessage = &MsgTimeout{}
)

// NewMsgTimeout constructs new MsgTimeout
func NewMsgTimeout(
	packet Packet, nextSequenceRecv uint64, proof commitmentexported.Proof,
	proofHeight uint64, signer sdk.AccAddress,
) (*MsgTimeout, error) {
	proofAny, err := proof.PackAny()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid timeout proof")
	}
	return &MsgTimeout{
		Packet:           packet,
		NextSequenceRecv: nextSequenceRecv,
		Proof:            *proofAny,
		ProofHeight:      proofHeight,
		Signer:           signer,
	}, nil
}

// Route implements sdk.Msg
func (msg MsgTimeout) Route() string {
	return host.RouterKey
}

// ValidateBasic implements sdk.Msg
func (msg MsgTimeout) ValidateBasic() error {
	proof := msg.GetProof()
	if proof == nil || proof.IsEmpty() {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if err := proof.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "timeout proof failed basic validation")
	}
	if msg.ProofHeight == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "proof height must be > 0")
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

// GetProof returns the cached value from Proof. It returns nil if the value
// is not cached or if the proof doesn't cast to a commitment Proof.
func (msg MsgTimeout) GetProof() commitmentexported.Proof {
	proof, _ := commitmenttypes.UnpackAnyProof(&msg.Proof)
	return proof
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgTimeout) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	var proof commitmentexported.Proof
	return unpacker.UnpackAny(&msg.Proof, &proof)
}

var (
	_ sdk.Msg                          = &MsgAcknowledgement{}
	_ cdctypes.UnpackInterfacesMessage = &MsgAcknowledgement{}
)

// NewMsgAcknowledgement constructs a new MsgAcknowledgement
func NewMsgAcknowledgement(
	packet Packet, ack []byte, proof commitmentexported.Proof, proofHeight uint64, signer sdk.AccAddress,
) (*MsgAcknowledgement, error) {
	proofAny, err := proof.PackAny()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid packet proof")
	}

	return &MsgAcknowledgement{
		Packet:          packet,
		Acknowledgement: ack,
		Proof:           *proofAny,
		ProofHeight:     proofHeight,
		Signer:          signer,
	}, nil
}

// Route implements sdk.Msg
func (msg MsgAcknowledgement) Route() string {
	return host.RouterKey
}

// ValidateBasic implements sdk.Msg
func (msg MsgAcknowledgement) ValidateBasic() error {
	proof := msg.GetProof()
	if proof == nil || proof.IsEmpty() {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if err := proof.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "acknowledgement proof failed basic validation")
	}
	if len(msg.Acknowledgement) > 100 {
		return sdkerrors.Wrap(ErrAcknowledgementTooLong, "acknowledgement cannot exceed 100 bytes")
	}
	if msg.ProofHeight == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "proof height must be > 0")
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

// GetProof returns the cached value from Proof. It returns nil if the value
// is not cached or if the proof doesn't cast to a commitment Proof.
func (msg MsgAcknowledgement) GetProof() commitmentexported.Proof {
	proof, _ := commitmenttypes.UnpackAnyProof(&msg.Proof)
	return proof
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgAcknowledgement) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	var proof commitmentexported.Proof
	return unpacker.UnpackAny(&msg.Proof, &proof)
}
