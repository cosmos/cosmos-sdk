package utils

import (
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

type (
	// CommunityPoolSpendProposalReq defines a community pool spend proposal request body.
	CommunityPoolSpendProposalReq struct {
		BaseReq rest.BaseReq `json:"base_req"`

		Title       string         `json:"title"`
		Description string         `json:"description"`
		Recipient   sdk.AccAddress `json:"recipient"`
		Amount      sdk.Coins      `json:"amount"`
		Proposer    sdk.AccAddress `json:"proposer"`
		Deposit     sdk.Coins      `json:"deposit"`
	}

	// CommunityPoolSpendProposalJSON defines a CommunityPoolSpendProposal with a deposit
	CommunityPoolSpendProposalJSON struct {
		Title       string         `json:"title"`
		Description string         `json:"description"`
		Recipient   sdk.AccAddress `json:"recipient"`
		Amount      sdk.Coins      `json:"amount"`
		Deposit     sdk.Coins      `json:"deposit"`
	}
)

// ParseCommunityPoolSpendProposalJSON reads and parses a CommunityPoolSpendProposalJSON from a file.
func ParseCommunityPoolSpendProposalJSON(cdc *codec.Codec, proposalFile string) (CommunityPoolSpendProposalJSON, error) {
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
