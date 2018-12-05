package rest

import (
	"bytes"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake"

	"github.com/gorilla/mux"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec, kb keys.Keybase) {
	r.HandleFunc(
		"/stake/delegators/{delegatorAddr}/delegations",
		putDelegationsHandlerFn(cdc, kb, cliCtx),
	).Methods("PUT")
	r.HandleFunc(
		"/stake/delegators/{delegatorAddr}/unbonding_delegations",
		postUnbondingDelegationsHandlerFn(cdc, kb, cliCtx),
	).Methods("POST")
	r.HandleFunc(
		"/stake/delegators/{delegatorAddr}/redelegations",
		postRedelegationsHandlerFn(cdc, kb, cliCtx),
	).Methods("POST")
}

type (
	msgDelegationsInput struct {
		BaseReq       utils.BaseReq `json:"base_req"`
		DelegatorAddr string        `json:"delegator_addr"` // in bech32
		ValidatorAddr string        `json:"validator_addr"` // in bech32
		Delegation    sdk.Coin      `json:"delegation"`
	}

	msgBeginRedelegateInput struct {
		BaseReq          utils.BaseReq `json:"base_req"`
		DelegatorAddr    string        `json:"delegator_addr"`     // in bech32
		ValidatorSrcAddr string        `json:"validator_src_addr"` // in bech32
		ValidatorDstAddr string        `json:"validator_dst_addr"` // in bech32
		SharesAmount     string        `json:"shares"`
	}

	msgBeginUnbondingInput struct {
		BaseReq       utils.BaseReq `json:"base_req"`
		DelegatorAddr string        `json:"delegator_addr"` // in bech32
		ValidatorAddr string        `json:"validator_addr"` // in bech32
		SharesAmount  string        `json:"shares"`
	}
)

// If not, we can just use CompleteAndBroadcastTxREST.
func putDelegationsHandlerFn(cdc *codec.Codec, kb keys.Keybase, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req msgDelegationsInput

		err := utils.ReadRESTReq(w, r, cdc, &req)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		info, err := kb.Get(baseReq.Name)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		delAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddr)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddr)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		if !bytes.Equal(info.GetPubKey().Address(), delAddr) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "Must use own delegator address")
			return
		}

		msg := stake.NewMsgDelegate(delAddr, valAddr, req.Delegation)
		err = msg.ValidateBasic()
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.CompleteAndBroadcastTxREST(w, r, cliCtx, baseReq, []sdk.Msg{msg}, cdc)
	}
}

func postRedelegationsHandlerFn(cdc *codec.Codec, kb keys.Keybase, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req msgBeginRedelegateInput

		err := utils.ReadRESTReq(w, r, cdc, &req)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		info, err := kb.Get(baseReq.Name)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		delAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddr)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		if !bytes.Equal(info.GetPubKey().Address(), delAddr) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "Must use own delegator address")
			return
		}

		valSrcAddr, err := sdk.ValAddressFromBech32(req.ValidatorSrcAddr)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		valDstAddr, err := sdk.ValAddressFromBech32(req.ValidatorDstAddr)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		shares, err := sdk.NewDecFromStr(req.SharesAmount)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		msg := stake.NewMsgBeginRedelegate(delAddr, valSrcAddr, valDstAddr, shares)
		err = msg.ValidateBasic()
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.CompleteAndBroadcastTxREST(w, r, cliCtx, baseReq, []sdk.Msg{msg}, cdc)
	}
}

func postUnbondingDelegationsHandlerFn(cdc *codec.Codec, kb keys.Keybase, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req msgBeginUnbondingInput

		err := utils.ReadRESTReq(w, r, cdc, &req)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		delAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddr)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		info, err := kb.Get(baseReq.Name)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		if !bytes.Equal(info.GetPubKey().Address(), delAddr) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "Must use own delegator address")
			return
		}

		valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddr)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		shares, err := sdk.NewDecFromStr(req.SharesAmount)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		msg := stake.NewMsgBeginUnbonding(delAddr, valAddr, shares)
		err = msg.ValidateBasic()
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.CompleteAndBroadcastTxREST(w, r, cliCtx, baseReq, []sdk.Msg{msg}, cdc)
	}
}
