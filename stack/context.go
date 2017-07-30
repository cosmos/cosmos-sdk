package stack

import (
	"github.com/pkg/errors"

	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/state"
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
func NewContext(chain string, height uint64, logger log.Logger) basecoin.Context {
	mock := MockContext(chain, height).(naiveContext)
	mock.Logger = logger
	return secureContext{
		naiveContext: mock,
	}
}

var _ basecoin.Context = secureContext{}

// WithPermissions will panic if they try to set permission without the proper app
func (c secureContext) WithPermissions(perms ...basecoin.Actor) basecoin.Context {
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

func (c secureContext) validPermisison(p basecoin.Actor) bool {
	// if app is set, then it must match
	if c.app != "" && c.app != p.App {
		return false
	}
	// if ibc, chain must be set, otherwise it must not
	return c.ibc == (p.ChainID != "")
}

// Reset should clear out all permissions,
// but carry on knowledge that this is a child
func (c secureContext) Reset() basecoin.Context {
	return secureContext{
		app:          c.app,
		ibc:          c.ibc,
		naiveContext: c.naiveContext.Reset().(naiveContext),
	}
}

// IsParent ensures that this is derived from the given secureClient
func (c secureContext) IsParent(other basecoin.Context) bool {
	so, ok := other.(secureContext)
	if !ok {
		return false
	}
	return c.naiveContext.IsParent(so.naiveContext)
}

// withApp is a private method that we can use to properly set the
// app controls in the middleware
func withApp(ctx basecoin.Context, app string) basecoin.Context {
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
func withIBC(ctx basecoin.Context) basecoin.Context {
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

func secureCheck(h basecoin.Checker, parent basecoin.Context) basecoin.Checker {
	next := func(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx) (res basecoin.CheckResult, err error) {
		if !parent.IsParent(ctx) {
			return res, errors.New("Passing in non-child Context")
		}
		return h.CheckTx(ctx, store, tx)
	}
	return basecoin.CheckerFunc(next)
}

func secureDeliver(h basecoin.Deliver, parent basecoin.Context) basecoin.Deliver {
	next := func(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx) (res basecoin.DeliverResult, err error) {
		if !parent.IsParent(ctx) {
			return res, errors.New("Passing in non-child Context")
		}
		return h.DeliverTx(ctx, store, tx)
	}
	return basecoin.DeliverFunc(next)
}
