package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/gov"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// REST Variable names
// nolint
const (
	RestProposalID     = "proposal-id"
	RestDepositer      = "depositer"
	RestVoter          = "voter"
	RestProposalStatus = "status"
	RestNumLatest      = "latest"
	storeName          = "gov"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *wire.Codec) {
	r.HandleFunc("/gov/proposals", postProposalHandlerFn(cdc, cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}/deposits", RestProposalID), depositHandlerFn(cdc, cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}/votes", RestProposalID), voteHandlerFn(cdc, cliCtx)).Methods("POST")

	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}", RestProposalID), queryProposalHandlerFn(cdc)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}/deposits/{%s}", RestProposalID, RestDepositer), queryDepositHandlerFn(cdc)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}/votes/{%s}", RestProposalID, RestVoter), queryVoteHandlerFn(cdc)).Methods("GET")

	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}/votes", RestProposalID), queryVotesOnProposalHandlerFn(cdc)).Methods("GET")

	r.HandleFunc("/gov/proposals", queryProposalsWithParameterFn(cdc)).Methods("GET")
}

type postProposalReq struct {
	BaseReq        baseReq          `json:"base_req"`
	Title          string           `json:"title"`           //  Title of the proposal
	Description    string           `json:"description"`     //  Description of the proposal
	ProposalType   gov.ProposalKind `json:"proposal_type"`   //  Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}
	Proposer       sdk.AccAddress   `json:"proposer"`        //  Address of the proposer
	InitialDeposit sdk.Coins        `json:"initial_deposit"` // Coins to add to the proposal's deposit
}

type depositReq struct {
	BaseReq   baseReq        `json:"base_req"`
	Depositer sdk.AccAddress `json:"depositer"` // Address of the depositer
	Amount    sdk.Coins      `json:"amount"`    // Coins to add to the proposal's deposit
}

type voteReq struct {
	BaseReq baseReq        `json:"base_req"`
	Voter   sdk.AccAddress `json:"voter"`  //  address of the voter
	Option  gov.VoteOption `json:"option"` //  option from OptionSet chosen by the voter
}

func postProposalHandlerFn(cdc *wire.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req postProposalReq
		err := buildReq(w, r, cdc, &req)
		if err != nil {
			return
		}

		if !req.BaseReq.baseReqValidate(w) {
			return
		}

		// create the message
		msg := gov.NewMsgSubmitProposal(req.Title, req.Description, req.ProposalType, req.Proposer, req.InitialDeposit)
		err = msg.ValidateBasic()
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		signAndBuild(w, r, cliCtx, req.BaseReq, msg, cdc)
	}
}

func depositHandlerFn(cdc *wire.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars[RestProposalID]

		if len(strProposalID) == 0 {
			err := errors.New("proposalId required but not specified")
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		proposalID, ok := parseInt64OrReturnBadRequest(strProposalID, w)
		if !ok {
			return
		}

		var req depositReq
		err := buildReq(w, r, cdc, &req)
		if err != nil {
			return
		}
		if !req.BaseReq.baseReqValidate(w) {
			return
		}

		// create the message
		msg := gov.NewMsgDeposit(req.Depositer, proposalID, req.Amount)
		err = msg.ValidateBasic()
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		signAndBuild(w, r, cliCtx, req.BaseReq, msg, cdc)
	}
}

func voteHandlerFn(cdc *wire.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars[RestProposalID]

		if len(strProposalID) == 0 {
			err := errors.New("proposalId required but not specified")
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		proposalID, ok := parseInt64OrReturnBadRequest(strProposalID, w)
		if !ok {
			return
		}

		var req voteReq
		err := buildReq(w, r, cdc, &req)
		if err != nil {
			return
		}
		if !req.BaseReq.baseReqValidate(w) {
			return
		}

		// create the message
		msg := gov.NewMsgVote(req.Voter, proposalID, req.Option)
		err = msg.ValidateBasic()
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		signAndBuild(w, r, cliCtx, req.BaseReq, msg, cdc)
	}
}

func queryProposalHandlerFn(cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars[RestProposalID]

		if len(strProposalID) == 0 {
			err := errors.New("proposalId required but not specified")
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		proposalID, ok := parseInt64OrReturnBadRequest(strProposalID, w)
		if !ok {
			return
		}

		cliCtx := context.NewCLIContext().WithCodec(cdc)

		params := gov.QueryProposalParams{
			ProposalID: proposalID,
		}

		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, err := cliCtx.QueryWithData("custom/gov/proposal", bz)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.Write(res)
	}
}

func queryDepositHandlerFn(cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars[RestProposalID]
		bechDepositerAddr := vars[RestDepositer]

		if len(strProposalID) == 0 {
			err := errors.New("proposalId required but not specified")
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		proposalID, ok := parseInt64OrReturnBadRequest(strProposalID, w)
		if !ok {
			return
		}

		if len(bechDepositerAddr) == 0 {
			err := errors.New("depositer address required but not specified")
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		depositerAddr, err := sdk.AccAddressFromBech32(bechDepositerAddr)
		if err != nil {
			err := errors.Errorf("'%s' needs to be bech32 encoded", RestDepositer)
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx := context.NewCLIContext().WithCodec(cdc)

		params := gov.QueryDepositParams{
			ProposalID: proposalID,
			Depositer:  depositerAddr,
		}

		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, err := cliCtx.QueryWithData("custom/gov/deposit", bz)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		var deposit gov.Deposit
		cdc.UnmarshalJSON(res, &deposit)
		if deposit.Empty() {
			res, err := cliCtx.QueryWithData("custom/gov/proposal", cdc.MustMarshalBinary(gov.QueryProposalParams{params.ProposalID}))
			if err != nil || len(res) == 0 {
				err := errors.Errorf("proposalID [%d] does not exist", proposalID)
				utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
				return
			}
			err = errors.Errorf("depositer [%s] did not deposit on proposalID [%d]", bechDepositerAddr, proposalID)
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		w.Write(res)
	}
}

func queryVoteHandlerFn(cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars[RestProposalID]
		bechVoterAddr := vars[RestVoter]

		if len(strProposalID) == 0 {
			err := errors.New("proposalId required but not specified")
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		proposalID, ok := parseInt64OrReturnBadRequest(strProposalID, w)
		if !ok {
			return
		}

		if len(bechVoterAddr) == 0 {
			err := errors.New("voter address required but not specified")
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		voterAddr, err := sdk.AccAddressFromBech32(bechVoterAddr)
		if err != nil {
			err := errors.Errorf("'%s' needs to be bech32 encoded", RestVoter)
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx := context.NewCLIContext().WithCodec(cdc)

		params := gov.QueryVoteParams{
			Voter:      voterAddr,
			ProposalID: proposalID,
		}
		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, err := cliCtx.QueryWithData("custom/gov/vote", bz)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		var vote gov.Vote
		cdc.UnmarshalJSON(res, &vote)
		if vote.Empty() {
			bz, err := cdc.MarshalJSON(gov.QueryProposalParams{params.ProposalID})
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			res, err := cliCtx.QueryWithData("custom/gov/proposal", bz)
			if err != nil || len(res) == 0 {
				err := errors.Errorf("proposalID [%d] does not exist", proposalID)
				utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
				return
			}
			err = errors.Errorf("voter [%s] did not deposit on proposalID [%d]", bechVoterAddr, proposalID)
			utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		w.Write(res)
	}
}

// nolint: gocyclo
// todo: Split this functionality into helper functions to remove the above
func queryVotesOnProposalHandlerFn(cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars[RestProposalID]

		if len(strProposalID) == 0 {
			err := errors.New("proposalId required but not specified")
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		proposalID, ok := parseInt64OrReturnBadRequest(strProposalID, w)
		if !ok {
			return
		}

		cliCtx := context.NewCLIContext().WithCodec(cdc)

		params := gov.QueryVotesParams{
			ProposalID: proposalID,
		}
		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, err := cliCtx.QueryWithData("custom/gov/votes", bz)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.Write(res)
	}
}

// nolint: gocyclo
// todo: Split this functionality into helper functions to remove the above
func queryProposalsWithParameterFn(cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bechVoterAddr := r.URL.Query().Get(RestVoter)
		bechDepositerAddr := r.URL.Query().Get(RestDepositer)
		strProposalStatus := r.URL.Query().Get(RestProposalStatus)
		strNumLatest := r.URL.Query().Get(RestNumLatest)

		params := gov.QueryProposalsParams{}

		if len(bechVoterAddr) != 0 {
			voterAddr, err := sdk.AccAddressFromBech32(bechVoterAddr)
			if err != nil {
				err := errors.Errorf("'%s' needs to be bech32 encoded", RestVoter)
				utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			params.Voter = voterAddr
		}

		if len(bechDepositerAddr) != 0 {
			depositerAddr, err := sdk.AccAddressFromBech32(bechDepositerAddr)
			if err != nil {
				err := errors.Errorf("'%s' needs to be bech32 encoded", RestDepositer)
				utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			params.Depositer = depositerAddr
		}

		if len(strProposalStatus) != 0 {
			proposalStatus, err := gov.ProposalStatusFromString(strProposalStatus)
			if err != nil {
				err := errors.Errorf("'%s' is not a valid Proposal Status", strProposalStatus)
				utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			params.ProposalStatus = proposalStatus
		}
		if len(strNumLatest) != 0 {
			numLatest, ok := parseInt64OrReturnBadRequest(strNumLatest, w)
			if !ok {
				return
			}
			params.NumLatestProposals = numLatest
		}

		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx := context.NewCLIContext().WithCodec(cdc)

		res, err := cliCtx.QueryWithData("custom/gov/proposals", bz)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.Write(res)
	}
}

// nolint: gocyclo
// todo: Split this functionality into helper functions to remove the above
func queryTallyOnProposalHandlerFn(cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		strProposalID := vars[RestProposalID]

		if len(strProposalID) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			err := errors.New("proposalId required but not specified")
			w.Write([]byte(err.Error()))

			return
		}

		proposalID, ok := parseInt64OrReturnBadRequest(strProposalID, w)
		if !ok {
			return
		}

		cliCtx := context.NewCLIContext().WithCodec(cdc)

		params := gov.QueryTallyParams{
			ProposalID: proposalID,
		}
		bz, err := cdc.MarshalJSON(params)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		res, err := cliCtx.QueryWithData("custom/gov/tally", bz)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(res)
	}
}
