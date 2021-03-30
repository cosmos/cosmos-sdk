package rest

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	clientrest "github.com/cosmos/cosmos-sdk/client/rest"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
)

// EncodeResp defines a tx encoding response.
type EncodeResp struct {
	Tx string `json:"tx" yaml:"tx"`
}

// ErrEncodeDecode is the error to show when encoding/decoding txs that are not
// amino-serializable (e.g. IBC txs).
var ErrEncodeDecode error = fmt.Errorf("this endpoint does not support txs that are not serializable"+
	" via Amino, such as txs that contain IBC `Msg`s. For more info, please refer to our"+
	" REST migration guide at %s", clientrest.DeprecationURL)

// EncodeTxRequestHandlerFn returns the encode tx REST handler. In particular,
// it takes a json-formatted transaction, encodes it to the Amino wire protocol,
// and responds with base64-encoded bytes.
func EncodeTxRequestHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req legacytx.StdTx

		body, err := ioutil.ReadAll(r.Body)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		// NOTE: amino is used intentionally here, don't migrate it
		err = clientCtx.LegacyAmino.UnmarshalJSON(body, &req)
		// If there's an unmarshalling error, we assume that it's because we're
		// using amino to unmarshal a non-amino tx.
		if err != nil {
			if rest.CheckBadRequestError(w, ErrEncodeDecode) {
				return
			}
		}

		// re-encode it in the chain's native binary format
		txBytes, err := tx.ConvertAndEncodeStdTx(clientCtx.TxConfig, req)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		// base64 encode the encoded tx bytes
		txBytesBase64 := base64.StdEncoding.EncodeToString(txBytes)

		response := EncodeResp{Tx: txBytesBase64}

		// NOTE: amino is set intentionally here, don't migrate it
		rest.PostProcessResponseBare(w, clientCtx, response)
	}
}
