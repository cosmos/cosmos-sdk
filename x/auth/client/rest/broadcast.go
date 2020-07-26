package rest

import (
	"io/ioutil"
	"net/http"

	"github.com/KiraCore/cosmos-sdk/client"
	"github.com/KiraCore/cosmos-sdk/types/rest"
	"github.com/KiraCore/cosmos-sdk/x/auth/types"
)

// BroadcastReq defines a tx broadcasting request.
type BroadcastReq struct {
	Tx   types.StdTx `json:"tx" yaml:"tx"`
	Mode string      `json:"mode" yaml:"mode"`
}

// BroadcastTxRequest implements a tx broadcasting handler that is responsible
// for broadcasting a valid and signed tx to a full node. The tx can be
// broadcasted via a sync|async|block mechanism.
func BroadcastTxRequest(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req BroadcastReq

		body, err := ioutil.ReadAll(r.Body)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		if err := clientCtx.JSONMarshaler.UnmarshalJSON(body, &req); rest.CheckBadRequestError(w, err) {
			return
		}

		txBytes, err := clientCtx.Codec.MarshalBinaryBare(req.Tx)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithBroadcastMode(req.Mode)

		res, err := clientCtx.BroadcastTx(txBytes)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		rest.PostProcessResponseBare(w, clientCtx, res)
	}
}
