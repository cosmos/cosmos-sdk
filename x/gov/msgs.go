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
	Title          string       //  Title of the proposal
	Description    string       //  Description of the proposal
	ProposalType   ProposalKind //  Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}
	Proposer       sdk.Address  //  Address of the proposer
	InitialDeposit sdk.Coins    //  Initial deposit paid by sender. Must be strictly positive.
}

func NewMsgSubmitProposal(title string, description string, proposalType ProposalKind, proposer sdk.Address, initialDeposit sdk.Coins) MsgSubmitProposal {
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
		return ErrInvalidProposalType(DefaultCodespace, ProposalTypeToString(msg.ProposalType))
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
	return fmt.Sprintf("MsgSubmitProposal{%v, %v, %v, %v}", msg.Title, msg.Description, ProposalTypeToString(msg.ProposalType), msg.InitialDeposit)
}

// Implements Msg.
func (msg MsgSubmitProposal) Get(key interface{}) (value interface{}) {
	return nil
}

// Implements Msg.
func (msg MsgSubmitProposal) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(struct {
		Title          string    `json:"title"`
		Description    string    `json:"description"`
		ProposalType   string    `json:"proposal_type"`
		Proposer       string    `json:"proposer"`
		InitialDeposit sdk.Coins `json:"deposit"`
	}{
		Title:          msg.Title,
		Description:    msg.Description,
		ProposalType:   ProposalTypeToString(msg.ProposalType),
		Proposer:       sdk.MustBech32ifyVal(msg.Proposer),
		InitialDeposit: msg.InitialDeposit,
	})
	if err != nil {
		panic(err)
	}
	return b
}

// Implements Msg.
func (msg MsgSubmitProposal) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Proposer}
}

//-----------------------------------------------------------
// MsgDeposit
type MsgDeposit struct {
	ProposalID int64       `json:"proposalID"` // ID of the proposal
	Depositer  sdk.Address `json:"depositer"`  // Address of the depositer
	Amount     sdk.Coins   `json:"amount"`     // Coins to add to the proposal's deposit
}

func NewMsgDeposit(depositer sdk.Address, proposalID int64, amount sdk.Coins) MsgDeposit {
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
	return fmt.Sprintf("MsgDeposit{%v=>%v: %v}", msg.Depositer, msg.ProposalID, msg.Amount)
}

// Implements Msg.
func (msg MsgDeposit) Get(key interface{}) (value interface{}) {
	return nil
}

// Implements Msg.
func (msg MsgDeposit) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(struct {
		ProposalID int64     `json:"proposalID"`
		Depositer  string    `json:"proposer"`
		Amount     sdk.Coins `json:"deposit"`
	}{
		ProposalID: msg.ProposalID,
		Depositer:  sdk.MustBech32ifyVal(msg.Depositer),
		Amount:     msg.Amount,
	})
	if err != nil {
		panic(err)
	}
	return b
}

// Implements Msg.
func (msg MsgDeposit) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Depositer}
}

//-----------------------------------------------------------
// MsgVote
type MsgVote struct {
	ProposalID int64       //  proposalID of the proposal
	Voter      sdk.Address //  address of the voter
	Option     VoteOption  //  option from OptionSet chosen by the voter
}

func NewMsgVote(voter sdk.Address, proposalID int64, option VoteOption) MsgVote {
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
		return ErrInvalidVote(DefaultCodespace, VoteOptionToString(msg.Option))
	}
	return nil
}

func (msg MsgVote) String() string {
	return fmt.Sprintf("MsgVote{%v - %v}", msg.ProposalID, msg.Option)
}

// Implements Msg.
func (msg MsgVote) Get(key interface{}) (value interface{}) {
	return nil
}

// Implements Msg.
func (msg MsgVote) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(struct {
		ProposalID int64  `json:"proposalID"`
		Voter      string `json:"voter"`
		Option     string `json:"option"`
	}{
		ProposalID: msg.ProposalID,
		Voter:      sdk.MustBech32ifyVal(msg.Voter),
		Option:     VoteOptionToString(msg.Option),
	})
	if err != nil {
		panic(err)
	}
	return b
}

// Implements Msg.
func (msg MsgVote) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Voter}
}
