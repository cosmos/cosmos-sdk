package rest

import "github.com/cosmos/cosmos-sdk/x/auth/exported"

// Concrete Swagger types used to generate REST documentation. Note, these types
// are not actually used but since all queries return a generic JSON raw message,
// they enabled typed documentation.
//
//nolint:deadcode,unused
type (
	accountQuery struct {
		Height int64            `json:"height"`
		Result exported.Account `json:"result"`
	}
)
