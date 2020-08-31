package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// QueryBalancesRequestHandlerFn returns a REST handler that queries for all
// account balances or a specific balance by denomination.
func QueryBalancesRequestHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		vars := mux.Vars(r)
		bech32addr := vars["address"]

		addr, err := sdk.AccAddressFromBech32(bech32addr)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		ctx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		var (
			params interface{}
			route  string
		)

		denom := r.FormValue("denom")
		if denom == "" {
			params = types.NewQueryAllBalancesRequest(addr, nil)
			route = fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryAllBalances)
		} else {
			params = types.NewQueryBalanceRequest(addr, denom)
			route = fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryBalance)
		}

		bz, err := ctx.LegacyAmino.MarshalJSON(params)
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
func totalSupplyHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, page, limit, err := rest.ParseHTTPArgsWithLimit(r, 0)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		params := types.NewQueryTotalSupplyParams(page, limit)
		bz, err := clientCtx.LegacyAmino.MarshalJSON(params)

		if rest.CheckBadRequestError(w, err) {
			return
		}

		res, height, err := clientCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryTotalSupply), bz)

		if rest.CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}

// HTTP request handler to query the supply of a single denom
func supplyOfHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		denom := mux.Vars(r)["denom"]
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		params := types.NewQuerySupplyOfParams(denom)
		bz, err := clientCtx.LegacyAmino.MarshalJSON(params)

		if rest.CheckBadRequestError(w, err) {
			return
		}

		res, height, err := clientCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QuerySupplyOf), bz)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}
