package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/errors"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// Message types for the IBC client
const (
	TypeMsgCreateClient    string = "create_client"
	TypeMsgUpdateClient    string = "update_client"
	TypeClientMisbehaviour string = "client_misbehaviour"
)

var _ sdk.Msg = MsgCreateClient{}

// MsgCreateClient defines a message to create an IBC client
type MsgCreateClient struct {
	ClientID       string                  `json:"client_id" yaml:"client_id"`
	ClientType     string                  `json:"client_type" yaml:"client_type"`
	ConsensusState exported.ConsensusState `json:"consensus_state" yaml:"consensus_address"`
	Signer         sdk.AccAddress          `json:"address" yaml:"address"`
}

// NewMsgCreateClient creates a new MsgCreateClient instance
func NewMsgCreateClient(id, clientType string, consensusState exported.ConsensusState, signer sdk.AccAddress) MsgCreateClient {
	return MsgCreateClient{
		ClientID:       id,
		ClientType:     clientType,
		ConsensusState: consensusState,
		Signer:         signer,
	}
}

// Route implements sdk.Msg
func (msg MsgCreateClient) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgCreateClient) Type() string {
	return TypeMsgCreateClient
}

// ValidateBasic implements sdk.Msg
func (msg MsgCreateClient) ValidateBasic() error {
	if clientType := exported.ClientTypeFromString(msg.ClientType); clientType == 0 {
		return sdkerrors.Wrap(errors.ErrInvalidClientType, msg.ClientType)
	}
	if msg.ConsensusState == nil {
		return errors.ErrInvalidConsensus
	}
	if msg.Signer.Empty() {
		return sdkerrors.ErrInvalidAddress
	}
	return host.DefaultClientIdentifierValidator(msg.ClientID)
}

// GetSignBytes implements sdk.Msg
func (msg MsgCreateClient) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgCreateClient) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

var _ sdk.Msg = MsgUpdateClient{}

// MsgUpdateClient defines a message to update an IBC client
type MsgUpdateClient struct {
	ClientID string          `json:"client_id" yaml:"client_id"`
	Header   exported.Header `json:"header" yaml:"header"`
	Signer   sdk.AccAddress  `json:"address" yaml:"address"`
}

// NewMsgUpdateClient creates a new MsgUpdateClient instance
func NewMsgUpdateClient(id string, header exported.Header, signer sdk.AccAddress) MsgUpdateClient {
	return MsgUpdateClient{
		ClientID: id,
		Header:   header,
		Signer:   signer,
	}
}

// Route implements sdk.Msg
func (msg MsgUpdateClient) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgUpdateClient) Type() string {
	return TypeMsgUpdateClient
}

// ValidateBasic implements sdk.Msg
func (msg MsgUpdateClient) ValidateBasic() error {
	if msg.Header == nil {
		return errors.ErrInvalidHeader
	}
	if msg.Signer.Empty() {
		return sdkerrors.ErrInvalidAddress
	}
	return host.DefaultClientIdentifierValidator(msg.ClientID)
}

// GetSignBytes implements sdk.Msg
func (msg MsgUpdateClient) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgUpdateClient) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

// MsgSubmitMisbehaviour defines a message to update an IBC client
type MsgSubmitMisbehaviour struct {
	ClientID string                    `json:"id" yaml:"id"`
	Evidence evidenceexported.Evidence `json:"evidence" yaml:"evidence"`
	Signer   sdk.AccAddress            `json:"address" yaml:"address"`
}

// NewMsgSubmitMisbehaviour creates a new MsgSubmitMisbehaviour instance
func NewMsgSubmitMisbehaviour(id string, evidence evidenceexported.Evidence, signer sdk.AccAddress) MsgSubmitMisbehaviour {
	return MsgSubmitMisbehaviour{
		ClientID: id,
		Evidence: evidence,
		Signer:   signer,
	}
}

// Route implements sdk.Msg
func (msg MsgSubmitMisbehaviour) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgSubmitMisbehaviour) Type() string {
	return "submit_misbehaviour"
}

// ValidateBasic implements sdk.Msg
func (msg MsgSubmitMisbehaviour) ValidateBasic() error {
	if msg.Evidence == nil {
		return sdkerrors.Wrap(errors.ErrInvalidEvidence, "evidence cannot be nil")
	}
	if err := msg.Evidence.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(errors.ErrInvalidEvidence, err.Error())
	}
	if msg.Signer.Empty() {
		return sdkerrors.ErrInvalidAddress
	}
	return host.DefaultClientIdentifierValidator(msg.ClientID)
}

// GetSignBytes implements sdk.Msg
func (msg MsgSubmitMisbehaviour) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgSubmitMisbehaviour) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
