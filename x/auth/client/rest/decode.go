package rest

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type (
	// DecodeReq defines a tx decoding request.
	DecodeReq struct {
		Tx string `json:"tx"`
	}

	// DecodeResp defines a tx decoding response.
	DecodeResp authtypes.StdTx
)

// DecodeTxRequestHandlerFn returns the decode tx REST handler. In particular,
// it takes base64-decoded bytes, decodes it from the Amino wire protocol,
// and responds with a json-formatted transaction.
func DecodeTxRequestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req DecodeReq

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

		txBytes, err := base64.StdEncoding.DecodeString(req.Tx)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var stdTx authtypes.StdTx
		err = cliCtx.Codec.UnmarshalBinaryLengthPrefixed(txBytes, &stdTx)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		response := DecodeResp(stdTx)
		rest.PostProcessResponse(w, cliCtx, response)
	}
}
