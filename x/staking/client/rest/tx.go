package rest

import (
	"bytes"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	clientrest "github.com/cosmos/cosmos-sdk/client/rest"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec, kb keys.Keybase) {
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/delegations",
		postDelegationsHandlerFn(cdc, kb, cliCtx),
	).Methods("POST")
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/unbonding_delegations",
		postUnbondingDelegationsHandlerFn(cdc, kb, cliCtx),
	).Methods("POST")
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/redelegations",
		postRedelegationsHandlerFn(cdc, kb, cliCtx),
	).Methods("POST")
}

type (
	msgDelegationsInput struct {
		BaseReq       rest.BaseReq   `json:"base_req"`
		DelegatorAddr sdk.AccAddress `json:"delegator_addr"` // in bech32
		ValidatorAddr sdk.ValAddress `json:"validator_addr"` // in bech32
		Delegation    sdk.Coin       `json:"delegation"`
	}

	msgBeginRedelegateInput struct {
		BaseReq          rest.BaseReq   `json:"base_req"`
		DelegatorAddr    sdk.AccAddress `json:"delegator_addr"`     // in bech32
		ValidatorSrcAddr sdk.ValAddress `json:"validator_src_addr"` // in bech32
		ValidatorDstAddr sdk.ValAddress `json:"validator_dst_addr"` // in bech32
		SharesAmount     sdk.Dec        `json:"shares"`
	}

	msgUndelegateInput struct {
		BaseReq       rest.BaseReq   `json:"base_req"`
		DelegatorAddr sdk.AccAddress `json:"delegator_addr"` // in bech32
		ValidatorAddr sdk.ValAddress `json:"validator_addr"` // in bech32
		SharesAmount  sdk.Dec        `json:"shares"`
	}
)

func postDelegationsHandlerFn(cdc *codec.Codec, kb keys.Keybase, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req msgDelegationsInput

		if !rest.ReadRESTReq(w, r, cdc, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		msg := staking.NewMsgDelegate(req.DelegatorAddr, req.ValidatorAddr, req.Delegation)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		if req.BaseReq.GenerateOnly {
			clientrest.WriteGenerateStdTxResponse(w, cdc, cliCtx, req.BaseReq, []sdk.Msg{msg})
			return
		}

		// derive the from account address and name from the Keybase
		fromAddress, fromName, err := context.GetFromFields(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx = cliCtx.WithFromName(fromName).WithFromAddress(fromAddress)

		if !bytes.Equal(cliCtx.GetFromAddress(), req.DelegatorAddr) {
			rest.WriteErrorResponse(w, http.StatusUnauthorized, "must use own delegator address")
			return
		}

		clientrest.CompleteAndBroadcastTxREST(w, r, cliCtx, req.BaseReq, []sdk.Msg{msg}, cdc)
	}
}

func postRedelegationsHandlerFn(cdc *codec.Codec, kb keys.Keybase, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req msgBeginRedelegateInput

		if !rest.ReadRESTReq(w, r, cdc, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		msg := staking.NewMsgBeginRedelegate(req.DelegatorAddr, req.ValidatorSrcAddr, req.ValidatorDstAddr, req.SharesAmount)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		if req.BaseReq.GenerateOnly {
			clientrest.WriteGenerateStdTxResponse(w, cdc, cliCtx, req.BaseReq, []sdk.Msg{msg})
			return
		}

		// derive the from account address and name from the Keybase
		fromAddress, fromName, err := context.GetFromFields(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx = cliCtx.WithFromName(fromName).WithFromAddress(fromAddress)

		if !bytes.Equal(cliCtx.GetFromAddress(), req.DelegatorAddr) {
			rest.WriteErrorResponse(w, http.StatusUnauthorized, "must use own delegator address")
			return
		}

		clientrest.CompleteAndBroadcastTxREST(w, r, cliCtx, req.BaseReq, []sdk.Msg{msg}, cdc)
	}
}

func postUnbondingDelegationsHandlerFn(cdc *codec.Codec, kb keys.Keybase, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req msgUndelegateInput

		if !rest.ReadRESTReq(w, r, cdc, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		msg := staking.NewMsgUndelegate(req.DelegatorAddr, req.ValidatorAddr, req.SharesAmount)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		if req.BaseReq.GenerateOnly {
			clientrest.WriteGenerateStdTxResponse(w, cdc, cliCtx, req.BaseReq, []sdk.Msg{msg})
			return
		}

		// derive the from account address and name from the Keybase
		fromAddress, fromName, err := context.GetFromFields(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx = cliCtx.WithFromName(fromName).WithFromAddress(fromAddress)

		if !bytes.Equal(cliCtx.GetFromAddress(), req.DelegatorAddr) {
			rest.WriteErrorResponse(w, http.StatusUnauthorized, "must use own delegator address")
			return
		}

		clientrest.CompleteAndBroadcastTxREST(w, r, cliCtx, req.BaseReq, []sdk.Msg{msg}, cdc)
	}
}
