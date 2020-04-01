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

		if ctx.Marshaler != nil {
			marshaler = ctx.Marshaler
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
