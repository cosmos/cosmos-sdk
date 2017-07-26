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
	// this exposes the log.Logger and all other methods we don't override
	naiveContext
}

// NewContext - create a new secureContext
func NewContext(chain string, height uint64, logger log.Logger) basecoin.Context {
	return secureContext{
		naiveContext: MockContext(chain, height).(naiveContext),
	}
}

var _ basecoin.Context = secureContext{}

// WithPermissions will panic if they try to set permission without the proper app
func (c secureContext) WithPermissions(perms ...basecoin.Actor) basecoin.Context {
	// the guard makes sure you only set permissions for the app you are inside
	for _, p := range perms {
		// TODO: also check chainID, limit only certain middleware can set IBC?
		if p.App != c.app {
			err := errors.Errorf("Cannot set permission for %s from %s", c.app, p.App)
			panic(err)
		}
	}

	return secureContext{
		app:          c.app,
		naiveContext: c.naiveContext.WithPermissions(perms...).(naiveContext),
	}
}

// Reset should clear out all permissions,
// but carry on knowledge that this is a child
func (c secureContext) Reset() basecoin.Context {
	return secureContext{
		app:          c.app,
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
		naiveContext: sc.naiveContext,
	}
}

func secureCheck(h basecoin.Checker, parent basecoin.Context) basecoin.Checker {
	next := func(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx) (res basecoin.Result, err error) {
		if !parent.IsParent(ctx) {
			return res, errors.New("Passing in non-child Context")
		}
		return h.CheckTx(ctx, store, tx)
	}
	return basecoin.CheckerFunc(next)
}

func secureDeliver(h basecoin.Deliver, parent basecoin.Context) basecoin.Deliver {
	next := func(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx) (res basecoin.Result, err error) {
		if !parent.IsParent(ctx) {
			return res, errors.New("Passing in non-child Context")
		}
		return h.DeliverTx(ctx, store, tx)
	}
	return basecoin.DeliverFunc(next)
}
