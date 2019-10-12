package rest

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank/internal/types"
)

// Concrete Swagger types used to generate REST documentation. Note, these types
// are not actually used but since all queries return a generic JSON raw message,
// they enabled typed documentation.
//
//nolint:deadcode,unused
type (
	queryBalance struct {
		Height int64     `json:"height"`
		Result sdk.Coins `json:"result"`
	}

	sendResponse struct {
		auth.StdTx
		Msgs []types.MsgSend `json:"msg"`
	}
)
