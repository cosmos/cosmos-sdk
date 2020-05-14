package rest

import (
	"fmt"
	"net/http"

	"github.com/KiraCore/cosmos-sdk/client/context"
	sdk "github.com/KiraCore/cosmos-sdk/types"
	"github.com/KiraCore/cosmos-sdk/types/rest"
	"github.com/KiraCore/cosmos-sdk/x/genutil/types"
)

// QueryGenesisTxs writes the genesis transactions to the response if no error
// occurs.
func QueryGenesisTxs(cliCtx context.CLIContext, w http.ResponseWriter) {
	resultGenesis, err := cliCtx.Client.Genesis()
	if err != nil {
		rest.WriteErrorResponse(
			w, http.StatusInternalServerError,
			fmt.Sprintf("failed to retrieve genesis from client: %s", err),
		)
		return
	}

	appState, err := types.GenesisStateFromGenDoc(cliCtx.Codec, *resultGenesis.Genesis)
	if err != nil {
		rest.WriteErrorResponse(
			w, http.StatusInternalServerError,
			fmt.Sprintf("failed to decode genesis doc: %s", err),
		)
		return
	}

	genState := types.GetGenesisStateFromAppState(cliCtx.Codec, appState)
	genTxs := make([]sdk.Tx, len(genState.GenTxs))
	for i, tx := range genState.GenTxs {
		err := cliCtx.Codec.UnmarshalJSON(tx, &genTxs[i])
		if err != nil {
			rest.WriteErrorResponse(
				w, http.StatusInternalServerError,
				fmt.Sprintf("failed to decode genesis transaction: %s", err),
			)
			return
		}
	}

	rest.PostProcessResponseBare(w, cliCtx, genTxs)
}
