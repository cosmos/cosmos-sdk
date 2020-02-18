package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// Message types for the IBC client
const (
	TypeMsgCreateClient    string = "create_client"
	TypeMsgUpdateClient    string = "update_client"
	TypeClientMisbehaviour string = "client_misbehaviour"
)

var _ clientexported.MsgCreateClient = MsgCreateClient{}
var _ clientexported.MsgUpdateClient = MsgUpdateClient{}

// MsgCreateClient defines a message to create an IBC client
type MsgCreateClient struct {
	ClientID        string         `json:"client_id" yaml:"client_id"`
	Header          Header         `json:"header" yaml:"header"`
	TrustingPeriod  time.Duration  `json:"trusting_period" yaml:"trusting_period"`
	UnbondingPeriod time.Duration  `json:"unbonding_period" yaml:"unbonding_period"`
	Signer          sdk.AccAddress `json:"address" yaml:"address"`
}

// NewMsgCreateClient creates a new MsgCreateClient instance
func NewMsgCreateClient(
	id string, header Header,
	trustingPeriod, unbondingPeriod time.Duration, signer sdk.AccAddress,
) MsgCreateClient {
	return MsgCreateClient{
		ClientID:        id,
		Header:          header,
		TrustingPeriod:  trustingPeriod,
		UnbondingPeriod: unbondingPeriod,
		Signer:          signer,
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
	if msg.TrustingPeriod == 0 {
		return sdkerrors.Wrap(ErrInvalidTrustingPeriod, "duration cannot be 0")
	}
	if msg.UnbondingPeriod == 0 {
		return sdkerrors.Wrap(ErrInvalidUnbondingPeriod, "duration cannot be 0")
	}
	if msg.Signer.Empty() {
		return sdkerrors.ErrInvalidAddress
	}
	// ValidateBasic of provided header with self-attested chain-id
	if msg.Header.ValidateBasic(msg.Header.ChainID) != nil {
		return sdkerrors.Wrap(ErrInvalidHeader, "header failed validatebasic with its own chain-id")
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

// GetClientID implements clientexported.MsgCreateClient
func (msg MsgCreateClient) GetClientID() string {
	return msg.ClientID
}

// GetClientType implements clientexported.MsgCreateClient
func (msg MsgCreateClient) GetClientType() string {
	return clientexported.ClientTypeTendermint
}

// GetConsensusState implements clientexported.MsgCreateClient
func (msg MsgCreateClient) GetConsensusState() clientexported.ConsensusState {
	// Construct initial consensus state from provided Header
	root := commitment.NewRoot(msg.Header.AppHash)
	return ConsensusState{
		Timestamp:    msg.Header.Time,
		Root:         root,
		Height:       uint64(msg.Header.Height),
		ValidatorSet: msg.Header.ValidatorSet,
	}
}

// MsgUpdateClient defines a message to update an IBC client
type MsgUpdateClient struct {
	ClientID string         `json:"client_id" yaml:"client_id"`
	Header   Header         `json:"header" yaml:"header"`
	Signer   sdk.AccAddress `json:"address" yaml:"address"`
}

// NewMsgUpdateClient creates a new MsgUpdateClient instance
func NewMsgUpdateClient(id string, header Header, signer sdk.AccAddress) MsgUpdateClient {
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

// GetClientID implements clientexported.MsgUpdateClient
func (msg MsgUpdateClient) GetClientID() string {
	return msg.ClientID
}

// GetHeader implements clientexported.MsgUpdateClient
func (msg MsgUpdateClient) GetHeader() clientexported.Header {
	return msg.Header
}
