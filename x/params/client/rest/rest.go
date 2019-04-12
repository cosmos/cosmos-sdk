package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	clientrest "github.com/cosmos/cosmos-sdk/client/rest"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
	"github.com/cosmos/cosmos-sdk/x/params"
)

func PropHandler(cliCtx context.CLIContext, cdc *codec.Codec) govrest.PropHandler {
	return govrest.PropHandler{
		Name:    "paramchange",
		Handler: postParamchangeProposalHandlerFn(cdc, cliCtx),
	}
}

type PostProposalReq struct {
	BaseReq        rest.BaseReq    `json:"base_req"`
	Title          string          `json:"title"`
	Description    string          `json:"description"`
	Changes        []params.Change `json:"changes"`
	Proposer       sdk.AccAddress  `json:"proposer"`
	InitialDeposit sdk.Coins       `json:"initial_deposit"`
}

func postParamchangeProposalHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PostProposalReq
		if !rest.ReadRESTReq(w, r, cdc, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		content := params.NewParamChangeProposal(req.Title, req.Description, req.Changes)

		msg := gov.NewMsgSubmitProposal(content, req.Proposer, req.InitialDeposit)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		clientrest.WriteGenerateStdTxResponse(w, cdc, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
