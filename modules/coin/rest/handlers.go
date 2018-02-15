package rest

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/client/commands/query"
	"github.com/cosmos/cosmos-sdk/client/commands/search"
	"github.com/cosmos/cosmos-sdk/modules/auth"
	"github.com/cosmos/cosmos-sdk/modules/base"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/modules/fee"
	"github.com/cosmos/cosmos-sdk/modules/nonce"
	"github.com/cosmos/cosmos-sdk/stack"
)

// SendInput is the request to send an amount from one actor to another.
// Note: Not using the `validator:""` tags here because SendInput has
// many fields so it would be nice to figure out all the invalid
// inputs and report them back to the caller, in one shot.
type SendInput struct {
	Fees     *coin.Coin `json:"fees"`
	Multi    bool       `json:"multi,omitempty"`
	Sequence uint32     `json:"sequence"`

	To     *sdk.Actor `json:"to"`
	From   *sdk.Actor `json:"from"`
	Amount coin.Coins `json:"amount"`
}

// doQueryAccount is the HTTP handlerfunc to query an account
// It expects a query string with
func doQueryAccount(w http.ResponseWriter, r *http.Request) {
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
	key := stack.PrefixedKey(coin.NameCoin, actor.Bytes())
	account := new(coin.Account)
	prove := !viper.GetBool(commands.FlagTrustNode)
	height, err := query.GetParsed(key, account, h, prove)
	if client.IsNoDataErr(err) {
		err := fmt.Errorf("account bytes are empty for address: %q", signature)
		sdk.WriteError(w, err)
		return
	} else if err != nil {
		sdk.WriteError(w, err)
		return
	}

	if err := query.FoutputProof(w, account, height); err != nil {
		sdk.WriteError(w, err)
	}
}

// doQueryAccount is the HTTP handlerfunc to search for
// all SendTx transactions with this account as sender
// or receiver
func doSearchSent(w http.ResponseWriter, r *http.Request) {
	args := mux.Vars(r)
	account := args["account"]
	actor, err := commands.ParseActor(account)
	if err != nil {
		sdk.WriteError(w, err)
		return
	}

	// TODO: handle minHeight...
	// var h int
	// qHeight := r.URL.Query().Get("height")
	// if qHeight != "" {
	// 	h, err = strconv.Atoi(qHeight)
	// 	if err != nil {
	// 		common.WriteError(w, err)
	// 		return
	// 	}
	// }

	findSender := fmt.Sprintf("coin.sender='%s'", actor)
	findReceiver := fmt.Sprintf("coin.receiver='%s'", actor)
	prove := !viper.GetBool(commands.FlagTrustNode)
	all, err := search.FindAnyTx(prove, findSender, findReceiver)
	if err != nil {
		sdk.WriteError(w, err)
		return
	}

	// format....
	output, err := search.FormatSearch(all, coin.ExtractCoinTx)
	if err != nil {
		sdk.WriteError(w, err)
		return
	}

	// display
	if err := search.Foutput(w, output); err != nil {
		sdk.WriteError(w, err)
	}
}

func PrepareSendTx(si *SendInput) sdk.Tx {
	tx := coin.NewSendOneTx(*si.From, *si.To, si.Amount)
	// fees are optional
	if si.Fees != nil && !si.Fees.IsZero() {
		tx = fee.NewFee(tx, *si.Fees, *si.From)
	}
	// only add the actual signer to the nonce
	signers := []sdk.Actor{*si.From}
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
	if err := sdk.ParseRequestAndValidateJSON(r, si); err != nil {
		sdk.WriteError(w, err)
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
		err := &sdk.ErrorResponse{
			Err:  strings.Join(errsList, ", "),
			Code: code,
		}
		sdk.WriteCode(w, err, code)
		return
	}

	tx := PrepareSendTx(si)
	sdk.WriteSuccess(w, tx)
}

// mux.Router registrars

// RegisterCoinSend is a mux.Router handler that exposes
// POST method access on route /build/send to create a
// transaction for sending money from one account to another.
func RegisterCoinSend(r *mux.Router) error {
	r.HandleFunc("/build/send", doSend).Methods("POST")
	return nil
}

// RegisterQueryAccount is a mux.Router handler that exposes GET
// method access on route /query/account/{signature} to query accounts
func RegisterQueryAccount(r *mux.Router) error {
	r.HandleFunc("/query/account/{signature}", doQueryAccount).Methods("GET")
	return nil
}

// RegisterSearchSent is a mux.Router handler that exposes GET
// method access on route /tx/coin/{account} to historical sendtx transactions
func RegisterSearchSent(r *mux.Router) error {
	r.HandleFunc("/tx/coin/{account}", doSearchSent).Methods("GET")
	return nil
}

// RegisterAll is a convenience function to
// register all the  handlers in this package.
func RegisterAll(r *mux.Router) error {
	funcs := []func(*mux.Router) error{
		RegisterCoinSend,
		RegisterQueryAccount,
		RegisterSearchSent,
	}

	for _, fn := range funcs {
		if err := fn(r); err != nil {
			return err
		}
	}
	return nil
}

// End of mux.Router registrars
