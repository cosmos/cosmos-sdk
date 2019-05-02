package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router,
	cdc *codec.Codec, queryRoute string) {

	// Get the total supply of a collection
	r.HandleFunc(
		"/nfts/supply/{denom}", supplyNFTHandler(cdc, cliCtx, queryRoute),
	).Methods("GET")

	// Get the collections of NFTs owned by an address
	r.HandleFunc(
		"/nfts/balance/{delegatorAddr}", getNFTsBalanceHandler(cdc, cliCtx, queryRoute),
	).Methods("GET")

	// Get the NFTs owned by an address from a given collection
	r.HandleFunc(
		"/nfts/balance/{delegatorAddr}/collection/{denom}", getNFTsBalanceHandler(cdc, cliCtx, queryRoute),
	).Methods("GET")

	// Get all the NFT from a given collection
	r.HandleFunc(
		"/nfts/collection/{denom}", getCollectionHandler(cdc, cliCtx, queryRoute),
	).Methods("GET")

	// Query a single NFT
	r.HandleFunc(
		"/nfts/collection/{denom}/nft/{id}", getNFTHandler(cdc, cliCtx, queryRoute),
	).Methods("GET")
}
