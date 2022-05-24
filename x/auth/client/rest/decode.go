package rest

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

type (
	// DecodeReq defines a tx decoding request.
	DecodeReq struct {
		Tx string `json:"tx"`
	}

	// DecodeResp defines a tx decoding response.
	DecodeResp legacytx.StdTx
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
		err = clientCtx.LegacyAmino.UnmarshalJSON(body, &req)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		txBytes, err := base64.StdEncoding.DecodeString(req.Tx)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		stdTx, err := convertToStdTx(w, clientCtx, txBytes)
		if err != nil {
			// Error is already returned by convertToStdTx.
			return
		}

		response := DecodeResp(stdTx)

		err = checkAminoMarshalError(clientCtx, response, "/cosmos/tx/v1beta1/txs/decode")
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())

			return
		}

		rest.PostProcessResponse(w, clientCtx, response)
	}
}

// convertToStdTx converts tx proto binary bytes retrieved from Tendermint into
// a StdTx. Returns the StdTx, as well as a flag denoting if the function
// successfully converted or not.
func convertToStdTx(w http.ResponseWriter, clientCtx client.Context, txBytes []byte) (legacytx.StdTx, error) {
	txI, err := clientCtx.TxConfig.TxDecoder()(txBytes)
	if rest.CheckBadRequestError(w, err) {
		return legacytx.StdTx{}, err
	}

	tx, ok := txI.(signing.Tx)
	if !ok {
		rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("%+v is not backwards compatible with %T", tx, legacytx.StdTx{}))
		return legacytx.StdTx{}, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "expected %T, got %T", (signing.Tx)(nil), txI)
	}

	stdTx, err := clienttx.ConvertTxToStdTx(clientCtx.LegacyAmino, tx)
	if rest.CheckBadRequestError(w, err) {
		return legacytx.StdTx{}, err
	}

	return stdTx, nil
}
