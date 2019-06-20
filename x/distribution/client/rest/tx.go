package rest

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/common"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router, queryRoute string) {
	// Withdraw all delegator rewards
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/rewards",
		withdrawDelegatorRewardsHandlerFn(cliCtx, queryRoute),
	).Methods("POST")

	// Withdraw delegation rewards
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/rewards/{validatorAddr}",
		withdrawDelegationRewardsHandlerFn(cliCtx),
	).Methods("POST")

	// Replace the rewards withdrawal address
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/withdraw_address",
		setDelegatorWithdrawalAddrHandlerFn(cliCtx),
	).Methods("POST")

	// Withdraw validator rewards and commission
	r.HandleFunc(
		"/distribution/validators/{validatorAddr}/rewards",
		withdrawValidatorRewardsHandlerFn(cliCtx),
	).Methods("POST")

}

type (
	withdrawRewardsReq struct {
		BaseReq rest.BaseReq `json:"base_req"`
	}

	setWithdrawalAddrReq struct {
		BaseReq         rest.BaseReq   `json:"base_req"`
		WithdrawAddress sdk.AccAddress `json:"withdraw_address"`
	}
)

// Withdraw delegator rewards
func withdrawDelegatorRewardsHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req withdrawRewardsReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
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

		msgs, err := common.WithdrawAllDelegatorRewards(cliCtx, queryRoute, delAddr)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, msgs)
	}
}

// Withdraw delegation rewards
func withdrawDelegationRewardsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req withdrawRewardsReq

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
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
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// Replace the rewards withdrawal address
func setDelegatorWithdrawalAddrHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req setWithdrawalAddrReq

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
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
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// Withdraw validator rewards and commission
func withdrawValidatorRewardsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req withdrawRewardsReq

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
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
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, msgs)
	}
}

// Auxiliary

func checkDelegatorAddressVar(w http.ResponseWriter, r *http.Request) (sdk.AccAddress, bool) {
	addr, err := sdk.AccAddressFromBech32(mux.Vars(r)["delegatorAddr"])
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return nil, false
	}

	return addr, true
}

func checkValidatorAddressVar(w http.ResponseWriter, r *http.Request) (sdk.ValAddress, bool) {
	addr, err := sdk.ValAddressFromBech32(mux.Vars(r)["validatorAddr"])
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return nil, false
	}

	return addr, true
}
