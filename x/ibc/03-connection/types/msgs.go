package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

var _ sdk.Msg = MsgConnectionOpenInit{}

// MsgConnectionOpenInit defines the msg sent by an account on Chain A to
// initialize a connection with Chain B.
type MsgConnectionOpenInit struct {
	ConnectionID string         `json:"connection_id"`
	ClientID     string         `json:"client_id"`
	Counterparty Counterparty   `json:"counterparty"`
	NextTimeout  uint64         `json:"next_timeout"` // TODO: Where is this defined?
	Signer       sdk.AccAddress `json:"signer"`
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
	return nil // TODO
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
	ConnectionID         string        `json:"connection_id"`
	ClientID             string        `json:"client_id"`
	Counterparty         Counterparty  `json:"counterparty"`
	CounterpartyVersions []string      `json:"counterparty_versions"` // TODO: why wasn't this defined previously?
	Proofs               []ics23.Proof `json:"proofs"`                // Contains a Proof of the initialization the connection on Chain A
	Height               uint64        `json:"height"`                // TODO: Rename to ProofHeight? Is this supposed to be the same as ConsensusHeight?
	// ConsensusHeight       uint64         `json:"consensus_height"`
	Signer sdk.AccAddress `json:"signer"`
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
	return nil // TODO
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
	ConnectionID    string         `json:"connection_id"`
	Timeout         uint64         `json:"timeout"`      // TODO: Where's this defined ?
	NextTimeout     uint64         `json:"next_timeout"` // TODO: Where's this defined ?
	Proofs          []ics23.Proof  `json:"proofs"`       // Contains a Proof for the change of the connection state on Chain B: `none -> TRYOPEN`
	ConsensusHeight uint64         `json:"consensus_height"`
	Signer          sdk.AccAddress `json:"signer"`
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
	return nil // TODO
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
	ConnectionID string         `json:"connection_id"`
	Timeout      uint64         `json:"timeout"` // TODO: Where's this defined ?
	Proofs       []ics23.Proof  `json:"proofs"`  // Contains a Proof for the change of the connection state on Chain A: `INIT -> OPEN`
	Height       uint64         `json:"height"`  // TODO: Rename to ProofHeight?
	Signer       sdk.AccAddress `json:"signer"`
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
	return nil // TODO
}

// GetSignBytes implements sdk.Msg
func (msg MsgConnectionOpenConfirm) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgConnectionOpenConfirm) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
