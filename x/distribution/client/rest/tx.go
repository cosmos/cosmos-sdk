package rest

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router,
	cdc *codec.Codec) {

	// Withdraw delegation rewards
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/rewards/{validatorAddr}",
		withdrawDelegationRewardsHandlerFn(cdc, cliCtx),
	).Methods("POST")

	// Withdraw validator rewards and commission
	r.HandleFunc(
		"/distribution/validators/{validatorAddr}/rewards",
		withdrawValidatorRewardsHandlerFn(cdc, cliCtx),
	).Methods("POST")

}

type withdrawRewardsReq struct {
	BaseReq utils.BaseReq `json:"base_req"`
}

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
		delAddr, abort := checkDelegatorAddressVar(w, r)
		if abort {
			return
		}

		valAddr, abort := checkValidatorAddressVar(w, r)
		if abort {
			return
		}

		// derive the from account address and name from the Keybase
		fromAddress, fromName, err := context.GetFromFields(req.BaseReq.From)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		cliCtx = cliCtx.WithFromName(fromName).WithFromAddress(fromAddress)

		msg := types.NewMsgWithdrawDelegatorReward(delAddr, valAddr)
		if err := msg.ValidateBasic(); err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		if req.BaseReq.GenerateOnly {
			utils.WriteGenerateStdTxResponse(w, cdc, req.BaseReq, []sdk.Msg{msg})
			return
		}

		utils.CompleteAndBroadcastTxREST(w, r, cliCtx, req.BaseReq, []sdk.Msg{msg}, cdc)
	}
}

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
		valAddr, abort := checkValidatorAddressVar(w, r)
		if abort {
			return
		}

		// derive the from account address and name from the Keybase
		fromAddress, fromName, err := context.GetFromFields(req.BaseReq.From)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		cliCtx = cliCtx.WithFromName(fromName).WithFromAddress(fromAddress)

		// build and validate MsgWithdrawValidatorCommission
		commissionMsg := types.NewMsgWithdrawValidatorCommission(valAddr)
		if err := commissionMsg.ValidateBasic(); err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// build and validate MsgWithdrawDelegatorReward
		delAddr := sdk.AccAddress(valAddr.Bytes())
		rewardMsg := types.NewMsgWithdrawDelegatorReward(delAddr, valAddr)
		if err := rewardMsg.ValidateBasic(); err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// prepare multi-message transaction
		msgs := []sdk.Msg{rewardMsg, commissionMsg}
		if req.BaseReq.GenerateOnly {
			utils.WriteGenerateStdTxResponse(w, cdc, req.BaseReq, msgs)
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
		return nil, true
	}
	return addr, false
}

func checkValidatorAddressVar(w http.ResponseWriter, r *http.Request) (sdk.ValAddress, bool) {
	addr, err := sdk.ValAddressFromBech32(mux.Vars(r)["validatorAddr"])
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return nil, true
	}
	return addr, false
}
