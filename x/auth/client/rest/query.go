package rest

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"github.com/cosmos/cosmos-sdk/x/auth"
)

// register REST routes
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec, storeName string) {
	r.HandleFunc(
		"/auth/accounts/{address}",
		QueryAccountRequestHandlerFn(storeName, cdc, context.GetAccountDecoder(cdc), cliCtx),
	).Methods("GET")

	r.HandleFunc(
		"/bank/balances/{address}",
		QueryBalancesRequestHandlerFn(storeName, cdc, context.GetAccountDecoder(cdc), cliCtx),
	).Methods("GET")
}

// query accountREST Handler
func QueryAccountRequestHandlerFn(
	storeName string, cdc *codec.Codec,
	decoder auth.AccountDecoder, cliCtx context.CLIContext,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bech32addr := vars["address"]

		addr, err := sdk.AccAddressFromBech32(bech32addr)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		res, err := cliCtx.QueryStore(auth.AddressStoreKey(addr), storeName)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// the query will return empty account if there is no data
		if len(res) == 0 {
			rest.PostProcessResponse(w, cdc, auth.BaseAccount{}, cliCtx.Indent)
			return
		}

		// decode the value
		account, err := decoder(res)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, account, cliCtx.Indent)
	}
}

// query accountREST Handler
func QueryBalancesRequestHandlerFn(
	storeName string, cdc *codec.Codec,
	decoder auth.AccountDecoder, cliCtx context.CLIContext,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		bech32addr := vars["address"]

		addr, err := sdk.AccAddressFromBech32(bech32addr)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		res, err := cliCtx.QueryStore(auth.AddressStoreKey(addr), storeName)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// the query will return empty if there is no data for this account
		if len(res) == 0 {
			rest.PostProcessResponse(w, cdc, sdk.Coins{}, cliCtx.Indent)
			return
		}

		// decode the value
		account, err := decoder(res)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, account.GetCoins(), cliCtx.Indent)
	}
}
