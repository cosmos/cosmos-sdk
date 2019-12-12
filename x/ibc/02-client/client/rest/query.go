package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/client/utils"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router, queryRoute string) {
	r.HandleFunc(fmt.Sprintf("/ibc/clients/{%s}/client-state", RestClientID), queryClientStateHandlerFn(cliCtx, queryRoute)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/ibc/clients/{%s}/consensus-state", RestClientID), queryConsensusStateHandlerFn(cliCtx, queryRoute)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/ibc/clients/{%s}/roots/{%s}", RestClientID, RestRootHeight), queryRootHandlerFn(cliCtx, queryRoute)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/ibc/clients/{%s}/committers/{%s}", RestClientID, RestRootHeight), queryCommitterHandlerFn(cliCtx, queryRoute)).Methods("GET")
	r.HandleFunc("/ibc/header", queryHeaderHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/ibc/node-state", queryNodeConsensusStateHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/ibc/path", queryPathHandlerFn(cliCtx)).Methods("GET")
}

// queryClientStateHandlerFn implements a client state querying route
//
// @Summary Query client state
// @Tags IBC
// @Produce  json
// @Param client-id path string true "Client ID"
// @Param prove query boolean false "Proof of result"
// @Success 200 {object} QueryClientState "OK"
// @Failure 400 {object} rest.ErrorResponse "Invalid client id"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/clients/{client-id}/client-state [get]
func queryClientStateHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		clientID := vars[RestClientID]
		prove := rest.ParseQueryProve(r)

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		clientStateRes, err := utils.QueryClientState(cliCtx, clientID, prove)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(int64(clientStateRes.ProofHeight))
		rest.PostProcessResponse(w, cliCtx, clientStateRes)
	}
}

// queryConsensusStateHandlerFn implements a consensus state querying route
//
// @Summary Query cliet consensus-state
// @Tags IBC
// @Produce  json
// @Param client-id path string true "Client ID"
// @Param prove query boolean false "Proof of result"
// @Success 200 {object} QueryConsensusState "OK"
// @Failure 400 {object} rest.ErrorResponse "Invalid client id"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/clients/{client-id}/consensus-state [get]
func queryConsensusStateHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		clientID := vars[RestClientID]
		prove := rest.ParseQueryProve(r)

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		csRes, err := utils.QueryConsensusStateProof(cliCtx, clientID, prove)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(int64(csRes.ProofHeight))
		rest.PostProcessResponse(w, cliCtx, csRes)
	}
}

// queryRootHandlerFn implements a root querying route
//
// @Summary Query client root
// @Tags IBC
// @Produce  json
// @Param client-id path string true "Client ID"
// @Param height path number true "Root height"
// @Success 200 {object} QueryRoot "OK"
// @Failure 400 {object} rest.ErrorResponse "Invalid client id or height"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/clients/{client-id}/roots/{height} [get]
func queryRootHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		clientID := vars[RestClientID]
		height, err := strconv.ParseUint(vars[RestRootHeight], 10, 64)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		prove := rest.ParseQueryProve(r)

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		rootRes, err := utils.QueryCommitmentRoot(cliCtx, clientID, height, prove)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(int64(rootRes.ProofHeight))
		rest.PostProcessResponse(w, cliCtx, rootRes)
	}
}

// queryCommitterHandlerFn implements a committer querying route
//
// @Summary Query committer
// @Tags IBC
// @Produce json
// @Param client-id path string true "Client ID"
// @Param height path number true "Committer height"
// @Success 200 {object} QueryCommitter "OK"
// @Failure 400 {object} rest.ErrorResponse "Invalid client id or height"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/clients/{client-id}/committers/{height} [get]
func queryCommitterHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		clientID := vars[RestClientID]
		height, err := strconv.ParseUint(vars[RestRootHeight], 10, 64)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		prove := rest.ParseQueryProve(r)

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		committerRes, err := utils.QueryCommitter(cliCtx, clientID, height, prove)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(int64(committerRes.ProofHeight))
		rest.PostProcessResponse(w, cliCtx, committerRes)
	}
}

// queryHeaderHandlerFn implements a header querying route
//
// @Summary Query header
// @Tags IBC
// @Produce  json
// @Success 200 {object} QueryHeader "OK"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/header [get]
func queryHeaderHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header, height, err := utils.QueryTendermintHeader(cliCtx)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		res := cliCtx.Codec.MustMarshalJSON(header)
		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// queryNodeConsensusStateHandlerFn implements a node consensus state querying route
//
// @Summary Query node consensus-state
// @Tags IBC
// @Produce  json
// @Success 200 {object} QueryNodeConsensusState "OK"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/node-state [get]
func queryNodeConsensusStateHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state, height, err := utils.QueryNodeConsensusState(cliCtx)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		}

		res := cliCtx.Codec.MustMarshalJSON(state)
		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// queryPathHandlerFn implements a node consensus path querying route
//
// @Summary Query IBC path
// @Tags IBC
// @Produce  json
// @Success 200 {object} QueryPath "OK"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/path [get]
func queryPathHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := commitment.NewPrefix([]byte("ibc"))
		res := cliCtx.Codec.MustMarshalJSON(path)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
