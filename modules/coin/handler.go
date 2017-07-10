package coin

import (
	"github.com/tendermint/go-wire/data"
	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/modules/auth"
	"github.com/tendermint/basecoin/state"
)

//NameCoin - name space of the coin module
const NameCoin = "coin"

// Handler includes an accountant
type Handler struct {
	Accountant
}

var _ basecoin.Handler = Handler{}

// NewHandler - new accountant handler for the coin module
func NewHandler() Handler {
	return Handler{
		Accountant: NewAccountant(""),
	}
}

// Name - return name space
func (Handler) Name() string {
	return NameCoin
}

// CheckTx checks if there is enough money in the account
func (h Handler) CheckTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	send, err := checkTx(ctx, tx)
	if err != nil {
		return res, err
	}

	// now make sure there is money
	for _, in := range send.Inputs {
		_, err = h.CheckCoins(store, in.Address, in.Coins.Negative(), in.Sequence)
		if err != nil {
			return res, err
		}
	}

	// otherwise, we are good
	return res, nil
}

// DeliverTx moves the money
func (h Handler) DeliverTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	send, err := checkTx(ctx, tx)
	if err != nil {
		return res, err
	}

	// deduct from all input accounts
	for _, in := range send.Inputs {
		_, err = h.ChangeCoins(store, in.Address, in.Coins.Negative(), in.Sequence)
		if err != nil {
			return res, err
		}
	}

	// add to all output accounts
	for _, out := range send.Outputs {
		// note: sequence number is ignored when adding coins, only checked for subtracting
		_, err = h.ChangeCoins(store, out.Address, out.Coins, 0)
		if err != nil {
			return res, err
		}
	}

	// a-ok!
	return basecoin.Result{}, nil
}

// SetOption - sets the genesis account balance
func (h Handler) SetOption(l log.Logger, store state.KVStore, module, key, value string) (log string, err error) {
	if module != NameCoin {
		return "", errors.ErrUnknownModule(module)
	}
	if key == "account" {
		var acc GenesisAccount
		err = data.FromJSON([]byte(value), &acc)
		if err != nil {
			return "", err
		}
		acc.Balance.Sort()
		addr, err := acc.GetAddr()
		if err != nil {
			return "", ErrInvalidAddress()
		}
		// this sets the permission for a public key signature, use that app
		actor := auth.SigPerm(addr)
		err = storeAccount(store, h.MakeKey(actor), acc.ToAccount())
		if err != nil {
			return "", err
		}
		return "Success", nil

	}
	return errors.ErrUnknownKey(key)
}

func checkTx(ctx basecoin.Context, tx basecoin.Tx) (send SendTx, err error) {
	// check if the tx is proper type and valid
	send, ok := tx.Unwrap().(SendTx)
	if !ok {
		return send, errors.ErrInvalidFormat(tx)
	}
	err = send.ValidateBasic()
	if err != nil {
		return send, err
	}

	// check if all inputs have permission
	for _, in := range send.Inputs {
		if !ctx.HasPermission(in.Address) {
			return send, errors.ErrUnauthorized()
		}
	}
	return send, nil
}
