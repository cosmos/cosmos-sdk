package rest

import (
	"errors"
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

// REST Variable names
// nolint
const (
	RestParamsType     = "type"
	RestProposalID     = "proposal-id"
	RestDepositor      = "depositor"
	RestVoter          = "voter"
	RestProposalStatus = "status"
	RestNumLimit       = "limit"
)

// ProposalRESTHandler defines a REST handler implemented in another module. The
// sub-route is mounted on the governance REST handler.
type ProposalRESTHandler struct {
	SubRoute string
	Handler  func(http.ResponseWriter, *http.Request)
}

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, phs []ProposalRESTHandler) {
	propSubRtr := r.PathPrefix("/gov/proposals").Subrouter()
	for _, ph := range phs {
		propSubRtr.HandleFunc(fmt.Sprintf("/%s", ph.SubRoute), ph.Handler).Methods("POST")
	}

	r.HandleFunc("/gov/proposals", postProposalHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}/deposits", RestProposalID), depositHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}/votes", RestProposalID), voteHandlerFn(cliCtx)).Methods("POST")

	r.HandleFunc(
		fmt.Sprintf("/gov/parameters/{%s}", RestParamsType),
		queryParamsHandlerFn(cliCtx),
	).Methods("GET")

	r.HandleFunc("/gov/proposals", queryProposalsWithParameterFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}", RestProposalID), queryProposalHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(
		fmt.Sprintf("/gov/proposals/{%s}/proposer", RestProposalID),
		queryProposerHandlerFn(cliCtx),
	).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}/deposits", RestProposalID), queryDepositsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}/deposits/{%s}", RestProposalID, RestDepositor), queryDepositHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}/tally", RestProposalID), queryTallyOnProposalHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}/votes", RestProposalID), queryVotesOnProposalHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}/votes/{%s}", RestProposalID, RestVoter), queryVoteHandlerFn(cliCtx)).Methods("GET")
}

// PostProposalReq defines the properties of a proposal request's body.
type PostProposalReq struct {
	BaseReq        rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Title          string         `json:"title" yaml:"title"`                     // Title of the proposal
	Description    string         `json:"description" yaml:"description"`         // Description of the proposal
	ProposalType   string         `json:"proposal_type" yaml:"proposal_type"`     // Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}
	Proposer       sdk.AccAddress `json:"proposer" yaml:"proposer"`               // Address of the proposer
	InitialDeposit sdk.Coins      `json:"initial_deposit" yaml:"initial_deposit"` // Coins to add to the proposal's deposit
}

// DepositReq defines the properties of a deposit request's body.
type DepositReq struct {
	BaseReq   rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Depositor sdk.AccAddress `json:"depositor" yaml:"depositor"` // Address of the depositor
	Amount    sdk.Coins      `json:"amount" yaml:"amount"`       // Coins to add to the proposal's deposit
}

// VoteReq defines the properties of a vote request's body.
type VoteReq struct {
	BaseReq rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Voter   sdk.AccAddress `json:"voter" yaml:"voter"`   // address of the voter
	Option  string         `json:"option" yaml:"option"` // option from OptionSet chosen by the voter
}

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

func depositHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars[RestProposalID]

		if len(strProposalID) == 0 {
			err := errors.New("proposalId required but not specified")
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
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

func voteHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars[RestProposalID]

		if len(strProposalID) == 0 {
			err := errors.New("proposalId required but not specified")
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
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

func queryParamsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		paramType := vars[RestParamsType]

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/gov/%s/%s", types.QueryParams, paramType), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryProposalHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars[RestProposalID]

		if len(strProposalID) == 0 {
			err := errors.New("proposalId required but not specified")
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		proposalID, ok := rest.ParseUint64OrReturnBadRequest(w, strProposalID)
		if !ok {
			return
		}

		cliCtx, ok = rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		params := types.NewQueryProposalParams(proposalID)

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, height, err := cliCtx.QueryWithData("custom/gov/proposal", bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryDepositsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars[RestProposalID]

		proposalID, ok := rest.ParseUint64OrReturnBadRequest(w, strProposalID)
		if !ok {
			return
		}

		cliCtx, ok = rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		params := types.NewQueryProposalParams(proposalID)

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData("custom/gov/proposal", bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		var proposal types.Proposal
		if err := cliCtx.Codec.UnmarshalJSON(res, &proposal); err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// For inactive proposals we must query the txs directly to get the deposits
		// as they're no longer in state.
		propStatus := proposal.Status
		if !(propStatus == types.StatusVotingPeriod || propStatus == types.StatusDepositPeriod) {
			res, err = gcutils.QueryDepositsByTxQuery(cliCtx, params)
		} else {
			res, _, err = cliCtx.QueryWithData("custom/gov/deposits", bz)
		}

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryProposerHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars[RestProposalID]

		proposalID, ok := rest.ParseUint64OrReturnBadRequest(w, strProposalID)
		if !ok {
			return
		}

		cliCtx, ok = rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, err := gcutils.QueryProposerByTxQuery(cliCtx, proposalID)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryDepositHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars[RestProposalID]
		bechDepositorAddr := vars[RestDepositor]

		if len(strProposalID) == 0 {
			err := errors.New("proposalId required but not specified")
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		proposalID, ok := rest.ParseUint64OrReturnBadRequest(w, strProposalID)
		if !ok {
			return
		}

		if len(bechDepositorAddr) == 0 {
			err := errors.New("depositor address required but not specified")
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		depositorAddr, err := sdk.AccAddressFromBech32(bechDepositorAddr)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx, ok = rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		params := types.NewQueryDepositParams(proposalID, depositorAddr)

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData("custom/gov/deposit", bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		var deposit types.Deposit
		if err := cliCtx.Codec.UnmarshalJSON(res, &deposit); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// For an empty deposit, either the proposal does not exist or is inactive in
		// which case the deposit would be removed from state and should be queried
		// for directly via a txs query.
		if deposit.Empty() {
			bz, err := cliCtx.Codec.MarshalJSON(types.NewQueryProposalParams(proposalID))
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}

			res, _, err = cliCtx.QueryWithData("custom/gov/proposal", bz)
			if err != nil || len(res) == 0 {
				err := fmt.Errorf("proposalID %d does not exist", proposalID)
				rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
				return
			}

			res, err = gcutils.QueryDepositByTxQuery(cliCtx, params)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryVoteHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars[RestProposalID]
		bechVoterAddr := vars[RestVoter]

		if len(strProposalID) == 0 {
			err := errors.New("proposalId required but not specified")
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		proposalID, ok := rest.ParseUint64OrReturnBadRequest(w, strProposalID)
		if !ok {
			return
		}

		if len(bechVoterAddr) == 0 {
			err := errors.New("voter address required but not specified")
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		voterAddr, err := sdk.AccAddressFromBech32(bechVoterAddr)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx, ok = rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		params := types.NewQueryVoteParams(proposalID, voterAddr)

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData("custom/gov/vote", bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		var vote types.Vote
		if err := cliCtx.Codec.UnmarshalJSON(res, &vote); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// For an empty vote, either the proposal does not exist or is inactive in
		// which case the vote would be removed from state and should be queried for
		// directly via a txs query.
		if vote.Empty() {
			bz, err := cliCtx.Codec.MarshalJSON(types.NewQueryProposalParams(proposalID))
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}

			res, _, err = cliCtx.QueryWithData("custom/gov/proposal", bz)
			if err != nil || len(res) == 0 {
				err := fmt.Errorf("proposalID %d does not exist", proposalID)
				rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
				return
			}

			res, err = gcutils.QueryVoteByTxQuery(cliCtx, params)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// todo: Split this functionality into helper functions to remove the above
func queryVotesOnProposalHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars[RestProposalID]

		if len(strProposalID) == 0 {
			err := errors.New("proposalId required but not specified")
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		proposalID, ok := rest.ParseUint64OrReturnBadRequest(w, strProposalID)
		if !ok {
			return
		}

		cliCtx, ok = rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		params := types.NewQueryProposalParams(proposalID)

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData("custom/gov/proposal", bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		var proposal types.Proposal
		if err := cliCtx.Codec.UnmarshalJSON(res, &proposal); err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// For inactive proposals we must query the txs directly to get the votes
		// as they're no longer in state.
		propStatus := proposal.Status
		if !(propStatus == types.StatusVotingPeriod || propStatus == types.StatusDepositPeriod) {
			res, err = gcutils.QueryVotesByTxQuery(cliCtx, params)
		} else {
			res, _, err = cliCtx.QueryWithData("custom/gov/votes", bz)
		}

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// todo: Split this functionality into helper functions to remove the above
func queryProposalsWithParameterFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bechVoterAddr := r.URL.Query().Get(RestVoter)
		bechDepositorAddr := r.URL.Query().Get(RestDepositor)
		strProposalStatus := r.URL.Query().Get(RestProposalStatus)
		strNumLimit := r.URL.Query().Get(RestNumLimit)

		params := types.QueryProposalsParams{}

		if len(bechVoterAddr) != 0 {
			voterAddr, err := sdk.AccAddressFromBech32(bechVoterAddr)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			params.Voter = voterAddr
		}

		if len(bechDepositorAddr) != 0 {
			depositorAddr, err := sdk.AccAddressFromBech32(bechDepositorAddr)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			params.Depositor = depositorAddr
		}

		if len(strProposalStatus) != 0 {
			proposalStatus, err := types.ProposalStatusFromString(gcutils.NormalizeProposalStatus(strProposalStatus))
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			params.ProposalStatus = proposalStatus
		}
		if len(strNumLimit) != 0 {
			numLimit, ok := rest.ParseUint64OrReturnBadRequest(w, strNumLimit)
			if !ok {
				return
			}
			params.Limit = numLimit
		}

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, height, err := cliCtx.QueryWithData("custom/gov/proposals", bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// todo: Split this functionality into helper functions to remove the above
func queryTallyOnProposalHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars[RestProposalID]

		if len(strProposalID) == 0 {
			err := errors.New("proposalId required but not specified")
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		proposalID, ok := rest.ParseUint64OrReturnBadRequest(w, strProposalID)
		if !ok {
			return
		}

		cliCtx, ok = rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		params := types.NewQueryProposalParams(proposalID)

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, height, err := cliCtx.QueryWithData("custom/gov/tally", bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
