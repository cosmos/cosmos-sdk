package types

import (
	"time"

	ics23 "github.com/confio/ics23/go"
	tmmath "github.com/tendermint/tendermint/libs/math"
	lite "github.com/tendermint/tendermint/lite2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// Message types for the IBC client
const (
	TypeMsgCreateClient             string = "create_client"
	TypeMsgUpdateClient             string = "update_client"
	TypeMsgSubmitClientMisbehaviour string = "submit_client_misbehaviour"
)

var (
	_ clientexported.MsgCreateClient     = &MsgCreateClient{}
	_ clientexported.MsgUpdateClient     = &MsgUpdateClient{}
	_ evidenceexported.MsgSubmitEvidence = &MsgSubmitClientMisbehaviour{}
)

// MsgCreateClient defines a message to create an IBC client
type MsgCreateClient struct {
	ClientID        string             `json:"client_id" yaml:"client_id"`
	Header          Header             `json:"header" yaml:"header"`
	TrustLevel      tmmath.Fraction    `json:"trust_level" yaml:"trust_level"`
	TrustingPeriod  time.Duration      `json:"trusting_period" yaml:"trusting_period"`
	UnbondingPeriod time.Duration      `json:"unbonding_period" yaml:"unbonding_period"`
	MaxClockDrift   time.Duration      `json:"max_clock_drift" yaml:"max_clock_drift"`
	ProofSpecs      []*ics23.ProofSpec `json:"proof_specs" yaml:"proof_specs"`
	Signer          sdk.AccAddress     `json:"address" yaml:"address"`
}

// this is a constant to satisfy the linter
const TODO = "TODO"

// dummy implementation of proto.Message
func (msg *MsgCreateClient) Reset()         {}
func (msg *MsgCreateClient) String() string { return TODO }
func (msg *MsgCreateClient) ProtoMessage()  {}

// NewMsgCreateClient creates a new MsgCreateClient instance
func NewMsgCreateClient(
	id string, header Header, trustLevel tmmath.Fraction,
	trustingPeriod, unbondingPeriod, maxClockDrift time.Duration,
	specs []*ics23.ProofSpec, signer sdk.AccAddress,
) *MsgCreateClient {

	return &MsgCreateClient{
		ClientID:        id,
		Header:          header,
		TrustLevel:      trustLevel,
		TrustingPeriod:  trustingPeriod,
		UnbondingPeriod: unbondingPeriod,
		MaxClockDrift:   maxClockDrift,
		ProofSpecs:      specs,
		Signer:          signer,
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
	if msg.TrustingPeriod == 0 {
		return sdkerrors.Wrap(ErrInvalidTrustingPeriod, "duration cannot be 0")
	}
	if err := lite.ValidateTrustLevel(msg.TrustLevel); err != nil {
		return err
	}
	if msg.UnbondingPeriod == 0 {
		return sdkerrors.Wrap(ErrInvalidUnbondingPeriod, "duration cannot be 0")
	}
	if msg.Signer.Empty() {
		return sdkerrors.ErrInvalidAddress
	}
	if msg.Header.SignedHeader.Header == nil {
		return sdkerrors.Wrap(ErrInvalidHeader, "header cannot be nil")
	}
	// ValidateBasic of provided header with self-attested chain-id
	if err := msg.Header.ValidateBasic(msg.Header.ChainID); err != nil {
		return sdkerrors.Wrapf(ErrInvalidHeader, "header failed validatebasic with its own chain-id: %v", err)
	}
	if msg.TrustingPeriod >= msg.UnbondingPeriod {
		return sdkerrors.Wrapf(
			ErrInvalidTrustingPeriod,
			"trusting period (%s) should be < unbonding period (%s)", msg.TrustingPeriod, msg.UnbondingPeriod,
		)
	}
	// Validate ProofSpecs
	if msg.ProofSpecs == nil {
		return sdkerrors.Wrap(ErrInvalidProofSpecs, "proof specs cannot be nil")
	}
	for _, spec := range msg.ProofSpecs {
		if spec == nil {
			return sdkerrors.Wrap(ErrInvalidProofSpecs, "proof spec cannot be nil")
		}
	}
	return host.ClientIdentifierValidator(msg.ClientID)
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
	root := commitmenttypes.NewMerkleRoot(msg.Header.AppHash)
	return ConsensusState{
		Timestamp:          msg.Header.Time,
		Root:               root,
		Height:             uint64(msg.Header.Height),
		NextValidatorsHash: msg.Header.NextValidatorsHash,
		ValidatorSet:       msg.Header.ValidatorSet,
	}
}

// MsgUpdateClient defines a message to update an IBC client
type MsgUpdateClient struct {
	ClientID string         `json:"client_id" yaml:"client_id"`
	Header   Header         `json:"header" yaml:"header"`
	Signer   sdk.AccAddress `json:"address" yaml:"address"`
}

// dummy implementation of proto.Message
func (msg *MsgUpdateClient) Reset()         {}
func (msg *MsgUpdateClient) String() string { return TODO }
func (msg *MsgUpdateClient) ProtoMessage()  {}

// NewMsgUpdateClient creates a new MsgUpdateClient instance
func NewMsgUpdateClient(id string, header Header, signer sdk.AccAddress) *MsgUpdateClient {
	return &MsgUpdateClient{
		ClientID: id,
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
		return sdkerrors.ErrInvalidAddress
	}
	return host.ClientIdentifierValidator(msg.ClientID)
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

// MsgSubmitClientMisbehaviour defines an sdk.Msg type that supports submitting
// Evidence for client misbehaviour.
type MsgSubmitClientMisbehaviour struct {
	Evidence  evidenceexported.Evidence `json:"evidence" yaml:"evidence"`
	Submitter sdk.AccAddress            `json:"submitter" yaml:"submitter"`
}

// dummy implementation of proto.Message
func (msg MsgSubmitClientMisbehaviour) Reset()         {}
func (msg MsgSubmitClientMisbehaviour) String() string { return TODO }
func (msg MsgSubmitClientMisbehaviour) ProtoMessage()  {}

// NewMsgSubmitClientMisbehaviour creates a new MsgSubmitClientMisbehaviour
// instance.
func NewMsgSubmitClientMisbehaviour(e evidenceexported.Evidence, s sdk.AccAddress) MsgSubmitClientMisbehaviour {
	return MsgSubmitClientMisbehaviour{Evidence: e, Submitter: s}
}

// Route returns the MsgSubmitClientMisbehaviour's route.
func (msg MsgSubmitClientMisbehaviour) Route() string { return host.RouterKey }

// Type returns the MsgSubmitClientMisbehaviour's type.
func (msg MsgSubmitClientMisbehaviour) Type() string {
	return TypeMsgSubmitClientMisbehaviour
}

// ValidateBasic performs basic (non-state-dependant) validation on a MsgSubmitClientMisbehaviour.
func (msg MsgSubmitClientMisbehaviour) ValidateBasic() error {
	if msg.Evidence == nil {
		return sdkerrors.Wrap(evidencetypes.ErrInvalidEvidence, "missing evidence")
	}
	if err := msg.Evidence.ValidateBasic(); err != nil {
		return err
	}
	if msg.Submitter.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Submitter.String())
	}

	return nil
}

// GetSignBytes returns the raw bytes a signer is expected to sign when submitting
// a MsgSubmitClientMisbehaviour message.
func (msg MsgSubmitClientMisbehaviour) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
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
