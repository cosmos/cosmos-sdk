package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tendermint/go-crypto/keys"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

func registerTxRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	r.HandleFunc(
		"/stake/delegations",
		editDelegationsRequestHandlerFn(cdc, kb, ctx),
	).Methods("POST")
}

type msgDelegateInput struct {
	DelegatorAddr string   `json:"delegator_addr"` // in bech32
	ValidatorAddr string   `json:"validator_addr"` // in bech32
	Bond          sdk.Coin `json:"bond"`
}
type msgUnbondInput struct {
	DelegatorAddr string `json:"delegator_addr"` // in bech32
	ValidatorAddr string `json:"validator_addr"` // in bech32
	Shares        string `json:"shares"`
}

type editDelegationsBody struct {
	LocalAccountName string             `json:"name"`
	Password         string             `json:"password"`
	ChainID          string             `json:"chain_id"`
	AccountNumber    int64              `json:"account_number"`
	Sequence         int64              `json:"sequence"`
	Gas              int64              `json:"gas"`
	Delegate         []msgDelegateInput `json:"delegate"`
	Unbond           []msgUnbondInput   `json:"unbond"`
}

func editDelegationsRequestHandlerFn(cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m editDelegationsBody
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		err = json.Unmarshal(body, &m)
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

		// build messages
		messages := make([]sdk.Msg, len(m.Delegate)+len(m.Unbond))
		i := 0
		for _, msg := range m.Delegate {
			delegatorAddr, err := sdk.GetAccAddressBech32(msg.DelegatorAddr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("couldn't decode delegator. Error: %s", err.Error())))
				return
			}
			validatorAddr, err := sdk.GetValAddressBech32(msg.ValidatorAddr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("couldn't decode validator. Error: %s", err.Error())))
				return
			}
			if !bytes.Equal(info.Address(), delegatorAddr) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("must use own delegator address"))
				return
			}
			messages[i] = stake.MsgDelegate{
				DelegatorAddr: delegatorAddr,
				ValidatorAddr: validatorAddr,
				Bond:          msg.Bond,
			}
			i++
		}
		for _, msg := range m.Unbond {
			delegatorAddr, err := sdk.GetAccAddressBech32(msg.DelegatorAddr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("couldn't decode delegator. Error: %s", err.Error())))
				return
			}
			validatorAddr, err := sdk.GetValAddressBech32(msg.ValidatorAddr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("couldn't decode validator. Error: %s", err.Error())))
				return
			}
			if !bytes.Equal(info.Address(), delegatorAddr) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("must use own delegator address"))
				return
			}
			messages[i] = stake.MsgUnbond{
				DelegatorAddr: delegatorAddr,
				ValidatorAddr: validatorAddr,
				Shares:        msg.Shares,
			}
			i++
		}

		// add gas to context
		ctx = ctx.WithGas(m.Gas)

		// sign messages
		signedTxs := make([][]byte, len(messages[:]))
		for i, msg := range messages {
			// increment sequence for each message
			ctx = ctx.WithAccountNumber(m.AccountNumber)
			ctx = ctx.WithSequence(m.Sequence)
			m.Sequence++

			txBytes, err := ctx.SignAndBuild(m.LocalAccountName, m.Password, msg, cdc)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(err.Error()))
				return
			}

			signedTxs[i] = txBytes
		}

		// send
		// XXX the operation might not be atomic if a tx fails
		//     should we have a sdk.MultiMsg type to make sending atomic?
		results := make([]*ctypes.ResultBroadcastTxCommit, len(signedTxs[:]))
		for i, txBytes := range signedTxs {
			res, err := ctx.BroadcastTx(txBytes)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			results[i] = res
		}

		output, err := json.MarshalIndent(results[:], "", "  ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}
