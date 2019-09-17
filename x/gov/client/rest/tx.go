package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	gcutils "github.com/cosmos/cosmos-sdk/x/gov/client/utils"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router, phs []ProposalRESTHandler) {
	propSubRtr := r.PathPrefix("/gov/proposals").Subrouter()
	for _, ph := range phs {
		propSubRtr.HandleFunc(fmt.Sprintf("/%s", ph.SubRoute), ph.Handler).Methods("POST")
	}

	r.HandleFunc("/gov/proposals", postProposalHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}/deposits", RestProposalID), depositHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}/votes", RestProposalID), voteHandlerFn(cliCtx)).Methods("POST")
}

// postProposalHandlerFn implements a proposal generation handler that
// is responsible for constructing a properly formatted proposal for signing.
//
// @Summary Generate a proposal transaction.
// @Description Generate a proposal transaction that is ready for signing.
// @Tags transactions
// @Accept  json
// @Produce  json
// @Param tx body PostProposalReq true "The data required to construct a proposal. Valid value of proposal_type can be text|parameter_change"
// @Success 200 {object} types.StdTx
// @Failure 400 {object} rest.ErrorResponse "Returned if the request is invalid."
// @Router /gov/proposals [post]
func postProposalHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PostProposalReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		proposalType := gcutils.NormalizeProposalType(req.ProposalType)
		content := types.ContentFromProposalType(req.Title, req.Description, proposalType)

		msg := types.NewMsgSubmitProposal(content, req.InitialDeposit, req.Proposer)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// depositHandlerFn implements a deposit generation handler that
// is responsible for constructing a properly formatted deposit for signing.
//
// @Summary Generate a deposit transaction.
// @Description Generate a deposit transaction that is ready for signing.
// @Tags transactions
// @Accept  json
// @Produce  json
// @Param proposalID path int true "The ID of the governance proposal to deposit to."
// @Param tx body DepositReq true "The data required to construct a deposit."
// @Success 200 {object} types.StdTx
// @Failure 400 {object} rest.ErrorResponse "Returned if the request is invalid."
// @Router /gov/proposals/{proposalID}/deposits [post]
func depositHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars[RestProposalID]

		if len(strProposalID) == 0 {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "proposalId required but not specified")
			return
		}

		proposalID, ok := rest.ParseUint64OrReturnBadRequest(w, strProposalID)
		if !ok {
			return
		}

		var req DepositReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		// create the message
		msg := types.NewMsgDeposit(req.Depositor, proposalID, req.Amount)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// voteHandlerFn implements a governance vote generation handler that
// is responsible for constructing a properly formatted governance vote for signing.
//
// @Summary Generate a vote transaction.
// @Description Generate a vote transaction that is ready for signing.
// @Tags transactions
// @Accept  json
// @Produce  json
// @Param proposalID path int true "The ID of the governance proposal to vote for."
// @Param tx body VoteReq true "The data required to construct a vote."
// @Success 200 {object} types.StdTx
// @Failure 400 {object} rest.ErrorResponse "Returned if the request is invalid."
// @Router /gov/proposals/{proposalID}/votes [post]
func voteHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars[RestProposalID]

		if len(strProposalID) == 0 {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "proposalId required but not specified")
			return
		}

		proposalID, ok := rest.ParseUint64OrReturnBadRequest(w, strProposalID)
		if !ok {
			return
		}

		var req VoteReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		voteOption, err := types.VoteOptionFromString(gcutils.NormalizeVoteOption(req.Option))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// create the message
		msg := types.NewMsgVote(req.Voter, proposalID, voteOption)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
