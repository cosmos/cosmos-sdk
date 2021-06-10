package v040

import (
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func (sup *SoftwareUpgradeProposal) ProposalRoute() string { return types.RouterKey }
func (sup *SoftwareUpgradeProposal) ProposalType() string  { return types.ProposalTypeSoftwareUpgrade }
func (sup *SoftwareUpgradeProposal) ValidateBasic() error {
	if err := sup.Plan.ValidateBasic(); err != nil {
		return err
	}
	return gov.ValidateAbstract(sup)
}

// Implements Proposal Interface
var _ gov.Content = &CancelSoftwareUpgradeProposal{}

func (csup *CancelSoftwareUpgradeProposal) ProposalRoute() string { return types.RouterKey }
func (csup *CancelSoftwareUpgradeProposal) ProposalType() string {
	return types.ProposalTypeCancelSoftwareUpgrade
}
func (csup *CancelSoftwareUpgradeProposal) ValidateBasic() error {
	return gov.ValidateAbstract(csup)
}
