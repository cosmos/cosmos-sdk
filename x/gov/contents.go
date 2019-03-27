package gov

import (
	"github.com/cosmos/cosmos-sdk/x/gov/proposal"
)

// Text Proposals
type TextProposal struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func NewTextProposal(title, description string) proposal.Content {
	return TextProposal{title, description}
}

// Implements Proposal Interface
var _ proposal.Content = TextProposal{}

// nolint
func (tp TextProposal) GetTitle() string       { return tp.Title }
func (tp TextProposal) GetDescription() string { return tp.Description }
func (tp TextProposal) ProposalRoute() string  { return RouterKey }
func (tp TextProposal) ProposalType() string   { return ProposalTypeText }

// Software Upgrade Proposals
// TODO: we have to add fields for SUP specific arguments e.g. commit hash
type SoftwareUpgradeProposal struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func NewSoftwareUpgradeProposal(title, description string) proposal.Content {
	return SoftwareUpgradeProposal{title, description}
}

// Implements Proposal Interface
var _ proposal.Content = SoftwareUpgradeProposal{}

// nolint
func (sup SoftwareUpgradeProposal) GetTitle() string       { return sup.Title }
func (sup SoftwareUpgradeProposal) GetDescription() string { return sup.Description }
func (sup SoftwareUpgradeProposal) ProposalRoute() string  { return RouterKey }
func (sup SoftwareUpgradeProposal) ProposalType() string   { return ProposalTypeSoftwareUpgrade }
