package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

var _ sdk.Msg = &MsgConnectionOpenInit{}

// NewMsgConnectionOpenInit creates a new MsgConnectionOpenInit instance
func NewMsgConnectionOpenInit(
	connectionID, clientID, counterpartyConnectionID,
	counterpartyClientID string, counterpartyPrefix commitmenttypes.MerklePrefix,
	signer sdk.AccAddress,
) *MsgConnectionOpenInit {
	counterparty := NewCounterparty(counterpartyClientID, counterpartyConnectionID, counterpartyPrefix)
	return &MsgConnectionOpenInit{
		ConnectionId: connectionID,
		ClientId:     clientID,
		Counterparty: counterparty,
		Signer:       signer,
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

// ValidateBasic implements sdk.Msg
func (msg MsgConnectionOpenInit) ValidateBasic() error {
	if err := host.ConnectionIdentifierValidator(msg.ConnectionId); err != nil {
		return sdkerrors.Wrap(err, "invalid connection ID")
	}
	if err := host.ClientIdentifierValidator(msg.ClientId); err != nil {
		return sdkerrors.Wrap(err, "invalid client ID")
	}
	if msg.Signer.Empty() {
		return sdkerrors.ErrInvalidAddress
	}
	return msg.Counterparty.ValidateBasic()
}

// GetSignBytes implements sdk.Msg
func (msg MsgConnectionOpenInit) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners implements sdk.Msg
func (msg MsgConnectionOpenInit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

var _ sdk.Msg = &MsgConnectionOpenTry{}

// NewMsgConnectionOpenTry creates a new MsgConnectionOpenTry instance
func NewMsgConnectionOpenTry(
	connectionID, clientID, counterpartyConnectionID,
	counterpartyClientID string, counterpartyClient clientexported.ClientState,
	counterpartyPrefix commitmenttypes.MerklePrefix, counterpartyVersions []string,
	proofInit, proofClient, proofConsensus []byte,
	proofHeight, consensusHeight uint64, signer sdk.AccAddress,
) *MsgConnectionOpenTry {
	counterparty := NewCounterparty(counterpartyClientID, counterpartyConnectionID, counterpartyPrefix)
	csAny, _ := clienttypes.PackClientState(counterpartyClient)
	return &MsgConnectionOpenTry{
		ConnectionId:         connectionID,
		ClientId:             clientID,
		ClientState:          csAny,
		Counterparty:         counterparty,
		CounterpartyVersions: counterpartyVersions,
		ProofInit:            proofInit,
		ProofClient:          proofClient,
		ProofConsensus:       proofConsensus,
		ProofHeight:          proofHeight,
		ConsensusHeight:      consensusHeight,
		Signer:               signer,
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
	if err := host.ConnectionIdentifierValidator(msg.ConnectionId); err != nil {
		return sdkerrors.Wrap(err, "invalid connection ID")
	}
	if err := host.ClientIdentifierValidator(msg.ClientId); err != nil {
		return sdkerrors.Wrap(err, "invalid client ID")
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
	if msg.ProofHeight == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "proof height must be > 0")
	}
	if msg.ConsensusHeight == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "consensus height must be > 0")
	}
	if msg.Signer.Empty() {
		return sdkerrors.ErrInvalidAddress
	}
	return msg.Counterparty.ValidateBasic()
}

// GetSignBytes implements sdk.Msg
func (msg MsgConnectionOpenTry) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners implements sdk.Msg
func (msg MsgConnectionOpenTry) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

var _ sdk.Msg = &MsgConnectionOpenAck{}

// NewMsgConnectionOpenAck creates a new MsgConnectionOpenAck instance
func NewMsgConnectionOpenAck(
	connectionID string, counterpartyClient clientexported.ClientState,
	proofTry, proofClient, proofConsensus []byte,
	proofHeight, consensusHeight uint64, version string,
	signer sdk.AccAddress,
) *MsgConnectionOpenAck {
	csAny, _ := clienttypes.PackClientState(counterpartyClient)
	return &MsgConnectionOpenAck{
		ConnectionId:    connectionID,
		ClientState:     csAny,
		ProofTry:        proofTry,
		ProofClient:     proofClient,
		ProofConsensus:  proofConsensus,
		ProofHeight:     proofHeight,
		ConsensusHeight: consensusHeight,
		Version:         version,
		Signer:          signer,
	}
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
	if err := host.ConnectionIdentifierValidator(msg.ConnectionId); err != nil {
		return sdkerrors.Wrap(err, "invalid connection ID")
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
	if msg.ProofHeight == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "proof height must be > 0")
	}
	if msg.ConsensusHeight == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "consensus height must be > 0")
	}
	if msg.Signer.Empty() {
		return sdkerrors.ErrInvalidAddress
	}
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgConnectionOpenAck) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners implements sdk.Msg
func (msg MsgConnectionOpenAck) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

var _ sdk.Msg = &MsgConnectionOpenConfirm{}

// NewMsgConnectionOpenConfirm creates a new MsgConnectionOpenConfirm instance
func NewMsgConnectionOpenConfirm(
	connectionID string, proofAck []byte, proofHeight uint64,
	signer sdk.AccAddress,
) *MsgConnectionOpenConfirm {
	return &MsgConnectionOpenConfirm{
		ConnectionId: connectionID,
		ProofAck:     proofAck,
		ProofHeight:  proofHeight,
		Signer:       signer,
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
	if err := host.ConnectionIdentifierValidator(msg.ConnectionId); err != nil {
		return sdkerrors.Wrap(err, "invalid connection ID")
	}
	if len(msg.ProofAck) == 0 {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof ack")
	}
	if msg.ProofHeight == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "proof height must be > 0")
	}
	if msg.Signer.Empty() {
		return sdkerrors.ErrInvalidAddress
	}
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgConnectionOpenConfirm) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners implements sdk.Msg
func (msg MsgConnectionOpenConfirm) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
