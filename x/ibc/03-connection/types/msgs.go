package types

import (
	"fmt"
	"strings"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

var _ sdk.Msg = &MsgConnectionOpenInit{}

// NewMsgConnectionOpenInit creates a new MsgConnectionOpenInit instance
func NewMsgConnectionOpenInit(
	connectionID, clientID, counterpartyConnectionID,
	counterpartyClientID string, counterpartyPrefix commitmentexported.Prefix,
	signer sdk.AccAddress,
) (*MsgConnectionOpenInit, error) {
	counterparty, err := NewCounterparty(counterpartyClientID, counterpartyConnectionID, counterpartyPrefix)
	if err != nil {
		return nil, err
	}

	return &MsgConnectionOpenInit{
		ConnectionID: connectionID,
		ClientID:     clientID,
		Counterparty: counterparty,
		Signer:       signer,
	}, nil
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

var (
	_ sdk.Msg                          = &MsgConnectionOpenTry{}
	_ cdctypes.UnpackInterfacesMessage = &MsgConnectionOpenTry{}
)

// NewMsgConnectionOpenTry creates a new MsgConnectionOpenTry instance
func NewMsgConnectionOpenTry(
	connectionID, clientID, counterpartyConnectionID,
	counterpartyClientID string, counterpartyPrefix commitmentexported.Prefix,
	counterpartyVersions []string, proofInit, proofConsensus commitmentexported.Proof,
	proofHeight, consensusHeight uint64, signer sdk.AccAddress,
) (*MsgConnectionOpenTry, error) {
	counterparty, err := NewCounterparty(counterpartyClientID, counterpartyConnectionID, counterpartyPrefix)
	if err != nil {
		return nil, err
	}

	proofInitAny, err := proofInit.PackAny()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid proof init")
	}

	proofConsensusAny, err := proofConsensus.PackAny()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid proof consensus")
	}

	return &MsgConnectionOpenTry{
		ConnectionID:         connectionID,
		ClientID:             clientID,
		Counterparty:         counterparty,
		CounterpartyVersions: counterpartyVersions,
		ProofInit:            *proofInitAny,
		ProofConsensus:       *proofConsensusAny,
		ProofHeight:          proofHeight,
		ConsensusHeight:      consensusHeight,
		Signer:               signer,
	}, nil
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

	proofInit := msg.GetProofInit()
	proofConsensus := msg.GetProofConsensus()

	if proofInit == nil || proofInit.IsEmpty() {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof init")
	}

	if proofConsensus == nil || proofConsensus.IsEmpty() {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof consensus")
	}
	if err := proofInit.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof init failed basic validation")
	}
	if err := proofConsensus.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof consensus failed basic validation")
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

// GetProofInit returns the cached value from ProofInit. It returns nil if the value
// is not cached or if the proof doesn't cast to a commitment Proof.
func (msg MsgConnectionOpenTry) GetProofInit() commitmentexported.Proof {
	proof, _ := commitmenttypes.UnpackAnyProof(&msg.ProofInit)
	return proof
}

// GetProofConsensus returns the cached value from ProofConsensus. It returns nil if the value
// is not cached or if the proof doesn't cast to a commitment Proof.
func (msg MsgConnectionOpenTry) GetProofConsensus() commitmentexported.Proof {
	proof, _ := commitmenttypes.UnpackAnyProof(&msg.ProofConsensus)
	return proof
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgConnectionOpenTry) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	var (
		proofInit      commitmentexported.Proof
		proofConsensus commitmentexported.Proof
	)
	err := unpacker.UnpackAny(&msg.ProofInit, &proofInit)
	if err != nil {
		return fmt.Errorf("proof init unpack failed: %w", err)
	}

	err = unpacker.UnpackAny(&msg.ProofConsensus, &proofConsensus)
	if err != nil {
		return fmt.Errorf("proof consensus unpack failed: %w", err)
	}
	return nil
}

var (
	_ sdk.Msg                          = &MsgConnectionOpenAck{}
	_ cdctypes.UnpackInterfacesMessage = &MsgConnectionOpenAck{}
)

// NewMsgConnectionOpenAck creates a new MsgConnectionOpenAck instance
func NewMsgConnectionOpenAck(
	connectionID string, proofTry, proofConsensus commitmentexported.Proof,
	proofHeight, consensusHeight uint64, version string,
	signer sdk.AccAddress,
) (*MsgConnectionOpenAck, error) {
	proofTryAny, err := proofTry.PackAny()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid proof try")
	}

	proofConsensusAny, err := proofConsensus.PackAny()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid proof consensus")
	}

	return &MsgConnectionOpenAck{
		ConnectionID:    connectionID,
		ProofTry:        *proofTryAny,
		ProofConsensus:  *proofConsensusAny,
		ProofHeight:     proofHeight,
		ConsensusHeight: consensusHeight,
		Version:         version,
		Signer:          signer,
	}, nil
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
	proofTry := msg.GetProofTry()
	proofConsensus := msg.GetProofConsensus()

	if proofTry == nil || proofTry.IsEmpty() {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof try")
	}

	if proofConsensus == nil || proofConsensus.IsEmpty() {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof consensus")
	}
	if err := proofTry.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof try failed basic validation")
	}
	if err := proofConsensus.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof consensus failed basic validation")
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

// GetProofTry returns the cached value from ProofTry. It returns nil if the value
// is not cached or if the proof doesn't cast to a commitment Proof.
func (msg MsgConnectionOpenAck) GetProofTry() commitmentexported.Proof {
	proof, _ := commitmenttypes.UnpackAnyProof(&msg.ProofTry)
	return proof
}

// GetProofConsensus returns the cached value from ProofConsensus. It returns nil if the value
// is not cached or if the proof doesn't cast to a commitment Proof.
func (msg MsgConnectionOpenAck) GetProofConsensus() commitmentexported.Proof {
	proof, _ := commitmenttypes.UnpackAnyProof(&msg.ProofConsensus)
	return proof
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgConnectionOpenAck) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	var (
		proofTry       commitmentexported.Proof
		proofConsensus commitmentexported.Proof
	)
	err := unpacker.UnpackAny(&msg.ProofTry, &proofTry)
	if err != nil {
		return fmt.Errorf("proof try unpack failed: %w", err)
	}

	err = unpacker.UnpackAny(&msg.ProofConsensus, &proofConsensus)
	if err != nil {
		return fmt.Errorf("proof consensus unpack failed: %w", err)
	}
	return nil
}

var (
	_ sdk.Msg                          = &MsgConnectionOpenConfirm{}
	_ cdctypes.UnpackInterfacesMessage = &MsgConnectionOpenConfirm{}
)

// NewMsgConnectionOpenConfirm creates a new MsgConnectionOpenConfirm instance
func NewMsgConnectionOpenConfirm(
	connectionID string, proofAck commitmentexported.Proof, proofHeight uint64,
	signer sdk.AccAddress,
) (*MsgConnectionOpenConfirm, error) {
	proofAckAny, err := proofAck.PackAny()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid proof ack")
	}

	return &MsgConnectionOpenConfirm{
		ConnectionID: connectionID,
		ProofAck:     *proofAckAny,
		ProofHeight:  proofHeight,
		Signer:       signer,
	}, nil
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
	proofAck := msg.GetProofAck()
	if proofAck == nil || proofAck.IsEmpty() {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof ack")
	}
	if err := proofAck.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof ack failed basic validation")
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

// GetProofAck returns the cached value from ProofAck. It returns nil if the value
// is not cached or if the proof doesn't cast to a commitment Proof.
func (msg MsgConnectionOpenConfirm) GetProofAck() commitmentexported.Proof {
	proof, _ := commitmenttypes.UnpackAnyProof(&msg.ProofAck)
	return proof
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgConnectionOpenConfirm) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	var proofAck commitmentexported.Proof
	err := unpacker.UnpackAny(&msg.ProofAck, &proofAck)
	if err != nil {
		return fmt.Errorf("proof ack unpack failed: %w", err)
	}

	return nil
}
