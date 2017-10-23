package util

import (
	"math/rand"

	"github.com/tendermint/tmlibs/log"

	sdk "github.com/cosmos/cosmos-sdk"
)

// store nonce as it's own type so no one can even try to fake it
type nonce int64

type naiveContext struct {
	id     nonce
	chain  string
	height uint64
	perms  []sdk.Actor
	log.Logger
}

// MockContext returns a simple, non-checking context for test cases.
//
// Always use NewContext() for production code to sandbox malicious code better
func MockContext(chain string, height uint64) sdk.Context {
	return naiveContext{
		id:     nonce(rand.Int63()),
		chain:  chain,
		height: height,
		Logger: log.NewNopLogger(),
	}
}

var _ sdk.Context = naiveContext{}

func (c naiveContext) ChainID() string {
	return c.chain
}

func (c naiveContext) BlockHeight() uint64 {
	return c.height
}

// WithPermissions will panic if they try to set permission without the proper app
func (c naiveContext) WithPermissions(perms ...sdk.Actor) sdk.Context {
	return naiveContext{
		id:     c.id,
		chain:  c.chain,
		height: c.height,
		perms:  append(c.perms, perms...),
		Logger: c.Logger,
	}
}

func (c naiveContext) HasPermission(perm sdk.Actor) bool {
	for _, p := range c.perms {
		if p.Equals(perm) {
			return true
		}
	}
	return false
}

func (c naiveContext) GetPermissions(chain, app string) (res []sdk.Actor) {
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
func (c naiveContext) IsParent(other sdk.Context) bool {
	nc, ok := other.(naiveContext)
	if !ok {
		return false
	}
	return c.id == nc.id
}

// Reset should clear out all permissions,
// but carry on knowledge that this is a child
func (c naiveContext) Reset() sdk.Context {
	return naiveContext{
		id:     c.id,
		chain:  c.chain,
		height: c.height,
		Logger: c.Logger,
	}
}
