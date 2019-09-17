package rest

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	gcutils "github.com/cosmos/cosmos-sdk/x/gov/client/utils"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/gov/parameters/{%s}", RestParamsType), queryParamsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/gov/proposals", queryProposalsWithParameterFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}", RestProposalID), queryProposalHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}/proposer", RestProposalID), queryProposerHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}/deposits", RestProposalID), queryDepositsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}/deposits/{%s}", RestProposalID, RestDepositor), queryDepositHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}/tally", RestProposalID), queryTallyOnProposalHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}/votes", RestProposalID), queryVotesOnProposalHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/gov/proposals/{%s}/votes/{%s}", RestProposalID, RestVoter), queryVoteHandlerFn(cliCtx)).Methods("GET")
}

// queryParamsHanderFun implements a governance params querying route
// for returning data on either deposit|tallying|voting parameters.
//
// @Summary Query governance parameters
// @Tags queries
// @Description Query either deposit|tallying|voting parameters of the gov module.
// @Produce json
// @Param type path string true "Type of param, deposit|tallying|voting"
// @Param height query string false "block height to execute query, defaults to chain tip."
// @Success 200 {object} types.ResponseWithHeight
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have a valid height."
// @Failure 404 {object} rest.ErrorResponse "Returned if the type of parameter isn't found in store."
// @Router /gov/parameters/{type} [get]
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

// queryProposalHanderFun implements a governance proposal querying route
// for returning data on an individual governance proposal.
//
// @Summary Query a governance proposal
// @Tags queries
// @Description Query an individual governance proposal
// @Produce json
// @Param proposalID path int true "The ID of the governance proposal."
// @Param height query string false "block height to execute query, defaults to chain tip."
// @Success 200 {object} types.ResponseWithHeight
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have a valid proposalID or height."
// @Failure 500 {object} rest.ErrorResponse "Returned if the proposalID isn't in the store."
// @Router /gov/proposals/{proposalID} [get]
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

// queryDepositsHandlerFn implements a governance deposits querying route
// for returning data on an individual governance proposal's deposits.
//
// @Summary Query a governance proposal's deposits.
// @Tags queries
// @Description Query an individual governance proposal's deposits.
// @Description NOTE: In order to query deposits for passed proposals, the transaction
// @Description record must be available otherwise the query will fail. This requires a
// @Description node that is not pruning transaction history.
// @Produce json
// @Param proposalID path int true "The ID of the governance proposal."
// @Param height query string false "block height to execute query, defaults to chain tip."
// @Success 200 {object} types.ResponseWithHeight
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have a valid proposalID or height."
// @Failure 500 {object} rest.ErrorResponse "Returned if the proposalID isn't in the store."
// @Router /gov/proposals/{proposalID}/deposits [get]
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

// queryProposerHandlerFn implements a governance proposal proposer querying route
// for returning data on individual governance proposal proposer.
//
// @Summary Query a governance proposal's proposer.
// @Tags queries
// @Description Query an individual governance proposal's proposer.
// @Produce json
// @Param proposalID path int true "The ID of the governance proposal."
// @Param height query string false "block height to execute query, defaults to chain tip."
// @Success 200 {object} types.ResponseWithHeight
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have a valid proposalID or height."
// @Failure 500 {object} rest.ErrorResponse "Returned if the proposalID isn't in the store."
// @Router /gov/proposals/{proposalID}/proposer [get]
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

// queryDepositHandlerFn implements a governance proposal deposit querying route
// for returning data on a governance proposal's individual deposits.
//
// @Summary Query a governance proposal's individual deposit.
// @Tags queries
// @Description Query an individual governance proposal's deposits.
// @Description NOTE: In order to query a deposit for a passed proposal, the transaction
// @Description record must be available otherwise the query will fail. This requires a
// @Description node that is not pruning transaction history.
// @Produce json
// @Param proposalID path int true "The ID of the governance proposal."
// @Param depositor path string true "The address of the depositor."
// @Param height query string false "block height to execute query, defaults to chain tip."
// @Success 200 {object} types.ResponseWithHeight
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have a valid proposalID or depositor."
// @Failure 404 {object} rest.ErrorResponse "Returned if the proposalID is not found."
// @Failure 500 {object} rest.ErrorResponse "Returned if the proposalID or depositor isn't in the store."
// @Router /gov/proposals/{proposalID}/deposits/{depositor} [get]
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

// queryVoteHandlerFn implements a governance proposal vote querying route
// for returning data on an individual governance proposal vote.
//
// @Summary Query a governance proposal's individual vote.
// @Tags queries
// @Description Query an individual governance proposal's vote.
// @Description NOTE: In order to query votes for passed proposals, the transaction
// @Description record must be available otherwise the query will fail. This requires a
// @Description node that is not pruning transaction history.
// @Produce json
// @Param proposalID path int true "The ID of the governance proposal."
// @Param voter path string true "The address of the voter."
// @Param height query string false "block height to execute query, defaults to chain tip."
// @Success 200 {object} types.ResponseWithHeight
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have a valid proposalID or v voter."
// @Failure 404 {object} rest.ErrorResponse "Returned if the proposalID is not found."
// @Failure 500 {object} rest.ErrorResponse "Returned if the proposalID or voter isn't in the store."
// @Router /gov/proposals/{proposalID}/votes/{voter} [get]
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
//
// queryVotesOnProposalHandlerFn implements a governance proposal votes querying route
// for returning data on an individual governance proposal's votes.
//
// @Summary Query a governance proposal's votes.
// @Tags queries
// @Description Query an individual governance proposal's votes.
// @Description NOTE: In order to query deposits for passed proposals, the transaction
// @Description record must be available otherwise the query will fail. This requires a
// @Description node that is not pruning transaction history.
// @Produce json
// @Param proposalID path int true "The ID of the governance proposal."
// @Param height query string false "block height to execute query, defaults to chain tip."
// @Success 200 {object} types.ResponseWithHeight
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have a valid proposalID or height."
// @Failure 500 {object} rest.ErrorResponse "Returned if the proposalID isn't in the store."
// @Router /gov/proposals/{proposalID}/votes [get]
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
//
// queryProposalsWithParameterFn implements governance proposals querying
// for returning data on a filtered list of proposals.
//
// @Summary Query for the list of governance proposals.
// @Tags queries
// @Description Query the list of governance proposals with optional filters for
// @Description proposal status, depositor, and/or voter.
// @Produce json
// @Param height query string false "block height to execute query, defaults to chain tip."
// @Param status query string false "filter proposals by proposal status: deposit_period|voting_period|passed|rejected."
// @Param depositor query string false "filter proposals by depositor address."
// @Param voter query string false "filter proposals by voter address."
// @Success 200 {object} types.ResponseWithHeight
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have a valid parameters."
// @Failure 500 {object} rest.ErrorResponse "Returned if store query errors."
// @Router /gov/proposals [get]
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
//
// queryTallyOnProposalHandlerFn implements a governance proposal tally querying route
// for returning data on an individual governance proposal tally.
//
// @Summary Query a governance proposal's individual tally.
// @Tags queries
// @Description Query an individual governance proposal's vote tally.
// @Produce json
// @Param proposalID path int true "The ID of the governance proposal."
// @Param height query string false "block height to execute query, defaults to chain tip."
// @Success 200 {object} types.ResponseWithHeight
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have a valid proposalID or height."
// @Failure 500 {object} rest.ErrorResponse "Returned if the proposalID or depositor isn't in the store."
// @Router /gov/proposals/{proposalID}/tally [get]
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
