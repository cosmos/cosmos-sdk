package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router, queryRoute string) {
	r.HandleFunc(fmt.Sprintf("/ibc/clients/{%s}/consensus-state", RestClientID), queryConsensusStateHandlerFn(cliCtx, queryRoute)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/ibc/clients/{%s}/client-state", RestClientID), queryClientStateHandlerFn(cliCtx, queryRoute)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/ibc/clients/{%s}/roots/{%s}", RestClientID, RestRootHeight), queryRootHandlerFn(cliCtx, queryRoute)).Methods("GET")
	r.HandleFunc("/ibc/header", queryHeaderHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/ibc/node-state", queryNodeConsensusStateHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/ibc/path", queryPathHandlerFn(cliCtx)).Methods("GET")
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

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		req := abci.RequestQuery{
			Path:  "store/ibc/key",
			Data:  types.KeyConsensusState(clientID),
			Prove: rest.ParseQueryProve(r),
		}

		res, err := cliCtx.QueryABCI(req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		var cs tendermint.ConsensusState
		if err := cliCtx.Codec.UnmarshalBinaryLengthPrefixed(res.Value, &cs); err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		}

		cliCtx = cliCtx.WithHeight(res.Height)
		rest.PostProcessResponse(w, cliCtx, types.NewConsensusStateResponse(clientID, cs, res.Proof, res.Height))
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
		header, err := utils.GetTendermintHeader(cliCtx)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, header)
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
func queryClientStateHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		clientID := vars[RestClientID]

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		req := abci.RequestQuery{
			Path:  "store/ibc/key",
			Data:  types.KeyClientState(clientID),
			Prove: rest.ParseQueryProve(r),
		}

		res, err := cliCtx.QueryABCI(req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		var state types.State
		if err := cliCtx.Codec.UnmarshalBinaryLengthPrefixed(res.Value, &state); err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		}

		cliCtx = cliCtx.WithHeight(res.Height)
		rest.PostProcessResponse(w, cliCtx, types.NewClientStateResponse(clientID, state, res.Proof, res.Height))
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

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		req := abci.RequestQuery{
			Path:  "store/ibc/key",
			Data:  types.KeyRoot(clientID, height),
			Prove: rest.ParseQueryProve(r),
		}

		res, err := cliCtx.QueryABCI(req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		var root commitment.Root
		if err := cliCtx.Codec.UnmarshalJSON(res.Value, &root); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(res.Height)
		rest.PostProcessResponse(w, cliCtx, types.NewRootResponse(clientID, height, root, res.Proof, res.Height))
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
		node, err := cliCtx.GetNode()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		info, err := node.ABCIInfo()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		height := info.Response.LastBlockHeight
		prevHeight := height - 1

		commit, err := node.Commit(&height)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		validators, err := node.Validators(&prevHeight)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		state := tendermint.ConsensusState{
			ChainID:          commit.ChainID,
			Height:           uint64(commit.Height),
			Root:             commitment.NewRoot(commit.AppHash),
			NextValidatorSet: tmtypes.NewValidatorSet(validators.Validators),
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
