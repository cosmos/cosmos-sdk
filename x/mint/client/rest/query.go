package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(
		"/minting/parameters",
		queryParamsHandlerFn(cliCtx),
	).Methods("GET")

	r.HandleFunc(
		"/minting/inflation",
		queryInflationHandlerFn(cliCtx),
	).Methods("GET")

	r.HandleFunc(
		"/minting/annual-provisions",
		queryAnnualProvisionsHandlerFn(cliCtx),
	).Methods("GET")
}

type mintParams struct { // nolint: deadcode unused
	Height int64        `json:"height"`
	Result types.Params `json:"result"`
}

// queryParamsHandlerFn implements a query route for params of the mint module
//
// @Summary Minting module parameters
// @Tags mint
// @Produce json
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} mintParams
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /minting/parameters [get]
func queryParamsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryParameters)

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.QueryWithData(route, nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

type mintInflation struct { // nolint: deadcode unused
	Height int64  `json:"height"`
	Result string `json:"result"`
}

// queryInflationHandlerFn implements a query for current minting inflation value
//
// @Summary Current minting inflation value
// @Tags mint
// @Produce json
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} mintInflation
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /minting/inflation [get]
func queryInflationHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryInflation)

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.QueryWithData(route, nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

type mintAnnualProvisions struct { // nolint: deadcode unused
	Height int64  `json:"height"`
	Result string `json:"result"`
}

// queryAnnualProvisionsHandlerFn implements a query for current minting annual provisions value
//
// @Summary Current minting annual provisions value
// @Tags mint
// @Produce json
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} mintAnnualProvisions
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /minting/inflation [get]
func queryAnnualProvisionsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryAnnualProvisions)

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.QueryWithData(route, nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
