package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/client/utils"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/ibc/clients", queryAllClientStatesFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/ibc/clients/{%s}/client-state", RestClientID), queryClientStateHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/ibc/clients/{%s}/consensus-state", RestClientID), queryConsensusStateHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/ibc/header", queryHeaderHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/ibc/node-state", queryNodeConsensusStateHandlerFn(cliCtx)).Methods("GET")
}

// queryAllClientStatesFn queries all available light clients
//
// @Summary Query client states
// @Tags IBC
// @Produce  json
// @Param page query int false "The page number to query" default(1)
// @Param limit query int false "The number of results per page" default(100)
// @Success 200 {object} QueryClientState "OK"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/clients [get]
func queryAllClientStatesFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, page, limit, err := rest.ParseHTTPArgsWithLimit(r, 0)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		clients, height, err := utils.QueryAllClientStates(cliCtx, page, limit)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, clients)
	}
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
func queryClientStateHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		clientID := vars[RestClientID]
		prove := rest.ParseQueryParamBool(r, flags.FlagProve)

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
// @Param height path number true "Height"
// @Param prove query boolean false "Proof of result"
// @Success 200 {object} QueryConsensusState "OK"
// @Failure 400 {object} rest.ErrorResponse "Invalid client id"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/clients/{client-id}/consensus-state [get]
func queryConsensusStateHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		clientID := vars[RestClientID]
		height, err := strconv.ParseUint(vars[RestRootHeight], 10, 64)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		prove := rest.ParseQueryParamBool(r, flags.FlagProve)

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		csRes, err := utils.QueryConsensusState(cliCtx, clientID, height, prove)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(int64(csRes.ProofHeight))
		rest.PostProcessResponse(w, cliCtx, csRes)
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
