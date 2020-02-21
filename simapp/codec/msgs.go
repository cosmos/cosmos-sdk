package codec

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	eviexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
)

// NewMsgSubmitEvidence returns a new MsgSubmitEvidence. The evidence adhering
// to the Evidence interface must be a reference.
func NewMsgSubmitEvidence(evidenceI eviexported.Evidence, s sdk.AccAddress) MsgSubmitEvidence {
	e := &Evidence{}
	e.SetEvidence(evidenceI)

	return MsgSubmitEvidence{Evidence: e, MsgSubmitEvidenceBase: evidence.NewMsgSubmitEvidenceBase(s)}
}

// ValidateBasic performs basic (non-state-dependant) validation on a
// MsgSubmitEvidence.
func (msg MsgSubmitEvidence) ValidateBasic() error {
	if err := msg.MsgSubmitEvidenceBase.ValidateBasic(); err != nil {
		return nil
	}
	if msg.Evidence == nil {
		return sdkerrors.Wrap(evidence.ErrInvalidEvidence, "missing evidence")
	}
	if err := msg.Evidence.GetEvidence().ValidateBasic(); err != nil {
		return err
	}

	return nil
}
