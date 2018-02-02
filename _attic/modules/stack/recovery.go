package stack

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/log"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/state"
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
func (Recovery) CheckTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx, next sdk.Checker) (res sdk.CheckResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = normalizePanic(r)
		}
	}()
	return next.CheckTx(ctx, store, tx)
}

// DeliverTx catches any panic and converts to error - fulfills Middlware interface
func (Recovery) DeliverTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx, next sdk.Deliver) (res sdk.DeliverResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = normalizePanic(r)
		}
	}()
	return next.DeliverTx(ctx, store, tx)
}

// InitState catches any panic and converts to error - fulfills Middlware interface
func (Recovery) InitState(l log.Logger, store state.SimpleDB, module, key, value string, next sdk.InitStater) (log string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = normalizePanic(r)
		}
	}()
	return next.InitState(l, store, module, key, value)
}

// InitValidate catches any panic and logs it
// TODO: return an error???
func (Recovery) InitValidate(l log.Logger, store state.SimpleDB,
	vals []*abci.Validator, next sdk.InitValidater) {

	defer func() {
		if r := recover(); r != nil {
			// TODO: return an error???
			err := normalizePanic(r)
			l.With("err", err).Error(err.Error())
		}
	}()
	next.InitValidate(l, store, vals)
}

// normalizePanic makes sure we can get a nice TMError (with stack) out of it
func normalizePanic(p interface{}) error {
	if err, isErr := p.(error); isErr {
		return errors.Wrap(err)
	}
	msg := fmt.Sprintf("%v", p)
	return errors.ErrInternal(msg)
}
