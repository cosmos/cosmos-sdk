package tx

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"io/ioutil"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/client/utils"
)

const (
	// Returns with the response from CheckTx.
	flagSync  = "sync"
	// Returns right away, with no response
	flagAsync = "async"
	// Only returns error if mempool.BroadcastTx errs (ie. problem with the app) or if we timeout waiting for tx to commit.
	flagBlock = "block"
)

// BroadcastBody Tx Broadcast Body
type BroadcastBody struct {
	TxBytes string `json:"tx"`
	Return string `json:"return"`
}

// BroadcastTxRequest REST Handler
// nolint: gocyclo
func BroadcastTxRequest(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m BroadcastBody
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		err = cdc.UnmarshalJSON(body, &m)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		var output []byte
		switch m.Return {
		case flagBlock:
			res, err := cliCtx.BroadcastTx([]byte(m.TxBytes))
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			output, err = cdc.MarshalJSON(res)
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
		case flagSync:
			res, err := cliCtx.BroadcastTxSync([]byte(m.TxBytes))
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			output, err = cdc.MarshalJSON(res)
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
		case flagAsync:
			res, err := cliCtx.BroadcastTxAsync([]byte(m.TxBytes))
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			output, err = cdc.MarshalJSON(res)
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
		default:
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "unsupported return type. supported types: block, sync, async")
			return
		}

		w.Write(output)
	}
}
