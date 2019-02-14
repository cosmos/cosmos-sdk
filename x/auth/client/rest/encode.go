package rest

import (
	"encoding/base64"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"github.com/cosmos/cosmos-sdk/x/auth"
)

type encodeReq struct {
	Tx auth.StdTx `json:"tx"`
}

type encodeResp struct {
	Tx string `json:"tx"`
}

// EncodeTxRequestHandlerFn returns the encode tx REST handler.  In particular, it takes a
// json-formatted transaction, encodes it to the Amino wire protocol, and responds with
// base64-encoded bytes
func EncodeTxRequestHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m encodeReq
		// Decode the transaction from JSON
		if ok := unmarshalBodyOrReturnBadRequest(cliCtx, w, r, &m); !ok {
			return
		}

		// Re-encode it to the wire protocol
		txBytes, err := cliCtx.Codec.MarshalBinaryLengthPrefixed(m.Tx)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Encode the bytes to base64
		txBytesBase64 := base64.StdEncoding.EncodeToString(txBytes)

		// Write it back
		response := encodeResp{Tx: txBytesBase64}
		rest.PostProcessResponse(w, cdc, response, cliCtx.Indent)
	}
}
