package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/gov"
)

const (
	ProposalTypeSoftwareUpgrade       string = "SoftwareUpgrade"
	ProposalTypeCancelSoftwareUpgrade string = "CancelSoftwareUpgrade"
)

// Software Upgrade Proposals
type SoftwareUpgradeProposal struct {
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description" yaml:"description"`
	Plan        Plan   `json:"plan" yaml:"plan"`
}

func NewSoftwareUpgradeProposal(title, description string, plan Plan) gov.Content {
	return SoftwareUpgradeProposal{title, description, plan}
}

// Implements Proposal Interface
var _ gov.Content = SoftwareUpgradeProposal{}

func init() {
	gov.RegisterProposalType(ProposalTypeSoftwareUpgrade)
	gov.RegisterProposalTypeCodec(SoftwareUpgradeProposal{}, "cosmos-sdk/SoftwareUpgradeProposal")
	gov.RegisterProposalType(ProposalTypeCancelSoftwareUpgrade)
	gov.RegisterProposalTypeCodec(CancelSoftwareUpgradeProposal{}, "cosmos-sdk/CancelSoftwareUpgradeProposal")
}

// nolint
func (sup SoftwareUpgradeProposal) GetTitle() string       { return sup.Title }
func (sup SoftwareUpgradeProposal) GetDescription() string { return sup.Description }
func (sup SoftwareUpgradeProposal) ProposalRoute() string  { return RouterKey }
func (sup SoftwareUpgradeProposal) ProposalType() string   { return ProposalTypeSoftwareUpgrade }
func (sup SoftwareUpgradeProposal) ValidateBasic() error {
	if err := sup.Plan.ValidateBasic(); err != nil {
		return err
	}
	return gov.ValidateAbstract(sup)
}

func (sup SoftwareUpgradeProposal) String() string {
	return fmt.Sprintf(`Software Upgrade Proposal:
  Title:       %s
  Description: %s
`, sup.Title, sup.Description)
}

// Cancel Software Upgrade Proposals
type CancelSoftwareUpgradeProposal struct {
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description" yaml:"description"`
}

func NewCancelSoftwareUpgradeProposal(title, description string) gov.Content {
	return CancelSoftwareUpgradeProposal{title, description}
}

// Implements Proposal Interface
var _ gov.Content = CancelSoftwareUpgradeProposal{}

// nolint
func (sup CancelSoftwareUpgradeProposal) GetTitle() string       { return sup.Title }
func (sup CancelSoftwareUpgradeProposal) GetDescription() string { return sup.Description }
func (sup CancelSoftwareUpgradeProposal) ProposalRoute() string  { return RouterKey }
func (sup CancelSoftwareUpgradeProposal) ProposalType() string {
	return ProposalTypeCancelSoftwareUpgrade
}
func (sup CancelSoftwareUpgradeProposal) ValidateBasic() error {
	return gov.ValidateAbstract(sup)
}

func (sup CancelSoftwareUpgradeProposal) String() string {
	return fmt.Sprintf(`Cancel Software Upgrade Proposal:
  Title:       %s
  Description: %s
`, sup.Title, sup.Description)
}
