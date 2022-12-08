package types

import (
	"fmt"

	gov "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeSoftwareUpgrade       string = "SoftwareUpgrade"
	ProposalTypeCancelSoftwareUpgrade string = "CancelSoftwareUpgrade"
)

// NewSoftwareUpgradeProposal creates a new SoftwareUpgradeProposal instance
func NewSoftwareUpgradeProposal(title, description string, plan Plan) gov.Content {
	return &SoftwareUpgradeProposal{title, description, plan}
}

// Implements Proposal Interface
var _ gov.Content = &SoftwareUpgradeProposal{}

func init() {
	gov.RegisterProposalType(ProposalTypeSoftwareUpgrade)
	gov.RegisterProposalType(ProposalTypeCancelSoftwareUpgrade)
}

// GetTitle gets the proposal's title
func (sup *SoftwareUpgradeProposal) GetTitle() string { return sup.Title }

// GetDescription gets the proposal's description
func (sup *SoftwareUpgradeProposal) GetDescription() string { return sup.Description }

// ProposalRoute gets the proposal's router key
func (sup *SoftwareUpgradeProposal) ProposalRoute() string { return RouterKey }

// ProposalType is "SoftwareUpgrade"
func (sup *SoftwareUpgradeProposal) ProposalType() string { return ProposalTypeSoftwareUpgrade }

// ValidateBasic validates the proposal
func (sup *SoftwareUpgradeProposal) ValidateBasic() error {
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

// NewCancelSoftwareUpgradeProposal creates a new CancelSoftwareUpgradeProposal instance
func NewCancelSoftwareUpgradeProposal(title, description string) gov.Content {
	return &CancelSoftwareUpgradeProposal{title, description}
}

// Implements Proposal Interface
var _ gov.Content = &CancelSoftwareUpgradeProposal{}

// GetTitle gets the proposal's title
func (csup *CancelSoftwareUpgradeProposal) GetTitle() string { return csup.Title }

// GetDescription gets the proposal's description
func (csup *CancelSoftwareUpgradeProposal) GetDescription() string { return csup.Description }

// ProposalRoute gets the proposal's router key
func (csup *CancelSoftwareUpgradeProposal) ProposalRoute() string { return RouterKey }

// ProposalType is "CancelSoftwareUpgrade"
func (csup *CancelSoftwareUpgradeProposal) ProposalType() string {
	return ProposalTypeCancelSoftwareUpgrade
}

// ValidateBasic validates the proposal
func (csup *CancelSoftwareUpgradeProposal) ValidateBasic() error {
	return gov.ValidateAbstract(csup)
}

func (csup CancelSoftwareUpgradeProposal) String() string {
	return fmt.Sprintf(`Cancel Software Upgrade Proposal:
  Title:       %s
  Description: %s
`, csup.Title, csup.Description)
}
