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

var _ sdk.ProposalContent = ProposalChange{}

func (pc ProposalChange) ProposalRoute() string { return RouterKey }
func (pc ProposalChange) ProposalType() string  { return ProposalTypeParameterChange }

const RouterKey = "params"
const ProposalTypeParameterChange = "ParameterChange"
