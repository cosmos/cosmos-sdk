package cli

import (
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// ParseCommunityPoolSpendProposalJSON reads and parses a CommunityPoolSpendProposalJSON from a file.
func ParseCommunityPoolSpendProposalJSON(cdc codec.JSONMarshaler, proposalFile string) (types.CommunityPoolSpendProposalJSON, error) {
	proposal := types.CommunityPoolSpendProposalJSON{}

	contents, err := ioutil.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err = cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	return proposal, nil
}
