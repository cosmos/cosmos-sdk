package fee_delegation

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

// RegisterRoutes registers staking-related REST handlers to a router
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	registerQueryRoutes(cliCtx, r)
}

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(
		"/fee-delegation/allowances/{granteeAddr}",
		getAllowFeesHandlerFn(cliCtx),
	).Methods("GET")
}

type QueryCapabilityParams struct {
	Grantee sdk.AccAddress
	Granter sdk.AccAddress
	Route   string
	Typ     string
}

func getAllowFeesHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		grantee := vars["granteeAddr"]
		route := fmt.Sprintf("custom/fee-delegation/%s/%s", QueryGetFeeAllowances, grantee)

		res, _, err := cliCtx.QueryWithData(route, []byte{})
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
