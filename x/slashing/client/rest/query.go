package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/rpc"
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
		signingInfoHandlerListFn(cliCtx, slashing.StoreKey, cdc),
	).Methods("GET").Queries("page", "{page}", "limit", "{limit}")

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

		rest.PostProcessResponse(w, cdc, signingInfo, cliCtx.Indent)
	}
}

// http request handler to query signing info
func signingInfoHandlerListFn(cliCtx context.CLIContext, storeName string, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var signingInfoList []slashing.ValidatorSigningInfo

		_, page, limit, err := rest.ParseHTTPArgs(r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		height, err := rpc.GetChainHeight(cliCtx)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		validators, err := rpc.GetValidators(cliCtx, &height)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		if len(validators.Validators) == 0 {
			rest.PostProcessResponse(w, cdc, signingInfoList, cliCtx.Indent)
			return
		}

		// TODO: this should happen when querying Validators from RPC,
		//  as soon as it's available this is not needed anymore
		// parameter page is (page-1) because ParseHTTPArgs starts with page 1, where our array start with 0
		start, end := adjustPagination(uint(len(validators.Validators)), uint(page)-1, uint(limit))
		for _, validator := range validators.Validators[start:end] {
			address := validator.Address
			signingInfo, code, err := getSigningInfo(cliCtx, storeName, cdc, address)
			if err != nil {
				rest.WriteErrorResponse(w, code, err.Error())
				return
			}
			signingInfoList = append(signingInfoList, signingInfo)
		}

		if len(signingInfoList) == 0 {
			rest.PostProcessResponse(w, cdc, signingInfoList, cliCtx.Indent)
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
		code = http.StatusOK
		return
	}

	err = cdc.UnmarshalBinaryLengthPrefixed(res, &signingInfo)
	if err != nil {
		code = http.StatusInternalServerError
		return
	}

	return
}

// Adjust pagination with page starting from 0
func adjustPagination(size, page, limit uint) (start uint, end uint) {
	// If someone asks for pages bigger than our dataset, just return everything
	if limit > size {
		return 0, size
	}

	// Do pagination when healthy, fallback to 0
	start = 0
	if page*limit < size {
		start = page * limit
	}

	// Do pagination only when healthy, fallback to len(dataset)
	end = size
	if start+limit <= size {
		end = start + limit
	}

	return start, end
}
