package rest

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	"github.com/gorilla/mux"
	"net/http"
)

// RegisterRoutes registers REST routes for the upgrade module under the path specified by routeName.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec, routeName string, storeName string) {
	r.HandleFunc(fmt.Sprintf("/%s", routeName), getUpgradePlanHandler(cdc, cliCtx, storeName)).Methods("GET")
}

func getUpgradePlanHandler(cdc *codec.Codec, cliCtx context.CLIContext, storeName string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, request *http.Request) {
		res, err := cliCtx.QueryStore([]byte(upgrade.PlanKey), storeName)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		if len(res) == 0 {
			http.NotFound(w, request)
			return
		}

		var plan upgrade.Plan
		err = cdc.UnmarshalBinaryBare(res, &plan)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, plan, cliCtx.Indent)
	}
}
