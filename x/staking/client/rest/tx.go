package rest

import (
	"bytes"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(
		"/staking/validators/{validatorAddr}",
		createValidatorHandlerFn(cliCtx),
	).Methods("POST")
	r.HandleFunc(
		"/staking/validators/{validatorAddr}",
		editValidatorHandlerFn(cliCtx),
	).Methods("PUT")
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
	// CreateValidatorRequest defines the properties of a create validator request's body.
	CreateValidatorRequest struct {
		BaseReq           rest.BaseReq          `json:"base_req" yaml:"base_req"`
		ValidatorAddress  sdk.ValAddress        `json:"validator_address" yaml:"validator_address"` // in bech32
		Pubkey            string                `json:"pubkey" yaml:"pubkey"`
		Amount             sdk.Coin              `json:"amount" yaml:"amount"`
		Description       types.Description     `json:"description" yaml:"description"`
		Commission        types.CommissionRates `json:"commission" yaml:"commission"`
		MinSelfDelegation sdk.Int               `json:"min_self_delegation" yaml:"min_self_delegation"`
	}

	// EditValidatorRequest defines the properties of a edit validator request's body.
	EditValidatorRequest struct {
		BaseReq           rest.BaseReq      `json:"base_req" yaml:"base_req"`
		ValidatorAddress  sdk.ValAddress    `json:"address" yaml:"address"`
		Description       types.Description `json:"description" yaml:"description"`
		CommissionRate    sdk.Dec           `json:"commission_rate" yaml:"commission_rate"`
		MinSelfDelegation sdk.Int           `json:"min_self_delegation" yaml:"min_self_delegation"`
	}

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

// createValidatorHandlerFn implements a create validator handler that is responsible
// for constructing a properly formatted create validator transaction for signing.
//
// @Summary Generate a create validator transaction
// @Description Generate a create validator transaction that is ready for signing
// @Tags staking
// @Accept  json
// @Produce  json
// @Param validatorAddr path string true "The validator address"
// @Param body body rest.CreateValidatorRequest true "The create validator request payload"
// @Success 200 {object} rest.createValidator "Returns the unsigned transaction"
// @Failure 400 {object} rest.ErrorResponse "Returned if the request is invalid"
// @Failure 401 {object} rest.ErrorResponse "Returned if chain-id required but not present"
// @Failure 402 {object} rest.ErrorResponse "Returned if fees or gas are invalid"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/validators/{validatorAddr} [post]
func createValidatorHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateValidatorRequest

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		pk, err := sdk.GetConsPubKeyBech32(req.Pubkey)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgCreateValidator(req.ValidatorAddress, pk, req.Value, req.Description, req.Commission, req.MinSelfDelegation)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// editValidatorHandlerFn implements a edit validator handler that is responsible
// for constructing a properly formatted edit validator transaction for signing.
//
// @Summary Generate a edit validator transaction
// @Description Generate a edit validator transaction that is ready for signing
// @Tags staking
// @Accept  json
// @Produce  json
// @Param validatorAddr path string true "The validator address"
// @Param body body rest.EditValidatorRequest true "The edit validator request payload"
// @Success 200 {object} rest.editValidator "Returns the unsigned transaction"
// @Failure 400 {object} rest.ErrorResponse "Returned if the request is invalid"
// @Failure 401 {object} rest.ErrorResponse "Returned if chain-id required but not present"
// @Failure 402 {object} rest.ErrorResponse "Returned if fees or gas are invalid"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/validators/{validatorAddr} [put]
func editValidatorHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req EditValidatorRequest

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		msg := types.NewMsgEditValidator(req.ValidatorAddress, req.Description, &req.CommissionRate, &req.MinSelfDelegation)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// postDelegationsHandlerFn implements a delegation handler that is responsible
// for constructing a properly formatted delegation transaction for signing.
//
// @Summary Generate a delegation transaction
// @Description Generate a delegation transaction that is ready for signing
// @Tags staking
// @Accept  json
// @Produce  json
// @Param delegatorAddr path string true "The delegator address"
// @Param body body rest.DelegateRequest true "The delegate request payload"
// @Success 200 {object} rest.postDelegation "Returns the unsigned transaction"
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

// postRedelegationsHandlerFn implements a redelegation handler that is responsible
// for constructing a properly formatted redelegation transaction for signing.
//
// @Summary Generate a redelegation transaction
// @Description Generate a redelegation transaction that is ready for signing
// @Tags staking
// @Accept  json
// @Produce  json
// @Param delegatorAddr path string true "The delegator address"
// @Param body body rest.RedelegateRequest true "The redelegate request payload"
// @Success 200 {object} rest.postRedelegation "Returns the unsigned transaction"
// @Failure 400 {object} rest.ErrorResponse "Returned if the request is invalid"
// @Failure 401 {object} rest.ErrorResponse "Returned if chain-id required but not present, or delegation address incorrect"
// @Failure 402 {object} rest.ErrorResponse "Returned if fees or gas are invalid"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/delegators/{delegatorAddr}/redelegations [post]
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

// postUnbondingDelegationsHandlerFn implements an unbonding_delegation handler that is responsible
// for constructing a properly formatted unbonding_delegation transaction for signing.
//
// @Summary Generate an unbonding transaction
// @Description Generate an unbonding transaction that is ready for signing
// @Tags staking
// @Accept  json
// @Produce  json
// @Param delegatorAddr path string true "The delegator address"
// @Param body body rest.UndelegateRequest true "The undelegate request payload"
// @Success 200 {object} rest.postUndelegate "Returns the unsigned transaction"
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
