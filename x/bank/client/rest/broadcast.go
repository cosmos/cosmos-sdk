package rest

import (
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

const (
	flagSync  = "sync"
	flagAsync = "async"
	flagBlock = "block"
)


type broadcastBody struct {
	Tx auth.StdTx `json:"tx"`
	Return string `json:"return"`
}

// BroadcastTxRequestHandlerFn returns the broadcast tx REST handler
func BroadcastTxRequestHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m broadcastBody
		if ok := unmarshalBodyOrReturnBadRequest(cliCtx, w, r, &m); !ok {
			return
		}

		txBytes, err := cliCtx.Codec.MarshalBinary(m.Tx)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		var res interface{}
		switch m.Return {
		case flagBlock:
			res, err = cliCtx.BroadcastTx(txBytes)
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
		case flagSync:
			res, err = cliCtx.BroadcastTxSync(txBytes)
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
		case flagAsync:
			res, err = cliCtx.BroadcastTxAsync(txBytes)
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
		default:
			utils.WriteErrorResponse(w, http.StatusInternalServerError,
				"unsupported return type. supported types: block, sync, async")
			return
		}

		output, err := codec.MarshalJSONIndent(cdc, res)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Write(output)
	}
}

func unmarshalBodyOrReturnBadRequest(cliCtx context.CLIContext, w http.ResponseWriter, r *http.Request, m *broadcastBody) bool {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return false
	}
	err = cliCtx.Codec.UnmarshalJSON(body, m)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return false
	}
	return true
}
