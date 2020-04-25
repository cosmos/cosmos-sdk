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
	_ sdk.Msg                    = MsgSubmitEvidence{}
	_ sdk.InterfaceMsg           = MsgSubmitEvidence{}
	_ exported.MsgSubmitEvidence = MsgSubmitEvidence{}
)

// NewMsgSubmitEvidence returns a new MsgSubmitEvidence with a signer/submitter.
func NewMsgSubmitEvidence(s sdk.AccAddress, evi exported.Evidence) (MsgSubmitEvidence, error) {
	any, err := sdk.NewAnyWithValue(evi)
	if err != nil {
		return MsgSubmitEvidence{Submitter: s}, err
	}
	return MsgSubmitEvidence{Submitter: s, Evidence: any}, nil
}

// Route returns the MsgSubmitEvidenceBase's route.
func (m MsgSubmitEvidence) Route() string { return RouterKey }

// Type returns the MsgSubmitEvidenceBase's type.
func (m MsgSubmitEvidence) Type() string { return TypeMsgSubmitEvidence }

// ValidateBasic performs basic (non-state-dependant) validation on a MsgSubmitEvidenceBase.
func (m MsgSubmitEvidence) ValidateBasic() error {
	if m.Submitter.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, m.Submitter.String())
	}

	return nil
}

// GetSignBytes returns the raw bytes a signer is expected to sign when submitting
// a MsgSubmitEvidenceBase message.
func (m MsgSubmitEvidence) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// GetSigners returns the single expected signer for a MsgSubmitEvidenceBase.
func (m MsgSubmitEvidence) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Submitter}
}

func (m MsgSubmitEvidence) GetEvidence() exported.Evidence {
	evi, ok := m.Evidence.CachedValue().(exported.Evidence)
	if !ok {
		return nil
	}
	return evi
}

func (m MsgSubmitEvidence) GetSubmitter() sdk.AccAddress {
	return m.Submitter
}

func (m MsgSubmitEvidence) Rehydrate(ctx sdk.InterfaceContext) error {
	var evi exported.Evidence
	return ctx.UnpackAny(m.Evidence, &evi)
}
