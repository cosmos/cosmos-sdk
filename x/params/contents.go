package params

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/proposal"
)

const RouterKey = "params"

// Struct for single parameter change
type Change struct {
	Space  string `json:"space"`
	Key    []byte `json:"key"`
	Subkey []byte `json:"subkey"`
	Value  []byte `json:"value"`
}

// Constructs new Change
func NewChange(space string, key, subkey, value []byte) Change {
	return Change{space, key, subkey, value}
}

func (c Change) String() string {
	var subkey string
	if len(c.Subkey) != 0 {
		subkey = "(" + string(c.Subkey) + ")"
	}
	return fmt.Sprintf("{%s%s := %X} (%s)", string(c.Key), subkey, c.Value, c.Space)
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
type ChangeProposal struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Changes     []Change `json:"changes"`
}

// Constructs new ChangeProposal
func NewChangeProposal(title string, description string, changes []Change) ChangeProposal {
	return ChangeProposal{
		Title:       title,
		Description: description,
		Changes:     changes,
	}
}

var _ proposal.Content = ChangeProposal{}

// nolint - Implements proposal.Content
func (pc ChangeProposal) GetTitle() string       { return pc.Title }
func (pc ChangeProposal) GetDescription() string { return pc.Description }
func (pc ChangeProposal) ProposalRoute() string  { return RouterKey }
func (pc ChangeProposal) ProposalType() string   { return "ParameterChange" }
func (pc ChangeProposal) ValidateBasic() sdk.Error {
	err := proposal.ValidateAbstract(DefaultCodespace, pc)
	if err != nil {
		return err
	}
	return ValidateChanges(pc.Changes)
}
func (pc ChangeProposal) String() string {
	return fmt.Sprintf("ParameterChangeProposal{%s, %s, %s}", pc.Title, pc.Description, pc.Changes)
}
