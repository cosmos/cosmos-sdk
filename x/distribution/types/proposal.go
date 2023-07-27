package types

import (
	"strings"

	errorsmod "cosmossdk.io/errors"
)

// GetTitle returns the title of a community pool spend proposal.
func (csp *CommunityPoolSpendProposal) GetTitle() string { return csp.Title }

// GetDescription returns the description of a community pool spend proposal.
func (csp *CommunityPoolSpendProposal) GetDescription() string { return csp.Description }

// GetDescription returns the routing key of a community pool spend proposal.
func (csp *CommunityPoolSpendProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a community pool spend proposal.
func (csp *CommunityPoolSpendProposal) ProposalType() string { return "CommunityPoolSpend" }

// ValidateBasic runs basic stateless validity checks
func (csp *CommunityPoolSpendProposal) ValidateBasic() error {
	err := validateAbstract(csp)
	if err != nil {
		return err
	}
	if !csp.Amount.IsValid() {
		return ErrInvalidProposalAmount
	}
	if csp.Recipient == "" {
		return ErrEmptyProposalRecipient
	}

	return nil
}

// validateAbstract is semantically duplicated from x/gov/types/v1beta1/proposal.go to avoid a cyclic dependency
// between these two modules (x/distribution and x/gov).
// See: https://github.com/cosmos/cosmos-sdk/blob/4a6a1e3cb8de459891cb0495052589673d14ef51/x/gov/types/v1beta1/proposal.go#L196-L214
func validateAbstract(c *CommunityPoolSpendProposal) error {
	const (
		MaxTitleLength       = 140
		MaxDescriptionLength = 10000
	)
	title := c.GetTitle()
	if len(strings.TrimSpace(title)) == 0 {
		return errorsmod.Wrap(ErrInvalidProposalContent, "proposal title cannot be blank")
	}
	if len(title) > MaxTitleLength {
		return errorsmod.Wrapf(ErrInvalidProposalContent, "proposal title is longer than max length of %d", MaxTitleLength)
	}

	description := c.GetDescription()
	if len(description) == 0 {
		return errorsmod.Wrap(ErrInvalidProposalContent, "proposal description cannot be blank")
	}
	if len(description) > MaxDescriptionLength {
		return errorsmod.Wrapf(ErrInvalidProposalContent, "proposal description is longer than max length of %d", MaxDescriptionLength)
	}

	return nil
}
