package cli

import (
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	// CommunityPoolSpendProposalJSON defines a CommunityPoolSpendProposal with a deposit
	CommunityPoolSpendProposalJSON struct {
		Title       string         `json:"title" yaml:"title"`
		Description string         `json:"description" yaml:"description"`
		Recipient   sdk.AccAddress `json:"recipient" yaml:"recipient"`
		Amount      string         `json:"amount" yaml:"amount"`
		Deposit     string         `json:"deposit" yaml:"deposit"`
	}
)

// ParseCommunityPoolSpendProposalJSON reads and parses a CommunityPoolSpendProposalJSON from a file.
func ParseCommunityPoolSpendProposalJSON(cdc codec.JSONMarshaler, proposalFile string) (CommunityPoolSpendProposalJSON, error) {
	proposal := CommunityPoolSpendProposalJSON{}

	contents, err := ioutil.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err := cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	return proposal, nil
}
