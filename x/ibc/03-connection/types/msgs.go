package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
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
	counterpartyClientID string, counterpartyPrefix commitment.PrefixI,
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
func (msg MsgConnectionOpenInit) ValidateBasic() sdk.Error {
	if err := host.DefaultConnectionIdentifierValidator(msg.ConnectionID); err != nil {
		return sdk.NewError(host.IBCCodeSpace, 1, fmt.Sprintf("invalid connection ID: %s", err.Error()))
	}
	if err := host.DefaultClientIdentifierValidator(msg.ClientID); err != nil {
		return sdk.NewError(host.IBCCodeSpace, 1, fmt.Sprintf("invalid client ID: %s", err.Error()))
	}
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress("missing signer address")
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
	ConnectionID         string            `json:"connection_id"`
	ClientID             string            `json:"client_id"`
	Counterparty         Counterparty      `json:"counterparty"`
	CounterpartyVersions []string          `json:"counterparty_versions"`
	ProofInit            commitment.ProofI `json:"proof_init"`      // proof of the initialization the connection on Chain A: `none -> INIT`
	ProofConsensus       commitment.ProofI `json:"proof_consensus"` // proof of client consensus stat
	ProofHeight          uint64            `json:"proof_height"`
	ConsensusHeight      uint64            `json:"consensus_height"`
	Signer               sdk.AccAddress    `json:"signer"`
}

// NewMsgConnectionOpenTry creates a new MsgConnectionOpenTry instance
func NewMsgConnectionOpenTry(
	connectionID, clientID, counterpartyConnectionID,
	counterpartyClientID string, counterpartyPrefix commitment.PrefixI,
	counterpartyVersions []string, proofInit, proofConsensus commitment.ProofI,
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
func (msg MsgConnectionOpenTry) ValidateBasic() sdk.Error {
	if err := host.DefaultConnectionIdentifierValidator(msg.ConnectionID); err != nil {
		return sdk.NewError(host.IBCCodeSpace, 1, fmt.Sprintf("invalid connection ID: %s", err.Error()))
	}
	if err := host.DefaultClientIdentifierValidator(msg.ClientID); err != nil {
		return sdk.NewError(host.IBCCodeSpace, 1, fmt.Sprintf("invalid client ID: %s", err.Error()))
	}
	if len(msg.CounterpartyVersions) == 0 {
		return ErrInvalidVersion(DefaultCodespace, "missing counterparty versions")
	}
	for _, version := range msg.CounterpartyVersions {
		if strings.TrimSpace(version) == "" {
			return ErrInvalidVersion(DefaultCodespace, "version can't be blank")
		}
	}
	if msg.ProofInit == nil {
		return ErrInvalidConnectionProof(DefaultCodespace, "proof init cannot be nil")
	}
	if msg.ProofConsensus == nil {
		return ErrInvalidConnectionProof(DefaultCodespace, "proof consensus cannot be nil")
	}
	if msg.ProofHeight == 0 {
		return ErrInvalidHeight(DefaultCodespace, "proof height must be > 0")
	}
	if msg.ConsensusHeight == 0 {
		return ErrInvalidHeight(DefaultCodespace, "consensus height must be > 0")
	}
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress("missing signer address")
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
	ConnectionID    string            `json:"connection_id"`
	ProofTry        commitment.ProofI `json:"proof_try"`       // proof for the change of the connection state on Chain B: `none -> TRYOPEN`
	ProofConsensus  commitment.ProofI `json:"proof_consensus"` // proof of client consensus stat
	ProofHeight     uint64            `json:"proof_height"`
	ConsensusHeight uint64            `json:"consensus_height"`
	Version         string            `json:"version"`
	Signer          sdk.AccAddress    `json:"signer"`
}

// NewMsgConnectionOpenAck creates a new MsgConnectionOpenAck instance
func NewMsgConnectionOpenAck(
	connectionID string, proofTry, proofConsensus commitment.ProofI,
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
func (msg MsgConnectionOpenAck) ValidateBasic() sdk.Error {
	if err := host.DefaultConnectionIdentifierValidator(msg.ConnectionID); err != nil {
		return sdk.NewError(host.IBCCodeSpace, 1, fmt.Sprintf("invalid connection ID: %s", err.Error()))
	}
	if strings.TrimSpace(msg.Version) == "" {
		return ErrInvalidVersion(DefaultCodespace, "version can't be blank")
	}
	if msg.ProofTry == nil {
		return ErrInvalidConnectionProof(DefaultCodespace, "proof try cannot be nil")
	}
	if msg.ProofConsensus == nil {
		return ErrInvalidConnectionProof(DefaultCodespace, "proof consensus cannot be nil")
	}
	if msg.ProofHeight == 0 {
		return ErrInvalidHeight(DefaultCodespace, "proof height must be > 0")
	}
	if msg.ConsensusHeight == 0 {
		return ErrInvalidHeight(DefaultCodespace, "consensus height must be > 0")
	}
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress("missing signer address")
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
	ConnectionID string            `json:"connection_id"`
	ProofAck     commitment.ProofI `json:"proof_ack"` // proof for the change of the connection state on Chain A: `INIT -> OPEN`
	ProofHeight  uint64            `json:"proof_height"`
	Signer       sdk.AccAddress    `json:"signer"`
}

// NewMsgConnectionOpenConfirm creates a new MsgConnectionOpenConfirm instance
func NewMsgConnectionOpenConfirm(
	connectionID string, proofAck commitment.ProofI, proofHeight uint64, signer sdk.AccAddress,
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
func (msg MsgConnectionOpenConfirm) ValidateBasic() sdk.Error {
	if err := host.DefaultConnectionIdentifierValidator(msg.ConnectionID); err != nil {
		return sdk.NewError(host.IBCCodeSpace, 1, fmt.Sprintf("invalid connection ID: %s", err.Error()))
	}
	if msg.ProofAck == nil {
		return ErrInvalidConnectionProof(DefaultCodespace, "proof ack cannot be nil")
	}
	if msg.ProofHeight == 0 {
		return ErrInvalidHeight(DefaultCodespace, "proof height must be > 0")
	}

	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress("missing signer address")
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
