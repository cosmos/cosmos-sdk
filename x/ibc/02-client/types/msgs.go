package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// message types for the IBC client
const (
	TypeMsgCreateClient       string = "create_client"
	TypeMsgUpdateClient       string = "update_client"
	TypeMsgSubmitMisbehaviour string = "submit_misbehaviour"
)

var (
	_ sdk.Msg = &MsgCreateClient{}
	_ sdk.Msg = &MsgUpdateClient{}
	_ sdk.Msg = &MsgSubmitMisbehaviour{}
)

// NewMsgCreateClient creates a new MsgCreateClient instance
func NewMsgCreateClient(
	id string, clientState exported.ClientState, consensusState exported.ConsensusState, signer sdk.AccAddress,
) (*MsgCreateClient, error) {

	anyClientState, err := PackClientState(clientState)
	if err != nil {
		return nil, err
	}

	anyConsensusState, err := PackConsensusState(consensusState)
	if err != nil {
		return nil, err
	}

	return &MsgCreateClient{
		ClientId:       id,
		ClientState:    anyClientState,
		ConsensusState: anyConsensusState,
		Signer:         signer,
	}, nil
}

// Route implements sdk.Msg
func (msg MsgCreateClient) Route() string {
	return host.RouterKey
}

// Type implements sdk.Msg
func (msg MsgCreateClient) Type() string {
	return TypeMsgCreateClient
}

// ValidateBasic implements sdk.Msg
func (msg MsgCreateClient) ValidateBasic() error {
	if msg.Signer.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "signer address cannot be empty")
	}
	clientState, err := UnpackClientState(msg.ClientState)
	if err != nil {
		return err
	}
	if err := clientState.Validate(); err != nil {
		return err
	}
	if clientState.ClientType() == exported.Localhost || msg.ClientId == exported.ClientTypeLocalHost {
		return sdkerrors.Wrap(ErrInvalidClient, "localhost client can only be created on chain initialization")
	}
	consensusState, err := UnpackConsensusState(msg.ConsensusState)
	if err != nil {
		return err
	}
	if err := consensusState.ValidateBasic(); err != nil {
		return err
	}
	return host.ClientIdentifierValidator(msg.ClientId)
}

// GetSignBytes implements sdk.Msg
func (msg MsgCreateClient) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners implements sdk.Msg
func (msg MsgCreateClient) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

// NewMsgUpdateClient creates a new MsgUpdateClient instance
func NewMsgUpdateClient(id string, header exported.Header, signer sdk.AccAddress) (*MsgUpdateClient, error) {
	anyHeader, err := PackHeader(header)
	if err != nil {
		return nil, err
	}

	return &MsgUpdateClient{
		ClientId: id,
		Header:   anyHeader,
		Signer:   signer,
	}, nil
}

// Route implements sdk.Msg
func (msg MsgUpdateClient) Route() string {
	return host.RouterKey
}

// Type implements sdk.Msg
func (msg MsgUpdateClient) Type() string {
	return TypeMsgUpdateClient
}

// ValidateBasic implements sdk.Msg
func (msg MsgUpdateClient) ValidateBasic() error {
	if msg.Signer.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "signer address cannot be empty")
	}
	header, err := UnpackHeader(msg.Header)
	if err != nil {
		return err
	}
	if err := header.ValidateBasic(); err != nil {
		return err
	}
	if msg.ClientId == exported.ClientTypeLocalHost {
		return sdkerrors.Wrap(ErrInvalidClient, "localhost client is only updated on ABCI BeginBlock")
	}
	return host.ClientIdentifierValidator(msg.ClientId)
}

// GetSignBytes implements sdk.Msg
func (msg MsgUpdateClient) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners implements sdk.Msg
func (msg MsgUpdateClient) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

// NewMsgSubmitMisbehaviour creates a new MsgSubmitMisbehaviour instance.
func NewMsgSubmitMisbehaviour(clientID string, misbehaviour exported.Misbehaviour, signer sdk.AccAddress) (*MsgSubmitMisbehaviour, error) {
	anyMisbehaviour, err := PackMisbehaviour(misbehaviour)
	if err != nil {
		return nil, err
	}

	return &MsgSubmitMisbehaviour{
		ClientId:     clientID,
		Misbehaviour: anyMisbehaviour,
		Signer:       signer,
	}, nil
}

// Route returns the MsgSubmitClientMisbehaviour's route.
func (msg MsgSubmitMisbehaviour) Route() string { return host.RouterKey }

// Type returns the MsgSubmitMisbehaviour's type.
func (msg MsgSubmitMisbehaviour) Type() string {
	return TypeMsgSubmitMisbehaviour
}

// ValidateBasic performs basic (non-state-dependant) validation on a MsgSubmitMisbehaviour.
func (msg MsgSubmitMisbehaviour) ValidateBasic() error {
	if msg.Signer.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "signer address cannot be empty")
	}
	misbehaviour, err := UnpackMisbehaviour(msg.Misbehaviour)
	if err != nil {
		return err
	}
	if err := misbehaviour.ValidateBasic(); err != nil {
		return err
	}
	if misbehaviour.GetClientID() != msg.ClientId {
		return sdkerrors.Wrapf(
			ErrInvalidMisbehaviour,
			"misbehaviour client-id doesn't match client-id from message (%s ≠ %s)",
			misbehaviour.GetClientID(), msg.ClientId,
		)
	}

	return host.ClientIdentifierValidator(msg.ClientId)
}

// GetSignBytes returns the raw bytes a signer is expected to sign when submitting
// a MsgSubmitMisbehaviour message.
func (msg MsgSubmitMisbehaviour) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the single expected signer for a MsgSubmitMisbehaviour.
func (msg MsgSubmitMisbehaviour) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
