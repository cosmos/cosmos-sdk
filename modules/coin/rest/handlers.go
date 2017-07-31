package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/client/commands"
	"github.com/tendermint/basecoin/client/commands/proofs"
	"github.com/tendermint/basecoin/modules/auth"
	"github.com/tendermint/basecoin/modules/base"
	"github.com/tendermint/basecoin/modules/coin"
	"github.com/tendermint/basecoin/modules/fee"
	"github.com/tendermint/basecoin/modules/nonce"
	"github.com/tendermint/basecoin/stack"
	lightclient "github.com/tendermint/light-client"
	"github.com/tendermint/tmlibs/common"
)

// SendInput is the request to send an amount from one actor to another.
// Note: Not using the `validator:""` tags here because SendInput has
// many fields so it would be nice to figure out all the invalid
// inputs and report them back to the caller, in one shot.
type SendInput struct {
	Fees     *coin.Coin `json:"fees"`
	Multi    bool       `json:"multi,omitempty"`
	Sequence uint32     `json:"sequence"`

	To     *basecoin.Actor `json:"to"`
	From   *basecoin.Actor `json:"from"`
	Amount coin.Coins      `json:"amount"`
}

func RegisterHandlers(r *mux.Router) error {
	r.HandleFunc("/build/send", doSend).Methods("POST")
	r.HandleFunc("/query/account/{signature}", doQueryAccount).Methods("GET")
	return nil
}

// doQueryAccount is the HTTP handlerfunc to query an account
// It expects a query string with
func doQueryAccount(w http.ResponseWriter, r *http.Request) {
	query := mux.Vars(r)
	signature := query["signature"]
	actor, err := commands.ParseActor(signature)
	if err != nil {
		common.WriteError(w, err)
		return
	}
	actor = coin.ChainAddr(actor)
	key := stack.PrefixedKey(coin.NameCoin, actor.Bytes())
	account := new(coin.Account)
	proof, err := proofs.GetAndParseAppProof(key, account)
	if lightclient.IsNoDataErr(err) {
		err := fmt.Errorf("account bytes are empty for address: %q", signature)
		common.WriteError(w, err)
		return
	} else if err != nil {
		common.WriteError(w, err)
		return
	}

	if err := proofs.FoutputProof(w, account, proof.BlockHeight()); err != nil {
		common.WriteError(w, err)
	}
}

func PrepareSendTx(si *SendInput) basecoin.Tx {
	tx := coin.NewSendOneTx(*si.From, *si.To, si.Amount)
	// fees are optional
	if si.Fees != nil && !si.Fees.IsZero() {
		tx = fee.NewFee(tx, *si.Fees, *si.From)
	}
	// only add the actual signer to the nonce
	signers := []basecoin.Actor{*si.From}
	tx = nonce.NewTx(si.Sequence, signers, tx)
	tx = base.NewChainTx(commands.GetChainID(), 0, tx)

	if si.Multi {
		tx = auth.NewMulti(tx).Wrap()
	} else {
		tx = auth.NewSig(tx).Wrap()
	}
	return tx
}

func doSend(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	si := new(SendInput)
	if err := common.ParseRequestAndValidateJSON(r, si); err != nil {
		common.WriteError(w, err)
		return
	}

	var errsList []string
	if si.From == nil {
		errsList = append(errsList, `"from" cannot be nil`)
	}
	if si.Sequence <= 0 {
		errsList = append(errsList, `"sequence" must be > 0`)
	}
	if si.To == nil {
		errsList = append(errsList, `"to" cannot be nil`)
	}
	if len(si.Amount) == 0 {
		errsList = append(errsList, `"amount" cannot be empty`)
	}
	if len(errsList) > 0 {
		code := http.StatusBadRequest
		err := &common.ErrorResponse{
			Err:  strings.Join(errsList, ", "),
			Code: code,
		}
		common.WriteCode(w, err, code)
		return
	}

	tx := PrepareSendTx(si)
	common.WriteSuccess(w, tx)
}
