package rest

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Concrete Swagger types used to generate REST documentation. Note, these types
// are not actually used but since all queries return a generic JSON raw message,
// they enabled typed documentation.
//
//nolint:deadcode,unused
type (
	createValidator struct {
		auth.StdTx
		Msgs []types.MsgCreateValidator `json:"msg"`
	}

	editValidator struct {
		auth.StdTx
		Msgs []types.MsgEditValidator `json:"msg"`
	}

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

	delegatorDelegations struct {
		Height int64                      `json:"height"`
		Result []types.DelegationResponse `json:"result"`
	}

	delegatorUnbondingDelegations struct {
		Height int64                       `json:"height"`
		Result []types.UnbondingDelegation `json:"result"`
	}

	// XXX: This type won't render correct documentation at the moment due to aliased
	// fields.
	//
	// Ref: https://github.com/swaggo/swag/issues/483
	delegatorTxs struct {
		Height int64                  `json:"height"`
		Result []*sdk.SearchTxsResult `json:"result" swaggertype:"array"`
	}

	delegatorValidators struct {
		Height int64             `json:"height"`
		Result []types.Validator `json:"result"`
	}

	delegatorValidator struct {
		Height int64           `json:"height"`
		Result types.Validator `json:"result"`
	}

	delegation struct {
		Height int64                    `json:"height"`
		Result types.DelegationResponse `json:"result"`
	}

	unbondingDelegation struct {
		Height int64                     `json:"height"`
		Result types.UnbondingDelegation `json:"result"`
	}

	redelegations struct {
		Height int64                       `json:"height"`
		Result types.RedelegationResponses `json:"result"`
	}

	validators struct {
		Height int64             `json:"height"`
		Result []types.Validator `json:"result"`
	}

	validator struct {
		Height int64           `json:"height"`
		Result types.Validator `json:"result"`
	}

	validatorDelegations struct {
		Height int64                     `json:"height"`
		Result types.DelegationResponses `json:"result"`
	}

	validatorUnbondingDelegations struct {
		Height int64                       `json:"height"`
		Result []types.UnbondingDelegation `json:"result"`
	}

	pool struct {
		Height int64      `json:"height"`
		Result types.Pool `json:"result"`
	}

	params struct {
		Height int64        `json:"height"`
		Result types.Params `json:"result"`
	}
)
