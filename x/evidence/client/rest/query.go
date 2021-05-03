package rest

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"

	"github.com/gorilla/mux"
)

func registerQueryRoutes(clientCtx client.Context, r *mux.Router) {
	r.HandleFunc(
		fmt.Sprintf("/evidence/{%s}", RestParamEvidenceHash),
		queryEvidenceHandler(clientCtx),
	).Methods(MethodGet)

	r.HandleFunc(
		"/evidence",
		queryAllEvidenceHandler(clientCtx),
	).Methods(MethodGet)
}

func queryEvidenceHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		evidenceHash := vars[RestParamEvidenceHash]

		if strings.TrimSpace(evidenceHash) == "" {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "evidence hash required but not specified")
			return
		}

		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		decodedHash, err := hex.DecodeString(evidenceHash)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "invalid evidence hash")
			return
		}

		params := types.NewQueryEvidenceRequest(decodedHash)
		bz, err := clientCtx.JSONCodec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal query params: %s", err))
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryEvidence)
		res, height, err := clientCtx.QueryWithData(route, bz)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}

func queryAllEvidenceHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, page, limit, err := rest.ParseHTTPArgsWithLimit(r, 0)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		params := types.NewQueryAllEvidenceParams(page, limit)
		bz, err := clientCtx.LegacyAmino.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal query params: %s", err))
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryAllEvidence)
		res, height, err := clientCtx.QueryWithData(route, bz)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}
