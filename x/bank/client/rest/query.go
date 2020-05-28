package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// QueryBalancesRequestHandlerFn returns a REST handler that queries for all
// account balances or a specific balance by denomination.
func QueryBalancesRequestHandlerFn(ctx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		vars := mux.Vars(r)
		bech32addr := vars["address"]

		addr, err := sdk.AccAddressFromBech32(bech32addr)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		ctx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, ctx, r)
		if !ok {
			return
		}

		var (
			params interface{}
			route  string
		)

		// TODO: Remove once client-side Protobuf migration has been completed.
		// ref: https://github.com/cosmos/cosmos-sdk/issues/5864
		var marshaler codec.JSONMarshaler

		if ctx.JSONMarshaler != nil {
			marshaler = ctx.JSONMarshaler
		} else {
			marshaler = ctx.Codec
		}

		denom := r.FormValue("denom")
		if denom == "" {
			params = types.NewQueryAllBalancesParams(addr)
			route = fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryAllBalances)
		} else {
			params = types.NewQueryBalanceParams(addr, denom)
			route = fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryBalance)
		}

		bz, err := marshaler.MarshalJSON(params)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		res, height, err := ctx.QueryWithData(route, bz)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		ctx = ctx.WithHeight(height)
		rest.PostProcessResponse(w, ctx, res)
	}
}

// HTTP request handler to query the total supply of coins
func totalSupplyHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, page, limit, err := rest.ParseHTTPArgsWithLimit(r, 0)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		params := types.NewQueryTotalSupplyParams(page, limit)
		bz, err := cliCtx.Codec.MarshalJSON(params)

		if rest.CheckBadRequestError(w, err) {
			return
		}

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryTotalSupply), bz)

		if rest.CheckInternalServerError(w, err) {
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// HTTP request handler to query the supply of a single denom
func supplyOfHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		denom := mux.Vars(r)["denom"]
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		params := types.NewQuerySupplyOfParams(denom)
		bz, err := cliCtx.Codec.MarshalJSON(params)

		if rest.CheckBadRequestError(w, err) {
			return
		}

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QuerySupplyOf), bz)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
