package stack

import (
	"fmt"

	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/state"
)

// nolint
const (
	NameRecovery = "rcvr"
)

// Recovery catches any panics and returns them as errors instead
type Recovery struct{}

// Name of the module - fulfills Middleware interface
func (Recovery) Name() string {
	return NameRecovery
}

var _ Middleware = Recovery{}

// CheckTx catches any panic and converts to error - fulfills Middlware interface
func (Recovery) CheckTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx, next basecoin.Checker) (res basecoin.CheckResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = normalizePanic(r)
		}
	}()
	return next.CheckTx(ctx, store, tx)
}

// DeliverTx catches any panic and converts to error - fulfills Middlware interface
func (Recovery) DeliverTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx, next basecoin.Deliver) (res basecoin.DeliverResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = normalizePanic(r)
		}
	}()
	return next.DeliverTx(ctx, store, tx)
}

// InitState catches any panic and converts to error - fulfills Middlware interface
func (Recovery) InitState(l log.Logger, store state.SimpleDB, module, key, value string, next basecoin.InitStater) (log string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = normalizePanic(r)
		}
	}()
	return next.InitState(l, store, module, key, value)
}

// normalizePanic makes sure we can get a nice TMError (with stack) out of it
func normalizePanic(p interface{}) error {
	if err, isErr := p.(error); isErr {
		return errors.Wrap(err)
	}
	msg := fmt.Sprintf("%v", p)
	return errors.ErrInternal(msg)
}
