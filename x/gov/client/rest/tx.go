package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	auth "github.com/cosmos/cosmos-sdk/x/auth"
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

// postProposal is used to generate documentation for postProposalHandlerFn
type postProposal struct {
	Msgs       []types.MsgSubmitProposal `json:"msg" yaml:"msg"`
	Fee        auth.StdFee               `json:"fee" yaml:"fee"`
	Signatures []auth.StdSignature       `json:"signatures" yaml:"signatures"`
	Memo       string                    `json:"memo" yaml:"memo"`
}

// postProposalHandlerFn implements a proposal generation handler that
// is responsible for constructing a properly formatted proposal for signing.
//
// @Summary Generate an unsigned proposal transaction
// @Description Generate a proposal transaction that is ready for signing
// @Tags governance
// @Accept  json
// @Produce  json
// @Param body body rest.PostProposalReq true "The data required to construct a proposal message, the proposal_type can be (text | parameter_change)"
// @Success 200 {object} rest.postProposal
// @Failure 400 {object} rest.ErrorResponse "Returned if the request is invalid"
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

// postDeposit is used to generate documentation for postDepositHandlerFn
type postDeposit struct {
	Msgs       []types.MsgDeposit  `json:"msg" yaml:"msg"`
	Fee        auth.StdFee         `json:"fee" yaml:"fee"`
	Signatures []auth.StdSignature `json:"signatures" yaml:"signatures"`
	Memo       string              `json:"memo" yaml:"memo"`
}

// depositHandlerFn implements a deposit generation handler that
// is responsible for constructing a properly formatted deposit for signing.
//
// @Summary Generate an unsigned deposit transaction
// @Description Generate a deposit transaction that is ready for signing.
// @Tags governance
// @Accept  json
// @Produce  json
// @Param proposalID path int true "The ID of the governance proposal to deposit to"
// @Param body body rest.DepositReq true "The data required to construct a deposit message"
// @Success 200 {object} rest.postDeposit
// @Failure 400 {object} rest.ErrorResponse "Returned if the request is invalid"
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

// postVote is used to generate documentation for postVoteHandlerFn
type postVote struct {
	Msgs       []types.MsgVote     `json:"msg" yaml:"msg"`
	Fee        auth.StdFee         `json:"fee" yaml:"fee"`
	Signatures []auth.StdSignature `json:"signatures" yaml:"signatures"`
	Memo       string              `json:"memo" yaml:"memo"`
}

// voteHandlerFn implements a governance vote generation handler that
// is responsible for constructing a properly formatted governance vote for signing.
//
// @Summary Generate an unsigned vote transaction
// @Description Generate a vote transaction that is ready for signing.
// @Tags governance
// @Accept  json
// @Produce  json
// @Param proposalID path int true "The ID of the governance proposal to vote for"
// @Param body body rest.VoteReq true "The data required to construct a vote message"
// @Success 200 {object} rest.postVote
// @Failure 400 {object} rest.ErrorResponse "Returned if the request is invalid"
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
