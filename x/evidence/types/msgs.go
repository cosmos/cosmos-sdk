package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
)

// Message types for the evidence module
const (
	TypeMsgSubmitEvidence = "submit_evidence"
)

var (
	_ sdk.Msg = MsgSubmitEvidence{}
)

func NewMsgSubmitEvidence(e exported.EvidenceI, s sdk.AccAddress) MsgSubmitEvidence {
	// Codec.MarshalEvidence(e)
	return MsgSubmitEvidence{Evidence: e, Submitter: s}
}

// Route returns the MsgSubmitEvidence's route.
func (m MsgSubmitEvidence) Route() string { return RouterKey }

// Type returns the MsgSubmitEvidence's type.
func (m MsgSubmitEvidence) Type() string { return TypeMsgSubmitEvidence }

// ValidateBasic performs basic (non-state-dependant) validation on a MsgSubmitEvidence.
func (m MsgSubmitEvidence) ValidateBasic() error {
	if m.Evidence.GetEvidenceI() == nil {
		return sdkerrors.Wrap(ErrInvalidEvidence, "missing evidence")
	}
	if err := m.Evidence.GetEvidenceI().ValidateBasic(); err != nil {
		return err
	}
	if m.Submitter.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, m.Submitter.String())
	}

	return nil
}

// GetSignBytes returns the raw bytes a signer is expected to sign when submitting
// a MsgSubmitEvidence message.
func (m MsgSubmitEvidence) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// GetSigners returns the single expected signer for a MsgSubmitEvidence.
func (m MsgSubmitEvidence) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Submitter}
}
