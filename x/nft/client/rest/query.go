package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/nft/internal/types"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec, queryRoute string) {

	// Get the total supply of a collection
	r.HandleFunc(
		"/nft/supply/{denom}", getSupply(cdc, cliCtx, queryRoute),
	).Methods("GET")

	// Get the collections of NFTs owned by an address
	r.HandleFunc(
		"/nft/owner/{delegatorAddr}", getOwner(cdc, cliCtx, queryRoute),
	).Methods("GET")

	// Get the NFTs owned by an address from a given collection
	r.HandleFunc(
		"/nft/owner/{delegatorAddr}/collection/{denom}", getOwnerByDenom(cdc, cliCtx, queryRoute),
	).Methods("GET")

	// Get all the NFT from a given collection
	r.HandleFunc(
		"/nft/collection/{denom}", getCollection(cdc, cliCtx, queryRoute),
	).Methods("GET")

	// Query all denoms
	r.HandleFunc(
		"/nft/denoms", getDenoms(cdc, cliCtx, queryRoute),
	).Methods("GET")

	// Query a single NFT
	r.HandleFunc(
		"/nft/collection/{denom}/nft/{id}", getNFT(cdc, cliCtx, queryRoute),
	).Methods("GET")
}

func getSupply(cdc *codec.Codec, cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		denom := mux.Vars(r)["denom"]

		params := types.NewQueryCollectionParams(denom)
		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/supply/%s", queryRoute, denom), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getOwner(cdc *codec.Codec, cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		address, err := sdk.AccAddressFromBech32(mux.Vars(r)["delegatorAddr"])
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		params := types.NewQueryBalanceParams(address, "")
		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/owner", queryRoute), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getOwnerByDenom(cdc *codec.Codec, cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		denom := vars["denom"]
		address, err := sdk.AccAddressFromBech32(vars["delegatorAddr"])
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		params := types.NewQueryBalanceParams(address, denom)
		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/ownerByDenom", queryRoute), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getCollection(cdc *codec.Codec, cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		denom := mux.Vars(r)["denom"]

		params := types.NewQueryCollectionParams(denom)
		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/collection", queryRoute), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getDenoms(cdc *codec.Codec, cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/denoms", queryRoute), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getNFT(cdc *codec.Codec, cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		denom := vars["denom"]
		id := vars["id"]

		params := types.NewQueryNFTParams(denom, id)
		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/nft", queryRoute), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
