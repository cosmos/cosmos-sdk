package types

import (
<<<<<<< HEAD
	gov "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeSoftwareUpgrade       string = "SoftwareUpgrade"
	ProposalTypeCancelSoftwareUpgrade string = "CancelSoftwareUpgrade"
)

// NewSoftwareUpgradeProposal creates a new SoftwareUpgradeProposal instance.
// Deprecated: this proposal is considered legacy and is deprecated in favor of
// Msg-based gov proposals. See MsgSoftwareUpgrade.
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

// NewCancelSoftwareUpgradeProposal creates a new CancelSoftwareUpgradeProposal
// instance. Deprecated: this proposal is considered legacy and is deprecated in
// favor of Msg-based gov proposals. See MsgCancelUpgrade.
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
=======
	"cosmossdk.io/x/gov/types"
	"cosmossdk.io/x/gov/types/v1beta1"
)

// GetTitle returns the proposal title
func (sp *SoftwareUpgradeProposal) GetTitle() string { return sp.Title }

// GetDescription returns the proposal description
func (sp *SoftwareUpgradeProposal) GetDescription() string { return sp.Description }

// ProposalRoute returns the proposal router key
func (sp *SoftwareUpgradeProposal) ProposalRoute() string { return types.RouterKey }

// ProposalType is "Text"
func (sp *SoftwareUpgradeProposal) ProposalType() string { return v1beta1.ProposalTypeText }

// ValidateBasic validates the content's title and description of the proposal
func (sp *SoftwareUpgradeProposal) ValidateBasic() error { return v1beta1.ValidateAbstract(sp) }
>>>>>>> 5581225a9 (fix(x/upgrade): register missing implementation for SoftwareUpgradeProposal (#23179))
