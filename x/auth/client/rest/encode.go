package rest

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// EncodeResp defines a tx encoding response.
type EncodeResp struct {
	Tx string `json:"tx" yaml:"tx"`
}

// EncodeTxRequestHandlerFn returns the encode tx REST handler. In particular,
// it takes a json-formatted transaction, encodes it to the Amino wire protocol,
// and responds with base64-encoded bytes.
func EncodeTxRequestHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.StdTx

		body, err := ioutil.ReadAll(r.Body)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		// NOTE: amino is used intentionally here, don't migrate it
		err = clientCtx.Codec.UnmarshalJSON(body, &req)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		// re-encode it in the chain's native binary format
		txBytes, err := convertAndEncodeStdTx(clientCtx.TxConfig, req)
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

func convertAndEncodeStdTx(txConfig client.TxConfig, stdTx types.StdTx) ([]byte, error) {
	builder := txConfig.NewTxBuilder()

	var theTx sdk.Tx

	// check if we need a StdTx anyway, in that case don't copy
	if _, ok := builder.GetTx().(types.StdTx); ok {
		theTx = stdTx
	} else {
		err := tx.CopyTx(stdTx, builder)
		if err != nil {
			return nil, err
		}

		theTx = builder.GetTx()
	}

	return txConfig.TxEncoder()(theTx)
}
