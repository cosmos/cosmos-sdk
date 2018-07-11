package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Type that represents VoteOption as a byte
type VoteOption = byte

//nolint
const (
	OptionEmpty      VoteOption = 0x00
	OptionYes        VoteOption = 0x01
	OptionAbstain    VoteOption = 0x02
	OptionNo         VoteOption = 0x03
	OptionNoWithVeto VoteOption = 0x04
)

// Vote
type Vote struct {
	Voter      sdk.AccAddress `json:"voter"`       //  address of the voter
	ProposalID int64          `json:"proposal_id"` //  proposalID of the proposal
	Option     VoteOption     `json:"option"`      //  option from OptionSet chosen by the voter
}

// Deposit
type Deposit struct {
	Depositer  sdk.AccAddress `json:"depositer"`   //  Address of the depositer
	ProposalID int64          `json:"proposal_id"` //  proposalID of the proposal
	Amount     sdk.Coins      `json:"amount"`      //  Deposit amount
}

// ProposalTypeToString for pretty prints of ProposalType
func VoteOptionToString(option VoteOption) string {
	switch option {
	case OptionYes:
		return "Yes"
	case OptionAbstain:
		return "Abstain"
	case OptionNo:
		return "No"
	case OptionNoWithVeto:
		return "NoWithVeto"
	default:
		return ""
	}
}

func validVoteOption(option VoteOption) bool {
	if option == OptionYes ||
		option == OptionAbstain ||
		option == OptionNo ||
		option == OptionNoWithVeto {
		return true
	}
	return false
}

// String to proposalType byte.  Returns ff if invalid.
func StringToVoteOption(str string) (VoteOption, sdk.Error) {
	switch str {
	case "Yes":
		return OptionYes, nil
	case "Abstain":
		return OptionAbstain, nil
	case "No":
		return OptionNo, nil
	case "NoWithVeto":
		return OptionNoWithVeto, nil
	default:
		return VoteOption(0xff), ErrInvalidVote(DefaultCodespace, str)
	}
}

//-----------------------------------------------------------
// REST

// Rest Deposits
type DepositRest struct {
	Depositer  sdk.AccAddress `json:"depositer"`   //  address of the depositer
	ProposalID int64          `json:"proposal_id"` //  proposalID of the proposal
	Amount     sdk.Coins      `json:"option"`
}

// Turn any Deposit to a DepositRest
func DepositToRest(deposit Deposit) DepositRest {
	return DepositRest{
		Depositer:  deposit.Depositer,
		ProposalID: deposit.ProposalID,
		Amount:     deposit.Amount,
	}
}

// Rest Votes
type VoteRest struct {
	Voter      sdk.AccAddress `json:"voter"`       //  address of the voter
	ProposalID int64          `json:"proposal_id"` //  proposalID of the proposal
	Option     string         `json:"option"`
}

// Turn any Vote to a VoteRest
func VoteToRest(vote Vote) VoteRest {
	return VoteRest{
		Voter:      vote.Voter,
		ProposalID: vote.ProposalID,
		Option:     VoteOptionToString(vote.Option),
	}
}
