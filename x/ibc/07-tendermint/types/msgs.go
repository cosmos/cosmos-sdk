package types

import (
	"time"

	ics23 "github.com/confio/ics23/go"
	"github.com/tendermint/tendermint/light"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

var (
	_ clientexported.MsgCreateClient     = (*MsgCreateClient)(nil)
	_ clientexported.MsgUpdateClient     = (*MsgUpdateClient)(nil)
	_ evidenceexported.MsgSubmitEvidence = (*MsgSubmitClientMisbehaviour)(nil)
)

// NewMsgCreateClient creates a new MsgCreateClient instance
func NewMsgCreateClient(
	id string, header *Header, trustLevel Fraction,
	trustingPeriod, unbondingPeriod, maxClockDrift time.Duration,
	specs []*ics23.ProofSpec, allowGovernanceOverrideAfterExpiry bool,
	allowGovernanceOverrideAfterMisbehaviour bool, signer sdk.AccAddress,
) *MsgCreateClient {

	return &MsgCreateClient{
		ClientId:                                 id,
		Header:                                   header,
		TrustLevel:                               trustLevel,
		TrustingPeriod:                           trustingPeriod,
		UnbondingPeriod:                          unbondingPeriod,
		MaxClockDrift:                            maxClockDrift,
		ProofSpecs:                               specs,
		Signer:                                   signer,
		AllowGovernanceOverrideAfterExpiry:       allowGovernanceOverrideAfterExpiry,
		AllowGovernanceOverrideAfterMisbehaviour: allowGovernanceOverrideAfterMisbehaviour,
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
	if msg.TrustingPeriod == 0 {
		return sdkerrors.Wrap(ErrInvalidTrustingPeriod, "duration cannot be 0")
	}
	if err := light.ValidateTrustLevel(msg.TrustLevel.ToTendermint()); err != nil {
		return sdkerrors.Wrap(err, "invalid trust level for tendermint light client")
	}
	if msg.UnbondingPeriod == 0 {
		return sdkerrors.Wrap(ErrInvalidUnbondingPeriod, "duration cannot be 0")
	}
	if msg.Signer.Empty() {
		return sdkerrors.ErrInvalidAddress
	}

	if msg.Header == nil || msg.Header.SignedHeader == nil || msg.Header.SignedHeader.Header == nil {
		return sdkerrors.Wrap(ErrInvalidHeader, "header cannot be nil")
	}
	// ValidateBasic of provided header with self-attested chain-id
	if err := msg.Header.ValidateBasic(msg.Header.Header.GetChainID()); err != nil {
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

// GetClientID implements clientexported.MsgCreateClient
func (msg MsgCreateClient) GetClientID() string {
	return msg.ClientId
}

// GetClientType implements clientexported.MsgCreateClient
func (msg MsgCreateClient) GetClientType() string {
	return clientexported.ClientTypeTendermint
}

// GetConsensusState implements clientexported.MsgCreateClient
func (msg MsgCreateClient) GetConsensusState() clientexported.ConsensusState {
	// Construct initial consensus state from provided Header
	root := commitmenttypes.NewMerkleRoot(msg.Header.Header.GetAppHash())
	return &ConsensusState{
		Timestamp:          msg.Header.GetTime(),
		Root:               root,
		Height:             msg.Header.GetHeight(),
		NextValidatorsHash: msg.Header.Header.NextValidatorsHash,
	}
}

// InitializeFromMsg creates a tendermint client state from a CreateClientMsg
func (msg MsgCreateClient) InitializeClientState() clientexported.ClientState {
	return NewClientState(msg.Header.Header.GetChainID(), msg.TrustLevel,
		msg.TrustingPeriod, msg.UnbondingPeriod, msg.MaxClockDrift,
		msg.Header.GetHeight(), msg.Header.GetTime(), msg.ProofSpecs, msg.AllowGovernanceOverrideAfterExpiry, msg.AllowGovernanceOverrideAfterMisbehaviour,
	)
}

// NewMsgUpdateClient creates a new MsgUpdateClient instance
func NewMsgUpdateClient(id string, header *Header, signer sdk.AccAddress) *MsgUpdateClient {
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
	return clientexported.TypeMsgUpdateClient
}

// ValidateBasic implements sdk.Msg
func (msg MsgUpdateClient) ValidateBasic() error {
	if msg.Signer.Empty() {
		return sdkerrors.ErrInvalidAddress
	}
	if msg.Header == nil || msg.Header.SignedHeader == nil || msg.Header.Header == nil {
		return sdkerrors.Wrap(ErrInvalidHeader, "header cannot be nil")
	}
	// ValidateBasic of provided header with self-attested chain-id
	if err := msg.Header.ValidateBasic(msg.Header.Header.GetChainID()); err != nil {
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
