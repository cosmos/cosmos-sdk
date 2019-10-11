package rest

import (
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Concrete Swagger types used to generate REST documentation. Note, these types
// are not actually used but since all queries return a generic JSON raw message,
// they enabled typed documentation.
//
//nolint:deadcode,unused
type (
	postDelegation struct {
		auth.StdTx
		Msgs []types.MsgDelegate `json:"msg"`
	}

	postRedelegation struct {
		auth.StdTx
		Msgs []types.MsgBeginRedelegate `json:"msg"`
	}

	postUndelegate struct {
		auth.StdTx
		Msgs []types.MsgUndelegate `json:"msg"`
	}
)
