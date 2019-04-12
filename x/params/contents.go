package params

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/proposal"
)

const RouterKey = "params"

const ProposalTypeChange = "ParameterChange"

// Struct for single parameter change
type Change struct {
	Space  string `json:"space"`
	Key    string `json:"key"`
	Subkey []byte `json:"subkey"`
	Value  []byte `json:"value"`
}

// Constructs new Change
func NewChange(space, key string, subkey, value []byte) Change {
	return Change{space, key, subkey, value}
}

func (c Change) String() string {
	var subkey string
	if len(c.Subkey) != 0 {
		subkey = "(" + string(c.Subkey) + ")"
	}
	return fmt.Sprintf("{%s%s := %X} (%s)", c.Key, subkey, c.Value, c.Space)
}

// ValidateChange checks whether the input data is empty or not
func ValidateChanges(changes []Change) sdk.Error {
	if len(changes) == 0 {
		return ErrEmptyChanges(DefaultCodespace)
	}

	for _, c := range changes {
		if len(c.Space) == 0 {
			return ErrEmptySpace(DefaultCodespace)
		}
		if len(c.Key) == 0 {
			return ErrEmptyKey(DefaultCodespace)
		}
		if len(c.Value) == 0 {
			return ErrEmptyValue(DefaultCodespace)
		}
	}
	return nil
}

// Proposal which contains multiple changes on proposals
type ParameterChangeProposal struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Changes     []Change `json:"changes"`
}

// Constructs new ParameterChangeProposal
func NewParameterChangeProposal(title string, description string, changes []Change) ParameterChangeProposal {
	return ParameterChangeProposal{
		Title:       title,
		Description: description,
		Changes:     changes,
	}
}

var _ proposal.Content = ParameterChangeProposal{}

// nolint - Implements proposal.Content
func (pc ParameterChangeProposal) GetTitle() string       { return pc.Title }
func (pc ParameterChangeProposal) GetDescription() string { return pc.Description }
func (pc ParameterChangeProposal) ProposalRoute() string  { return RouterKey }
func (pc ParameterChangeProposal) ProposalType() string   { return ProposalTypeChange }
func (pc ParameterChangeProposal) ValidateBasic() sdk.Error {
	err := proposal.ValidateAbstract(DefaultCodespace, pc)
	if err != nil {
		return err
	}
	return ValidateChanges(pc.Changes)
}
func (pc ParameterChangeProposal) String() string {
	return fmt.Sprintf("ParameterParameterChangeProposal{%s, %s, %s}", pc.Title, pc.Description, pc.Changes)
}
