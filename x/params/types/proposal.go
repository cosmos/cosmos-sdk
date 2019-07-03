package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	// ProposalTypeChange defines the type for a ParameterChangeProposal
	ProposalTypeChange = "ParameterChange"
)

// Assert ParameterChangeProposal implements govtypes.Content at compile-time
var _ govtypes.Content = ParameterChangeProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeChange)
	govtypes.RegisterProposalTypeCodec(ParameterChangeProposal{}, "cosmos-sdk/ParameterChangeProposal")
}

// ParameterChangeProposal defines a proposal which contains multiple parameter
// changes.
type ParameterChangeProposal struct {
	Title       string        `json:"title" yaml:"title"`
	Description string        `json:"description" yaml:"description"`
	Changes     []ParamChange `json:"changes" yaml:"changes"`
}

func NewParameterChangeProposal(title, description string, changes []ParamChange) ParameterChangeProposal {
	return ParameterChangeProposal{title, description, changes}
}

// GetTitle returns the title of a parameter change proposal.
func (pcp ParameterChangeProposal) GetTitle() string { return pcp.Title }

// GetDescription returns the description of a parameter change proposal.
func (pcp ParameterChangeProposal) GetDescription() string { return pcp.Description }

// GetDescription returns the routing key of a parameter change proposal.
func (pcp ParameterChangeProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a parameter change proposal.
func (pcp ParameterChangeProposal) ProposalType() string { return ProposalTypeChange }

// ValidateBasic validates the parameter change proposal
func (pcp ParameterChangeProposal) ValidateBasic() sdk.Error {
	err := govtypes.ValidateAbstract(DefaultCodespace, pcp)
	if err != nil {
		return err
	}

	return ValidateChanges(pcp.Changes)
}

// String implements the Stringer interface.
func (pcp ParameterChangeProposal) String() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf(`Parameter Change Proposal:
  Title:       %s
  Description: %s
  Changes:
`, pcp.Title, pcp.Description))

	for _, pc := range pcp.Changes {
		b.WriteString(fmt.Sprintf(`    Param Change:
      Subspace: %s
      Key:      %s
      Subkey:   %X
      Value:    %X
`, pc.Subspace, pc.Key, pc.Subkey, pc.Value))
	}

	return b.String()
}

// ParamChange defines a parameter change.
type ParamChange struct {
	Subspace string `json:"subspace" yaml:"subspace"`
	Key      string `json:"key" yaml:"key"`
	Subkey   string `json:"subkey,omitempty" yaml:"subkey,omitempty"`
	Value    string `json:"value" yaml:"value"`
}

func NewParamChange(subspace, key, value string) ParamChange {
	return ParamChange{subspace, key, "", value}
}

func NewParamChangeWithSubkey(subspace, key, subkey, value string) ParamChange {
	return ParamChange{subspace, key, subkey, value}
}

// String implements the Stringer interface.
func (pc ParamChange) String() string {
	return fmt.Sprintf(`Param Change:
  Subspace: %s
  Key:      %s
  Subkey:   %X
  Value:    %X
`, pc.Subspace, pc.Key, pc.Subkey, pc.Value)
}

// ValidateChange performs basic validation checks over a set of ParamChange. It
// returns an error if any ParamChange is invalid.
func ValidateChanges(changes []ParamChange) sdk.Error {
	if len(changes) == 0 {
		return ErrEmptyChanges(DefaultCodespace)
	}

	for _, pc := range changes {
		if len(pc.Subspace) == 0 {
			return ErrEmptySubspace(DefaultCodespace)
		}
		if len(pc.Key) == 0 {
			return ErrEmptyKey(DefaultCodespace)
		}
		if len(pc.Value) == 0 {
			return ErrEmptyValue(DefaultCodespace)
		}
	}

	return nil
}
