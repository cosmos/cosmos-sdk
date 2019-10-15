package rest

import (
	"bytes"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/slashing/internal/types"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(
		"/slashing/validators/{validatorAddr}/unjail",
		unjailRequestHandlerFn(cliCtx),
	).Methods("POST")
}

// Unjail TX body
type UnjailReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

// unjailRequestHandlerFn - unjail a jailed validator
//
// @Summary Generate an unjail transaction
// @Description Generate an unjail transaction that is ready for signing
// @Tags slashing
// @Accept  json
// @Produce  json
// @Param validatorAddr path string true "The validator address"
// @Param body body rest.UnjailReq true "The unjail request payload"
// @Success 200 {object} rest.postUnjail
// @Failure 400 {object} rest.ErrorResponse "Invalid validator address or base_req"
// @Failure 401 {object} rest.ErrorResponse "Validator address incorrect"
// @Failure 500 {object} rest.ErrorResponse "Internal on server error"
// @Router /slashing/validators/{validatorAddr}/unjail [post]
func unjailRequestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		bech32validator := vars["validatorAddr"]

		var req UnjailReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		valAddr, err := sdk.ValAddressFromBech32(bech32validator)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		if !bytes.Equal(fromAddr, valAddr) {
			rest.WriteErrorResponse(w, http.StatusUnauthorized, "must use own validator address")
			return
		}

		msg := types.NewMsgUnjail(valAddr)
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
