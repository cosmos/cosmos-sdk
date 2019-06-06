package rest

import (
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
)
