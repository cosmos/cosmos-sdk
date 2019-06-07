package utils

import (
	"encoding/json"
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/params"
)

type (
	// ParamChangesJSON defines a slice of ParamChangeJSON objects which can be
	// converted to a slice of ParamChange objects.
	ParamChangesJSON []ParamChangeJSON

	// ParamChangeJSON defines a parameter change used in JSON input. This
	// allows values to be specified in raw JSON instead of being string encoded.
	ParamChangeJSON struct {
		Subspace string          `json:"subspace"`
		Key      string          `json:"key"`
		Subkey   string          `json:"subkey,omitempty"`
		Value    json.RawMessage `json:"value"`
	}

	// ParamChangeProposalJSON defines a ParameterChangeProposal with a deposit used
	// to parse parameter change proposals from a JSON file.
	ParamChangeProposalJSON struct {
		Title       string           `json:"title"`
		Description string           `json:"description"`
		Changes     ParamChangesJSON `json:"changes"`
		Deposit     sdk.Coins        `json:"deposit"`
	}

	// ParamChangeProposalReq defines a parameter change proposal request body.
	ParamChangeProposalReq struct {
		BaseReq rest.BaseReq `json:"base_req"`

		Title       string           `json:"title"`
		Description string           `json:"description"`
		Changes     ParamChangesJSON `json:"changes"`
		Proposer    sdk.AccAddress   `json:"proposer"`
		Deposit     sdk.Coins        `json:"deposit"`
	}
)

func NewParamChangeJSON(subspace, key, subkey string, value json.RawMessage) ParamChangeJSON {
	return ParamChangeJSON{subspace, key, subkey, value}
}

// ToParamChange converts a ParamChangeJSON object to ParamChange.
func (pcj ParamChangeJSON) ToParamChange() params.ParamChange {
	return params.NewParamChangeWithSubkey(pcj.Subspace, pcj.Key, pcj.Subkey, string(pcj.Value))
}

// ToParamChanges converts a slice of ParamChangeJSON objects to a slice of
// ParamChange.
func (pcj ParamChangesJSON) ToParamChanges() []params.ParamChange {
	res := make([]params.ParamChange, len(pcj))
	for i, pc := range pcj {
		res[i] = pc.ToParamChange()
	}
	return res
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
