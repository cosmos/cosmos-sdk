package rest

import (
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
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
//
// @Summary Broadcast a signed transaction
// @Description Broadcast a signed transaction with the broadcasting mode. The
// @Description mode must either be sync, async, or block. The use of block mode
// @Description is not advised. The sync mode will broadcast and wait for a
// @Description CheckTx response, whereas async mode will broadcast and return
// @Description immediately.
// @Tags transactions
// @Accept  json
// @Produce  json
// @Param tx body rest.BroadcastReq true "Signed transaction along with the broadcasting mode"
// @Success 200 {object} types.TxResponse
// @Failure 400 {object} rest.ErrorResponse "Returned if the request is invalid."
// @Failure 500 {object} rest.ErrorResponse "Returned if the transaction cannot be decoded."
// @Router /txs [post]
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

		cliCtx = cliCtx.WithBroadcastMode(req.Mode)

		res, err := cliCtx.BroadcastTx(txBytes)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponseBare(w, cliCtx, res)
	}
}
