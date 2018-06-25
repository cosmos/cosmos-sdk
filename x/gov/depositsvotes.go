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
	Voter      sdk.Address `json:"voter"`       //  address of the voter
	ProposalID int64       `json:"proposal_id"` //  proposalID of the proposal
	Option     VoteOption  `json:"option"`      //  option from OptionSet chosen by the voter
}

// Deposit
type Deposit struct {
	Depositer sdk.Address `json:"depositer"` //  Address of the depositer
	Amount    sdk.Coins   `json:"amount"`    //  Deposit amount
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
// Rest Votes
type VoteRest struct {
	Voter      string `json:"voter"`       //  address of the voter
	ProposalID int64  `json:"proposal_id"` //  proposalID of the proposal
	Option     string `json:"option"`
}

// Turn any Vote to a ProposalRest
func VoteToRest(vote Vote) VoteRest {
	bechAddr, _ := sdk.Bech32ifyAcc(vote.Voter)
	return VoteRest{
		Voter:      bechAddr,
		ProposalID: vote.ProposalID,
		Option:     VoteOptionToString(vote.Option),
	}
}
