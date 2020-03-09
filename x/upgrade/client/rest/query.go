package rest

import (
	"encoding/binary"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// RegisterRoutes registers REST routes for the upgrade module under the path specified by routeName.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/upgrade/current", getCurrentPlanHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/upgrade/applied/{name}", getDonePlanHandler(cliCtx)).Methods("GET")
	registerTxRoutes(cliCtx, r)
}

func getCurrentPlanHandler(cliCtx context.CLIContext) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, request *http.Request) {
		// ignore height for now
		res, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s", types.QuerierKey, types.QueryCurrent))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		if len(res) == 0 {
			http.NotFound(w, request)
			return
		}

		var plan types.Plan
		err = cliCtx.Codec.UnmarshalBinaryBare(res, &plan)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, plan)
	}
}

func getDonePlanHandler(cliCtx context.CLIContext) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]

		params := types.NewQueryAppliedParams(name)
		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierKey, types.QueryApplied), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
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
		rest.PostProcessResponse(w, cliCtx, applied)
	}
}
