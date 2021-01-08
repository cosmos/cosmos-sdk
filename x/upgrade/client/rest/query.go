package rest

import (
	"encoding/binary"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func registerQueryRoutes(clientCtx client.Context, r *mux.Router) {
	r.HandleFunc(
		"/upgrade/current", getCurrentPlanHandler(clientCtx),
	).Methods("GET")
	r.HandleFunc(
		"/upgrade/applied/{name}", getDonePlanHandler(clientCtx),
	).Methods("GET")
}

func getCurrentPlanHandler(clientCtx client.Context) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, request *http.Request) {
		// ignore height for now
		res, _, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s", types.QuerierKey, types.QueryCurrent))
		if rest.CheckInternalServerError(w, err) {
			return
		}
		if len(res) == 0 {
			http.NotFound(w, request)
			return
		}

		var plan types.Plan
		err = clientCtx.LegacyAmino.UnmarshalJSON(res, &plan)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		rest.PostProcessResponse(w, clientCtx, plan)
	}
}

func getDonePlanHandler(clientCtx client.Context) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]

		params := types.QueryAppliedPlanRequest{Name: name}
		bz, err := clientCtx.LegacyAmino.MarshalJSON(params)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		res, _, err := clientCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierKey, types.QueryApplied), bz)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		if len(res) == 0 {
			http.NotFound(w, r)
			return
		}
		if len(res) != 8 {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, "unknown format for applied-upgrade")
		}

		applied := int64(binary.BigEndian.Uint64(res))
		fmt.Println(applied)
		rest.PostProcessResponse(w, clientCtx, applied)
	}
}
