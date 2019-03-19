package params

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/proposal"
)

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
type ProposalChange struct {
	proposal.Abstract `json:"abstract"`
	Changes           []Change `json:"changes"`
}

// Constructs new ProposalChange
func NewProposalChange(title string, description string, changes []Change) ProposalChange {
	return ProposalChange{
		Abstract: proposal.NewAbstract(title, description),
		Changes:  changes,
	}
}

var _ proposal.Content = ProposalChange{}

// nolint - Implements proposal.Content
func (pc ProposalChange) ProposalRoute() string { return RouterKey }
func (pc ProposalChange) ProposalType() string  { return "ParameterChange" }
