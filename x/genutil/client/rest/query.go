package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// QueryGenesisTxs writes the genesis transactions to the response if no error
// occurs.
func QueryGenesisTxs(clientCtx client.Context, w http.ResponseWriter) {
	resultGenesis, err := clientCtx.Client.Genesis()
	if err != nil {
		rest.WriteErrorResponse(
			w, http.StatusInternalServerError,
			fmt.Sprintf("failed to retrieve genesis from client: %s", err),
		)
		return
	}

	appState, err := types.GenesisStateFromGenDoc(clientCtx.Codec, *resultGenesis.Genesis)
	if err != nil {
		rest.WriteErrorResponse(
			w, http.StatusInternalServerError,
			fmt.Sprintf("failed to decode genesis doc: %s", err),
		)
		return
	}

	genState := types.GetGenesisStateFromAppState(clientCtx.Codec, appState)
	genTxs := make([]sdk.Tx, len(genState.GenTxs))
	for i, tx := range genState.GenTxs {
		err := clientCtx.Codec.UnmarshalJSON(tx, &genTxs[i])
		if err != nil {
			rest.WriteErrorResponse(
				w, http.StatusInternalServerError,
				fmt.Sprintf("failed to decode genesis transaction: %s", err),
			)
			return
		}
	}

	rest.PostProcessResponseBare(w, clientCtx, genTxs)
}
