package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/slashing"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
	r.HandleFunc(
		"/slashing/validators/{validatorPubKey}/signing_info",
		signingInfoHandlerFn(cliCtx, slashing.StoreKey, cdc),
	).Methods("GET")

	r.HandleFunc(
		"/slashing/signing_infos",
		signingInfoHandlerListFn(cliCtx, cdc),
	).Methods("GET")

	r.HandleFunc(
		"/slashing/parameters",
		queryParamsHandlerFn(cdc, cliCtx),
	).Methods("GET")
}

// http request handler to query signing info
func signingInfoHandlerFn(cliCtx context.CLIContext, storeName string, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		pk, err := sdk.GetConsPubKeyBech32(vars["validatorPubKey"])
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		signingInfo, code, err := getSigningInfo(cliCtx, storeName, cdc, sdk.ConsAddress(pk.Address()))
		if err != nil {
			rest.WriteErrorResponse(w, code, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, signingInfo, cliCtx.Indent)
	}
}

// http request handler to query signing info
func signingInfoHandlerListFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, page, limit, err := rest.ParseHTTPArgs(r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Override default limit if it wasn't provided as the querier will set
		// it to the correct default limit.
		if l := r.FormValue("limit"); l == "" {
			limit = 0
		}

		params := slashing.NewQuerySigningInfosParams(page, limit)
		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		route := fmt.Sprintf("custom/%s/%s", slashing.QuerierRoute, slashing.QuerySigningInfos)
		res, err := cliCtx.QueryWithData(route, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func queryParamsHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route := fmt.Sprintf("custom/%s/parameters", slashing.QuerierRoute)

		res, err := cliCtx.QueryWithData(route, nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func getSigningInfo(
	cliCtx context.CLIContext, storeName string, cdc *codec.Codec, consAddr sdk.ConsAddress,
) (signingInfo slashing.ValidatorSigningInfo, code int, err error) {
	key := slashing.GetValidatorSigningInfoKey(consAddr)

	res, err := cliCtx.QueryStore(key, storeName)
	if err != nil {
		code = http.StatusInternalServerError
		return signingInfo, code, err
	}

	if len(res) == 0 {
		code = http.StatusOK
		return signingInfo, code, err
	}

	err = cdc.UnmarshalBinaryLengthPrefixed(res, &signingInfo)
	if err != nil {
		code = http.StatusInternalServerError
		return signingInfo, code, err
	}

	return signingInfo, code, nil
}
