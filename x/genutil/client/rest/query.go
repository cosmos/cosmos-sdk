package query

import (
	"encoding/json"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// QueryGenesisTxs implements a REST handler that returns genesis transactions.
func QueryGenesisTxs(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		genDoc, err := cliCtx.Client.HistoryClient.Genesis()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError,
				sdk.AppendMsgToErr("could not retrieve genesis doc", err.Error()))
			return
		}

		appState, err := types.GenesisStateFromGenDoc(cliCtx.Codec, genDoc)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError,
				sdk.AppendMsgToErr("could not decode genesis doc", err.Error()))
			return
		}

		genState, err := types.GetGenesisStateFromAppState(cliCtx.Codec, appState)
		genTxs := make([]sdk.Tx, len(genState.GenTxs))
		for i, tx := range genState.GenTxs {
			err := cliCtx.Codec.UnmarshalJSON(tx, &genTxs[i])
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError,
					sdk.AppendMsgToErr("could not decode genesis transaction", err.Error()))
				return
			}
		}

		rest.PostProcessResponse(w, cliCtx, genTxs)
	}
}
