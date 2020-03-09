package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Message types for the evidence module
const (
	TypeMsgSubmitEvidence = "submit_evidence"
)

var (
	_ sdk.Msg = MsgSubmitEvidenceBase{}
)

// NewMsgSubmitEvidenceBase returns a new MsgSubmitEvidenceBase with a signer/submitter.
// Note, the MsgSubmitEvidenceBase is not to be used as an actual message, but
// rather to be extended with Evidence.
func NewMsgSubmitEvidenceBase(s sdk.AccAddress) MsgSubmitEvidenceBase {
	return MsgSubmitEvidenceBase{Submitter: s}
}

// Route returns the MsgSubmitEvidenceBase's route.
func (m MsgSubmitEvidenceBase) Route() string { return RouterKey }

// Type returns the MsgSubmitEvidenceBase's type.
func (m MsgSubmitEvidenceBase) Type() string { return TypeMsgSubmitEvidence }

// ValidateBasic performs basic (non-state-dependant) validation on a MsgSubmitEvidenceBase.
func (m MsgSubmitEvidenceBase) ValidateBasic() error {
	if m.Submitter.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, m.Submitter.String())
	}

	return nil
}

// GetSignBytes returns the raw bytes a signer is expected to sign when submitting
// a MsgSubmitEvidenceBase message.
func (m MsgSubmitEvidenceBase) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// GetSigners returns the single expected signer for a MsgSubmitEvidenceBase.
func (m MsgSubmitEvidenceBase) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Submitter}
}
