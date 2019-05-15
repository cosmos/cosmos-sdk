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
	"github.com/cosmos/cosmos-sdk/x/nft/querier"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router,
	cdc *codec.Codec, queryRoute string) {

	// Get the total supply of a collection
	r.HandleFunc(
		"/nft/supply/{denom}", supplyNFTHandler(cdc, cliCtx, queryRoute),
	).Methods("GET")

	// Get the collections of NFTs owned by an address
	r.HandleFunc(
		"/nft/balance/{delegatorAddr}", getBalanceHandler(cdc, cliCtx, queryRoute),
	).Methods("GET")

	// Get the NFTs owned by an address from a given collection
	r.HandleFunc(
		"/nft/balance/{delegatorAddr}/collection/{denom}", getBalanceCollectionHandler(cdc, cliCtx, queryRoute),
	).Methods("GET")

	// Get all the NFT from a given collection
	r.HandleFunc(
		"/nft/collection/{denom}", getCollectionHandler(cdc, cliCtx, queryRoute),
	).Methods("GET")

	// Query a single NFT
	r.HandleFunc(
		"/nft/collection/{denom}/nft/{id}", getNFTHandler(cdc, cliCtx, queryRoute),
	).Methods("GET")
}

//TODO: query with data

func supplyNFTHandler(cdc *codec.Codec, cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		denom := mux.Vars(r)["denom"]

		params := querier.NewQueryCollectionParams(denom)
		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/supply/%s", queryRoute, denom), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func getBalanceHandler(cdc *codec.Codec, cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		address, err := sdk.AccAddressFromBech32(mux.Vars(r)["delegatorAddr"])
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		params := querier.NewQueryBalanceParams(address, "")
		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/balance", queryRoute), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func getBalanceCollectionHandler(cdc *codec.Codec, cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		denom := vars["denom"]
		address, err := sdk.AccAddressFromBech32(vars["delegatorAddr"])
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		params := querier.NewQueryBalanceParams(address, denom)
		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/balance", queryRoute), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func getCollectionHandler(cdc *codec.Codec, cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		denom := mux.Vars(r)["denom"]

		params := querier.NewQueryCollectionParams(denom)
		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/collection", queryRoute), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func getNFTHandler(cdc *codec.Codec, cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		denom := vars["denom"]
		tokenID := vars["id"]

		id, err := strconv.ParseUint(tokenID, 10, 64)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		params := querier.NewQueryNFTParams(denom, id)
		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/nft", queryRoute), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}
