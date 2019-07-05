package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Governance message types and routes
const (
	TypeMsgDeposit        = "deposit"
	TypeMsgVote           = "vote"
	TypeMsgSubmitProposal = "submit_proposal"
)

var _, _, _ sdk.Msg = MsgSubmitProposal{}, MsgDeposit{}, MsgVote{}

// MsgSubmitProposal
type MsgSubmitProposal struct {
	Content        Content        `json:"content" yaml:"content"`
	InitialDeposit sdk.Coins      `json:"initial_deposit" yaml:"initial_deposit"` //  Initial deposit paid by sender. Must be strictly positive
	Proposer       sdk.AccAddress `json:"proposer" yaml:"proposer"`               //  Address of the proposer
}

func NewMsgSubmitProposal(content Content, initialDeposit sdk.Coins, proposer sdk.AccAddress) MsgSubmitProposal {
	return MsgSubmitProposal{content, initialDeposit, proposer}
}

//nolint
func (msg MsgSubmitProposal) Route() string { return RouterKey }
func (msg MsgSubmitProposal) Type() string  { return TypeMsgSubmitProposal }

// Implements Msg.
func (msg MsgSubmitProposal) ValidateBasic() sdk.Error {
	if msg.Content == nil {
		return ErrInvalidProposalContent(DefaultCodespace, "missing content")
	}
	if msg.Content.ProposalType() == ProposalTypeSoftwareUpgrade {
		// Disable software upgrade proposals as they are currently equivalent
		// to text proposals. Re-enable once a valid software upgrade proposal
		// handler is implemented.
		return ErrInvalidProposalType(DefaultCodespace, msg.Content.ProposalType())
	}
	if msg.Proposer.Empty() {
		return sdk.ErrInvalidAddress(msg.Proposer.String())
	}
	if !msg.InitialDeposit.IsValid() {
		return sdk.ErrInvalidCoins(msg.InitialDeposit.String())
	}
	if msg.InitialDeposit.IsAnyNegative() {
		return sdk.ErrInvalidCoins(msg.InitialDeposit.String())
	}
	if !IsValidProposalType(msg.Content.ProposalType()) {
		return ErrInvalidProposalType(DefaultCodespace, msg.Content.ProposalType())
	}

	return msg.Content.ValidateBasic()
}

func (msg MsgSubmitProposal) String() string {
	return fmt.Sprintf(`Submit Proposal Message:
  Content:         %s
  Initial Deposit: %s
`, msg.Content.String(), msg.InitialDeposit)
}

// Implements Msg.
func (msg MsgSubmitProposal) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// Implements Msg.
func (msg MsgSubmitProposal) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Proposer}
}

// MsgDeposit
type MsgDeposit struct {
	ProposalID uint64         `json:"proposal_id" yaml:"proposal_id"` // ID of the proposal
	Depositor  sdk.AccAddress `json:"depositor" yaml:"depositor"`     // Address of the depositor
	Amount     sdk.Coins      `json:"amount" yaml:"amount"`           // Coins to add to the proposal's deposit
}

func NewMsgDeposit(depositor sdk.AccAddress, proposalID uint64, amount sdk.Coins) MsgDeposit {
	return MsgDeposit{proposalID, depositor, amount}
}

// Implements Msg.
// nolint
func (msg MsgDeposit) Route() string { return RouterKey }
func (msg MsgDeposit) Type() string  { return TypeMsgDeposit }

// Implements Msg.
func (msg MsgDeposit) ValidateBasic() sdk.Error {
	if msg.Depositor.Empty() {
		return sdk.ErrInvalidAddress(msg.Depositor.String())
	}
	if !msg.Amount.IsValid() {
		return sdk.ErrInvalidCoins(msg.Amount.String())
	}
	if msg.Amount.IsAnyNegative() {
		return sdk.ErrInvalidCoins(msg.Amount.String())
	}

	return nil
}

func (msg MsgDeposit) String() string {
	return fmt.Sprintf(`Deposit Message:
  Depositer:   %s
  Proposal ID: %d
  Amount:      %s
`, msg.Depositor, msg.ProposalID, msg.Amount)
}

// Implements Msg.
func (msg MsgDeposit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// Implements Msg.
func (msg MsgDeposit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Depositor}
}

// MsgVote
type MsgVote struct {
	ProposalID uint64         `json:"proposal_id" yaml:"proposal_id"` // ID of the proposal
	Voter      sdk.AccAddress `json:"voter" yaml:"voter"`             //  address of the voter
	Option     VoteOption     `json:"option" yaml:"option"`           //  option from OptionSet chosen by the voter
}

func NewMsgVote(voter sdk.AccAddress, proposalID uint64, option VoteOption) MsgVote {
	return MsgVote{proposalID, voter, option}
}

// Implements Msg.
// nolint
func (msg MsgVote) Route() string { return RouterKey }
func (msg MsgVote) Type() string  { return TypeMsgVote }

// Implements Msg.
func (msg MsgVote) ValidateBasic() sdk.Error {
	if msg.Voter.Empty() {
		return sdk.ErrInvalidAddress(msg.Voter.String())
	}
	if !ValidVoteOption(msg.Option) {
		return ErrInvalidVote(DefaultCodespace, msg.Option)
	}

	return nil
}

func (msg MsgVote) String() string {
	return fmt.Sprintf(`Vote Message:
  Proposal ID: %d
  Option:      %s
`, msg.ProposalID, msg.Option)
}

// Implements Msg.
func (msg MsgVote) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// Implements Msg.
func (msg MsgVote) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Voter}
}
