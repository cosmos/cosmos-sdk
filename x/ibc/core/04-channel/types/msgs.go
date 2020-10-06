package types

import (
	"encoding/base64"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
)

var _ sdk.Msg = &MsgChannelOpenInit{}

// NewMsgChannelOpenInit creates a new MsgChannelOpenInit
//nolint:interfacer
func NewMsgChannelOpenInit(
	portID, channelID string, version string, channelOrder Order, connectionHops []string,
	counterpartyPortID, counterpartyChannelID string, signer sdk.AccAddress,
) *MsgChannelOpenInit {
	counterparty := NewCounterparty(counterpartyPortID, counterpartyChannelID)
	channel := NewChannel(INIT, channelOrder, counterparty, connectionHops, version)
	return &MsgChannelOpenInit{
		PortId:    portID,
		ChannelId: channelID,
		Channel:   channel,
		Signer:    signer.String(),
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
	if err := host.PortIdentifierValidator(msg.PortId); err != nil {
		return sdkerrors.Wrap(err, "invalid port ID")
	}
	if err := host.ChannelIdentifierValidator(msg.ChannelId); err != nil {
		return sdkerrors.Wrap(err, "invalid channel ID")
	}
	// Signer can be empty
	return msg.Channel.ValidateBasic()
}

// GetSignBytes implements sdk.Msg
func (msg MsgChannelOpenInit) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners implements sdk.Msg
func (msg MsgChannelOpenInit) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

var _ sdk.Msg = &MsgChannelOpenTry{}

// NewMsgChannelOpenTry creates a new MsgChannelOpenTry instance
//nolint:interfacer
func NewMsgChannelOpenTry(
	portID, desiredChannelID, counterpartyChosenChannelID, version string, channelOrder Order, connectionHops []string,
	counterpartyPortID, counterpartyChannelID, counterpartyVersion string,
	proofInit []byte, proofHeight clienttypes.Height, signer sdk.AccAddress,
) *MsgChannelOpenTry {
	counterparty := NewCounterparty(counterpartyPortID, counterpartyChannelID)
	channel := NewChannel(INIT, channelOrder, counterparty, connectionHops, version)
	return &MsgChannelOpenTry{
		PortId:                      portID,
		DesiredChannelId:            desiredChannelID,
		CounterpartyChosenChannelId: counterpartyChosenChannelID,
		Channel:                     channel,
		CounterpartyVersion:         counterpartyVersion,
		ProofInit:                   proofInit,
		ProofHeight:                 proofHeight,
		Signer:                      signer.String(),
	}
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
	if err := host.PortIdentifierValidator(msg.PortId); err != nil {
		return sdkerrors.Wrap(err, "invalid port ID")
	}
	if err := host.ChannelIdentifierValidator(msg.DesiredChannelId); err != nil {
		return sdkerrors.Wrap(err, "invalid desired channel ID")
	}
	if msg.CounterpartyChosenChannelId != "" && msg.CounterpartyChosenChannelId != msg.DesiredChannelId {
		return sdkerrors.Wrap(ErrInvalidChannelIdentifier, "counterparty chosen channel ID must be empty or equal to desired channel ID")
	}
	if len(msg.ProofInit) == 0 {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof init")
	}
	if msg.ProofHeight.IsZero() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "proof height must be non-zero")
	}
	// Signer can be empty
	return msg.Channel.ValidateBasic()
}

// GetSignBytes implements sdk.Msg
func (msg MsgChannelOpenTry) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners implements sdk.Msg
func (msg MsgChannelOpenTry) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

var _ sdk.Msg = &MsgChannelOpenAck{}

// NewMsgChannelOpenAck creates a new MsgChannelOpenAck instance
//nolint:interfacer
func NewMsgChannelOpenAck(
	portID, channelID, counterpartyChannelID string, cpv string, proofTry []byte, proofHeight clienttypes.Height,
	signer sdk.AccAddress,
) *MsgChannelOpenAck {
	return &MsgChannelOpenAck{
		PortId:                portID,
		ChannelId:             channelID,
		CounterpartyChannelId: counterpartyChannelID,
		CounterpartyVersion:   cpv,
		ProofTry:              proofTry,
		ProofHeight:           proofHeight,
		Signer:                signer.String(),
	}
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
	if err := host.PortIdentifierValidator(msg.PortId); err != nil {
		return sdkerrors.Wrap(err, "invalid port ID")
	}
	if err := host.ChannelIdentifierValidator(msg.ChannelId); err != nil {
		return sdkerrors.Wrap(err, "invalid channel ID")
	}
	if err := host.ChannelIdentifierValidator(msg.CounterpartyChannelId); err != nil {
		return sdkerrors.Wrap(err, "invalid counterparty channel ID")
	}
	if len(msg.ProofTry) == 0 {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof try")
	}
	if msg.ProofHeight.IsZero() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "proof height must be non-zero")
	}
	// Signer can be empty
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgChannelOpenAck) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners implements sdk.Msg
func (msg MsgChannelOpenAck) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

var _ sdk.Msg = &MsgChannelOpenConfirm{}

// NewMsgChannelOpenConfirm creates a new MsgChannelOpenConfirm instance
//nolint:interfacer
func NewMsgChannelOpenConfirm(
	portID, channelID string, proofAck []byte, proofHeight clienttypes.Height,
	signer sdk.AccAddress,
) *MsgChannelOpenConfirm {
	return &MsgChannelOpenConfirm{
		PortId:      portID,
		ChannelId:   channelID,
		ProofAck:    proofAck,
		ProofHeight: proofHeight,
		Signer:      signer.String(),
	}
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
	if err := host.PortIdentifierValidator(msg.PortId); err != nil {
		return sdkerrors.Wrap(err, "invalid port ID")
	}
	if err := host.ChannelIdentifierValidator(msg.ChannelId); err != nil {
		return sdkerrors.Wrap(err, "invalid channel ID")
	}
	if len(msg.ProofAck) == 0 {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof ack")
	}
	if msg.ProofHeight.IsZero() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "proof height must be non-zero")
	}
	// Signer can be empty
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgChannelOpenConfirm) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners implements sdk.Msg
func (msg MsgChannelOpenConfirm) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

var _ sdk.Msg = &MsgChannelCloseInit{}

// NewMsgChannelCloseInit creates a new MsgChannelCloseInit instance
//nolint:interfacer
func NewMsgChannelCloseInit(
	portID string, channelID string, signer sdk.AccAddress,
) *MsgChannelCloseInit {
	return &MsgChannelCloseInit{
		PortId:    portID,
		ChannelId: channelID,
		Signer:    signer.String(),
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
	if err := host.PortIdentifierValidator(msg.PortId); err != nil {
		return sdkerrors.Wrap(err, "invalid port ID")
	}
	if err := host.ChannelIdentifierValidator(msg.ChannelId); err != nil {
		return sdkerrors.Wrap(err, "invalid channel ID")
	}
	// Signer can be empty
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgChannelCloseInit) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners implements sdk.Msg
func (msg MsgChannelCloseInit) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

var _ sdk.Msg = &MsgChannelCloseConfirm{}

// NewMsgChannelCloseConfirm creates a new MsgChannelCloseConfirm instance
//nolint:interfacer
func NewMsgChannelCloseConfirm(
	portID, channelID string, proofInit []byte, proofHeight clienttypes.Height,
	signer sdk.AccAddress,
) *MsgChannelCloseConfirm {
	return &MsgChannelCloseConfirm{
		PortId:      portID,
		ChannelId:   channelID,
		ProofInit:   proofInit,
		ProofHeight: proofHeight,
		Signer:      signer.String(),
	}
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
	if err := host.PortIdentifierValidator(msg.PortId); err != nil {
		return sdkerrors.Wrap(err, "invalid port ID")
	}
	if err := host.ChannelIdentifierValidator(msg.ChannelId); err != nil {
		return sdkerrors.Wrap(err, "invalid channel ID")
	}
	if len(msg.ProofInit) == 0 {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof init")
	}
	if msg.ProofHeight.IsZero() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "proof height must be non-zero")
	}
	// Signer can be empty
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgChannelCloseConfirm) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners implements sdk.Msg
func (msg MsgChannelCloseConfirm) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

var _ sdk.Msg = &MsgRecvPacket{}

// NewMsgRecvPacket constructs new MsgRecvPacket
//nolint:interfacer
func NewMsgRecvPacket(
	packet Packet, proof []byte, proofHeight clienttypes.Height,
	signer sdk.AccAddress,
) *MsgRecvPacket {
	return &MsgRecvPacket{
		Packet:      packet,
		Proof:       proof,
		ProofHeight: proofHeight,
		Signer:      signer.String(),
	}
}

// Route implements sdk.Msg
func (msg MsgRecvPacket) Route() string {
	return host.RouterKey
}

// ValidateBasic implements sdk.Msg
func (msg MsgRecvPacket) ValidateBasic() error {
	if len(msg.Proof) == 0 {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if msg.ProofHeight.IsZero() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "proof height must be non-zero")
	}
	if msg.Signer == "" {
		return sdkerrors.ErrInvalidAddress
	}

	return msg.Packet.ValidateBasic()
}

// GetSignBytes implements sdk.Msg
func (msg MsgRecvPacket) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(&msg))
}

// GetDataSignBytes returns the base64-encoded bytes used for the
// data field when signing the packet.
func (msg MsgRecvPacket) GetDataSignBytes() []byte {
	s := "\"" + base64.StdEncoding.EncodeToString(msg.Packet.Data) + "\""
	return []byte(s)
}

// GetSigners implements sdk.Msg
func (msg MsgRecvPacket) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

// Type implements sdk.Msg
func (msg MsgRecvPacket) Type() string {
	return "recv_packet"
}

var _ sdk.Msg = &MsgTimeout{}

// NewMsgTimeout constructs new MsgTimeout
//nolint:interfacer
func NewMsgTimeout(
	packet Packet, nextSequenceRecv uint64, proof []byte,
	proofHeight clienttypes.Height, signer sdk.AccAddress,
) *MsgTimeout {
	return &MsgTimeout{
		Packet:           packet,
		NextSequenceRecv: nextSequenceRecv,
		Proof:            proof,
		ProofHeight:      proofHeight,
		Signer:           signer.String(),
	}
}

// Route implements sdk.Msg
func (msg MsgTimeout) Route() string {
	return host.RouterKey
}

// ValidateBasic implements sdk.Msg
func (msg MsgTimeout) ValidateBasic() error {
	if len(msg.Proof) == 0 {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if msg.ProofHeight.IsZero() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "proof height must be non-zero")
	}
	if msg.Signer == "" {
		return sdkerrors.ErrInvalidAddress
	}

	return msg.Packet.ValidateBasic()
}

// GetSignBytes implements sdk.Msg
func (msg MsgTimeout) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners implements sdk.Msg
func (msg MsgTimeout) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

// Type implements sdk.Msg
func (msg MsgTimeout) Type() string {
	return "timeout_packet"
}

// NewMsgTimeoutOnClose constructs new MsgTimeoutOnClose
//nolint:interfacer
func NewMsgTimeoutOnClose(
	packet Packet, nextSequenceRecv uint64,
	proof, proofClose []byte,
	proofHeight clienttypes.Height, signer sdk.AccAddress,
) *MsgTimeoutOnClose {
	return &MsgTimeoutOnClose{
		Packet:           packet,
		NextSequenceRecv: nextSequenceRecv,
		Proof:            proof,
		ProofClose:       proofClose,
		ProofHeight:      proofHeight,
		Signer:           signer.String(),
	}
}

// Route implements sdk.Msg
func (msg MsgTimeoutOnClose) Route() string {
	return host.RouterKey
}

// ValidateBasic implements sdk.Msg
func (msg MsgTimeoutOnClose) ValidateBasic() error {
	if len(msg.Proof) == 0 {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if len(msg.ProofClose) == 0 {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof of closed counterparty channel end")
	}
	if msg.ProofHeight.IsZero() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "proof height must be non-zero")
	}
	if msg.Signer == "" {
		return sdkerrors.ErrInvalidAddress
	}

	return msg.Packet.ValidateBasic()
}

// GetSignBytes implements sdk.Msg
func (msg MsgTimeoutOnClose) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners implements sdk.Msg
func (msg MsgTimeoutOnClose) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

// Type implements sdk.Msg
func (msg MsgTimeoutOnClose) Type() string {
	return "timeout_on_close_packet"
}

var _ sdk.Msg = &MsgAcknowledgement{}

// NewMsgAcknowledgement constructs a new MsgAcknowledgement
//nolint:interfacer
func NewMsgAcknowledgement(
	packet Packet, ack []byte, proof []byte, proofHeight clienttypes.Height, signer sdk.AccAddress) *MsgAcknowledgement {
	return &MsgAcknowledgement{
		Packet:          packet,
		Acknowledgement: ack,
		Proof:           proof,
		ProofHeight:     proofHeight,
		Signer:          signer.String(),
	}
}

// Route implements sdk.Msg
func (msg MsgAcknowledgement) Route() string {
	return host.RouterKey
}

// ValidateBasic implements sdk.Msg
func (msg MsgAcknowledgement) ValidateBasic() error {
	if len(msg.Proof) == 0 {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if msg.ProofHeight.IsZero() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "proof height must be non-zero")
	}
	if msg.Signer == "" {
		return sdkerrors.ErrInvalidAddress
	}

	return msg.Packet.ValidateBasic()
}

// GetSignBytes implements sdk.Msg
func (msg MsgAcknowledgement) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners implements sdk.Msg
func (msg MsgAcknowledgement) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

// Type implements sdk.Msg
func (msg MsgAcknowledgement) Type() string {
	return "acknowledge_packet"
}
