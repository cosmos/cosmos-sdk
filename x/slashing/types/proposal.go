package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	// ProposalTypeSlashValidator defines the type for a SlashValidatorProposal
	ProposalTypeSlashValidator = "SlashValidator"
)

// String implements the Stringer interface.
func (svp SlashValidatorProposal) String() string {
	return fmt.Sprintf(
		`Slash Validator Proposal:
  Title:       %s
  Description: %s
  Validator Address: %s
  Slash By Percent: %s%%
`, svp.Title, svp.Description, svp.ValidatorAddress, svp.SlashFactor.Mul(sdk.NewDec(100)).String())
}

// GetTitle returns the title of a parameter change proposal.
func (svp SlashValidatorProposal) GetTitle() string { return svp.Title }

// GetDescription returns the description of a parameter change proposal.
func (svp SlashValidatorProposal) GetDescription() string { return svp.Description }

// ProposalRoute returns the routing key of a parameter change proposal.
func (svp SlashValidatorProposal) ProposalRoute() string { return ProposalTypeSlashValidator }

// ProposalType returns the type of a parameter change proposal.
func (svp SlashValidatorProposal) ProposalType() string { return ProposalTypeSlashValidator }

// ValidateBasic validates the parameter change proposal
func (svp SlashValidatorProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(svp)
	if err != nil {
		return err
	}
	// todo validate percentage is >0  <=100
	return nil
}
