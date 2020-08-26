package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

var (
	_ clientexported.MsgCreateClient     = &MsgCreateClient{}
	_ clientexported.MsgUpdateClient     = &MsgUpdateClient{}
	_ evidenceexported.MsgSubmitEvidence = &MsgSubmitClientMisbehaviour{}
)

// NewMsgCreateClient creates a new MsgCreateClient instance
func NewMsgCreateClient(id string, consensusState *ConsensusState) *MsgCreateClient {
	return &MsgCreateClient{
		ClientId:       id,
		ConsensusState: consensusState,
	}
}

// Route implements sdk.Msg
func (msg MsgCreateClient) Route() string {
	return host.RouterKey
}

// Type implements sdk.Msg
func (msg MsgCreateClient) Type() string {
	return clientexported.TypeMsgCreateClient
}

// ValidateBasic implements sdk.Msg
func (msg MsgCreateClient) ValidateBasic() error {
	if msg.ConsensusState == nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "consensus state cannot be nil")
	}
	if err := msg.ConsensusState.ValidateBasic(); err != nil {
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
	return []sdk.AccAddress{sdk.AccAddress(msg.ConsensusState.GetPubKey().Address())}
}

// GetClientID implements clientexported.MsgCreateClient
func (msg MsgCreateClient) GetClientID() string {
	return msg.ClientId
}

// GetClientType implements clientexported.MsgCreateClient
func (msg MsgCreateClient) GetClientType() string {
	return clientexported.ClientTypeSoloMachine
}

// GetConsensusState implements clientexported.MsgCreateClient
func (msg MsgCreateClient) GetConsensusState() clientexported.ConsensusState {
	return msg.ConsensusState
}

// InitializeFromMsg creates a solo machine client state from a MsgCreateClient
func (msg MsgCreateClient) InitializeClientState() clientexported.ClientState {
	return NewClientState(msg.ConsensusState)
}

// NewMsgUpdateClient creates a new MsgUpdateClient instance
func NewMsgUpdateClient(id string, header *Header) *MsgUpdateClient {
	return &MsgUpdateClient{
		ClientId: id,
		Header:   header,
	}
}

// Route implements sdk.Msg
func (msg MsgUpdateClient) Route() string {
	return host.RouterKey
}

// Type implements sdk.Msg
func (msg MsgUpdateClient) Type() string {
	return clientexported.TypeMsgUpdateClient
}

// ValidateBasic implements sdk.Msg
func (msg MsgUpdateClient) ValidateBasic() error {
	if msg.Header == nil {
		return sdkerrors.Wrap(ErrInvalidHeader, "header cannot be nil")
	}
	if err := msg.Header.ValidateBasic(); err != nil {
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
	return []sdk.AccAddress{sdk.AccAddress(msg.Header.GetPubKey().Address())}
}

// GetClientID implements clientexported.MsgUpdateClient
func (msg MsgUpdateClient) GetClientID() string {
	return msg.ClientId
}

// GetHeader implements clientexported.MsgUpdateClient
func (msg MsgUpdateClient) GetHeader() clientexported.Header {
	return msg.Header
}

// NewMsgSubmitClientMisbehaviour creates a new MsgSubmitClientMisbehaviour
// instance.
func NewMsgSubmitClientMisbehaviour(e *Evidence, s sdk.AccAddress) *MsgSubmitClientMisbehaviour {
	return &MsgSubmitClientMisbehaviour{Evidence: e, Submitter: s}
}

// Route returns the MsgSubmitClientMisbehaviour's route.
func (msg MsgSubmitClientMisbehaviour) Route() string { return host.RouterKey }

// Type returns the MsgSubmitClientMisbehaviour's type.
func (msg MsgSubmitClientMisbehaviour) Type() string {
	return clientexported.TypeMsgSubmitClientMisbehaviour
}

// ValidateBasic performs basic (non-state-dependent) validation on a MsgSubmitClientMisbehaviour.
func (msg MsgSubmitClientMisbehaviour) ValidateBasic() error {
	if err := msg.Evidence.ValidateBasic(); err != nil {
		return err
	}
	if msg.Submitter.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "submitter address cannot be empty")
	}

	return nil
}

// GetSignBytes returns the raw bytes a signer is expected to sign when submitting
// a MsgSubmitClientMisbehaviour message.
func (msg MsgSubmitClientMisbehaviour) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the single expected signer for a MsgSubmitClientMisbehaviour.
func (msg MsgSubmitClientMisbehaviour) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Submitter}
}

func (msg MsgSubmitClientMisbehaviour) GetEvidence() evidenceexported.Evidence {
	return msg.Evidence
}

func (msg MsgSubmitClientMisbehaviour) GetSubmitter() sdk.AccAddress {
	return msg.Submitter
}
