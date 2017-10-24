package util

import (
	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
)

// CheckDecorator returns an error if the tx doesn't have auth of this
// Required Actor, otherwise passes along the call untouched
type CheckDecorator struct {
	Required sdk.Actor
}

var _ sdk.Decorator = CheckDecorator{}

func (p CheckDecorator) CheckTx(ctx sdk.Context, store sdk.SimpleDB,
	tx interface{}, next sdk.Checker) (res sdk.CheckResult, err error) {

	if !ctx.HasPermission(p.Required) {
		return res, errors.ErrUnauthorized()
	}
	return next.CheckTx(ctx, store, tx)
}

func (p CheckDecorator) DeliverTx(ctx sdk.Context, store sdk.SimpleDB,
	tx interface{}, next sdk.Deliverer) (res sdk.DeliverResult, err error) {

	if !ctx.HasPermission(p.Required) {
		return res, errors.ErrUnauthorized()
	}
	return next.DeliverTx(ctx, store, tx)
}

// GrantDecorator tries to set the permission to this Actor, which may be prohibited
type GrantDecorator struct {
	Auth sdk.Actor
}

var _ sdk.Decorator = GrantDecorator{}

func (g GrantDecorator) CheckTx(ctx sdk.Context, store sdk.SimpleDB,
	tx interface{}, next sdk.Checker) (res sdk.CheckResult, err error) {

	ctx = ctx.WithPermissions(g.Auth)
	return next.CheckTx(ctx, store, tx)
}

func (g GrantDecorator) DeliverTx(ctx sdk.Context, store sdk.SimpleDB,
	tx interface{}, next sdk.Deliverer) (res sdk.DeliverResult, err error) {

	ctx = ctx.WithPermissions(g.Auth)
	return next.DeliverTx(ctx, store, tx)
}
