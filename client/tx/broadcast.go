package tx

import (
	"encoding/json"
	"net/http"

	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/gin-gonic/gin"
	"io/ioutil"
)

const (
	flagSync  = "sync"
	flagAsync = "async"
	flagBlock = "block"
)

// Tx Broadcast Body
type BroadcastTxBody struct {
	TxBytes string `json:"tx"`
}

// BroadcastTx REST Handler
func BroadcastTxRequestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m BroadcastTxBody

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&m)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte(err.Error()))
			return
		}

		res, err := cliCtx.BroadcastTx([]byte(m.TxBytes))
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write([]byte(string(res.Height)))
	}
}

// BroadcastBody contains the data of tx and specify how to broadcast tx
type BroadcastBody struct {
	Transaction string `json:"transaction"`
	Return      string `json:"return"`
}

// BroadcastTxRequest - Handler of broadcast tx
// nolint: gocyclo
func BroadcastTxRequest(cdc *wire.Codec, ctx context.CLIContext) gin.HandlerFunc {
	return func(gtx *gin.Context) {
		var txBody BroadcastBody
		body, err := ioutil.ReadAll(gtx.Request.Body)
		if err != nil {
			utils.NewError(gtx, http.StatusBadRequest, err)
			return
		}
		err = cdc.UnmarshalJSON(body, &txBody)
		if err != nil {
			utils.NewError(gtx, http.StatusBadRequest, err)
			return
		}
		var output []byte
		switch txBody.Return {
		case flagBlock:
			res, err := ctx.BroadcastTx([]byte(txBody.Transaction))
			if err != nil {
				utils.NewError(gtx, http.StatusInternalServerError, err)
				return
			}
			output, err = cdc.MarshalJSON(res)
			if err != nil {
				utils.NewError(gtx, http.StatusInternalServerError, err)
				return
			}
		case flagSync:
			res, err := ctx.BroadcastTxSync([]byte(txBody.Transaction))
			if err != nil {
				utils.NewError(gtx, http.StatusInternalServerError, err)
				return
			}
			output, err = cdc.MarshalJSON(res)
			if err != nil {
				utils.NewError(gtx, http.StatusInternalServerError, err)
				return
			}
		case flagAsync:
			res, err := ctx.BroadcastTxAsync([]byte(txBody.Transaction))
			if err != nil {
				utils.NewError(gtx, http.StatusInternalServerError, err)
				return
			}
			output, err = cdc.MarshalJSON(res)
			if err != nil {
				utils.NewError(gtx, http.StatusInternalServerError, err)
				return
			}
		default:
			utils.NewError(gtx, http.StatusBadRequest, fmt.Errorf("unsupported return type. supported types: block, sync, async"))
			return
		}
		utils.NormalResponse(gtx, output)
	}
}
