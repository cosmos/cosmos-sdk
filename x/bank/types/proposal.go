package types

import (
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeUpdateDenomMetadata string = "UpdateDenomMetadata"
)

var (
	_ govtypesv1beta1.Content = &UpdateDenomMetadataProposal{}
)

func init() {
	govtypesv1beta1.RegisterProposalType(ProposalTypeUpdateDenomMetadata)
}

func NewUpdateDenomMetadataProposal(
	title string,
	description string,
	metadata Metadata,
) *UpdateDenomMetadataProposal {
	return &UpdateDenomMetadataProposal{
		Title:       title,
		Description: description,
		Metadata:    metadata,
	}
}

func (p UpdateDenomMetadataProposal) ProposalRoute() string { return RouterKey }

func (p UpdateDenomMetadataProposal) ProposalType() string {
	return ProposalTypeUpdateDenomMetadata
}

func (p UpdateDenomMetadataProposal) ValidateBasic() error {
	return govtypesv1beta1.ValidateAbstract(&p)
}
