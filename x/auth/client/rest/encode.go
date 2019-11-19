package rest

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
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
func EncodeTxRequestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.StdTx

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

		// re-encode it via the Amino wire protocol
		txBytes, err := cliCtx.Codec.MarshalBinaryLengthPrefixed(req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// base64 encode the encoded tx bytes
		txBytesBase64 := base64.StdEncoding.EncodeToString(txBytes)

		response := EncodeResp{Tx: txBytesBase64}
		rest.PostProcessResponseBare(w, cliCtx, response)
	}
}
