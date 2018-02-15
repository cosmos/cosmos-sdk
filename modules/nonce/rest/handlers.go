package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	wire "github.com/tendermint/go-wire"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/client/commands/query"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/modules/nonce"
	"github.com/cosmos/cosmos-sdk/stack"
)

// doQueryNonce is the HTTP handlerfunc to query a nonce
func doQueryNonce(w http.ResponseWriter, r *http.Request) {
	args := mux.Vars(r)
	signature := args["signature"]
	actor, err := commands.ParseActor(signature)
	if err != nil {
		sdk.WriteError(w, err)
		return
	}

	var h int64
	qHeight := r.URL.Query().Get("height")
	if qHeight != "" {
		_h, err := strconv.Atoi(qHeight)
		if err != nil {
			sdk.WriteError(w, err)
			return
		}
		h = int64(_h)
	}

	actor = coin.ChainAddr(actor)
	key := nonce.GetSeqKey([]sdk.Actor{actor})
	key = stack.PrefixedKey(nonce.NameNonce, key)

	prove := !viper.GetBool(commands.FlagTrustNode)

	// query sequence number
	data, height, err := query.Get(key, h, prove)
	if client.IsNoDataErr(err) {
		err = fmt.Errorf("nonce empty for address: %q", signature)
		sdk.WriteError(w, err)
		return
	} else if err != nil {
		sdk.WriteError(w, err)
		return
	}

	// unmarshal sequence number
	var seq uint32
	err = wire.ReadBinaryBytes(data, &seq)
	if err != nil {
		msg := fmt.Sprintf("Error reading sequence for %X", key)
		sdk.WriteError(w, errors.ErrInternal(msg))
		return
	}

	// write result to client
	if err := query.FoutputProof(w, seq, height); err != nil {
		sdk.WriteError(w, err)
	}
}

// mux.Router registrars

// RegisterQueryNonce is a mux.Router handler that exposes GET
// method access on route /query/nonce/{signature} to query nonces
func RegisterQueryNonce(r *mux.Router) error {
	r.HandleFunc("/query/nonce/{signature}", doQueryNonce).Methods("GET")
	return nil
}

// End of mux.Router registrars
