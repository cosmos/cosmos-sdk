package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

var _ sdk.Msg = MsgConnectionOpenInit{}

// MsgConnectionOpenInit defines the msg sent by an account on Chain A to
// initialize a connection with Chain B.
type MsgConnectionOpenInit struct {
	ConnectionID string         `json:"connection_id"`
	ClientID     string         `json:"client_id"`
	Counterparty Counterparty   `json:"counterparty"`
	Signer       sdk.AccAddress `json:"signer"`
}

// NewMsgConnectionOpenInit creates a new MsgConnectionOpenInit instance
func NewMsgConnectionOpenInit(
	connectionID, clientID, counterpartyConnectionID,
	counterpartyClientID string, counterpartyPrefix commitmentexported.Prefix,
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
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgConnectionOpenInit) Type() string {
	return "connection_open_init"
}

// ValidateBasic implements sdk.Msg
func (msg MsgConnectionOpenInit) ValidateBasic() error {
	if err := host.DefaultConnectionIdentifierValidator(msg.ConnectionID); err != nil {
		return sdkerrors.Wrapf(err, "invalid connection ID: %s", msg.ConnectionID)
	}
	if err := host.DefaultClientIdentifierValidator(msg.ClientID); err != nil {
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

// MsgConnectionOpenTry defines a msg sent by a Relayer to try to open a connection
// on Chain B.
type MsgConnectionOpenTry struct {
	ConnectionID         string                   `json:"connection_id"`
	ClientID             string                   `json:"client_id"`
	Counterparty         Counterparty             `json:"counterparty"`
	CounterpartyVersions []string                 `json:"counterparty_versions"`
	ProofInit            commitmentexported.Proof `json:"proof_init"`      // proof of the initialization the connection on Chain A: `none -> INIT`
	ProofConsensus       commitmentexported.Proof `json:"proof_consensus"` // proof of client consensus state
	ProofHeight          uint64                   `json:"proof_height"`
	ConsensusHeight      uint64                   `json:"consensus_height"`
	Signer               sdk.AccAddress           `json:"signer"`
}

// NewMsgConnectionOpenTry creates a new MsgConnectionOpenTry instance
func NewMsgConnectionOpenTry(
	connectionID, clientID, counterpartyConnectionID,
	counterpartyClientID string, counterpartyPrefix commitmentexported.Prefix,
	counterpartyVersions []string, proofInit, proofConsensus commitmentexported.Proof,
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
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgConnectionOpenTry) Type() string {
	return "connection_open_try"
}

// ValidateBasic implements sdk.Msg
func (msg MsgConnectionOpenTry) ValidateBasic() error {
	if err := host.DefaultConnectionIdentifierValidator(msg.ConnectionID); err != nil {
		return sdkerrors.Wrapf(err, "invalid connection ID: %s", msg.ConnectionID)
	}
	if err := host.DefaultClientIdentifierValidator(msg.ClientID); err != nil {
		return sdkerrors.Wrapf(err, "invalid client ID: %s", msg.ClientID)
	}
	if len(msg.CounterpartyVersions) == 0 {
		return sdkerrors.Wrap(ibctypes.ErrInvalidVersion, "missing counterparty versions")
	}
	for _, version := range msg.CounterpartyVersions {
		if strings.TrimSpace(version) == "" {
			return sdkerrors.Wrap(ibctypes.ErrInvalidVersion, "version can't be blank")
		}
	}
	if msg.ProofInit == nil || msg.ProofConsensus == nil {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if err := msg.ProofInit.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof init cannot be nil")
	}
	if err := msg.ProofConsensus.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof consensus cannot be nil")
	}
	if msg.ProofHeight == 0 {
		return sdkerrors.Wrap(ibctypes.ErrInvalidHeight, "proof height must be > 0")
	}
	if msg.ConsensusHeight == 0 {
		return sdkerrors.Wrap(ibctypes.ErrInvalidHeight, "consensus height must be > 0")
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

// MsgConnectionOpenAck defines a msg sent by a Relayer to Chain A to acknowledge
// the change of connection state to TRYOPEN on Chain B.
type MsgConnectionOpenAck struct {
	ConnectionID    string                   `json:"connection_id"`
	ProofTry        commitmentexported.Proof `json:"proof_try"`       // proof for the change of the connection state on Chain B: `none -> TRYOPEN`
	ProofConsensus  commitmentexported.Proof `json:"proof_consensus"` // proof of client consensus state
	ProofHeight     uint64                   `json:"proof_height"`
	ConsensusHeight uint64                   `json:"consensus_height"`
	Version         string                   `json:"version"`
	Signer          sdk.AccAddress           `json:"signer"`
}

// NewMsgConnectionOpenAck creates a new MsgConnectionOpenAck instance
func NewMsgConnectionOpenAck(
	connectionID string, proofTry, proofConsensus commitmentexported.Proof,
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
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgConnectionOpenAck) Type() string {
	return "connection_open_ack"
}

// ValidateBasic implements sdk.Msg
func (msg MsgConnectionOpenAck) ValidateBasic() error {
	if err := host.DefaultConnectionIdentifierValidator(msg.ConnectionID); err != nil {
		return sdkerrors.Wrap(err, "invalid connection ID")
	}
	if strings.TrimSpace(msg.Version) == "" {
		return sdkerrors.Wrap(ibctypes.ErrInvalidVersion, "version can't be blank")
	}
	if msg.ProofTry == nil || msg.ProofConsensus == nil {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if err := msg.ProofTry.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof try cannot be nil")
	}
	if err := msg.ProofConsensus.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof consensus cannot be nil")
	}
	if msg.ProofHeight == 0 {
		return sdkerrors.Wrap(ibctypes.ErrInvalidHeight, "proof height must be > 0")
	}
	if msg.ConsensusHeight == 0 {
		return sdkerrors.Wrap(ibctypes.ErrInvalidHeight, "consensus height must be > 0")
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

// MsgConnectionOpenConfirm defines a msg sent by a Relayer to Chain B to acknowledge
// the change of connection state to OPEN on Chain A.
type MsgConnectionOpenConfirm struct {
	ConnectionID string                   `json:"connection_id"`
	ProofAck     commitmentexported.Proof `json:"proof_ack"` // proof for the change of the connection state on Chain A: `INIT -> OPEN`
	ProofHeight  uint64                   `json:"proof_height"`
	Signer       sdk.AccAddress           `json:"signer"`
}

// NewMsgConnectionOpenConfirm creates a new MsgConnectionOpenConfirm instance
func NewMsgConnectionOpenConfirm(
	connectionID string, proofAck commitmentexported.Proof, proofHeight uint64,
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
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgConnectionOpenConfirm) Type() string {
	return "connection_open_confirm"
}

// ValidateBasic implements sdk.Msg
func (msg MsgConnectionOpenConfirm) ValidateBasic() error {
	if err := host.DefaultConnectionIdentifierValidator(msg.ConnectionID); err != nil {
		return sdkerrors.Wrap(err, "invalid connection ID")
	}
	if msg.ProofAck == nil {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if err := msg.ProofAck.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof ack cannot be nil")
	}
	if msg.ProofHeight == 0 {
		return sdkerrors.Wrap(ibctypes.ErrInvalidHeight, "proof height must be > 0")
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
