package coin

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/types"
)

const (
	NameCoin = "coin"
)

// Handler writes
type Handler struct{}

var _ basecoin.Handler = Handler{}

func (_ Handler) Name() string {
	return NameCoin
}

// CheckTx checks if there is enough money in the account
func (h Handler) CheckTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	_, err = checkTx(ctx, tx)
	if err != nil {
		return res, err
	}

	// now make sure there is money

	// otherwise, we are good
	return res, nil
}

// DeliverTx moves the money
func (h Handler) DeliverTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	_, err = checkTx(ctx, tx)
	if err != nil {
		return res, err
	}

	// now move the money
	return basecoin.Result{}, nil
}

func checkTx(ctx basecoin.Context, tx basecoin.Tx) (*SendTx, error) {
	// check if the tx is proper type and valid
	send, ok := tx.Unwrap().(*SendTx)
	if !ok {
		return nil, errors.UnknownTxType(tx)
	}
	err := send.ValidateBasic()
	if err != nil {
		return nil, err
	}

	// check if all inputs have permission
	for _, in := range send.Inputs {
		if !ctx.HasPermission(in.Address) {
			return nil, errors.Unauthorized()
		}
	}
	return send, nil
}
