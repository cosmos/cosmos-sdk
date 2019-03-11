package params

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Change struct {
	Key    []byte      `json:"key"`
	Subkey []byte      `json:"subkey"`
	Value  interface{} `json:"value"`
}

type ProposalChange struct {
	sdk.ProposalAbstract `json:"proposal_abstract"`
	Space                string   `json:"space"`
	Changes              []Change `json:"changes"`
}

func NewProposalChange(title string, description string, space string, changes []Change) ProposalChange {
	return ProposalChange{
		ProposalAbstract: sdk.NewProposalAbstract(title, description),
		Space:            space,
		Changes:          changes,
	}
}

var _ sdk.ProposalContent = ProposalChange{}

func (pc ProposalChange) ProposalRoute() string { return RouteKey }
func (pc ProposalChange) ProposalType() string  { return "ParameterChange" }
