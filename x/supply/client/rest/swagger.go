package rest

import (
	"github.com/cosmos/cosmos-sdk/types"
)

// Concrete Swagger types used to generate REST documentation. Note, these types
// are not actually used but since all queries return a generic JSON raw message,
// they enabled typed documentation.
//
//nolint:deadcode,unused
type (
	totalSupply struct {
		Height int64        `json:"height"`
		Result []types.Coin `json:"result"`
	}

	totalDenomSupply struct {
		Height int64  `json:"height"`
		Result string `json:"result"`
	}
)
