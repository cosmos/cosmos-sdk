package utils

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authctx "github.com/cosmos/cosmos-sdk/x/auth/client/context"
	"github.com/gin-gonic/gin"
	"net/http"
)

// SendTx implements a auxiliary handler that facilitates sending a series of
// messages in a signed transaction given a TxContext and a QueryContext. It
// ensures that the account exists, has a proper number and sequence set. In
// addition, it builds and signs a transaction with the supplied messages.
// Finally, it broadcasts the signed transaction to a node.
func SendTx(txCtx authctx.TxContext, cliCtx context.CLIContext, msgs []sdk.Msg) error {
	if err := cliCtx.EnsureAccountExists(); err != nil {
		return err
	}

	from, err := cliCtx.GetFromAddress()
	if err != nil {
		return err
	}

	// TODO: (ref #1903) Allow for user supplied account number without
	// automatically doing a manual lookup.
	if txCtx.AccountNumber == 0 {
		accNum, err := cliCtx.GetAccountNumber(from)
		if err != nil {
			return err
		}

		txCtx = txCtx.WithAccountNumber(accNum)
	}

	// TODO: (ref #1903) Allow for user supplied account sequence without
	// automatically doing a manual lookup.
	if txCtx.Sequence == 0 {
		accSeq, err := cliCtx.GetAccountSequence(from)
		if err != nil {
			return err
		}

		txCtx = txCtx.WithSequence(accSeq)
	}

	passphrase, err := keys.GetPassphrase(cliCtx.FromAddressName)
	if err != nil {
		return err
	}

	// build and sign the transaction
	txBytes, err := txCtx.BuildAndSign(cliCtx.FromAddressName, passphrase, msgs)
	if err != nil {
		return err
	}

	// broadcast to a Tendermint node
	return cliCtx.EnsureBroadcastTx(txBytes)
}

func NewError(ctx *gin.Context, errCode int, err error) {
	errorResponse := HTTPError{
		Api:	"2.0",
		Code:   errCode,
		ErrMsg: err.Error(),
	}
	ctx.JSON(errCode, errorResponse)
}

func Response(ctx *gin.Context, data interface{}) {
	response := HTTPResponse{
		Api:	"2.0",
		Code:   0,
		Result: data,
	}
	ctx.JSON(http.StatusOK, response)
}

type HTTPResponse struct {
	Api 	string 		`json:"rest api" example:"2.0"`
	Code    int    		`json:"code" example:"0"`
	Result 	interface{} `json:"result"`
}

type HTTPError struct {
	Api 	string 		`json:"rest api" example:"2.0"`
	Code    int    		`json:"code" example:"500"`
	ErrMsg 	string 		`json:"error message"`
}