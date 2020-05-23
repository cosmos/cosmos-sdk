package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

var _ sdk.Msg = MsgConnectionOpenInit{}

// NewMsgConnectionOpenInit creates a new MsgConnectionOpenInit instance
func NewMsgConnectionOpenInit(
	connectionID, clientID, counterpartyConnectionID,
	counterpartyClientID string, counterpartyPrefix commitmentexported.Prefix,
	signer sdk.AccAddress,
) (MsgConnectionOpenInit, error) {
	counterparty, err := NewCounterparty(counterpartyClientID, counterpartyConnectionID, counterpartyPrefix)
	if err != nil {
		return MsgConnectionOpenInit{}, err
	}

	return MsgConnectionOpenInit{
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

var _ sdk.Msg = MsgConnectionOpenTry{}

// NewMsgConnectionOpenTry creates a new MsgConnectionOpenTry instance
func NewMsgConnectionOpenTry(
	connectionID, clientID, counterpartyConnectionID,
	counterpartyClientID string, counterpartyPrefix commitmentexported.Prefix,
	counterpartyVersions []string, proofInit, proofConsensus commitmentexported.Proof,
	proofHeight, consensusHeight uint64, signer sdk.AccAddress,
) (MsgConnectionOpenTry, error) {
	counterparty, err := NewCounterparty(counterpartyClientID, counterpartyConnectionID, counterpartyPrefix)
	if err != nil {
		return MsgConnectionOpenTry{}, err
	}

	proofInitAny, err := proofInit.PackAny()
	if err != nil {
		return MsgConnectionOpenTry{}, sdkerrors.Wrap(err, "proof init")
	}

	proofConsensusAny, err := proofConsensus.PackAny()
	if err != nil {
		return MsgConnectionOpenTry{}, sdkerrors.Wrap(err, "proof consensus")
	}

	return MsgConnectionOpenTry{
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
	proofInit, ok := msg.ProofInit.GetCachedValue().(commitmentexported.Proof)
	if !ok {
		return sdkerrors.Wrap(sdkerrors.ErrProtobufAny, "merkle proof init is not cached")
	}
	proofConsensus, ok := msg.ProofConsensus.GetCachedValue().(commitmentexported.Proof)
	if !ok {
		return sdkerrors.Wrap(sdkerrors.ErrProtobufAny, "merkle proof consensus is not cached")
	}
	if proofInit.IsEmpty() || proofConsensus.IsEmpty() {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if err := proofInit.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof init")
	}
	if err := proofConsensus.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof consensus")
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

var _ sdk.Msg = MsgConnectionOpenAck{}

// NewMsgConnectionOpenAck creates a new MsgConnectionOpenAck instance
func NewMsgConnectionOpenAck(
	connectionID string, proofTry, proofConsensus commitmentexported.Proof,
	proofHeight, consensusHeight uint64, version string,
	signer sdk.AccAddress,
) (MsgConnectionOpenAck, error) {
	proofTryAny, err := proofTry.PackAny()
	if err != nil {
		return MsgConnectionOpenAck{}, sdkerrors.Wrap(err, "proof try")
	}

	proofConsensusAny, err := proofConsensus.PackAny()
	if err != nil {
		return MsgConnectionOpenAck{}, sdkerrors.Wrap(err, "proof consensus")
	}

	return MsgConnectionOpenAck{
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
	proofTry, ok := msg.ProofTry.GetCachedValue().(commitmentexported.Proof)
	if !ok {
		return sdkerrors.Wrap(sdkerrors.ErrProtobufAny, "merkle proof try is not cached")
	}
	proofConsensus, ok := msg.ProofConsensus.GetCachedValue().(commitmentexported.Proof)
	if !ok {
		return sdkerrors.Wrap(sdkerrors.ErrProtobufAny, "merkle proof consensus is not cached")
	}
	if proofTry.IsEmpty() || proofConsensus.IsEmpty() {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if err := proofTry.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof try")
	}
	if err := proofConsensus.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof consensus")
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

var _ sdk.Msg = MsgConnectionOpenConfirm{}

// NewMsgConnectionOpenConfirm creates a new MsgConnectionOpenConfirm instance
func NewMsgConnectionOpenConfirm(
	connectionID string, proofAck commitmentexported.Proof, proofHeight uint64,
	signer sdk.AccAddress,
) (MsgConnectionOpenConfirm, error) {
	proofAckAny, err := proofAck.PackAny()
	if err != nil {
		return MsgConnectionOpenConfirm{}, sdkerrors.Wrap(err, "proof ack")
	}

	return MsgConnectionOpenConfirm{
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
	proofAck, ok := msg.ProofAck.GetCachedValue().(commitmentexported.Proof)
	if !ok {
		return sdkerrors.Wrap(sdkerrors.ErrProtobufAny, "merkle proof ack is not cached")
	}
	if proofAck.IsEmpty() {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof")
	}
	if err := proofAck.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proof ack")
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
