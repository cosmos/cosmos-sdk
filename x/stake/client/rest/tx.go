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
		postDelegationsHandlerFn(cdc, kb, cliCtx),
	).Methods("POST")
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
		BaseReq       utils.BaseReq  `json:"base_req"`
		DelegatorAddr sdk.AccAddress `json:"delegator_addr"` // in bech32
		ValidatorAddr sdk.ValAddress `json:"validator_addr"` // in bech32
		Delegation    sdk.Coin       `json:"delegation"`
	}

	msgBeginRedelegateInput struct {
		BaseReq          utils.BaseReq  `json:"base_req"`
		DelegatorAddr    sdk.AccAddress `json:"delegator_addr"`     // in bech32
		ValidatorSrcAddr sdk.ValAddress `json:"validator_src_addr"` // in bech32
		ValidatorDstAddr sdk.ValAddress `json:"validator_dst_addr"` // in bech32
		SharesAmount     sdk.Dec        `json:"shares"`
	}

	msgBeginUnbondingInput struct {
		BaseReq       utils.BaseReq  `json:"base_req"`
		DelegatorAddr sdk.AccAddress `json:"delegator_addr"` // in bech32
		ValidatorAddr sdk.ValAddress `json:"validator_addr"` // in bech32
		SharesAmount  sdk.Dec        `json:"shares"`
	}
)

func postDelegationsHandlerFn(cdc *codec.Codec, kb keys.Keybase, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req msgDelegationsInput

		err := utils.ReadRESTReq(w, r, cdc, &req)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx = cliCtx.WithGenerateOnly(req.BaseReq.GenerateOnly)
		cliCtx = cliCtx.WithSimulation(req.BaseReq.Simulate)

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w, cliCtx) {
			return
		}

		info, err := kb.Get(baseReq.Name)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		if !bytes.Equal(info.GetPubKey().Address(), req.DelegatorAddr) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "Must use own delegator address")
			return
		}

		msg := stake.NewMsgDelegate(req.DelegatorAddr, req.ValidatorAddr, req.Delegation)
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

		cliCtx = cliCtx.WithGenerateOnly(req.BaseReq.GenerateOnly)
		cliCtx = cliCtx.WithSimulation(req.BaseReq.Simulate)

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w, cliCtx) {
			return
		}

		info, err := kb.Get(baseReq.Name)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		if !bytes.Equal(info.GetPubKey().Address(), req.DelegatorAddr) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "Must use own delegator address")
			return
		}

		msg := stake.NewMsgBeginRedelegate(req.DelegatorAddr, req.ValidatorSrcAddr, req.ValidatorDstAddr, req.SharesAmount)
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

		cliCtx = cliCtx.WithGenerateOnly(req.BaseReq.GenerateOnly)
		cliCtx = cliCtx.WithSimulation(req.BaseReq.Simulate)

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w, cliCtx) {
			return
		}

		info, err := kb.Get(baseReq.Name)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		if !bytes.Equal(info.GetPubKey().Address(), req.DelegatorAddr) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "Must use own delegator address")
			return
		}

		msg := stake.NewMsgBeginUnbonding(req.DelegatorAddr, req.ValidatorAddr, req.SharesAmount)
		err = msg.ValidateBasic()
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.CompleteAndBroadcastTxREST(w, r, cliCtx, baseReq, []sdk.Msg{msg}, cdc)
	}
}
