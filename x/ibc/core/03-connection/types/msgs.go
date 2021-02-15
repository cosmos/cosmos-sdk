package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

var (
	_ sdk.Msg = &MsgConnectionOpenInit{}
	_ sdk.Msg = &MsgConnectionOpenConfirm{}
	_ sdk.Msg = &MsgConnectionOpenAck{}
	_ sdk.Msg = &MsgConnectionOpenTry{}

	_ codectypes.UnpackInterfacesMessage = MsgConnectionOpenTry{}
	_ codectypes.UnpackInterfacesMessage = MsgConnectionOpenAck{}
)

// NewMsgConnectionOpenInit creates a new MsgConnectionOpenInit instance. It sets the
// counterparty connection identifier to be empty.
//nolint:interfacer
func NewMsgConnectionOpenInit(
	clientID, counterpartyClientID string,
	counterpartyPrefix commitmenttypes.MerklePrefix,
	version *Version, delayPeriod uint64, signer sdk.AccAddress,
) *MsgConnectionOpenInit {
	// counterparty must have the same delay period
	counterparty := NewCounterparty(counterpartyClientID, "", counterpartyPrefix)
	return &MsgConnectionOpenInit{
		ClientId:     clientID,
		Counterparty: counterparty,
		Version:      version,
		DelayPeriod:  delayPeriod,
		Signer:       signer.String(),
	}
}

// Route implements sdk.Msg
func (msg MsgConnectionOpenInit) Route() string {
	return host.RouterKey
}

// Type implements sdk.Msg
func (msg MsgConnectionOpenInit) Type() string {
	return "connection_open_init"
}

// ValidateBasic implements sdk.Msg.
func (msg MsgConnectionOpenInit) ValidateBasic() error {
	if err := host.ClientIdentifierValidator(msg.ClientId); err != nil {
		return sdkerrors.Wrap(err, "invalid client ID")
	}
	if msg.Counterparty.ConnectionId != "" {
		return sdkerrors.Wrap(ErrInvalidCounterparty, "counterparty connection identifier must be empty")
	}

	// NOTE: Version can be nil on MsgConnectionOpenInit
	if msg.Version != nil {
		if err := ValidateVersion(msg.Version); err != nil {
			return sdkerrors.Wrap(err, "basic validation of the provided version failed")
		}
	}
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "string could not be parsed as address: %v", err)
	}
	return msg.Counterparty.ValidateBasic()
}

// GetSignBytes implements sdk.Msg. The function will panic since it is used
// for amino transaction verification which IBC does not support.
func (msg MsgConnectionOpenInit) GetSignBytes() []byte {
	panic("IBC messages do not support amino")
}

// GetSigners implements sdk.Msg
func (msg MsgConnectionOpenInit) GetSigners() []sdk.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{accAddr}
}

// NewMsgConnectionOpenTry creates a new MsgConnectionOpenTry instance
//nolint:interfacer
func NewMsgConnectionOpenTry(
	previousConnectionID, clientID, counterpartyConnectionID,
	counterpartyClientID string, counterpartyClient exported.ClientState,
	counterpartyPrefix commitmenttypes.MerklePrefix,
	counterpartyVersions []*Version, delayPeriod uint64,
	proofInit, proofClient, proofConsensus []byte,
	proofHeight, consensusHeight clienttypes.Height, signer sdk.AccAddress,
) *MsgConnectionOpenTry {
	counterparty := NewCounterparty(counterpartyClientID, counterpartyConnectionID, counterpartyPrefix)
	csAny, _ := clienttypes.PackClientState(counterpartyClient)
	return &MsgConnectionOpenTry{
		PreviousConnectionId: previousConnectionID,
		ClientId:             clientID,
		ClientState:          csAny,
		Counterparty:         counterparty,
		CounterpartyVersions: counterpartyVersions,
		DelayPeriod:          delayPeriod,
		ProofInit:            proofInit,
		ProofClient:          proofClient,
		ProofConsensus:       proofConsensus,
		ProofHeight:          proofHeight,
		ConsensusHeight:      consensusHeight,
		Signer:               signer.String(),
	}
}

// Route implements sdk.Msg
func (msg MsgConnectionOpenTry) Route() string {
	return host.RouterKey
}

// Type implements sdk.Msg
func (msg MsgConnectionOpenTry) Type() string {
	return "connection_open_try"
}

// ValidateBasic implements sdk.Msg
func (msg MsgConnectionOpenTry) ValidateBasic() error {
	// an empty connection identifier indicates that a connection identifier should be generated
	if msg.PreviousConnectionId != "" {
		if !IsValidConnectionID(msg.PreviousConnectionId) {
			return sdkerrors.Wrap(ErrInvalidConnectionIdentifier, "invalid previous connection ID")
		}
	}
	if err := host.ClientIdentifierValidator(msg.ClientId); err != nil {
		return sdkerrors.Wrap(err, "invalid client ID")
	}
	// counterparty validate basic allows empty counterparty connection identifiers
	if err := host.ConnectionIdentifierValidator(msg.Counterparty.ConnectionId); err != nil {
		return sdkerrors.Wrap(err, "invalid counterparty connection ID")
	}
	if msg.ClientState == nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidClient, "counterparty client is nil")
	}
	clientState, err := clienttypes.UnpackClientState(msg.ClientState)
	if err != nil {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidClient, "unpack err: %v", err)
	}
	if err := clientState.Validate(); err != nil {
		return sdkerrors.Wrap(err, "counterparty client is invalid")
	}
	if len(msg.CounterpartyVersions) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidVersion, "empty counterparty versions")
	}
	for i, version := range msg.CounterpartyVersions {
		if err := ValidateVersion(version); err != nil {
			return sdkerrors.Wrapf(err, "basic validation failed on version with index %d", i)
		}
	}
	if len(msg.ProofInit) == 0 {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof init")
	}
	if len(msg.ProofClient) == 0 {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit empty proof client")
	}
	if len(msg.ProofConsensus) == 0 {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof of consensus state")
	}
	if msg.ProofHeight.IsZero() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "proof height must be non-zero")
	}
	if msg.ConsensusHeight.IsZero() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "consensus height must be non-zero")
	}
	_, err = sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "string could not be parsed as address: %v", err)
	}
	return msg.Counterparty.ValidateBasic()
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgConnectionOpenTry) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return unpacker.UnpackAny(msg.ClientState, new(exported.ClientState))
}

// GetSignBytes implements sdk.Msg. The function will panic since it is used
// for amino transaction verification which IBC does not support.
func (msg MsgConnectionOpenTry) GetSignBytes() []byte {
	panic("IBC messages do not support amino")
}

// GetSigners implements sdk.Msg
func (msg MsgConnectionOpenTry) GetSigners() []sdk.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{accAddr}
}

// NewMsgConnectionOpenAck creates a new MsgConnectionOpenAck instance
//nolint:interfacer
func NewMsgConnectionOpenAck(
	connectionID, counterpartyConnectionID string, counterpartyClient exported.ClientState,
	proofTry, proofClient, proofConsensus []byte,
	proofHeight, consensusHeight clienttypes.Height,
	version *Version,
	signer sdk.AccAddress,
) *MsgConnectionOpenAck {
	csAny, _ := clienttypes.PackClientState(counterpartyClient)
	return &MsgConnectionOpenAck{
		ConnectionId:             connectionID,
		CounterpartyConnectionId: counterpartyConnectionID,
		ClientState:              csAny,
		ProofTry:                 proofTry,
		ProofClient:              proofClient,
		ProofConsensus:           proofConsensus,
		ProofHeight:              proofHeight,
		ConsensusHeight:          consensusHeight,
		Version:                  version,
		Signer:                   signer.String(),
	}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgConnectionOpenAck) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return unpacker.UnpackAny(msg.ClientState, new(exported.ClientState))
}

// Route implements sdk.Msg
func (msg MsgConnectionOpenAck) Route() string {
	return host.RouterKey
}

// Type implements sdk.Msg
func (msg MsgConnectionOpenAck) Type() string {
	return "connection_open_ack"
}

// ValidateBasic implements sdk.Msg
func (msg MsgConnectionOpenAck) ValidateBasic() error {
	if !IsValidConnectionID(msg.ConnectionId) {
		return ErrInvalidConnectionIdentifier
	}
	if err := host.ConnectionIdentifierValidator(msg.CounterpartyConnectionId); err != nil {
		return sdkerrors.Wrap(err, "invalid counterparty connection ID")
	}
	if err := ValidateVersion(msg.Version); err != nil {
		return err
	}
	if msg.ClientState == nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidClient, "counterparty client is nil")
	}
	clientState, err := clienttypes.UnpackClientState(msg.ClientState)
	if err != nil {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidClient, "unpack err: %v", err)
	}
	if err := clientState.Validate(); err != nil {
		return sdkerrors.Wrap(err, "counterparty client is invalid")
	}
	if len(msg.ProofTry) == 0 {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof try")
	}
	if len(msg.ProofClient) == 0 {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit empty proof client")
	}
	if len(msg.ProofConsensus) == 0 {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof of consensus state")
	}
	if msg.ProofHeight.IsZero() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "proof height must be non-zero")
	}
	if msg.ConsensusHeight.IsZero() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "consensus height must be non-zero")
	}
	_, err = sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "string could not be parsed as address: %v", err)
	}
	return nil
}

// GetSignBytes implements sdk.Msg. The function will panic since it is used
// for amino transaction verification which IBC does not support.
func (msg MsgConnectionOpenAck) GetSignBytes() []byte {
	panic("IBC messages do not support amino")
}

// GetSigners implements sdk.Msg
func (msg MsgConnectionOpenAck) GetSigners() []sdk.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{accAddr}
}

// NewMsgConnectionOpenConfirm creates a new MsgConnectionOpenConfirm instance
//nolint:interfacer
func NewMsgConnectionOpenConfirm(
	connectionID string, proofAck []byte, proofHeight clienttypes.Height,
	signer sdk.AccAddress,
) *MsgConnectionOpenConfirm {
	return &MsgConnectionOpenConfirm{
		ConnectionId: connectionID,
		ProofAck:     proofAck,
		ProofHeight:  proofHeight,
		Signer:       signer.String(),
	}
}

// Route implements sdk.Msg
func (msg MsgConnectionOpenConfirm) Route() string {
	return host.RouterKey
}

// Type implements sdk.Msg
func (msg MsgConnectionOpenConfirm) Type() string {
	return "connection_open_confirm"
}

// ValidateBasic implements sdk.Msg
func (msg MsgConnectionOpenConfirm) ValidateBasic() error {
	if !IsValidConnectionID(msg.ConnectionId) {
		return ErrInvalidConnectionIdentifier
	}
	if len(msg.ProofAck) == 0 {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof ack")
	}
	if msg.ProofHeight.IsZero() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "proof height must be non-zero")
	}
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "string could not be parsed as address: %v", err)
	}
	return nil
}

// GetSignBytes implements sdk.Msg. The function will panic since it is used
// for amino transaction verification which IBC does not support.
func (msg MsgConnectionOpenConfirm) GetSignBytes() []byte {
	panic("IBC messages do not support amino")
}

// GetSigners implements sdk.Msg
func (msg MsgConnectionOpenConfirm) GetSigners() []sdk.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{accAddr}
}
