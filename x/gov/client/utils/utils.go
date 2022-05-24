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

// NormalizeWeightedVoteOptions - normalize vote options param string
func NormalizeWeightedVoteOptions(options string) string {
	newOptions := []string{}
	for _, option := range strings.Split(options, ",") {
		fields := strings.Split(option, "=")
		fields[0] = NormalizeVoteOption(fields[0])
		if len(fields) < 2 {
			fields = append(fields, "1")
		}
		newOptions = append(newOptions, strings.Join(fields, "="))
	}
	return strings.Join(newOptions, ",")
}

// NormalizeProposalType - normalize user specified proposal type.
func NormalizeProposalType(proposalType string) string {
	switch proposalType {
	case "Text", "text":
		return types.ProposalTypeText

	default:
		return ""
	}
}

// NormalizeProposalStatus - normalize user specified proposal status.
func NormalizeProposalStatus(status string) string {
	switch status {
	case "DepositPeriod", "deposit_period":
		return types.StatusDepositPeriod.String()
	case "VotingPeriod", "voting_period":
		return types.StatusVotingPeriod.String()
	case "Passed", "passed":
		return types.StatusPassed.String()
	case "Rejected", "rejected":
		return types.StatusRejected.String()
	default:
		return status
	}
}
