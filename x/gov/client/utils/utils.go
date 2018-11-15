package utils

// NormalizeVoteOption - normalize user specified vote option
func NormalizeVoteOption(option string) string {
	switch option {
	case "Yes", "yes":
		return "Yes"
	case "Abstain", "abstain":
		return "Abstain"
	case "No", "no":
		return "No"
	case "NoWithVeto", "no_with_veto":
		return "NoWithVeto"
	}
	return ""
}

//NormalizeProposalType - normalize user specified proposal type
func NormalizeProposalType(proposalType string) string {
	switch proposalType {
	case "Text", "text":
		return "Text"
	case "ParameterChange", "parameter_change":
		return "ParameterChange"
	case "SoftwareUpgrade", "software_upgrade":
		return "SoftwareUpgrade"
	}
	return ""
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
