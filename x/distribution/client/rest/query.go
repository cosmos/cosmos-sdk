package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/common"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/mux"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router,
	cdc *codec.Codec, queryRoute string) {

	// Get the total rewards balance from all delegations
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/rewards",
		delegatorRewardsHandlerFn(cliCtx, cdc, queryRoute),
	).Methods("GET")

	// Query a delegation reward
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/rewards/{validatorAddr}",
		delegatorRewardsHandlerFn(cliCtx, cdc, queryRoute),
	).Methods("GET")

	// Get the rewards withdrawal address
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/withdraw_address",
		delegatorWithdrawalAddrHandlerFn(cliCtx, cdc, queryRoute),
	).Methods("GET")

	// Validator distribution information
	r.HandleFunc(
		"/distribution/validator/{validatorAddr}",
		validatorInfoHandlerFn(cliCtx, cdc, queryRoute),
	).Methods("GET")

	// Commission and self-delegation rewards of a single a validator
	r.HandleFunc(
		"/distribution/validator/{validatorAddr}/rewards",
		validatorRewardsHandlerFn(cliCtx, cdc, queryRoute),
	).Methods("GET")

	// Get the current distribution parameter values
	r.HandleFunc(
		"/distribution/parameters",
		paramsHandlerFn(cliCtx, cdc, queryRoute),
	).Methods("GET")

	// Get the current distribution pool
	r.HandleFunc(
		"/distribution/pool",
		poolHandlerFn(cliCtx, cdc, queryRoute),
	).Methods("GET")
}

// HTTP request handler to query delegators rewards
func delegatorRewardsHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// query for rewards from a particular delegator
		res, err := common.QueryRewards(cliCtx, cdc, queryRoute, mux.Vars(r)["delegatorAddr"], "")
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		utils.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

// HTTP request handler to query a delegation rewards
func delegationRewardsHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// query for rewards from a particular delegation
		res, err := common.QueryRewards(cliCtx, cdc, queryRoute,
			mux.Vars(r)["delegatorAddr"], mux.Vars(r)["validatorAddr"])
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		utils.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

// HTTP request handler to query a delegation rewards
func delegatorWithdrawalAddrHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		delegatorAddr, err := sdk.AccAddressFromBech32(mux.Vars(r)["delegatorAddr"])
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		params := distribution.NewQueryDelegatorWithdrawAddrParams(delegatorAddr)
		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/withdraw_addr", queryRoute), bz)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		utils.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

// HTTP request handler to query validator's distribution information
func validatorInfoHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {}
}

// HTTP request handler to query validator's commission and self-delegation rewards
func validatorRewardsHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {}
}

// HTTP request handler to query the distribution params values
func paramsHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		params, err := common.QueryParams(cliCtx, queryRoute)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		utils.PostProcessResponse(w, cdc, params, cliCtx.Indent)
	}
}

// HTTP request handler to query the pool information
func poolHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/params/fee_pool", queryRoute), []byte{})
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		utils.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}
