package utils

import "github.com/cosmos/cosmos-sdk/x/gov"

// NormalizeVoteOption - normalize user specified vote option
func NormalizeVoteOption(option string) string {
	switch option {
	case "Yes", "yes":
		return gov.OptionYes.String()

	case "Abstain", "abstain":
		return gov.OptionAbstain.String()

	case "No", "no":
		return gov.OptionNo.String()

	case "NoWithVeto", "no_with_veto":
		return gov.OptionNoWithVeto.String()

	default:
		return ""
	}
}

//NormalizeProposalType - normalize user specified proposal type
func NormalizeProposalType(proposalType string) string {
	switch proposalType {
	case "Text", "text":
		return gov.ProposalTypeText

	case "SoftwareUpgrade", "software_upgrade":
		return gov.ProposalTypeSoftwareUpgrade

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
	}
	return ""
}
