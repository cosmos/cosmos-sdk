package client

// NormalizeVoteOption - normalize user specified vote option
func NormalizeVoteOption(option string) string {
	switch option {
	case "yes":
		return "Yes"
	case "abstain":
		return "Abstain"
	case "no":
		return "No"
	case "no_with_veto":
		return "NoWithVeto"
	default:
		return ""
	}
}

//NormalizeProposalType - normalize user specified proposal type
func NormalizeProposalType(proposalType string) string {
	switch proposalType {
	case "text":
		return "Text"
	case "parameter_change":
		return "ParameterChange"
	case "software_upgrade":
		return "SoftwareUpgrade"
	default:
		return ""
	}
}

//NormalizeProposalStatus - normalize user specified proposal status
func NormalizeProposalStatus(status string) string {
	switch status {
	case "deposit_period":
		return "DepositPeriod"
	case "voting_period":
		return "VotingPeriod"
	case "passed":
		return "Passed"
	case "rejected":
		return "Rejected"
	default:
		return ""
	}
}

