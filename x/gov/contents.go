package gov

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/proposal"
)

//nolint
const (
	ProposalTypeText            string = "Text"
	ProposalTypeSoftwareUpgrade string = "SoftwareUpgrade"
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
func (tp TextProposal) ValidateBasic() sdk.Error {
	return proposal.ValidateAbstract(DefaultCodespace, tp)
}
func (tp TextProposal) String() string {
	return fmt.Sprintf("TextProposal{%s, %s}", tp.Title, tp.Description)
}

// Software Upgrade Proposals
// TODO: we have to add fields for SUP specific arguments e.g. commit hash, upgrade date, etc.
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
func (sup SoftwareUpgradeProposal) ValidateBasic() sdk.Error {
	return proposal.ValidateAbstract(DefaultCodespace, sup)
}
func (sup SoftwareUpgradeProposal) String() string {
	return fmt.Sprintf("SoftwareUpgradeProposal{%s, %s}", sup.Title, sup.Description)
}
