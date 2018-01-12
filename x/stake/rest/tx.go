package rest

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/tmlibs/common"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/modules/auth"
	"github.com/cosmos/cosmos-sdk/modules/base"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/modules/fee"
	"github.com/cosmos/cosmos-sdk/modules/nonce"
	"github.com/cosmos/gaia/modules/stake"
)

const (
	//parameters used in urls
	paramPubKey = "pubkey"
	paramAmount = "amount"
	paramShares = "shares"

	paramName    = "name"
	paramKeybase = "keybase"
	paramWebsite = "website"
	paramDetails = "details"
)

type delegateInput struct {
	Fees     *coin.Coin `json:"fees"`
	Sequence uint32     `json:"sequence"`

	Pubkey crypto.PubKey `json:"pub_key"`
	From   *sdk.Actor    `json:"from"`
	Amount coin.Coin     `json:"amount"`
}

type unbondInput struct {
	Fees     *coin.Coin `json:"fees"`
	Sequence uint32     `json:"sequence"`

	Pubkey crypto.PubKey `json:"pub_key"`
	From   *sdk.Actor    `json:"from"`
	Shares string        `json:"amount"`
}

// RegisterDelegate is a mux.Router handler that exposes
// POST method access on route /tx/stake/delegate to create a
// transaction for delegate to a candidaate/validator
func RegisterDelegate(r *mux.Router) error {
	r.HandleFunc("/build/stake/delegate", delegate).Methods("POST")
	return nil
}

// RegisterUnbond is a mux.Router handler that exposes
// POST method access on route /tx/stake/unbond to create a
// transaction for unbonding delegated coins
func RegisterUnbond(r *mux.Router) error {
	r.HandleFunc("/build/stake/unbond", unbond).Methods("POST")
	return nil
}

func prepareDelegateTx(di *delegateInput) sdk.Tx {
	tx := stake.NewTxDelegate(di.Amount, di.Pubkey)
	// fees are optional
	if di.Fees != nil && !di.Fees.IsZero() {
		tx = fee.NewFee(tx, *di.Fees, *di.From)
	}
	// only add the actual signer to the nonce
	signers := []sdk.Actor{*di.From}
	tx = nonce.NewTx(di.Sequence, signers, tx)
	tx = base.NewChainTx(commands.GetChainID(), 0, tx)

	tx = auth.NewSig(tx).Wrap()
	return tx
}

func delegate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	di := new(delegateInput)
	if err := common.ParseRequestAndValidateJSON(r, di); err != nil {
		common.WriteError(w, err)
		return
	}

	var errsList []string
	if di.From == nil {
		errsList = append(errsList, `"from" cannot be nil`)
	}
	if di.Sequence <= 0 {
		errsList = append(errsList, `"sequence" must be > 0`)
	}
	if di.Pubkey.Empty() {
		errsList = append(errsList, `"pubkey" cannot be empty`)
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

	tx := prepareDelegateTx(di)
	common.WriteSuccess(w, tx)
}

func prepareUnbondTx(ui *unbondInput) sdk.Tx {
	tx := stake.NewTxUnbond(ui.Shares, ui.Pubkey)
	// fees are optional
	if ui.Fees != nil && !ui.Fees.IsZero() {
		tx = fee.NewFee(tx, *ui.Fees, *ui.From)
	}
	// only add the actual signer to the nonce
	signers := []sdk.Actor{*ui.From}
	tx = nonce.NewTx(ui.Sequence, signers, tx)
	tx = base.NewChainTx(commands.GetChainID(), 0, tx)

	tx = auth.NewSig(tx).Wrap()
	return tx
}

func unbond(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ui := new(unbondInput)
	if err := common.ParseRequestAndValidateJSON(r, ui); err != nil {
		common.WriteError(w, err)
		return
	}

	var errsList []string
	if ui.From == nil {
		errsList = append(errsList, `"from" cannot be nil`)
	}
	if ui.Sequence <= 0 {
		errsList = append(errsList, `"sequence" must be > 0`)
	}
	if ui.Pubkey.Empty() {
		errsList = append(errsList, `"pubkey" cannot be empty`)
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

	tx := prepareUnbondTx(ui)
	common.WriteSuccess(w, tx)
}
