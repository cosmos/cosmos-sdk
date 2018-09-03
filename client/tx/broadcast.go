package tx

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/wire"
	"io/ioutil"
)

const (
	flagSync  = "sync"
	flagAsync = "async"
	flagBlock = "block"
)

// Tx Broadcast Body
// BroadcastBody contains the data of tx and specify how to broadcast tx
type BroadcastBody struct {
	Transaction string `json:"transaction"`
	Return      string `json:"return"`
}

// BroadcastTx REST Handler
func BroadcastTxRequestHandlerFn(cdc *wire.Codec, ctx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var txBody BroadcastBody
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		err = cdc.UnmarshalJSON(body, &txBody)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		var output []byte
		switch txBody.Return {
		case flagBlock:
			res, err := ctx.BroadcastTx([]byte(txBody.Transaction))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			output, err = cdc.MarshalJSON(res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
		case flagSync:
			res, err := ctx.BroadcastTxSync([]byte(txBody.Transaction))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			output, err = cdc.MarshalJSON(res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
		case flagAsync:
			res, err := ctx.BroadcastTxAsync([]byte(txBody.Transaction))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			output, err = cdc.MarshalJSON(res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
		default:
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unsupported return type. supported types: block, sync, async"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(output)
	}
}
