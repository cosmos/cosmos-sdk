package rest

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/examples/covenantcoin/x/covenant"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/gorilla/mux"
	"github.com/tendermint/go-crypto/keys"
	"io/ioutil"
	"net/http"
)

func RegisterRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	r.HandleFunc("/covenant/create", CreateHandlerFn(ctx, cdc, kb)).Methods("POST")
	r.HandleFunc("/covenant/settle", SettleHandlerFn(ctx, cdc, kb)).Methods("POST")
	r.HandleFunc("/covenant/get/{id}", GetHandlerFn(ctx, cdc)).Methods("GET")
}

type authedBody struct {
	LocalAccountName string `json:"name"`
	Password         string `json:"password"`
	ChainID          string `json:"chain_id"`
	Sequence         int64  `json:"sequence"`
}

type createBody struct {
	Amount           sdk.Coins     `json:"amount"`
	Settlers         []sdk.Address `json:"settlers"`
	Receivers        []sdk.Address `json:"receivers"`
	LocalAccountName string        `json:"name"`
	Password         string        `json:"password"`
	ChainID          string        `json:"chain_id"`
	Sequence         int64         `json:"sequence"`
}

type createResponse struct {
	CovID int64 `json:"covid"`
}

func CreateHandlerFn(ctx context.CoreContext, cdc *wire.Codec, kb keys.Keybase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m createBody
		body, _ := ioutil.ReadAll(r.Body)
		_ = cdc.UnmarshalJSON(body, m)
		info, _ := kb.Get(m.LocalAccountName)
		msg := covenant.MsgCreateCovenant{info.Address(), m.Settlers, m.Receivers, m.Amount}
		txBytes, _ := ctx.SignAndBuild(m.LocalAccountName, m.Password, msg, cdc)
		res, _ := ctx.BroadcastTx(txBytes)
		newCovID := new(int64)
		_ = cdc.UnmarshalBinary(res.DeliverTx.Data, newCovID)
		resp := createResponse{*newCovID}
		bz, _ := cdc.MarshalJSON(resp)
		w.Write(bz)
	}
}

type settleBody struct {
	CovID            int64       `json:"covid"`
	Receiver         sdk.Address `json:"receiver"`
	LocalAccountName string      `json:"name"`
	Password         string      `json:"password"`
	ChainID          string      `json:"chain_id"`
	Sequence         int64       `json:"sequence"`
}

func SettleHandlerFn(ctx context.CoreContext, cdc *wire.Codec, kb keys.Keybase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m settleBody
		body, _ := ioutil.ReadAll(r.Body)
		_ = cdc.UnmarshalJSON(body, m)
		info, _ := kb.Get(m.LocalAccountName)
		msg := covenant.MsgSettleCovenant{m.CovID, info.Address(), m.Receiver}
		txBytes, _ := ctx.SignAndBuild(m.LocalAccountName, m.Password, msg, cdc)
		_, _ = ctx.BroadcastTx(txBytes)
	}
}

func GetHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
