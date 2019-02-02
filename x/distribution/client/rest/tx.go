package rest

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/common"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router,
	cdc *codec.Codec, queryRoute string) {

	// Withdraw all delegator rewards
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/rewards",
		withdrawDelegatorRewardsHandlerFn(cdc, cliCtx, queryRoute),
	).Methods("POST")

	// Withdraw delegation rewards
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/rewards/{validatorAddr}",
		withdrawDelegationRewardsHandlerFn(cdc, cliCtx),
	).Methods("POST")

	// Replace the rewards withdrawal address
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/withdraw_address",
		setDelegatorWithdrawalAddrHandlerFn(cdc, cliCtx),
	).Methods("POST")

	// Withdraw validator rewards and commission
	r.HandleFunc(
		"/distribution/validators/{validatorAddr}/rewards",
		withdrawValidatorRewardsHandlerFn(cdc, cliCtx),
	).Methods("POST")

}

type (
	withdrawRewardsReq struct {
		BaseReq utils.BaseReq `json:"base_req"`
	}

	setWithdrawalAddrReq struct {
		BaseReq         utils.BaseReq  `json:"base_req"`
		WithdrawAddress sdk.AccAddress `json:"withdraw_address"`
	}
)

// Withdraw delegator rewards
func withdrawDelegatorRewardsHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext,
	queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req withdrawRewardsReq
		if err := utils.ReadRESTReq(w, r, cdc, &req); err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		// read and validate URL's variables
		delAddr, ok := checkDelegatorAddressVar(w, r)
		if !ok {
			return
		}

		msgs, err := common.WithdrawAllDelegatorRewards(cliCtx, cdc, queryRoute, delAddr)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		if req.BaseReq.GenerateOnly {
			utils.WriteGenerateStdTxResponse(w, cdc, cliCtx, req.BaseReq, msgs)
			return
		}

		utils.CompleteAndBroadcastTxREST(w, r, cliCtx, req.BaseReq, msgs, cdc)
	}
}

// Withdraw delegation rewards
func withdrawDelegationRewardsHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req withdrawRewardsReq

		if err := utils.ReadRESTReq(w, r, cdc, &req); err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		// read and validate URL's variables
		delAddr, ok := checkDelegatorAddressVar(w, r)
		if !ok {
			return
		}

		valAddr, ok := checkValidatorAddressVar(w, r)
		if !ok {
			return
		}

		msg := types.NewMsgWithdrawDelegatorReward(delAddr, valAddr)
		if err := msg.ValidateBasic(); err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		if req.BaseReq.GenerateOnly {
			utils.WriteGenerateStdTxResponse(w, cdc, cliCtx, req.BaseReq, []sdk.Msg{msg})
			return
		}

		utils.CompleteAndBroadcastTxREST(w, r, cliCtx, req.BaseReq, []sdk.Msg{msg}, cdc)
	}
}

// Replace the rewards withdrawal address
func setDelegatorWithdrawalAddrHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req setWithdrawalAddrReq

		if err := utils.ReadRESTReq(w, r, cdc, &req); err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		// read and validate URL's variables
		delAddr, ok := checkDelegatorAddressVar(w, r)
		if !ok {
			return
		}

		msg := types.NewMsgSetWithdrawAddress(delAddr, req.WithdrawAddress)
		if err := msg.ValidateBasic(); err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		if req.BaseReq.GenerateOnly {
			utils.WriteGenerateStdTxResponse(w, cdc, cliCtx, req.BaseReq, []sdk.Msg{msg})
			return
		}

		utils.CompleteAndBroadcastTxREST(w, r, cliCtx, req.BaseReq, []sdk.Msg{msg}, cdc)
	}
}

// Withdraw validator rewards and commission
func withdrawValidatorRewardsHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req withdrawRewardsReq

		if err := utils.ReadRESTReq(w, r, cdc, &req); err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		// read and validate URL's variable
		valAddr, ok := checkValidatorAddressVar(w, r)
		if !ok {
			return
		}

		// prepare multi-message transaction
		msgs, err := common.WithdrawValidatorRewardsAndCommission(valAddr)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		if req.BaseReq.GenerateOnly {
			utils.WriteGenerateStdTxResponse(w, cdc, cliCtx, req.BaseReq, msgs)
			return
		}

		utils.CompleteAndBroadcastTxREST(w, r, cliCtx, req.BaseReq, msgs, cdc)
	}
}

// Auxiliary

func checkDelegatorAddressVar(w http.ResponseWriter, r *http.Request) (sdk.AccAddress, bool) {
	addr, err := sdk.AccAddressFromBech32(mux.Vars(r)["delegatorAddr"])
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return nil, false
	}
	return addr, true
}

func checkValidatorAddressVar(w http.ResponseWriter, r *http.Request) (sdk.ValAddress, bool) {
	addr, err := sdk.ValAddressFromBech32(mux.Vars(r)["validatorAddr"])
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return nil, false
	}
	return addr, true
}
