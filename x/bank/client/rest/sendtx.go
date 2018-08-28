package rest

import (
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
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
	"github.com/tendermint/tendermint/libs/bech32"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	r.HandleFunc("/bank/transfers", SendRequestHandlerFn(cdc, kb, cliCtx)).Methods("POST")
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
	Generate		 bool      `json:"generate"`
	EnsureAccAndSeq  bool 	   `json:"ensure_account_sequence"`
}

var msgCdc = wire.NewCodec()

func init() {
	bank.RegisterWire(msgCdc)
}

// SendRequestHandlerFn - http request handler to send coins to a address
func SendRequestHandlerFn(cdc *wire.Codec, kb keys.Keybase, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var transferBody transferBody
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			utils.WriteErrorResponse(&w, http.StatusBadRequest, err.Error())
			return
		}
		err = msgCdc.UnmarshalJSON(body, &transferBody)
		if err != nil {
			utils.WriteErrorResponse(&w, http.StatusBadRequest, err.Error())
			return
		}
		transferBody, errCode, err := paramPreprocess(transferBody, kb)
		if err != nil {
			utils.WriteErrorResponse(&w, errCode, err.Error())
			return
		}

		txForSign, _, errMsg := composeTx(cdc, cliCtx, transferBody)
		if err != nil {
			if errMsg.Code() == sdk.CodeInternal {
				utils.WriteErrorResponse(&w, http.StatusInternalServerError, err.Error())
			} else {
				utils.WriteErrorResponse(&w, http.StatusBadRequest, err.Error())
			}
			return
		}

		if transferBody.Generate {
			w.Write(txForSign.Bytes())
			return
		}

		output, errCode, err := signAndBroadcase(cdc, cliCtx, txForSign, transferBody, kb)
		if err != nil {
			w.WriteHeader(errCode)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(output)
	}
}

// registerSwaggerTxRoutes - Central function to define routes that get registered by the main application
func registerSwaggerTxRoutes(routerGroup *gin.RouterGroup, ctx context.CLIContext, cdc *wire.Codec, kb keys.Keybase) {
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
		transferBody, errCode, err := paramPreprocess(transferBody, kb)
		if err != nil {
			httputils.NewError(gtx, errCode, err)
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

		if transferBody.Generate {
			httputils.NormalResponse(gtx, txForSign.Bytes())
			return
		}

		output, errCode, err := signAndBroadcase(cdc, ctx, txForSign, transferBody, kb)
		if err != nil {
			httputils.NewError(gtx, errCode, err)
			return
		}
		httputils.NormalResponse(gtx, output)
	}
}

// paramPreprocess performs transferBody preprocess
func paramPreprocess(body transferBody, kb keys.Keybase) (transferBody, int, error) {
	if body.Name == "" {
		if !body.Generate {
			return transferBody{}, http.StatusBadRequest, fmt.Errorf("missing key name, can't sign transaction")
		}
		if body.FromAddress == "" {
			return transferBody{}, http.StatusBadRequest, fmt.Errorf("both the key name and fromAddreass are missed")
		}
	}

	if body.Name != "" {
		info, err := kb.Get(body.Name)
		if err != nil {
			return transferBody{}, http.StatusBadRequest, err
		}
		if body.FromAddress == "" {
			addressFromPubKey, err := bech32.ConvertAndEncode(sdk.Bech32PrefixAccAddr, info.GetPubKey().Address().Bytes())
			if err != nil {
				return transferBody{}, http.StatusInternalServerError, err
			}
			body.FromAddress = addressFromPubKey
		} else {
			fromAddress, err := sdk.AccAddressFromBech32(body.FromAddress)
			if err != nil {
				return transferBody{}, http.StatusBadRequest, err
			}

			if !bytes.Equal(info.GetPubKey().Address(), fromAddress) {
				return transferBody{}, http.StatusBadRequest, fmt.Errorf("the fromAddress doesn't equal to the address of sign key")
			}
		}
	}
	return body, 0, nil
}

// signAndBroadcase perform transaction sign and broadcast operation
func signAndBroadcase(cdc *wire.Codec, ctx context.CLIContext, txForSign auth.StdSignMsg, transferBody transferBody, kb keys.Keybase) ([]byte, int, error) {

	if transferBody.Name == "" || transferBody.Password == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("missing key name or password in signning transaction")
	}

	sig, pubkey, err := kb.Sign(transferBody.Name, transferBody.Password, txForSign.Bytes())
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	sigs := []auth.StdSignature{{
		AccountNumber: txForSign.AccountNumber,
		Sequence:      txForSign.Sequence,
		PubKey:        pubkey,
		Signature:     sig,
	}}

	txBytes, err := ctx.Codec.MarshalBinary(auth.NewStdTx(txForSign.Msgs, txForSign.Fee, sigs, txForSign.Memo))
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	res, err := ctx.BroadcastTx(txBytes)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	output, err := wire.MarshalJSONIndent(cdc, res)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return output, 0, nil
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

	if txCtx.Gas == 0 {
		newCtx, err := utils.EnrichCtxWithGas(txCtx, ctx, transferBody.Name, transferBody.Password, []sdk.Msg{msg})
		if err != nil {
			return emptyMsg, emptyTxContext, sdk.ErrInternal(err.Error())
		}
		txCtx = newCtx
	}

	txForSign, err := txCtx.Build([]sdk.Msg{msg})
	if err != nil {
		return emptyMsg, emptyTxContext, sdk.ErrInternal(err.Error())
	}

	return txForSign, txCtx, nil
}