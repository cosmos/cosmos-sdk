package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// message types for the IBC client
const (
	TypeMsgCreateClient       string = "create_client"
	TypeMsgUpdateClient       string = "update_client"
	TypeMsgSubmitMisbehaviour string = "submit_misbehaviour"
)

// NewMsgCreateClient creates a new MsgCreateClient instance
func NewMsgCreateClient(
	id string, clientState, consensusState *codectypes.Any, signer sdk.AccAddress,
) *MsgCreateClient {
	return &MsgCreateClient{
		ClientId:       id,
		ClientState:    clientState,
		ConsensusState: consensusState,
		Signer:         signer,
	}
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
func NewMsgUpdateClient(id string, header *codectypes.Any, signer sdk.AccAddress) *MsgUpdateClient {
	return &MsgUpdateClient{
		ClientId: id,
		Header:   header,
		Signer:   signer,
	}
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
	header, err := UnpackConsensusState(msg.Header)
	if err != nil {
		return err
	}
	if err := header.ValidateBasic(); err != nil {
		return err
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

// NewMsgSubmitMisbehaviour creates a new MsgSubmitMisbehaviour
// instance.
func NewMsgSubmitMisbehaviour(clientID string, misbehaviour *codectypes.Any, signer sdk.AccAddress) *MsgSubmitMisbehaviour {
	return &MsgSubmitMisbehaviour{
		ClientId:     clientID,
		Misbehaviour: misbehaviour,
		Signer:       signer,
	}
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

// GetEvidence implements the Evidence interface
func (msg MsgSubmitMisbehaviour) GetEvidence() evidenceexported.Evidence {
	return MustUnpackMisbehaviour(msg.Misbehaviour)
}

// GetSubmitter implements the Evidence interface
func (msg MsgSubmitMisbehaviour) GetSubmitter() sdk.AccAddress {
	return msg.Signer
}
