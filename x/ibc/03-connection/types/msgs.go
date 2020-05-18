package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

var _ sdk.Msg = MsgConnectionOpenInit{}

// NewMsgConnectionOpenInit creates a new MsgConnectionOpenInit instance
func NewMsgConnectionOpenInit(
	connectionID, clientID, counterpartyConnectionID,
	counterpartyClientID string, counterpartyPrefix commitmenttypes.MerklePrefix,
	signer sdk.AccAddress,
) MsgConnectionOpenInit {
	counterparty := NewCounterparty(counterpartyClientID, counterpartyConnectionID, counterpartyPrefix)
	return MsgConnectionOpenInit{
		ConnectionID: connectionID,
		ClientID:     clientID,
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
	if err := host.ConnectionIdentifierValidator(msg.ConnectionID); err != nil {
		return sdkerrors.Wrapf(err, "invalid connection ID: %s", msg.ConnectionID)
	}
	if err := host.ClientIdentifierValidator(msg.ClientID); err != nil {
		return sdkerrors.Wrapf(err, "invalid client ID: %s", msg.ClientID)
	}
	if msg.Signer.Empty() {
		return sdkerrors.ErrInvalidAddress
	}
	return msg.Counterparty.ValidateBasic()
}

// GetSignBytes implements sdk.Msg
func (msg MsgConnectionOpenInit) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgConnectionOpenInit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

var _ sdk.Msg = MsgConnectionOpenTry{}

// NewMsgConnectionOpenTry creates a new MsgConnectionOpenTry instance
func NewMsgConnectionOpenTry(
	connectionID, clientID, counterpartyConnectionID,
	counterpartyClientID string, counterpartyPrefix commitmenttypes.MerklePrefix,
	counterpartyVersions []string, proofInit, proofConsensus commitmenttypes.MerkleProof,
	proofHeight, consensusHeight uint64, signer sdk.AccAddress,
) MsgConnectionOpenTry {
	counterparty := NewCounterparty(counterpartyClientID, counterpartyConnectionID, counterpartyPrefix)
	return MsgConnectionOpenTry{
		ConnectionID:         connectionID,
		ClientID:             clientID,
		Counterparty:         counterparty,
		CounterpartyVersions: counterpartyVersions,
		ProofInit:            proofInit,
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
	if err := host.ConnectionIdentifierValidator(msg.ConnectionID); err != nil {
		return sdkerrors.Wrapf(err, "invalid connection ID: %s", msg.ConnectionID)
	}
	if err := host.ClientIdentifierValidator(msg.ClientID); err != nil {
		return sdkerrors.Wrapf(err, "invalid client ID: %s", msg.ClientID)
	}
	if len(msg.CounterpartyVersions) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidVersion, "missing counterparty versions")
	}
	for _, version := range msg.CounterpartyVersions {
		if strings.TrimSpace(version) == "" {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidVersion, "version can't be blank")
		}
	}
	if msg.ProofInit.IsEmpty() || msg.ProofConsensus.IsEmpty() {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if err := msg.ProofInit.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof init cannot be nil")
	}
	if err := msg.ProofConsensus.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof consensus cannot be nil")
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
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgConnectionOpenTry) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

var _ sdk.Msg = MsgConnectionOpenAck{}

// NewMsgConnectionOpenAck creates a new MsgConnectionOpenAck instance
func NewMsgConnectionOpenAck(
	connectionID string, proofTry, proofConsensus commitmenttypes.MerkleProof,
	proofHeight, consensusHeight uint64, version string,
	signer sdk.AccAddress,
) MsgConnectionOpenAck {
	return MsgConnectionOpenAck{
		ConnectionID:    connectionID,
		ProofTry:        proofTry,
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
	if err := host.ConnectionIdentifierValidator(msg.ConnectionID); err != nil {
		return sdkerrors.Wrap(err, "invalid connection ID")
	}
	if strings.TrimSpace(msg.Version) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidVersion, "version can't be blank")
	}
	if msg.ProofTry.IsEmpty() || msg.ProofConsensus.IsEmpty() {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if err := msg.ProofTry.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof try cannot be nil")
	}
	if err := msg.ProofConsensus.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof consensus cannot be nil")
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
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgConnectionOpenAck) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

var _ sdk.Msg = MsgConnectionOpenConfirm{}

// NewMsgConnectionOpenConfirm creates a new MsgConnectionOpenConfirm instance
func NewMsgConnectionOpenConfirm(
	connectionID string, proofAck commitmenttypes.MerkleProof, proofHeight uint64,
	signer sdk.AccAddress,
) MsgConnectionOpenConfirm {
	return MsgConnectionOpenConfirm{
		ConnectionID: connectionID,
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
	if err := host.ConnectionIdentifierValidator(msg.ConnectionID); err != nil {
		return sdkerrors.Wrap(err, "invalid connection ID")
	}
	if msg.ProofAck.IsEmpty() {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if err := msg.ProofAck.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof ack cannot be nil")
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
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgConnectionOpenConfirm) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
