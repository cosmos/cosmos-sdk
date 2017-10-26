package util

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
)

// Recovery catches any panics and returns them as errors instead
type Recovery struct{}

var _ sdk.Decorator = Recovery{}

// CheckTx catches any panic and converts to error - fulfills Middlware interface
func (Recovery) CheckTx(ctx sdk.Context, store sdk.SimpleDB,
	tx interface{}, next sdk.Checker) (res sdk.CheckResult, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = normalizePanic(r)
		}
	}()
	return next.CheckTx(ctx, store, tx)
}

// DeliverTx catches any panic and converts to error - fulfills Middlware interface
func (Recovery) DeliverTx(ctx sdk.Context, store sdk.SimpleDB,
	tx interface{}, next sdk.Deliverer) (res sdk.DeliverResult, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = normalizePanic(r)
		}
	}()
	return next.DeliverTx(ctx, store, tx)
}

// normalizePanic makes sure we can get a nice TMError (with stack) out of it
func normalizePanic(p interface{}) error {
	if err, isErr := p.(error); isErr {
		return errors.Wrap(err)
	}
	msg := fmt.Sprintf("%v", p)
	return errors.ErrInternal(msg)
}
