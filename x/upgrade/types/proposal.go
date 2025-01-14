package types

import (
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
