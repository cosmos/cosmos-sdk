package rest

import (
	"bytes"
	"encoding/json"
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

type editDelegationsBody struct {
	LocalAccountName string              `json:"name"`
	Password         string              `json:"password"`
	ChainID          string              `json:"chain_id"`
	Sequence         int64               `json:"sequence"`
	Delegate         []stake.MsgDelegate `json:"delegate"`
	Unbond           []stake.MsgUnbond   `json:"unbond"`
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
		messages := make([]sdk.Msg, 0, len(m.Delegate)+len(m.Unbond))
		for _, msg := range m.Delegate {
			if !bytes.Equal(info.Address(), msg.DelegatorAddr) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Must use own delegator address"))
				return
			}
			messages = append(messages, msg)
		}
		for _, msg := range m.Unbond {
			if !bytes.Equal(info.Address(), msg.DelegatorAddr) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Must use own delegator address"))
				return
			}
			messages = append(messages, msg)
		}

		// sign messages
		signedTxs := make([][]byte, 0, len(messages))
		for _, msg := range messages {
			// increment sequence for each message
			ctx = ctx.WithSequence(m.Sequence)
			m.Sequence++

			txBytes, err := ctx.SignAndBuild(m.LocalAccountName, m.Password, msg, cdc)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(err.Error()))
				return
			}

			signedTxs = append(signedTxs, txBytes)
		}

		// send
		// XXX the operation might not be atomic if a tx fails
		//     should we have a sdk.MultiMsg type to make sending atomic?
		results := make([]*ctypes.ResultBroadcastTxCommit, 0, len(signedTxs))
		for _, txBytes := range signedTxs {
			res, err := ctx.BroadcastTx(txBytes)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			results = append(results, res)
		}

		output, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}
