package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/KiraCore/cosmos-sdk/client"
	"github.com/KiraCore/cosmos-sdk/client/flags"
	"github.com/KiraCore/cosmos-sdk/types/rest"
	"github.com/KiraCore/cosmos-sdk/x/ibc/02-client/client/utils"
)

func registerQueryRoutes(clientCtx client.Context, r *mux.Router) {
	r.HandleFunc("/ibc/clients", queryAllClientStatesFn(clientCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/ibc/clients/{%s}/client-state", RestClientID), queryClientStateHandlerFn(clientCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/ibc/clients/{%s}/consensus-state/{%s}", RestClientID, RestRootHeight), queryConsensusStateHandlerFn(clientCtx)).Methods("GET")
	r.HandleFunc("/ibc/header", queryHeaderHandlerFn(clientCtx)).Methods("GET")
	r.HandleFunc("/ibc/node-state", queryNodeConsensusStateHandlerFn(clientCtx)).Methods("GET")
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
func queryAllClientStatesFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, page, limit, err := rest.ParseHTTPArgsWithLimit(r, 0)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		clients, height, err := utils.QueryAllClientStates(clientCtx, page, limit)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, clients)
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
func queryClientStateHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		clientID := vars[RestClientID]
		prove := rest.ParseQueryParamBool(r, flags.FlagProve)

		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		clientStateRes, err := utils.QueryClientState(clientCtx, clientID, prove)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(int64(clientStateRes.ProofHeight))
		rest.PostProcessResponse(w, clientCtx, clientStateRes)
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
// @Router /ibc/clients/{client-id}/consensus-state/{height} [get]
func queryConsensusStateHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		clientID := vars[RestClientID]
		height, err := strconv.ParseUint(vars[RestRootHeight], 10, 64)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		prove := rest.ParseQueryParamBool(r, flags.FlagProve)

		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		csRes, err := utils.QueryConsensusState(clientCtx, clientID, height, prove)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(int64(csRes.ProofHeight))
		rest.PostProcessResponse(w, clientCtx, csRes)
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
func queryHeaderHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header, height, err := utils.QueryTendermintHeader(clientCtx)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		res := clientCtx.JSONMarshaler.MustMarshalJSON(header)
		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
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
func queryNodeConsensusStateHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state, height, err := utils.QueryNodeConsensusState(clientCtx)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		}

		res := clientCtx.JSONMarshaler.MustMarshalJSON(state)
		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}
