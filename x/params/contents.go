package params

import (
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

// Proposal which contains multiple changes on proposals
type ProposalChange struct {
	proposal.Abstract `json:"proposal_abstract"`
	Changes           []Change `json:"changes"`
}

// Constructs new ProposalChange
func NewProposalChange(title string, description string, changes []Change) ProposalChange {
	return ProposalChange{
		Abstract: proposal.NewAbstract(title, description),
		Changes:  changes,
	}
}

// Constructs proposal.Proto
func ProposalChangeProto(changes []Change) proposal.Proto {
	return func(title, description string) proposal.Content {
		return NewProposalChange(title, description, changes)
	}
}

var _ proposal.Content = ProposalChange{}

// Implements proposal.Content
func (pc ProposalChange) ProposalRoute() string { return RouterKey }
func (pc ProposalChange) ProposalType() string  { return "ParameterChange" }
