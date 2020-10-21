package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

// message types for the IBC client
const (
	TypeMsgCreateClient       string = "create_client"
	TypeMsgUpdateClient       string = "update_client"
	TypeMsgUpgradeClient      string = "upgrade_client"
	TypeMsgSubmitMisbehaviour string = "submit_misbehaviour"
)

var (
	_ sdk.Msg = &MsgCreateClient{}
	_ sdk.Msg = &MsgUpdateClient{}
	_ sdk.Msg = &MsgSubmitMisbehaviour{}
	_ sdk.Msg = &MsgUpgradeClient{}

	_ codectypes.UnpackInterfacesMessage = MsgCreateClient{}
	_ codectypes.UnpackInterfacesMessage = MsgUpdateClient{}
	_ codectypes.UnpackInterfacesMessage = MsgSubmitMisbehaviour{}
	_ codectypes.UnpackInterfacesMessage = MsgUpgradeClient{}
)

// NewMsgCreateClient creates a new MsgCreateClient instance
//nolint:interfacer
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
		Signer:         signer.String(),
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
	if msg.Signer == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "signer address cannot be empty")
	}
	clientState, err := UnpackClientState(msg.ClientState)
	if err != nil {
		return err
	}
	if err := clientState.Validate(); err != nil {
		return err
	}
	if clientState.ClientType() == exported.Localhost || msg.ClientId == exported.Localhost {
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
	accAddr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{accAddr}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgCreateClient) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var clientState exported.ClientState
	err := unpacker.UnpackAny(msg.ClientState, &clientState)
	if err != nil {
		return err
	}

	var consensusState exported.ConsensusState
	return unpacker.UnpackAny(msg.ConsensusState, &consensusState)
}

// NewMsgUpdateClient creates a new MsgUpdateClient instance
//nolint:interfacer
func NewMsgUpdateClient(id string, header exported.Header, signer sdk.AccAddress) (*MsgUpdateClient, error) {
	anyHeader, err := PackHeader(header)
	if err != nil {
		return nil, err
	}

	return &MsgUpdateClient{
		ClientId: id,
		Header:   anyHeader,
		Signer:   signer.String(),
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
	if msg.Signer == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "signer address cannot be empty")
	}
	header, err := UnpackHeader(msg.Header)
	if err != nil {
		return err
	}
	if err := header.ValidateBasic(); err != nil {
		return err
	}
	if msg.ClientId == exported.Localhost {
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
	accAddr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{accAddr}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgUpdateClient) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var header exported.Header
	return unpacker.UnpackAny(msg.Header, &header)
}

// NewMsgUpgradeClient creates a new MsgUpgradeClient instance
// nolint: interfacer
func NewMsgUpgradeClient(clientID string, clientState exported.ClientState, upgradeHeight exported.Height, proofUpgrade []byte, signer sdk.AccAddress) (*MsgUpgradeClient, error) {
	anyClient, err := PackClientState(clientState)
	if err != nil {
		return nil, err
	}

	height, ok := upgradeHeight.(Height)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidHeight, "invalid height type. expected: %T, got: %T", &Height{}, upgradeHeight)
	}

	return &MsgUpgradeClient{
		ClientId:      clientID,
		ClientState:   anyClient,
		ProofUpgrade:  proofUpgrade,
		UpgradeHeight: &height,
		Signer:        signer.String(),
	}, nil
}

// Route implements sdk.Msg
func (msg MsgUpgradeClient) Route() string {
	return host.RouterKey
}

// Type implements sdk.Msg
func (msg MsgUpgradeClient) Type() string {
	return TypeMsgUpgradeClient
}

// ValidateBasic implements sdk.Msg
func (msg MsgUpgradeClient) ValidateBasic() error {
	clientState, err := UnpackClientState(msg.ClientState)
	if err != nil {
		return err
	}
	if err := clientState.Validate(); err != nil {
		return err
	}
	if len(msg.ProofUpgrade) == 0 {
		return sdkerrors.Wrap(ErrInvalidUpgradeClient, "proof of upgrade cannot be empty")
	}
	if msg.UpgradeHeight == nil {
		return sdkerrors.Wrap(ErrInvalidUpgradeClient, "upgrade height cannot be nil")
	}
	if msg.UpgradeHeight.IsZero() {
		return sdkerrors.Wrap(ErrInvalidUpgradeClient, "upgrade height cannot be zero")
	}
	_, err = sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "string could not be parsed as address: %v", err)
	}
	return host.ClientIdentifierValidator(msg.ClientId)
}

// GetSignBytes implements sdk.Msg
func (msg MsgUpgradeClient) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners implements sdk.Msg
func (msg MsgUpgradeClient) GetSigners() []sdk.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{accAddr}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgUpgradeClient) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var clientState exported.ClientState
	return unpacker.UnpackAny(msg.ClientState, &clientState)
}

// NewMsgSubmitMisbehaviour creates a new MsgSubmitMisbehaviour instance.
//nolint:interfacer
func NewMsgSubmitMisbehaviour(clientID string, misbehaviour exported.Misbehaviour, signer sdk.AccAddress) (*MsgSubmitMisbehaviour, error) {
	anyMisbehaviour, err := PackMisbehaviour(misbehaviour)
	if err != nil {
		return nil, err
	}

	return &MsgSubmitMisbehaviour{
		ClientId:     clientID,
		Misbehaviour: anyMisbehaviour,
		Signer:       signer.String(),
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
	if msg.Signer == "" {
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
	accAddr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{accAddr}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgSubmitMisbehaviour) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var misbehaviour exported.Misbehaviour
	return unpacker.UnpackAny(msg.Misbehaviour, &misbehaviour)
}
