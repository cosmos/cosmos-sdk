package std

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	eviexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feeexported "github.com/cosmos/cosmos-sdk/x/feegrant/exported"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

var (
	_ eviexported.MsgSubmitEvidence    = MsgSubmitEvidence{}
	_ gov.MsgSubmitProposalI           = MsgSubmitProposal{}
	_ feeexported.MsgGrantFeeAllowance = MsgGrantFeeAllowance{}
)

// NewMsgSubmitEvidence returns a new MsgSubmitEvidence.
func NewMsgSubmitEvidence(evidenceI eviexported.Evidence, s sdk.AccAddress) (MsgSubmitEvidence, error) {
	e := &Evidence{}
	if err := e.SetEvidence(evidenceI); err != nil {
		return MsgSubmitEvidence{}, err
	}

	return MsgSubmitEvidence{
		Evidence:              e,
		MsgSubmitEvidenceBase: evidence.NewMsgSubmitEvidenceBase(s),
	}, nil
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

// NewMsgSubmitProposal returns a new MsgSubmitProposal.
func NewMsgSubmitProposal(c gov.Content, d sdk.Coins, p sdk.AccAddress) (MsgSubmitProposal, error) {
	content := &Content{}
	if err := content.SetContent(c); err != nil {
		return MsgSubmitProposal{}, err
	}

	return MsgSubmitProposal{
		Content:               content,
		MsgSubmitProposalBase: gov.NewMsgSubmitProposalBase(d, p),
	}, nil
}

// ValidateBasic performs basic (non-state-dependant) validation on a
// MsgSubmitProposal.
func (msg MsgSubmitProposal) ValidateBasic() error {
	if err := msg.MsgSubmitProposalBase.ValidateBasic(); err != nil {
		return nil
	}
	if msg.Content == nil {
		return sdkerrors.Wrap(gov.ErrInvalidProposalContent, "missing content")
	}
	if !gov.IsValidProposalType(msg.Content.GetContent().ProposalType()) {
		return sdkerrors.Wrap(gov.ErrInvalidProposalType, msg.Content.GetContent().ProposalType())
	}
	if err := msg.Content.GetContent().ValidateBasic(); err != nil {
		return err
	}

	return nil
}

// nolint
func (msg MsgSubmitProposal) GetContent() gov.Content      { return msg.Content.GetContent() }
func (msg MsgSubmitProposal) GetInitialDeposit() sdk.Coins { return msg.InitialDeposit }
func (msg MsgSubmitProposal) GetProposer() sdk.AccAddress  { return msg.Proposer }

func (msg MsgGrantFeeAllowance) NewMsgGrantFeeAllowance(feeAllowanceI feeexported.FeeAllowance, granter, grantee sdk.AccAddress) (MsgGrantFeeAllowance, error) {
	feeallowance := &FeeAllowance{}

	if err := feeallowance.SetFeeAllowance(feeAllowanceI); err != nil {
		return MsgGrantFeeAllowance{}, err
	}

	return MsgGrantFeeAllowance{
		Allowance:                feeallowance,
		MsgGrantFeeAllowanceBase: feegrant.NewMsgGrantFeeAllowanceBase(granter, grantee),
	}, nil
}

func (msg MsgGrantFeeAllowance) ValidateBasic() error {
	//TODO
	return nil
}

func (msg MsgGrantFeeAllowance) GetFeeGrant() feeexported.FeeAllowance {
	return msg.Allowance.GetFeeAllowance()
}

func (msg MsgGrantFeeAllowance) GetGrantee() sdk.AccAddress {
	return msg.Grantee
}

func (msg MsgGrantFeeAllowance) GetGranter() sdk.AccAddress {
	return msg.Granter
}

// PrepareForExport will make all needed changes to the allowance to prepare to be
// re-imported at height 0, and return a copy of this grant.
func (a MsgGrantFeeAllowance) PrepareForExport(dumpTime time.Time, dumpHeight int64) feeexported.FeeAllowanceGrant {
	err := a.GetFeeGrant().PrepareForExport(dumpTime, dumpHeight)
	if err != nil {
		//TODO handle this error
	}
	return a
}

// ValidateBasic performs basic validation on
// FeeAllowanceGrant
func (a FeeAllowanceGrant) ValidateBasic() error {
	if a.Granter.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing granter address")
	}
	if a.Grantee.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing grantee address")
	}
	if a.Grantee.Equals(a.Granter) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "cannot self-grant fee authorization")
	}

	return a.ValidateBasic()
}

func (grant FeeAllowanceGrant) GetFeeGrant() feeexported.FeeAllowance {
	return grant.Allowance.GetFeeAllowance()
}

func (grant FeeAllowanceGrant) GetGrantee() sdk.AccAddress {
	return grant.Grantee
}

func (grant FeeAllowanceGrant) GetGranter() sdk.AccAddress {
	return grant.Granter
}

// PrepareForExport will make all needed changes to the allowance to prepare to be
// re-imported at height 0, and return a copy of this grant.
func (a FeeAllowanceGrant) PrepareForExport(dumpTime time.Time, dumpHeight int64) feeexported.FeeAllowanceGrant {
	err := a.GetFeeGrant().PrepareForExport(dumpTime, dumpHeight)
	if err != nil {
		//TODO handle this error
	}
	return a
}
