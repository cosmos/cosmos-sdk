package std

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var (
	_ gov.MsgSubmitProposalI = &MsgSubmitProposal{}
)

// NewMsgSubmitProposal returns a new MsgSubmitProposal.
func NewMsgSubmitProposal(c gov.Content, d sdk.Coins, p sdk.AccAddress) (gov.MsgSubmitProposalI, error) {
	content := &Content{}
	if err := content.SetContent(c); err != nil {
		return nil, err
	}

	return &MsgSubmitProposal{
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

func (msg *MsgSubmitProposal) GetContent() gov.Content { return msg.Content.GetContent() }
func (msg *MsgSubmitProposal) SetContent(content gov.Content) error {
	stdContent := &Content{}
	err := stdContent.SetContent(content)
	if err != nil {
		return err
	}
	msg.Content = stdContent
	return nil
}
