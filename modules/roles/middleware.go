package roles

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
)

type Middleware struct {
	stack.PassOption
}

var _ stack.Middleware = Middleware{}

func NewMiddleware() Middleware {
	return Middleware{}
}

// Name - return name space
func (Middleware) Name() string {
	return NameRole
}

// CheckTx checks if this is valid
func (m Middleware) CheckTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx, next basecoin.Checker) (res basecoin.Result, err error) {
	assume, err := checkMiddleTx(ctx, tx)
	if err != nil {
		return res, err
	}

	ctx, err = assumeRole(ctx, store, assume)
	if err != nil {
		return res, err
	}

	return next.CheckTx(ctx, store, assume.Tx)
}

// DeliverTx moves the money
func (m Middleware) DeliverTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx, next basecoin.Deliver) (res basecoin.Result, err error) {
	assume, err := checkMiddleTx(ctx, tx)
	if err != nil {
		return res, err
	}

	ctx, err = assumeRole(ctx, store, assume)
	if err != nil {
		return res, err
	}

	return next.DeliverTx(ctx, store, assume.Tx)
}

func checkMiddleTx(ctx basecoin.Context, tx basecoin.Tx) (assume AssumeRoleTx, err error) {
	// check if the tx is proper type and valid
	assume, ok := tx.Unwrap().(AssumeRoleTx)
	if !ok {
		return assume, errors.ErrInvalidFormat(tx)
	}
	err = assume.ValidateBasic()
	if err != nil {
		return assume, err
	}

	// load it up and check it out...
	return assume, err
}

func assumeRole(ctx basecoin.Context, store state.KVStore, assume AssumeRoleTx) (basecoin.Context, error) {
	role, err := loadRole(store, MakeKey(assume.Role))
	if err != nil {
		return nil, err
	}

	if !role.IsAuthorized(ctx) {
		return nil, ErrInsufficientSigs()
	}
	ctx = ctx.WithPermissions(NewPerm(assume.Role))
	return ctx, nil
}
