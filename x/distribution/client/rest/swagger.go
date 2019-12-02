package rest

import (
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/common"
)

// Concrete Swagger types used to generate REST documentation. Note, these types
// are not actually used but since all queries return a generic JSON raw message,
// they enabled typed documentation.

//nolint:deadcode,unused
type (
	coinsReturn struct {
		Height int64       `json:"height"`
		Result types.Coins `json:"result"`
	}
	// delegatorWithdrawalAddr helps generate documentation for delegatorWithdrawalAddrHandlerFn
	delegatorWithdrawalAddr struct {
		Height int64         `json:"height"`
		Result types.Address `json:"result"`
	}

	// validatorInfo helps generate documentation for validatorInfoHandlerFn
	validatorInfo struct {
		Height int64             `json:"height"`
		Result ValidatorDistInfo `json:"result"`
	}

	// params helps generate documentation for paramsHandlerFn
	distrParams struct {
		Height int64               `json:"height"`
		Result common.PrettyParams `json:"result"`
	}
)
