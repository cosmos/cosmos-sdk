package rest

import (
	"bytes"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/delegations",
		postDelegationsHandlerFn(cliCtx),
	).Methods("POST")
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/unbonding_delegations",
		postUnbondingDelegationsHandlerFn(cliCtx),
	).Methods("POST")
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/redelegations",
		postRedelegationsHandlerFn(cliCtx),
	).Methods("POST")
}

type (
	// DelegateRequest defines the properties of a delegation request's body.
	DelegateRequest struct {
		BaseReq          rest.BaseReq   `json:"base_req" yaml:"base_req"`
		DelegatorAddress sdk.AccAddress `json:"delegator_address" yaml:"delegator_address"` // in bech32
		ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address"` // in bech32
		Amount           sdk.Coin       `json:"amount" yaml:"amount"`
	}

	// RedelegateRequest defines the properties of a redelegate request's body.
	RedelegateRequest struct {
		BaseReq             rest.BaseReq   `json:"base_req" yaml:"base_req"`
		DelegatorAddress    sdk.AccAddress `json:"delegator_address" yaml:"delegator_address"`         // in bech32
		ValidatorSrcAddress sdk.ValAddress `json:"validator_src_address" yaml:"validator_src_address"` // in bech32
		ValidatorDstAddress sdk.ValAddress `json:"validator_dst_address" yaml:"validator_dst_address"` // in bech32
		Amount              sdk.Coin       `json:"amount" yaml:"amount"`
	}

	// UndelegateRequest defines the properties of a undelegate request's body.
	UndelegateRequest struct {
		BaseReq          rest.BaseReq   `json:"base_req" yaml:"base_req"`
		DelegatorAddress sdk.AccAddress `json:"delegator_address" yaml:"delegator_address"` // in bech32
		ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address"` // in bech32
		Amount           sdk.Coin       `json:"amount" yaml:"amount"`
	}
)

// postDelegations is used to generate documentation for postDelegationsHandlerFn
type postDelegations struct {
	Msgs       []types.MsgDelegate `json:"msg" yaml:"msg"`
	Fee        auth.StdFee         `json:"fee" yaml:"fee"`
	Signatures []auth.StdSignature `json:"signatures" yaml:"signatures"`
	Memo       string              `json:"memo" yaml:"memo"`
}

// postDelegationsHandlerFn implements a delegation handler that is responsible
// for constructing a properly formatted delegation transaction for signing.
//
// @Summary Generate a delegation transaction
// @Description Generate a delegation transaction that is ready for signing
// @Tags staking
// @Accept  json
// @Produce  json
// @Param delegatorAddr path string true "delegator address"
// @Param body body rest.DelegateRequest true "The data required to delegate"
// @Success 200 {object} rest.postDelegations
// @Failure 400 {object} rest.ErrorResponse "Returned if the request is invalid"
// @Failure 401 {object} rest.ErrorResponse "Returned if chain-id required but not present, or delegation address incorrect"
// @Failure 402 {object} rest.ErrorResponse "Returned if fees or gas are invalid"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/delegators/{delegatorAddr}/delegations [post]
func postDelegationsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req DelegateRequest

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		msg := types.NewMsgDelegate(req.DelegatorAddress, req.ValidatorAddress, req.Amount)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		if !bytes.Equal(fromAddr, req.DelegatorAddress) {
			rest.WriteErrorResponse(w, http.StatusUnauthorized, "must use own delegator address")
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// postRedelegations is used to generate documentation for postRedelegationsHandlerFn
type postRedelegations struct {
	Msgs       []types.MsgBeginRedelegate `json:"msg" yaml:"msg"`
	Fee        auth.StdFee                `json:"fee" yaml:"fee"`
	Signatures []auth.StdSignature        `json:"signatures" yaml:"signatures"`
	Memo       string                     `json:"memo" yaml:"memo"`
}

// postRedelegationsHandlerFn implements a redelegation handler that is responsible
// for constructing a properly formatted redelegation transaction for signing.
//
// @Summary Generate a redelegation transaction
// @Description Generate a redelegation transaction that is ready for signing
// @Tags staking
// @Accept  json
// @Produce  json
// @Param delegatorAddr path string true "delegator address"
// @Param body body rest.RedelegateRequest true "The data required to delegate"
// @Success 200 {object} rest.postRedelegations
// @Failure 400 {object} rest.ErrorResponse "Returned if the request is invalid"
// @Failure 401 {object} rest.ErrorResponse "Returned if chain-id required but not present, or delegation address incorrect"
// @Failure 402 {object} rest.ErrorResponse "Returned if fees or gas are invalid"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/delegators/{delegatorAddr}/delegations [post]
func postRedelegationsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RedelegateRequest

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		msg := types.NewMsgBeginRedelegate(req.DelegatorAddress, req.ValidatorSrcAddress, req.ValidatorDstAddress, req.Amount)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		if !bytes.Equal(fromAddr, req.DelegatorAddress) {
			rest.WriteErrorResponse(w, http.StatusUnauthorized, "must use own delegator address")
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// postUnbondingDelegations is used to generate documentation for postUnbondingDelegationsHandlerFn
type postUnbondingDelegations struct {
	Msgs       []types.MsgUndelegate `json:"msg" yaml:"msg"`
	Fee        auth.StdFee           `json:"fee" yaml:"fee"`
	Signatures []auth.StdSignature   `json:"signatures" yaml:"signatures"`
	Memo       string                `json:"memo" yaml:"memo"`
}

// postUnbondingDelegationsHandlerFn implements an unbonding_delegation handler that is responsible
// for constructing a properly formatted unbonding_delegation transaction for signing.
//
// @Summary Generate an unbonding_delegation transaction
// @Description Generate an unbonding_delegation transaction that is ready for signing
// @Tags staking
// @Accept  json
// @Produce  json
// @Param delegatorAddr path string true "delegator address"
// @Param body body rest.UndelegateRequest true "The data required to undelegate"
// @Success 200 {object} rest.postUnbondingDelegations
// @Failure 400 {object} rest.ErrorResponse "Returned if the request is invalid"
// @Failure 401 {object} rest.ErrorResponse "Returned if chain-id required but not present, or delegation address incorrect"
// @Failure 402 {object} rest.ErrorResponse "Returned if fees or gas are invalid"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/delegators/{delegatorAddr}/unbonding_delegations [post]
func postUnbondingDelegationsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req UndelegateRequest

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		msg := types.NewMsgUndelegate(req.DelegatorAddress, req.ValidatorAddress, req.Amount)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		if !bytes.Equal(fromAddr, req.DelegatorAddress) {
			rest.WriteErrorResponse(w, http.StatusUnauthorized, "must use own delegator address")
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
