package v036

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	v036gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v036"
)

const (
	// ModuleName defines the name of the module
	ModuleName = "params"

	// RouterKey defines the routing key for a ParameterChangeProposal
	RouterKey = "params"
)

const (
	// ProposalTypeChange defines the type for a ParameterChangeProposal
	ProposalTypeChange = "ParameterChange"
)

// Param module codespace constants
const (
	DefaultCodespace = "params"

	CodeUnknownSubspace  = 1
	CodeSettingParameter = 2
	CodeEmptyData        = 3
)

// Assert ParameterChangeProposal implements v036gov.Content at compile-time
var _ v036gov.Content = ParameterChangeProposal{}

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
func (pcp ParameterChangeProposal) ValidateBasic() error {
	err := v036gov.ValidateAbstract(pcp)
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
func ValidateChanges(changes []ParamChange) error {
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

// ErrUnknownSubspace returns an unknown subspace error.
func ErrUnknownSubspace(codespace string, space string) error {
	return fmt.Errorf("unknown subspace %s", space)
}

// ErrSettingParameter returns an error for failing to set a parameter.
func ErrSettingParameter(codespace string, key, subkey, value, msg string) error {
	return fmt.Errorf("error setting parameter %s on %s (%s): %s", value, key, subkey, msg)
}

// ErrEmptyChanges returns an error for empty parameter changes.
func ErrEmptyChanges(codespace string) error {
	return fmt.Errorf("submitted parameter changes are empty")
}

// ErrEmptySubspace returns an error for an empty subspace.
func ErrEmptySubspace(codespace string) error {
	return fmt.Errorf("parameter subspace is empty")
}

// ErrEmptyKey returns an error for when an empty key is given.
func ErrEmptyKey(codespace string) error {
	return fmt.Errorf("parameter key is empty")
}

// ErrEmptyValue returns an error for when an empty key is given.
func ErrEmptyValue(codespace string) error {
	return fmt.Errorf("parameter value is empty")
}

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(ParameterChangeProposal{}, "cosmos-sdk/ParameterChangeProposal", nil)
}
