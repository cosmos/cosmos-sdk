package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Governance message types and routes
const (
	TypeMsgDeposit        = "deposit"
	TypeMsgVote           = "vote"
	TypeMsgSubmitProposal = "submit_proposal"
)

var _, _ sdk.Msg = MsgDeposit{}, MsgVote{}

type MsgSubmitProposal interface {
	sdk.Msg

	GetContent() Content
	SetContent(content Content) error

	GetInitialDeposit() sdk.Coins
	SetInitialDeposit(coins sdk.Coins)

	GetProposer() sdk.AccAddress
	SetProposer(proposer sdk.AccAddress)
}

// NewMsgSubmitProposal creates a new MsgSubmitProposal instance
//func NewMsgSubmitProposal(content Content, initialDeposit sdk.Coins, proposer sdk.AccAddress) MsgSubmitProposal {
//	return MsgSubmitProposal{content, initialDeposit, proposer}
//}

// Route implements Msg
func (msg MsgSubmitProposalBase) Route() string { return RouterKey }

// Type implements Msg
func (msg MsgSubmitProposalBase) Type() string { return TypeMsgSubmitProposal }

// ValidateBasic implements Msg
func (msg MsgSubmitProposalBase) ValidateBasic() error {
	//if msg.Content == nil {
	//	return ErrInvalidProposalContent(DefaultCodespace, "missing content")
	//}
	if msg.Proposer.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Proposer.String())
	}
	initialDeposit := sdk.Coins(msg.IntialDeposit)
	if !initialDeposit.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, initialDeposit.String())
	}
	if initialDeposit.IsAnyNegative() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, initialDeposit.String())
	}
	//if !IsValidProposalType(msg.Content.ProposalType()) {
	//	return ErrInvalidProposalType(DefaultCodespace, msg.Content.ProposalType())
	//}
	//
	//return msg.Content.ValidateBasic()
	return nil
}

// String implements the Stringer interface
//func (msg MsgSubmitProposal) String() string {
//	return fmt.Sprintf(`Submit Proposal Message:
//  Content:         %s
//  Initial Deposit: %s
//`, msg.Content.String(), msg.InitialDeposit)
//}

// GetSigners implements Msg
func (msg MsgSubmitProposalBase) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Proposer}
}

func (m MsgSubmitProposalBase) GetInitialDeposit() sdk.Coins {
	return m.IntialDeposit
}

func (m MsgSubmitProposalBase) GetProposer() sdk.AccAddress {
	return m.Proposer
}

// NewMsgDeposit creates a new MsgDeposit instance
func NewMsgDeposit(depositor sdk.AccAddress, proposalID uint64, amount sdk.Coins) MsgDeposit {
	return MsgDeposit{proposalID, depositor, amount}
}

// Route implements Msg
func (msg MsgDeposit) Route() string { return RouterKey }

// Type implements Msg
func (msg MsgDeposit) Type() string { return TypeMsgDeposit }

// ValidateBasic implements Msg
func (msg MsgDeposit) ValidateBasic() error {
	if msg.Depositor.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Depositor.String())
	}
	amount := sdk.Coins(msg.Amount)
	if !amount.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, amount.String())
	}
	if amount.IsAnyNegative() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, amount.String())
	}

	return nil
}

// String implements the Stringer interface
func (msg MsgDeposit) String() string {
	return fmt.Sprintf(`Deposit Message:
  Depositer:   %s
  Proposal ID: %d
  Amount:      %s
`, msg.Depositor, msg.ProposalID, msg.Amount)
}

// GetSignBytes implements Msg
//func (msg MsgDeposit) GetSignBytes() []byte {
//	bz := ModuleCdc.MustMarshalJSON(msg)
//	return sdk.MustSortJSON(bz)
//}

// GetSigners implements Msg
func (msg MsgDeposit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Depositor}
}

// NewMsgVote creates a message to cast a vote on an active proposal
func NewMsgVote(voter sdk.AccAddress, proposalID uint64, option VoteOption) MsgVote {
	return MsgVote{proposalID, voter, option}
}

// Route implements Msg
func (msg MsgVote) Route() string { return RouterKey }

// Type implements Msg
func (msg MsgVote) Type() string { return TypeMsgVote }

// ValidateBasic implements Msg
func (msg MsgVote) ValidateBasic() error {
	if msg.Voter.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Voter.String())
	}
	if !ValidVoteOption(msg.Option) {
		return sdkerrors.Wrap(ErrInvalidVote, msg.Option.String())
	}

	return nil
}

// String implements the Stringer interface
func (msg MsgVote) String() string {
	return fmt.Sprintf(`Vote Message:
  Proposal ID: %d
  Option:      %s
`, msg.ProposalID, msg.Option)
}

// GetSignBytes implements Msg
//func (msg MsgVote) GetSignBytes() []byte {
//	bz := ModuleCdc.MustMarshalJSON(msg)
//	return sdk.MustSortJSON(bz)
//}

// GetSigners implements Msg
func (msg MsgVote) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Voter}
}
