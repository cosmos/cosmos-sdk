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
	"github.com/cosmos/cosmos-sdk/client/httputils"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"bytes"
	"fmt"
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
	Name             string    `json:"name"`
	Password         string    `json:"password"`
	FromAddress		 string	   `json:"from_address"`
	ToAddress		 string	   `json:"to_address"`
	Amount           sdk.Coins `json:"amount"`
	ChainID          string    `json:"chain_id"`
	AccountNumber    int64     `json:"account_number"`
	Sequence         int64     `json:"sequence"`
	Gas              int64     `json:"gas"`
	Fee		         string    `json:"fee"`
	Signed		     bool      `json:"signed"` // true by default
	EnsureAccAndSeq  bool 	   `json:"ensure_account_sequence"`
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
	routerGroup.POST("/bank/transfers", transferRequestFn(cdc, ctx, kb))
}

// handler of creating transfer transaction
func transferRequestFn(cdc *wire.Codec, ctx context.CLIContext, kb keys.Keybase) gin.HandlerFunc {
	return func(gtx *gin.Context) {
		var transferBody transferBody
		body, err := ioutil.ReadAll(gtx.Request.Body)
		if err != nil {
			httputils.NewError(gtx, http.StatusBadRequest, err)
			return
		}
		err = cdc.UnmarshalJSON(body, &transferBody)
		if err != nil {
			httputils.NewError(gtx, http.StatusBadRequest, err)
			return
		}

		txForSign, _, errMsg := composeTx(cdc, ctx, transferBody)
		if err != nil {
			if errMsg.Code() == sdk.CodeInternal {
				httputils.NewError(gtx, http.StatusInternalServerError, err)
			} else {
				httputils.NewError(gtx, http.StatusBadRequest, err)
			}
			return
		}

		if !transferBody.Signed {
			httputils.NormalResponse(gtx, txForSign.Bytes())
			return
		}

		signAndBroadcase(gtx, cdc, ctx, txForSign, transferBody, kb)
	}
}

// signAndBroadcase perform transaction sign and broadcast operation
func signAndBroadcase(gtx *gin.Context, cdc *wire.Codec, ctx context.CLIContext, txForSign auth.StdSignMsg, transferBody transferBody, kb keys.Keybase) {

	if transferBody.Name == "" || transferBody.Password == "" {
		httputils.NewError(gtx, http.StatusBadRequest, fmt.Errorf("missing key name or password in signning transaction"))
		return
	}

	info, err := kb.Get(transferBody.Name)
	if err != nil {
		httputils.NewError(gtx, http.StatusBadRequest, err)
		return
	}

	fromAddress, err := sdk.AccAddressFromBech32(transferBody.FromAddress)
	if err != nil {
		httputils.NewError(gtx, http.StatusBadRequest, err)
		return
	}

	if !bytes.Equal(info.GetPubKey().Address(), fromAddress) {
		httputils.NewError(gtx, http.StatusBadRequest, fmt.Errorf("the fromAddress doesn't equal to the address of sign key"))
		return
	}

	sig, pubkey, err := kb.Sign(transferBody.Name, transferBody.Password, txForSign.Bytes())
	if err != nil {
		httputils.NewError(gtx, http.StatusInternalServerError, err)
		return
	}

	sigs := []auth.StdSignature{{
		AccountNumber: txForSign.AccountNumber,
		Sequence:      txForSign.Sequence,
		PubKey:        pubkey,
		Signature:     sig,
	}}

	txBytes, err := ctx.Codec.MarshalBinary(auth.NewStdTx(txForSign.Msgs, txForSign.Fee, sigs, txForSign.Memo))
	if err != nil {
		httputils.NewError(gtx, http.StatusInternalServerError, err)
		return
	}

	res, err := ctx.BroadcastTx(txBytes)
	if err != nil {
		httputils.NewError(gtx, http.StatusInternalServerError, err)
		return
	}

	output, err := wire.MarshalJSONIndent(cdc, res)
	if err != nil {
		httputils.NewError(gtx, http.StatusInternalServerError, err)
		return
	}

	httputils.NormalResponse(gtx, output)
}

// composeTx perform StdSignMsg building operation
func composeTx(cdc *wire.Codec, ctx context.CLIContext, transferBody transferBody) (auth.StdSignMsg, authctx.TxContext, sdk.Error) {

	emptyMsg := auth.StdSignMsg{}
	emptyTxContext := authctx.TxContext{}

	fromAddress, err := sdk.AccAddressFromBech32(transferBody.FromAddress)
	if err != nil {
		return emptyMsg, emptyTxContext, sdk.ErrInvalidAddress(err.Error())
	}

	toAddress, err := sdk.AccAddressFromBech32(transferBody.ToAddress)
	if err != nil {
		return emptyMsg, emptyTxContext, sdk.ErrInvalidAddress(err.Error())
	}

	// build message
	msg := client.BuildMsg(fromAddress, toAddress, transferBody.Amount)

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