package rest

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

type (
	// DecodeReq defines a tx decoding request.
	DecodeReq struct {
		Tx string `json:"tx"`
	}

	// DecodeResp defines a tx decoding response.
	DecodeResp types.StdTx
)

// DecodeTxRequestHandlerFn returns the decode tx REST handler. In particular,
// it takes base64-decoded bytes, decodes it from the Amino wire protocol,
// and responds with a json-formatted transaction.
func DecodeTxRequestHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req DecodeReq

		body, err := ioutil.ReadAll(r.Body)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		// NOTE: amino is used intentionally here, don't migrate it
		err = clientCtx.Codec.UnmarshalJSON(body, &req)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		txBytes, err := base64.StdEncoding.DecodeString(req.Tx)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		tx, err := clientCtx.TxConfig.TxDecoder()(txBytes)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		stdTx, err := clienttx.ConvertTxToStdTx(clientCtx.Codec, tx)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		response := DecodeResp(stdTx)

		// NOTE: amino is set intentionally here, don't migrate it
		clientCtx = clientCtx.WithJSONMarshaler(clientCtx.Codec)
		rest.PostProcessResponse(w, clientCtx, response)
	}
}
