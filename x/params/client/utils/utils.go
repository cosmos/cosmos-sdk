package utils

import (
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// ParamChangeProposalJSON defines a ParameterChangeProposal with a deposit used
// to parse  parameter change proposals from a JSON file.
type ParamChangeProposalJSON struct {
	params.ParameterChangeProposal

	Deposit sdk.Coins `json:"deposit"`
}

// ParamChangeProposalReq defines a parameter change proposal request body.
type ParamChangeProposalReq struct {
	BaseReq rest.BaseReq `json:"base_req"`

	params.ParameterChangeProposal

	Proposer sdk.AccAddress `json:"proposer"`
	Deposit  sdk.Coins      `json:"deposit"`
}

// ParseParamChangeProposalJSON reads and parses a ParamChangeProposalJSON from
// file.
func ParseParamChangeProposalJSON(cdc *codec.Codec, proposalFile string) (ParamChangeProposalJSON, error) {
	proposal := ParamChangeProposalJSON{}

	contents, err := ioutil.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err := cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	return proposal, nil
}
