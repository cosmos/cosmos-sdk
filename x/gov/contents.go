package gov

import (
	"github.com/cosmos/cosmos-sdk/x/gov/proposal"
)

// Text Proposals
type TextProposal struct {
	proposal.Abstract `json:"abstract"`
}

func NewTextProposal(title, description string) proposal.Content {
	return TextProposal{proposal.NewAbstract(title, description)}
}

// Implements Proposal Interface
var _ proposal.Content = TextProposal{}

// nolint
func (tp TextProposal) ProposalRoute() string { return RouterKey }
func (tp TextProposal) ProposalType() string  { return ProposalTypeText }

// Software Upgrade Proposals
type SoftwareUpgradeProposal struct {
	proposal.Abstract `json:"abstract"`
}

func NewSoftwareUpgradeProposal(title, description string) proposal.Content {
	return SoftwareUpgradeProposal{proposal.NewAbstract(title, description)}
}

// Implements Proposal Interface
var _ proposal.Content = SoftwareUpgradeProposal{}

// nolint

func (sup SoftwareUpgradeProposal) ProposalRoute() string { return RouterKey }
func (sup SoftwareUpgradeProposal) ProposalType() string  { return ProposalTypeSoftwareUpgrade }
