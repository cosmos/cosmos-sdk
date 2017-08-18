package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/client/commands"
	"github.com/tendermint/basecoin/client/commands/query"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/modules/coin"
	"github.com/tendermint/basecoin/modules/nonce"
	"github.com/tendermint/basecoin/stack"
	wire "github.com/tendermint/go-wire"
	lightclient "github.com/tendermint/light-client"
	"github.com/tendermint/tmlibs/common"
)

// doQueryNonce is the HTTP handlerfunc to query a nonce
func doQueryNonce(w http.ResponseWriter, r *http.Request) {
	args := mux.Vars(r)
	signature := args["signature"]
	actor, err := commands.ParseActor(signature)
	if err != nil {
		common.WriteError(w, err)
		return
	}
	actor = coin.ChainAddr(actor)
	key := nonce.GetSeqKey([]basecoin.Actor{actor})
	key = stack.PrefixedKey(nonce.NameNonce, key)

	prove := !viper.GetBool(commands.FlagTrustNode)

	// query sequence number
	data, height, err := query.Get(key, prove)
	if lightclient.IsNoDataErr(err) {
		err = fmt.Errorf("nonce empty for address: %q", signature)
		common.WriteError(w, err)
		return
	} else if err != nil {
		common.WriteError(w, err)
		return
	}

	// unmarshal sequence number
	var seq uint32
	err = wire.ReadBinaryBytes(data, &seq)
	if err != nil {
		msg := fmt.Sprintf("Error reading sequence for %X", key)
		common.WriteError(w, errors.ErrInternal(msg))
		return
	}

	// write result to client
	if err := query.FoutputProof(w, seq, height); err != nil {
		common.WriteError(w, err)
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
