package params

import (
	"github.com/cosmos/cosmos-sdk/x/gov/proposal"
)

type Change struct {
	Key    []byte      `json:"key"`
	Subkey []byte      `json:"subkey"`
	Value  interface{} `json:"value"`
}

type ProposalChange struct {
	proposal.Abstract `json:"proposal_abstract"`
	Space             string   `json:"space"`
	Changes           []Change `json:"changes"`
}

func NewProposalChange(title string, description string, space string, changes []Change) ProposalChange {
	return ProposalChange{
		Abstract: proposal.NewAbstract(title, description),
		Space:    space,
		Changes:  changes,
	}
}

func ProposalChangeProto(space string, changes []Change) proposal.Proto {
	return func(title, description string) proposal.Content {
		return NewProposalChange(title, description, space, changes)
	}
}

var _ proposal.Content = ProposalChange{}

func (pc ProposalChange) ProposalRoute() string { return RouteKey }
func (pc ProposalChange) ProposalType() string  { return "ParameterChange" }
