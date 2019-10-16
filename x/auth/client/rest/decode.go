package rest

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
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
//
// @Summary Decode a transaction from the Amino wire format
// @Description Decode a transaction (signed or not) from base64-encoded Amino serialized bytes to JSON
// @Tags transactions
// @Accept json
// @Produce json
// @Param tx body rest.DecodeReq true "The transaction to decode"
// @Success 200 {object} rest.DecodeResp
// @Failure 400 {object} rest.ErrorResponse "The transaction was malformated"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /txs/decode [post]
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

		var stdTx types.StdTx

		err = cliCtx.Codec.UnmarshalBinaryLengthPrefixed(txBytes, &stdTx)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		response := DecodeResp(stdTx)
		rest.PostProcessResponseBare(w, cliCtx, response)
	}
}
