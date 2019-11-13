package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router, queryRoute string) {
	r.HandleFunc(fmt.Sprintf("/ibc/channel/ports/{%s}/channels/{%s}", RestChannelID, RestPortID), queryChannelHandlerFn(cliCtx, queryRoute)).Methods("GET")
}

func queryChannelHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		portID := vars[RestPortID]
		channelID := vars[RestChannelID]

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		// return proof if the prove query param is set to true
		proveStr := r.FormValue("prove")
		prove := false
		if strings.ToLower(strings.TrimSpace(proveStr)) == "true" {
			prove = true
		}

		bz, err := cliCtx.Codec.MarshalJSON(types.NewQueryChannelParams(portID, channelID))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		req := abci.RequestQuery{
			Path:  fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryChannel),
			Data:  bz,
			Prove: prove,
		}

		res, err := cliCtx.QueryABCI(req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(res.Height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
