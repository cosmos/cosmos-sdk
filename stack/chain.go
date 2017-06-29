package stack

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/txs"
	"github.com/tendermint/basecoin/types"
)

const (
	NameChain = "chan"
)

// Chain enforces that this tx was bound to the named chain
type Chain struct {
	ChainID string
}

func (_ Chain) Name() string {
	return NameRecovery
}

var _ Middleware = Chain{}

func (c Chain) CheckTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx, next basecoin.Checker) (res basecoin.Result, err error) {
	stx, err := c.checkChain(tx)
	if err != nil {
		return res, err
	}
	return next.CheckTx(ctx, store, stx)
}

func (c Chain) DeliverTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx, next basecoin.Deliver) (res basecoin.Result, err error) {
	stx, err := c.checkChain(tx)
	if err != nil {
		return res, err
	}
	return next.DeliverTx(ctx, store, stx)
}

// checkChain makes sure the tx is a txs.Chain and
func (c Chain) checkChain(tx basecoin.Tx) (basecoin.Tx, error) {
	ctx, ok := tx.Unwrap().(*txs.Chain)
	if !ok {
		return tx, errors.NoChain()
	}
	if ctx.ChainID != c.ChainID {
		return tx, errors.WrongChain(ctx.ChainID)
	}
	return ctx.Tx, nil
}
