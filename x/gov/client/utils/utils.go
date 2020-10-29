package utils

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// NormalizeVoteOption - normalize user specified vote option
func NormalizeVoteOption(option string) string {
	switch option {
	case "Yes", "yes":
		return types.OptionYes.String()

	case "Abstain", "abstain":
		return types.OptionAbstain.String()

	case "No", "no":
		return types.OptionNo.String()

	case "NoWithVeto", "no_with_veto":
		return types.OptionNoWithVeto.String()

	default:
		return option
	}
}

// NormalizeSubVotes - normalize subvotes
func NormalizeSubVotes(subvotes string) string {
	newSubVotes := []string{}
	for _, subvote := range strings.Split(subvotes, ",") {
		fields := strings.Split(subvote, "=")
		fields[0] = NormalizeVoteOption(fields[0])
		if len(fields) < 2 {
			fields = append(fields, "1")
		}
		newSubVotes = append(newSubVotes, strings.Join(fields, "="))
	}
	return strings.Join(newSubVotes, ",")
}

//NormalizeProposalType - normalize user specified proposal type
func NormalizeProposalType(proposalType string) string {
	switch proposalType {
	case "Text", "text":
		return types.ProposalTypeText

	default:
		return ""
	}
}

//NormalizeProposalStatus - normalize user specified proposal status
func NormalizeProposalStatus(status string) string {
	switch status {
	case "DepositPeriod", "deposit_period":
		return "DepositPeriod"
	case "VotingPeriod", "voting_period":
		return "VotingPeriod"
	case "Passed", "passed":
		return "Passed"
	case "Rejected", "rejected":
		return "Rejected"
	default:
		return status
	}
}
