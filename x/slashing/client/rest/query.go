package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/slashing"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
	r.HandleFunc(
		"/slashing/validators/signing_info/{validatorPubKey}",
		signingInfoHandlerFn(cliCtx, slashing.StoreKey, cdc),
	).Methods("GET")

	r.HandleFunc(
		"/slashing/validators/signing_info",
		signingInfoHandlerListFn(cliCtx, slashing.StoreKey, cdc),
	).
		Methods("GET").
		Queries("page", "{page}", "pageSize", "{pageSize}")

	r.HandleFunc(
		"/slashing/validators/signing_info",
		signingInfoHandlerListFn(cliCtx, slashing.StoreKey, cdc),
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

		signingInfo, code, err := getSigningInfo(cliCtx, storeName, cdc, pk.Address())

		if err != nil {
			rest.WriteErrorResponse(w, code, err.Error())
			return
		}

		if code == http.StatusNoContent {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		rest.PostProcessResponse(w, cdc, signingInfo, cliCtx.Indent)
	}
}

// http request handler to query signing info
func signingInfoHandlerListFn(cliCtx context.CLIContext, storeName string, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var signingInfoList []slashing.ValidatorSigningInfo

		res, err := cliCtx.QueryWithData("custom/staking/validators", nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		if len(res) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		var validators []sdk.Validator
		err = cdc.UnmarshalBinaryLengthPrefixed(res, &validators)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Pagination should happen at QueryWithData
		pageParam := r.FormValue("page")
		pageSizeParam := r.FormValue("pageSize")

		// If we are in the not-paginated route return everything
		start := 0
		end := len(validators)
		if pageParam != "" && pageSizeParam != "" {
			page, errPage := strconv.Atoi(pageParam)
			pageSize, errPageSize := strconv.Atoi(pageSizeParam)

			// Quit if pages are non valid integers or dummy values
			if errPage != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, errPage.Error())
				return
			}
			if errPageSize != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, errPageSize.Error())
				return
			}
			if page < 0 {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, "Page should be greater than 0")
				return
			}
			if pageSize < 1 {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, "PageSize should be greater than 1")
				return
			}

			// If someone asks for pages bigger than our dataset, just return everything
			if pageSize > end {
				pageSize = end
			}

			// Do pagination only when healthy, fallback to 0
			if page*pageSize < end {
				start = page * pageSize
			}

			// Do pagination only when healthy, fallback to len(dataset)
			if start+pageSize < end {
				end = start + pageSize
			}
		}

		for _, validator := range validators[start:end] {
			pubKey := validator.GetConsPubKey()
			address := pubKey.Address()
			signingInfo, code, err := getSigningInfo(cliCtx, storeName, cdc, address)
			if err != nil {
				rest.WriteErrorResponse(w, code, err.Error())
				return
			}
			signingInfoList = append(signingInfoList, signingInfo)
		}

		if len(signingInfoList) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		rest.PostProcessResponse(w, cdc, signingInfoList, cliCtx.Indent)
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

func getSigningInfo(cliCtx context.CLIContext, storeName string, cdc *codec.Codec, address []byte) (signingInfo slashing.ValidatorSigningInfo, code int, err error) {
	key := slashing.GetValidatorSigningInfoKey(sdk.ConsAddress(address))

	res, err := cliCtx.QueryStore(key, storeName)
	if err != nil {
		code = http.StatusInternalServerError
		return
	}

	if len(res) == 0 {
		code = http.StatusNoContent
		return
	}

	err = cdc.UnmarshalBinaryLengthPrefixed(res, &signingInfo)
	if err != nil {
		code = http.StatusInternalServerError
		return
	}

	return
}
