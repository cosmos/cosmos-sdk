package stack

import (
	"bytes"
	"math/rand"

	"github.com/pkg/errors"

	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/state"
)

// store nonce as it's own type so no one can even try to fake it
type nonce int64

type secureContext struct {
	id    nonce
	chain string
	app   string
	perms []basecoin.Actor
	log.Logger
}

// NewContext - create a new secureContext
func NewContext(chain string, logger log.Logger) basecoin.Context {
	return secureContext{
		id:     nonce(rand.Int63()),
		chain:  chain,
		Logger: logger,
	}
}

var _ basecoin.Context = secureContext{}

func (c secureContext) ChainID() string {
	return c.chain
}

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
		id:     c.id,
		chain:  c.chain,
		app:    c.app,
		perms:  append(c.perms, perms...),
		Logger: c.Logger,
	}
}

func (c secureContext) HasPermission(perm basecoin.Actor) bool {
	for _, p := range c.perms {
		if perm.App == p.App && bytes.Equal(perm.Address, p.Address) {
			return true
		}
	}
	return false
}

func (c secureContext) GetPermissions(chain, app string) (res []basecoin.Actor) {
	for _, p := range c.perms {
		if chain == p.ChainID {
			if app == "" || app == p.App {
				res = append(res, p)
			}
		}
	}
	return res
}

// IsParent ensures that this is derived from the given secureClient
func (c secureContext) IsParent(other basecoin.Context) bool {
	so, ok := other.(secureContext)
	if !ok {
		return false
	}
	return c.id == so.id
}

// Reset should clear out all permissions,
// but carry on knowledge that this is a child
func (c secureContext) Reset() basecoin.Context {
	return secureContext{
		id:     c.id,
		chain:  c.chain,
		app:    c.app,
		perms:  nil,
		Logger: c.Logger,
	}
}

// withApp is a private method that we can use to properly set the
// app controls in the middleware
func withApp(ctx basecoin.Context, app string) basecoin.Context {
	sc, ok := ctx.(secureContext)
	if !ok {
		return ctx
	}
	return secureContext{
		id:     sc.id,
		chain:  sc.chain,
		app:    app,
		perms:  sc.perms,
		Logger: sc.Logger,
	}
}

func secureCheck(h basecoin.Checker, parent basecoin.Context) basecoin.Checker {
	next := func(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
		if !parent.IsParent(ctx) {
			return res, errors.New("Passing in non-child Context")
		}
		return h.CheckTx(ctx, store, tx)
	}
	return basecoin.CheckerFunc(next)
}

func secureDeliver(h basecoin.Deliver, parent basecoin.Context) basecoin.Deliver {
	next := func(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
		if !parent.IsParent(ctx) {
			return res, errors.New("Passing in non-child Context")
		}
		return h.DeliverTx(ctx, store, tx)
	}
	return basecoin.DeliverFunc(next)
}
