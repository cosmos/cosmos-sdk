package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/slashing/internal/types"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(
		"/slashing/validators/{validatorPubKey}/signing_info",
		signingInfoHandlerFn(cliCtx),
	).Methods("GET")

	r.HandleFunc(
		"/slashing/signing_infos",
		signingInfoHandlerListFn(cliCtx),
	).Methods("GET")

	r.HandleFunc(
		"/slashing/parameters",
		queryParamsHandlerFn(cliCtx),
	).Methods("GET")
}

// http request handler to query signing info for a specific validator
//
// @Summary Get the signing info of a given validator
// @Description Get the signing info of a given validator by public key
// @Tags slashing
// @Produce json
// @Param validatorPubKey path string true "Bech32 validator public key"
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} rest.validatorSignInfo
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have a valid height or invalid validator public key "
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /slashing/validators/{validatorPubKey}/signing_info [get]
func signingInfoHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		pk, err := sdk.GetConsPubKeyBech32(vars["validatorPubKey"])
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		params := types.NewQuerySigningInfoParams(sdk.ConsAddress(pk.Address()))

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QuerySigningInfo)
		res, height, err := cliCtx.QueryWithData(route, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// http request handler to query signing info for all validators
//
// @Summary Get the signing info of all validators
// @Description Get the signing info of all validators
// @Tags slashing
// @Produce json
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} rest.validatorsSigningInfo
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /slashing/signing_infos [get]
func signingInfoHandlerListFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, page, limit, err := rest.ParseHTTPArgsWithLimit(r, 0)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		params := types.NewQuerySigningInfosParams(page, limit)
		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QuerySigningInfos)
		res, height, err := cliCtx.QueryWithData(route, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// http request handler to query params for the slashing module
//
// @Summary Get the current slashing parameters
// @Description Get the current slashing parameters
// @Tags slashing
// @Produce json
// @Param height query string false "Block height to execute query (defaults to chain tip)"
// @Success 200 {object} rest.slashingParams
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /slashing/parameters [get]
func queryParamsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		route := fmt.Sprintf("custom/%s/parameters", types.QuerierRoute)

		res, height, err := cliCtx.QueryWithData(route, nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
