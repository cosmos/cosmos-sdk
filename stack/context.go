package stack

import (
	"github.com/pkg/errors"

	"github.com/tendermint/tmlibs/log"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
)

// store nonce as it's own type so no one can even try to fake it
type nonce int64

type secureContext struct {
	app string
	ibc bool
	// this exposes the log.Logger and all other methods we don't override
	naiveContext
}

// NewContext - create a new secureContext
func NewContext(chain string, height uint64, logger log.Logger) sdk.Context {
	mock := MockContext(chain, height).(naiveContext)
	mock.Logger = logger
	return secureContext{
		naiveContext: mock,
	}
}

var _ sdk.Context = secureContext{}

// WithPermissions will panic if they try to set permission without the proper app
func (c secureContext) WithPermissions(perms ...sdk.Actor) sdk.Context {
	// the guard makes sure you only set permissions for the app you are inside
	for _, p := range perms {
		if !c.validPermisison(p) {
			err := errors.Errorf("Cannot set permission for %s/%s on (app=%s, ibc=%b)",
				p.ChainID, p.App, c.app, c.ibc)
			panic(err)
		}
	}

	return secureContext{
		app:          c.app,
		ibc:          c.ibc,
		naiveContext: c.naiveContext.WithPermissions(perms...).(naiveContext),
	}
}

func (c secureContext) validPermisison(p sdk.Actor) bool {
	// if app is set, then it must match
	if c.app != "" && c.app != p.App {
		return false
	}
	// if ibc, chain must be set, otherwise it must not
	return c.ibc == (p.ChainID != "")
}

// Reset should clear out all permissions,
// but carry on knowledge that this is a child
func (c secureContext) Reset() sdk.Context {
	return secureContext{
		app:          c.app,
		ibc:          c.ibc,
		naiveContext: c.naiveContext.Reset().(naiveContext),
	}
}

// IsParent ensures that this is derived from the given secureClient
func (c secureContext) IsParent(other sdk.Context) bool {
	so, ok := other.(secureContext)
	if !ok {
		return false
	}
	return c.naiveContext.IsParent(so.naiveContext)
}

// withApp is a private method that we can use to properly set the
// app controls in the middleware
func withApp(ctx sdk.Context, app string) sdk.Context {
	sc, ok := ctx.(secureContext)
	if !ok {
		return ctx
	}
	return secureContext{
		app:          app,
		ibc:          false,
		naiveContext: sc.naiveContext,
	}
}

// withIBC is a private method so we can securely allow IBC permissioning
func withIBC(ctx sdk.Context) sdk.Context {
	sc, ok := ctx.(secureContext)
	if !ok {
		return ctx
	}
	return secureContext{
		app:          "",
		ibc:          true,
		naiveContext: sc.naiveContext,
	}
}

func secureCheck(h sdk.Checker, parent sdk.Context) sdk.Checker {
	next := func(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx) (res sdk.CheckResult, err error) {
		if !parent.IsParent(ctx) {
			return res, errors.New("Passing in non-child Context")
		}
		return h.CheckTx(ctx, store, tx)
	}
	return sdk.CheckerFunc(next)
}

func secureDeliver(h sdk.Deliver, parent sdk.Context) sdk.Deliver {
	next := func(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx) (res sdk.DeliverResult, err error) {
		if !parent.IsParent(ctx) {
			return res, errors.New("Passing in non-child Context")
		}
		return h.DeliverTx(ctx, store, tx)
	}
	return sdk.DeliverFunc(next)
}
