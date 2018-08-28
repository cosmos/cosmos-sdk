package tx

import (
	"encoding/json"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/httputils"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"fmt"
)

const (
	FlagSync     = "sync"
	FlagAsync     = "async"
	FlagBlock     = "block"
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

type TxBody struct {
	Transaction string `json:"transaction"`
	Return string `json:"return"`
}

// BroadcastTxRequest REST Handler
func BroadcastTxRequest(cdc *wire.Codec, ctx context.CLIContext) gin.HandlerFunc {
	return func(gtx *gin.Context) {
		var txBody TxBody
		body, err := ioutil.ReadAll(gtx.Request.Body)
		if err != nil {
			httputils.NewError(gtx, http.StatusBadRequest, err)
			return
		}
		err = cdc.UnmarshalJSON(body, &txBody)
		if err != nil {
			httputils.NewError(gtx, http.StatusBadRequest, err)
			return
		}
		var output []byte
		switch txBody.Return {
		case FlagBlock:
			res, err := ctx.BroadcastTx([]byte(txBody.Transaction))
			if err != nil {
				httputils.NewError(gtx, http.StatusInternalServerError, err)
				return
			}
			output, err = cdc.MarshalJSON(res)
			if err != nil {
				httputils.NewError(gtx, http.StatusInternalServerError, err)
				return
			}
		case FlagSync:
			res, err := ctx.BroadcastTxSync([]byte(txBody.Transaction))
			if err != nil {
				httputils.NewError(gtx, http.StatusInternalServerError, err)
				return
			}
			output, err = cdc.MarshalJSON(res)
			if err != nil {
				httputils.NewError(gtx, http.StatusInternalServerError, err)
				return
			}
		case FlagAsync:
			res, err := ctx.BroadcastTxAsync([]byte(txBody.Transaction))
			if err != nil {
				httputils.NewError(gtx, http.StatusInternalServerError, err)
				return
			}
			output, err = cdc.MarshalJSON(res)
			if err != nil {
				httputils.NewError(gtx, http.StatusInternalServerError, err)
				return
			}
		default:
			httputils.NewError(gtx, http.StatusBadRequest, fmt.Errorf("unsupported return type. supported types: block, sync, async"))
			return
		}
		httputils.NormalResponse(gtx,output)
	}
}