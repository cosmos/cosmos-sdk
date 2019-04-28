package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/context"
	clientrest "github.com/cosmos/cosmos-sdk/client/rest"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/nfts/types"

	"github.com/gorilla/mux"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router,
	cdc *codec.Codec, queryRoute string) {

	// Transfer an NFT to an address
	r.HandleFunc(
		"/nfts/transfer", transferNFTHandler(cdc, cliCtx),
	).Methods("POST")

	// Mint an NFT
	r.HandleFunc(
		"/nfts/collection/{denom}", mintNFTHandler(cdc, cliCtx),
	).Methods("POST")

	// Burn an NFT
	r.HandleFunc(
		"/nfts/collection/{denom}/nft/{id}", burnNFTHandler(cdc, cliCtx),
	).Methods("DELETE")

	// Update an NFT metadata
	r.HandleFunc(
		"/nfts/collection/{denom}/nft/{id}/metadata", editNFTMetadataHandler(cdc, cliCtx),
	).Methods("PUT")
}

func metadataHandler(cdc *codec.Codec, cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		denom := vars["denom"]
		tokenID := vars["tokenID"]
		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/metadata/%s/%s", storeName, denom, tokenID), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

type transferNFTReq struct {
	BaseReq   rest.BaseReq `json:"base_req"`
	Denom     string       `json:"denom"`
	TokenID   string       `json:"tokenID"`
	Recipient string       `json:"recipient"`
}

func transferNFTHandler(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req transferNFTReq
		if !rest.ReadRESTReq(w, r, cdc, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}
		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}
		recipient, err := sdk.AccAddressFromBech32(req.Recipient)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		tokenID, err := strconv.ParseUint(req.TokenID, 10, 64)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		// create the message
		msg := types.NewMsgTransferNFT(cliCtx.GetFromAddress(), recipient, types.Denom(req.Denom), types.TokenID(tokenID))
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		clientrest.WriteGenerateStdTxResponse(w, cdc, cliCtx, baseReq, []sdk.Msg{msg})
	}
}

type editNFTMetadataReq struct {
	BaseReq     rest.BaseReq `json:"base_req"`
	Denom       string       `json:"denom"`
	TokenID     string       `json:"tokenID"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Image       string       `json:"image"`
	TokenURI    string       `json:"tokenURI"`
}

func editNFTMetadataHandler(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req editNFTMetadataReq
		if !rest.ReadRESTReq(w, r, cdc, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}
		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}
		tokenID, err := strconv.ParseUint(req.TokenID, 10, 64)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		editName := req.Name != ""
		editDescription := req.Description != ""
		editImage := req.Image != ""
		editTokenURI := req.TokenURI != ""
		// create the message
		msg := types.NewMsgEditNFTMetadata(cliCtx.GetFromAddress(), types.Denom(req.Denom), types.TokenID(tokenID),
			editName, editDescription, editImage, editTokenURI,
			req.Name, req.Description, req.Image, req.TokenURI,
		)
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		clientrest.WriteGenerateStdTxResponse(w, cdc, cliCtx, baseReq, []sdk.Msg{msg})
	}
}

type mintNFTReq struct {
	BaseReq     rest.BaseReq `json:"base_req"`
	Recipient   string       `json:"recipient"`
	Denom       string       `json:"denom"`
	TokenID     string       `json:"tokenID"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Image       string       `json:"image"`
	TokenURI    string       `json:"tokenURI"`
}

func mintNFTHandler(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req mintNFTReq
		if !rest.ReadRESTReq(w, r, cdc, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}
		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}
		recipient, err := sdk.AccAddressFromBech32(req.Recipient)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		tokenID, err := strconv.ParseUint(req.TokenID, 10, 64)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		// create the message
		msg := types.NewMsgMintNFT(cliCtx.GetFromAddress(), recipient, types.TokenID(tokenID), types.Denom(req.Denom),
			req.Name, req.Description, req.Image, req.TokenURI,
		)
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		clientrest.WriteGenerateStdTxResponse(w, cdc, cliCtx, baseReq, []sdk.Msg{msg})
	}
}

type burnNFTReq struct {
	BaseReq rest.BaseReq `json:"base_req"`
	Denom   string       `json:"denom"`
	TokenID string       `json:"tokenID"`
}

func burnNFTHandler(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req burnNFTReq
		if !rest.ReadRESTReq(w, r, cdc, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}
		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}
		tokenID, err := strconv.ParseUint(req.TokenID, 10, 64)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		// create the message
		msg := types.NewMsgBurnNFT(cliCtx.GetFromAddress(), types.TokenID(tokenID), types.Denom(req.Denom))
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		clientrest.WriteGenerateStdTxResponse(w, cdc, cliCtx, baseReq, []sdk.Msg{msg})
	}
}
