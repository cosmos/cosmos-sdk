package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/keyerror"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
)

// SignBody defines the properties of a sign request's body.
type SignBody struct {
	Tx        auth.StdTx    `json:"tx"`
	AppendSig bool          `json:"append_sig"`
	BaseReq   utils.BaseReq `json:"base_req"`
}

// nolint: unparam
// sign tx REST handler
func SignTxRequestHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m SignBody

		if err := utils.ReadRESTReq(w, r, cdc, &m); err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		if !m.BaseReq.ValidateBasic(w) {
			return
		}

		// validate tx
		// discard error if it's CodeNoSignatures as the tx comes with no signatures
		if err := m.Tx.ValidateBasic(); err != nil && err.Code() != sdk.CodeNoSignatures {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		txBldr := authtxb.NewTxBuilder(
			utils.GetTxEncoder(cdc),
			m.BaseReq.AccountNumber,
			m.BaseReq.Sequence,
			m.Tx.Fee.Gas,
			1.0,
			false,
			m.BaseReq.ChainID,
			m.Tx.GetMemo(),
			m.Tx.Fee.Amount,
			nil,
		)

		signedTx, err := txBldr.SignStdTx(m.BaseReq.Name, m.BaseReq.Password, m.Tx, m.AppendSig)
		if keyerror.IsErrKeyNotFound(err) {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		} else if keyerror.IsErrWrongPassword(err) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		} else if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		utils.PostProcessResponse(w, cdc, signedTx, cliCtx.Indent)
	}
}
