package rest

import (
	"encoding/hex"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// BroadcastReq defines a tx broadcasting request.
type BroadcastReq struct {
	Tx   types.StdTx `json:"tx" yaml:"tx"`
	Mode string      `json:"mode" yaml:"mode"`
}

// BroadcastTxRequest implements a tx broadcasting handler that is responsible
// for broadcasting a valid and signed tx to a full node. The tx can be
// broadcasted via a sync|async|block mechanism.
func BroadcastTxRequest(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req BroadcastReq

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		err = cliCtx.Codec.UnmarshalJSON(body, &req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		txBytes, err := cliCtx.Codec.MarshalBinaryLengthPrefixed(req.Tx)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		if req.Mode == flags.BroadcastSimulate {
			rawRes, _, err := cliCtx.QueryWithData("/app/simulate", txBytes)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}

			var simulationResult sdk.Result
			if err := cliCtx.Codec.UnmarshalBinaryLengthPrefixed(rawRes, &simulationResult); err != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			txResponse := sdk.TxResponse{
				Code:      uint32(simulationResult.Code),
				Data:      hex.EncodeToString(simulationResult.Data),
				Codespace: string(simulationResult.Codespace),
				RawLog:    simulationResult.Log,
				GasWanted: int64(simulationResult.GasWanted),
				GasUsed:   int64(simulationResult.GasUsed),
			}
			rest.PostProcessResponseBare(w, cliCtx, txResponse)
			return
		}

		cliCtx = cliCtx.WithBroadcastMode(req.Mode)

		res, err := cliCtx.BroadcastTx(txBytes)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponseBare(w, cliCtx, res)
	}
}
