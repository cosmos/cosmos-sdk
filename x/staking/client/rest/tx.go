package rest

import (
	"bytes"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func registerTxHandlers(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/delegations",
		newPostDelegationsHandlerFn(cliCtx),
	).Methods("POST")
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/unbonding_delegations",
		newPostUnbondingDelegationsHandlerFn(cliCtx),
	).Methods("POST")
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/redelegations",
		newPostRedelegationsHandlerFn(cliCtx),
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

func newPostDelegationsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
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
		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		if !bytes.Equal(fromAddr, req.DelegatorAddress) {
			rest.WriteErrorResponse(w, http.StatusUnauthorized, "must use own delegator address")
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

func newPostRedelegationsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
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
		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		if !bytes.Equal(fromAddr, req.DelegatorAddress) {
			rest.WriteErrorResponse(w, http.StatusUnauthorized, "must use own delegator address")
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

func newPostUnbondingDelegationsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
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
		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		if !bytes.Equal(fromAddr, req.DelegatorAddress) {
			rest.WriteErrorResponse(w, http.StatusUnauthorized, "must use own delegator address")
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// ---------------------------------------------------------------------------
// Deprecated
//
// TODO: Remove once client-side Protobuf migration has been completed.
// ---------------------------------------------------------------------------
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
		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		if !bytes.Equal(fromAddr, req.DelegatorAddress) {
			rest.WriteErrorResponse(w, http.StatusUnauthorized, "must use own delegator address")
			return
		}

		authclient.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

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
		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		if !bytes.Equal(fromAddr, req.DelegatorAddress) {
			rest.WriteErrorResponse(w, http.StatusUnauthorized, "must use own delegator address")
			return
		}

		authclient.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

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
		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		if !bytes.Equal(fromAddr, req.DelegatorAddress) {
			rest.WriteErrorResponse(w, http.StatusUnauthorized, "must use own delegator address")
			return
		}

		authclient.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
