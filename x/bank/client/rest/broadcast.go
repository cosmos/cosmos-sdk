package rest

import (
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

type broadcastBody struct {
	Tx auth.StdTx `json:"tx"`
}

// BroadcastTxRequestHandlerFn returns the broadcast tx REST handler
func BroadcastTxRequestHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m broadcastBody
		if ok := unmarshalBodyOrReturnBadRequest(cliCtx, w, r, &m); !ok {
			return
		}

		txBytes, err := cliCtx.Codec.MarshalBinaryLengthPrefixed(m.Tx)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		res, err := cliCtx.BroadcastTx(txBytes)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		utils.PostProcessResponse(w, cdc, res, cliCtx.Indent)
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
