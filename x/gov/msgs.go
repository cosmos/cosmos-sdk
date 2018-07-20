package gov

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// name to idetify transaction types
const MsgType = "gov"

//-----------------------------------------------------------
// MsgSubmitProposal
type MsgSubmitProposal struct {
	Title          string         //  Title of the proposal
	Description    string         //  Description of the proposal
	ProposalType   ProposalKind   //  Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}
	Proposer       sdk.AccAddress //  Address of the proposer
	InitialDeposit sdk.Coins      //  Initial deposit paid by sender. Must be strictly positive.
}

func NewMsgSubmitProposal(title string, description string, proposalType ProposalKind, proposer sdk.AccAddress, initialDeposit sdk.Coins) MsgSubmitProposal {
	return MsgSubmitProposal{
		Title:          title,
		Description:    description,
		ProposalType:   proposalType,
		Proposer:       proposer,
		InitialDeposit: initialDeposit,
	}
}

// Implements Msg.
func (msg MsgSubmitProposal) Type() string { return MsgType }

// Implements Msg.
func (msg MsgSubmitProposal) ValidateBasic() sdk.Error {
	if len(msg.Title) == 0 {
		return ErrInvalidTitle(DefaultCodespace, msg.Title) // TODO: Proper Error
	}
	if len(msg.Description) == 0 {
		return ErrInvalidDescription(DefaultCodespace, msg.Description) // TODO: Proper Error
	}
	if !validProposalType(msg.ProposalType) {
		return ErrInvalidProposalType(DefaultCodespace, msg.ProposalType)
	}
	if len(msg.Proposer) == 0 {
		return sdk.ErrInvalidAddress(msg.Proposer.String())
	}
	if !msg.InitialDeposit.IsValid() {
		return sdk.ErrInvalidCoins(msg.InitialDeposit.String())
	}
	if !msg.InitialDeposit.IsNotNegative() {
		return sdk.ErrInvalidCoins(msg.InitialDeposit.String())
	}
	return nil
}

func (msg MsgSubmitProposal) String() string {
	return fmt.Sprintf("MsgSubmitProposal{%s, %s, %s, %v}", msg.Title, msg.Description, msg.ProposalType, msg.InitialDeposit)
}

// Implements Msg.
func (msg MsgSubmitProposal) Get(key interface{}) (value interface{}) {
	return nil
}

// Implements Msg.
func (msg MsgSubmitProposal) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// Implements Msg.
func (msg MsgSubmitProposal) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Proposer}
}

//-----------------------------------------------------------
// MsgDeposit
type MsgDeposit struct {
	ProposalID int64          `json:"proposalID"` // ID of the proposal
	Depositer  sdk.AccAddress `json:"depositer"`  // Address of the depositer
	Amount     sdk.Coins      `json:"amount"`     // Coins to add to the proposal's deposit
}

func NewMsgDeposit(depositer sdk.AccAddress, proposalID int64, amount sdk.Coins) MsgDeposit {
	return MsgDeposit{
		ProposalID: proposalID,
		Depositer:  depositer,
		Amount:     amount,
	}
}

// Implements Msg.
func (msg MsgDeposit) Type() string { return MsgType }

// Implements Msg.
func (msg MsgDeposit) ValidateBasic() sdk.Error {
	if len(msg.Depositer) == 0 {
		return sdk.ErrInvalidAddress(msg.Depositer.String())
	}
	if !msg.Amount.IsValid() {
		return sdk.ErrInvalidCoins(msg.Amount.String())
	}
	if !msg.Amount.IsNotNegative() {
		return sdk.ErrInvalidCoins(msg.Amount.String())
	}
	if msg.ProposalID < 0 {
		return ErrUnknownProposal(DefaultCodespace, msg.ProposalID)
	}
	return nil
}

func (msg MsgDeposit) String() string {
	return fmt.Sprintf("MsgDeposit{%s=>%v: %v}", msg.Depositer, msg.ProposalID, msg.Amount)
}

// Implements Msg.
func (msg MsgDeposit) Get(key interface{}) (value interface{}) {
	return nil
}

// Implements Msg.
func (msg MsgDeposit) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// Implements Msg.
func (msg MsgDeposit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Depositer}
}

//-----------------------------------------------------------
// MsgVote
type MsgVote struct {
	ProposalID int64          //  proposalID of the proposal
	Voter      sdk.AccAddress //  address of the voter
	Option     VoteOption     //  option from OptionSet chosen by the voter
}

func NewMsgVote(voter sdk.AccAddress, proposalID int64, option VoteOption) MsgVote {
	return MsgVote{
		ProposalID: proposalID,
		Voter:      voter,
		Option:     option,
	}
}

// Implements Msg.
func (msg MsgVote) Type() string { return MsgType }

// Implements Msg.
func (msg MsgVote) ValidateBasic() sdk.Error {
	if len(msg.Voter.Bytes()) == 0 {
		return sdk.ErrInvalidAddress(msg.Voter.String())
	}
	if msg.ProposalID < 0 {
		return ErrUnknownProposal(DefaultCodespace, msg.ProposalID)
	}
	if !validVoteOption(msg.Option) {
		return ErrInvalidVote(DefaultCodespace, msg.Option)
	}
	return nil
}

func (msg MsgVote) String() string {
	return fmt.Sprintf("MsgVote{%v - %s}", msg.ProposalID, msg.Option)
}

// Implements Msg.
func (msg MsgVote) Get(key interface{}) (value interface{}) {
	return nil
}

// Implements Msg.
func (msg MsgVote) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// Implements Msg.
func (msg MsgVote) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Voter}
}
