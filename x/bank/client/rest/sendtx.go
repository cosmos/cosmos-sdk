package rest

import (
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authctx "github.com/cosmos/cosmos-sdk/x/auth/client/context"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/bank/client"

	"github.com/gorilla/mux"
	"github.com/gin-gonic/gin"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"encoding/base64"
	"github.com/cosmos/cosmos-sdk/client/httputils"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"errors"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	r.HandleFunc("/accounts/{address}/send", SendRequestHandlerFn(cdc, kb, cliCtx)).Methods("POST")
}

type sendBody struct {
	// fees is not used currently
	// Fees             sdk.Coin  `json="fees"`
	Amount           sdk.Coins `json:"amount"`
	LocalAccountName string    `json:"name"`
	Password         string    `json:"password"`
	ChainID          string    `json:"chain_id"`
	AccountNumber    int64     `json:"account_number"`
	Sequence         int64     `json:"sequence"`
	Gas              int64     `json:"gas"`
	Fee				 string    `json:"fee"`
}

type transferBody struct {
	ChainID         string  `json:"chain_id"`
	FromAddress		string	`json:"from_address"`
	ToAddress		string	`json:"to_address"`
	Amount			sdk.Int `json:"amount"`
	Denomination 	string 	`json:"denomination"`
	AccountNumber	int64	`json:"account_number"`
	Sequence		int64	`json:"sequence"`
	EnsureAccAndSeq bool 	`json:"ensure_account_sequence"`
	Gas				int64	`json:"gas"`
	Fee				string  `json:"fee"`
}

type signedBody struct {
	TransferBody	transferBody	`json:"transfer_body"`
	Signature		[]byte			`json:"signature_list"`
	PublicKey		[]byte			`json:"public_key_list"`
}

var msgCdc = wire.NewCodec()

func init() {
	bank.RegisterWire(msgCdc)
}

// SendRequestHandlerFn - http request handler to send coins to a address
func SendRequestHandlerFn(cdc *wire.Codec, kb keys.Keybase, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// collect data
		vars := mux.Vars(r)
		bech32addr := vars["address"]

		to, err := sdk.AccAddressFromBech32(bech32addr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		var m sendBody
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		err = msgCdc.UnmarshalJSON(body, &m)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		info, err := kb.Get(m.LocalAccountName)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		// build message
		msg := client.BuildMsg(sdk.AccAddress(info.GetPubKey().Address()), to, m.Amount)
		if err != nil { // XXX rechecking same error ?
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		txCtx := authctx.TxContext{
			Codec:         cdc,
			Gas:           m.Gas,
			Fee:           m.Fee,
			ChainID:       m.ChainID,
			AccountNumber: m.AccountNumber,
			Sequence:      m.Sequence,
		}

		txBytes, err := txCtx.BuildAndSign(m.LocalAccountName, m.Password, []sdk.Msg{msg})
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		res, err := cliCtx.BroadcastTx(txBytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		output, err := wire.MarshalJSONIndent(cdc, res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}

// RegisterSwaggerRoutes - Central function to define routes that get registered by the main application
func RegisterSwaggerRoutes(routerGroup *gin.RouterGroup, ctx context.CLIContext, cdc *wire.Codec, kb keys.Keybase) {
	routerGroup.POST("/accounts/:address/send", sendRequestFn(cdc, ctx, kb))
	routerGroup.POST("/create_transfer", createTransferTxForSignFn(cdc, ctx))
	routerGroup.POST("/signed_transfer", composeAndBroadcastSignedTransferTxFn(cdc, ctx))
}

func composeTx(cdc *wire.Codec, ctx context.CLIContext, transferBody transferBody) (auth.StdSignMsg, authctx.TxContext, sdk.Error) {

	emptyMsg := auth.StdSignMsg{}
	emptyTxContext := authctx.TxContext{}

	amount := sdk.NewCoin(transferBody.Denomination, transferBody.Amount)
	var amounts sdk.Coins
	amounts = append(amounts, amount)

	fromAddress, err := sdk.AccAddressFromBech32(transferBody.FromAddress)
	if err != nil {
		return emptyMsg, emptyTxContext, sdk.ErrInvalidAddress(err.Error())
	}

	toAddress, err := sdk.AccAddressFromBech32(transferBody.ToAddress)
	if err != nil {
		return emptyMsg, emptyTxContext, sdk.ErrInvalidAddress(err.Error())
	}

	// build message
	msg := client.BuildMsg(fromAddress, toAddress, amounts)

	accountNumber := transferBody.AccountNumber
	sequence := transferBody.Sequence
	gas := transferBody.Gas
	fee := transferBody.Fee

	if transferBody.EnsureAccAndSeq {
		if ctx.AccDecoder == nil {
			ctx = ctx.WithAccountDecoder(authcmd.GetAccountDecoder(cdc))
		}
		accountNumber, err = ctx.GetAccountNumber(fromAddress)
		if err != nil {
			return emptyMsg, emptyTxContext, sdk.ErrInternal(err.Error())
		}

		sequence, err = ctx.GetAccountSequence(fromAddress)
		if err != nil {
			return emptyMsg, emptyTxContext, sdk.ErrInternal(err.Error())
		}
	}

	txCtx := authctx.TxContext{
		Codec:         cdc,
		Gas:           gas,
		Fee:           fee,
		ChainID:       transferBody.ChainID,
		AccountNumber: accountNumber,
		Sequence:      sequence,
	}

	txForSign, err := txCtx.Build([]sdk.Msg{msg})
	if err != nil {
		return emptyMsg, emptyTxContext, sdk.ErrInternal(err.Error())
	}

	return txForSign, txCtx, nil
}

// handler of creating transfer transaction
func createTransferTxForSignFn(cdc *wire.Codec, ctx context.CLIContext) gin.HandlerFunc {
	return func(gtx *gin.Context) {
		var transferBody transferBody
		if err := gtx.BindJSON(&transferBody); err != nil {
			httputils.NewError(gtx, http.StatusBadRequest, err)
			return
		}

		txForSign, _, err := composeTx(cdc, ctx, transferBody)
		if err != nil {
			if err.Code() == sdk.CodeInternal {
				httputils.NewError(gtx, http.StatusInternalServerError, err)
			} else {
				httputils.NewError(gtx, http.StatusBadRequest, err)
			}
			return
		}

		base64TxData := make([]byte, base64.StdEncoding.EncodedLen(len(txForSign.Bytes())))
		base64.StdEncoding.Encode(base64TxData,txForSign.Bytes())

		httputils.NormalResponse(gtx,string(base64TxData))
	}
}

// handler of composing and broadcasting transactions in swagger rest server
func composeAndBroadcastSignedTransferTxFn(cdc *wire.Codec, ctx context.CLIContext) gin.HandlerFunc {
	return func(gtx *gin.Context) {
		var signedTransaction signedBody
		if err := gtx.BindJSON(&signedTransaction); err != nil {
			httputils.NewError(gtx, http.StatusBadRequest, err)
			return
		}

		if signedTransaction.Signature == nil || signedTransaction.PublicKey == nil {
			httputils.NewError(gtx, http.StatusBadRequest, errors.New("signature or public key is empty"))
			return
		}

		signature, err := base64.StdEncoding.DecodeString(string(signedTransaction.Signature))
		if err != nil {
			httputils.NewError(gtx, http.StatusBadRequest, err)
			return
		}
		publicKey, err := base64.StdEncoding.DecodeString(string(signedTransaction.PublicKey))
		if err != nil {
			httputils.NewError(gtx, http.StatusBadRequest, err)
			return
		}

		txForSign, txCtx, errMsg := composeTx(cdc, ctx, signedTransaction.TransferBody)
		if errMsg != nil {
			if errMsg.Code() == sdk.CodeInternal {
				httputils.NewError(gtx, http.StatusInternalServerError, errMsg)
			} else {
				httputils.NewError(gtx, http.StatusBadRequest, errMsg)
			}
		}

		txDataForBroadcast, err := txCtx.BuildTxWithSignature(cdc, txForSign, signature, publicKey)
		if err != nil {
			httputils.NewError(gtx, http.StatusInternalServerError, err)
			return
		}

		res, err := ctx.BroadcastTx(txDataForBroadcast)
		if err != nil {
			httputils.NewError(gtx, http.StatusInternalServerError, err)
			return
		}

		httputils.NormalResponse(gtx,res)
	}
}

// handler of sending tokens in swagger rest server
func sendRequestFn(cdc *wire.Codec, ctx context.CLIContext, kb keys.Keybase) gin.HandlerFunc {
	return func(gtx *gin.Context) {

		bech32addr := gtx.Param("address")

		address, err := sdk.AccAddressFromBech32(bech32addr)
		if err != nil {
			httputils.NewError(gtx, http.StatusBadRequest, err)
			return
		}

		var m sendBody
		if err := gtx.BindJSON(&m); err != nil {
			httputils.NewError(gtx, http.StatusBadRequest, err)
			return
		}

		info, err := kb.Get(m.LocalAccountName)
		if err != nil {
			httputils.NewError(gtx, http.StatusUnauthorized, err)
			return
		}

		from := sdk.AccAddress(info.GetPubKey().Address())

		to, err := sdk.AccAddressFromBech32(address.String())
		if err != nil {
			httputils.NewError(gtx, http.StatusBadRequest, err)
			return
		}

		// build message
		msg := client.BuildMsg(from, to, m.Amount)
		if err != nil { // XXX rechecking same error ?
			httputils.NewError(gtx, http.StatusInternalServerError, err)
			return
		}

		txCtx := authctx.TxContext{
			Codec:         cdc,
			Gas:           m.Gas,
			Fee:           m.Fee,
			ChainID:       m.ChainID,
			AccountNumber: m.AccountNumber,
			Sequence:      m.Sequence,
		}

		txBytes, err := txCtx.BuildAndSign(m.LocalAccountName, m.Password, []sdk.Msg{msg})
		if err != nil {
			httputils.NewError(gtx, http.StatusUnauthorized, err)
			return
		}

		res, err := ctx.BroadcastTx(txBytes)
		if err != nil {
			httputils.NewError(gtx, http.StatusInternalServerError, err)
			return
		}

		httputils.NormalResponse(gtx,res)
	}
}