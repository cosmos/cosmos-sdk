//nolint
package stack

import (
	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/state"
)

const (
	NameCheck = "check"
	NameGrant = "grant"
)

// CheckMiddleware returns an error if the tx doesn't have auth of this
// Required Actor, otherwise passes along the call untouched
type CheckMiddleware struct {
	Required sdk.Actor
	PassInitState
	PassInitValidate
}

var _ Middleware = CheckMiddleware{}

func (_ CheckMiddleware) Name() string {
	return NameCheck
}

func (p CheckMiddleware) CheckTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx, next sdk.Checker) (res sdk.CheckResult, err error) {
	if !ctx.HasPermission(p.Required) {
		return res, errors.ErrUnauthorized()
	}
	return next.CheckTx(ctx, store, tx)
}

func (p CheckMiddleware) DeliverTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx, next sdk.Deliver) (res sdk.DeliverResult, err error) {
	if !ctx.HasPermission(p.Required) {
		return res, errors.ErrUnauthorized()
	}
	return next.DeliverTx(ctx, store, tx)
}

// GrantMiddleware tries to set the permission to this Actor, which may be prohibited
type GrantMiddleware struct {
	Auth sdk.Actor
	PassInitState
	PassInitValidate
}

var _ Middleware = GrantMiddleware{}

func (_ GrantMiddleware) Name() string {
	return NameGrant
}

func (g GrantMiddleware) CheckTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx, next sdk.Checker) (res sdk.CheckResult, err error) {
	ctx = ctx.WithPermissions(g.Auth)
	return next.CheckTx(ctx, store, tx)
}

func (g GrantMiddleware) DeliverTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx, next sdk.Deliver) (res sdk.DeliverResult, err error) {
	ctx = ctx.WithPermissions(g.Auth)
	return next.DeliverTx(ctx, store, tx)
}
