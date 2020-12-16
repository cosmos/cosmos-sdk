package rest

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	clientrest "github.com/cosmos/cosmos-sdk/client/rest"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
)

// BroadcastReq defines a tx broadcasting request.
type BroadcastReq struct {
	Tx   legacytx.StdTx `json:"tx" yaml:"tx"`
	Mode string         `json:"mode" yaml:"mode"`
}

var _ codectypes.UnpackInterfacesMessage = BroadcastReq{}

// UnpackInterfaces implements the UnpackInterfacesMessage interface.
func (m BroadcastReq) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return m.Tx.UnpackInterfaces(unpacker)
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

		// NOTE: amino is used intentionally here, don't migrate it!
		err = clientCtx.LegacyAmino.UnmarshalJSON(body, &req)
		if err != nil {
			err := fmt.Errorf("this transaction cannot be broadcasted via legacy REST endpoints, because it does not support"+
				" Amino serialization. Please either use CLI, gRPC, gRPC-gateway, or directly query the Tendermint RPC"+
				" endpoint to broadcast this transaction. The new REST endpoint (via gRPC-gateway) is POST /cosmos/tx/v1beta1/txs."+
				" Please also see the REST endpoints migration guide at %s for more info", clientrest.DeprecationURL)
			if rest.CheckBadRequestError(w, err) {
				return
			}
		}

		txBytes, err := tx.ConvertAndEncodeStdTx(clientCtx.TxConfig, req.Tx)
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
