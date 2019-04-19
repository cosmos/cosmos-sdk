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
	"github.com/cosmos/cosmos-sdk/x/nfts"

	"github.com/gorilla/mux"
)

// TODO: sdk-application-tutorial is missing one GET route in RegisterRoutes in rest.go

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec, storeName string) {
	r.HandleFunc(fmt.Sprintf("/%s/transfer", storeName), transferNFTHandler(cdc, cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/metadata", storeName), editNFTMetadataHandler(cdc, cliCtx)).Methods("PUT")
	r.HandleFunc(fmt.Sprintf("/%s/mint", storeName), mintNFTHandler(cdc, cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/burn", storeName), burnNFTHandler(cdc, cliCtx)).Methods("DELETE")

	r.HandleFunc(fmt.Sprintf("/%s/balanceOf", storeName), balanceOfHandler(cdc, cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/ownerOf", storeName), ownerOfHandler(cdc, cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/metadata", storeName), metadataHandler(cdc, cliCtx, storeName)).Methods("GET")
}

func balanceOfHandler(cdc *codec.Codec, cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		denom := vars["denom"]
		account := vars["account"]
		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/balanceOf/%s/%s", storeName, denom, account), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func ownerOfHandler(cdc *codec.Codec, cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		denom := vars["denom"]
		tokenID := vars["tokenID"]
		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/ownerOf/%s/%s", storeName, denom, tokenID), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
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
		msg := nfts.NewMsgTransferNFT(cliCtx.GetFromAddress(), recipient, nfts.Denom(req.Denom), nfts.TokenID(tokenID))
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
		msg := nfts.NewMsgEditNFTMetadata(cliCtx.GetFromAddress(), nfts.Denom(req.Denom), nfts.TokenID(tokenID),
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
		msg := nfts.NewMsgMintNFT(cliCtx.GetFromAddress(), recipient, nfts.TokenID(tokenID), nfts.Denom(req.Denom),
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
		msg := nfts.NewMsgBurnNFT(cliCtx.GetFromAddress(), nfts.TokenID(tokenID), nfts.Denom(req.Denom))
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		clientrest.WriteGenerateStdTxResponse(w, cdc, cliCtx, baseReq, []sdk.Msg{msg})
	}
}
