package rest

import (
	"encoding/binary"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/ibc/ports/{%s}/channels/{%s}/next-sequence", RestChannelID, RestPortID), queryChannelHandlerFn(cliCtx)).Methods("GET")
}

func queryChannelHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		portID := vars[RestPortID]
		channelID := vars[RestChannelID]

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		req := abci.RequestQuery{
			Path:  "store/ibc/key",
			Data:  channel.KeyNextSequenceRecv(portID, channelID),
			Prove: true,
		}

		res, err := cliCtx.QueryABCI(req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		sequence := binary.BigEndian.Uint64(res.Value)

		cliCtx = cliCtx.WithHeight(res.Height)
		rest.PostProcessResponse(w, cliCtx, sequence)
	}
}
