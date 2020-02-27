package codec

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	eviexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
)

var _ eviexported.MsgSubmitEvidence = MsgSubmitEvidence{}

// NewMsgSubmitEvidence returns a new MsgSubmitEvidence.
func NewMsgSubmitEvidence(evidenceI eviexported.Evidence, s sdk.AccAddress) (MsgSubmitEvidence, error) {
	e := &Evidence{}
	if err := e.SetEvidence(evidenceI); err != nil {
		return MsgSubmitEvidence{}, err
	}

	return MsgSubmitEvidence{Evidence: e, MsgSubmitEvidenceBase: evidence.NewMsgSubmitEvidenceBase(s)}, nil
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

// nolint
func (msg MsgSubmitEvidence) GetEvidence() eviexported.Evidence { return msg.Evidence.GetEvidence() }
func (msg MsgSubmitEvidence) GetSubmitter() sdk.AccAddress      { return msg.Submitter }
