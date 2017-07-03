package stack

import (
	"fmt"

	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/types"
)

const (
	NameRecovery = "rcvr"
)

// Recovery catches any panics and returns them as errors instead
type Recovery struct{}

func (_ Recovery) Name() string {
	return NameRecovery
}

var _ Middleware = Recovery{}

func (_ Recovery) CheckTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx, next basecoin.Checker) (res basecoin.Result, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = normalizePanic(r)
		}
	}()
	return next.CheckTx(ctx, store, tx)
}

func (_ Recovery) DeliverTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx, next basecoin.Deliver) (res basecoin.Result, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = normalizePanic(r)
		}
	}()
	return next.DeliverTx(ctx, store, tx)
}

func (_ Recovery) SetOption(l log.Logger, store types.KVStore, key, value string, next basecoin.SetOptioner) (log string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = normalizePanic(r)
		}
	}()
	return next.SetOption(l, store, key, value)
}

// normalizePanic makes sure we can get a nice TMError (with stack) out of it
func normalizePanic(p interface{}) error {
	if err, isErr := p.(error); isErr {
		return errors.Wrap(err)
	}
	msg := fmt.Sprintf("%v", p)
	return errors.ErrInternal(msg)
}
