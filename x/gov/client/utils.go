package client

// NormalizeVoteOption - normalize user specified vote option
func NormalizeVoteOption(option string) string {
	switch option {
	case "yes":
	case "Yes":
		return "Yes"
	case "abstain":
	case "Abstain":
		return "Abstain"
	case "no":
	case "No":
		return "No"
	case "no_with_veto":
	case "NoWithVeto":
		return "NoWithVeto"
	}
	return ""
}

//NormalizeProposalType - normalize user specified proposal type
func NormalizeProposalType(proposalType string) string {
	switch proposalType {
	case "text":
	case "Text":
		return "Text"
	case "parameter_change":
	case "ParameterChange":
		return "ParameterChange"
	case "software_upgrade":
	case "SoftwareUpgrade":
		return "SoftwareUpgrade"
	}
	return ""
}

//NormalizeProposalStatus - normalize user specified proposal status
func NormalizeProposalStatus(status string) string {
	switch status {
	case "deposit_period":
	case "DepositPeriod":
		return "DepositPeriod"
	case "voting_period":
	case "VotingPeriod":
		return "VotingPeriod"
	case "passed":
	case "Passed":
		return "Passed"
	case "rejected":
	case "Rejected":
		return "Rejected"
	}
	return ""
}

