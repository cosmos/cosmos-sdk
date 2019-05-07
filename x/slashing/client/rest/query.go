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

		signingInfo, code, err := getSigningInfo(cliCtx, storeName, cdc, sdk.ConsAddress(pk.Address()))
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

		valResult, err := rpc.GetValidators(cliCtx, &height)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		if len(valResult.Validators) == 0 {
			rest.PostProcessResponse(w, cdc, []slashing.ValidatorSigningInfo{}, cliCtx.Indent)
			return
		}

		fmt.Println("NUM VALIDATORS IN RESULT:", len(valResult.Validators))

		// TODO: this should happen when querying Validators from RPC,
		//  as soon as it's available this is not needed anymore
		// parameter page is (page-1) because ParseHTTPArgs starts with page 1, where our array start with 0
		start, end := adjustPagination(uint(len(valResult.Validators)), uint(page)-1, uint(limit))
		for _, validator := range valResult.Validators[start:end] {
			consAddr := validator.Address
			signingInfo, code, err := getSigningInfo(cliCtx, storeName, cdc, consAddr)
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
