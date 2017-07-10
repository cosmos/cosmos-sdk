package stack

import (
	"bytes"
	"math/rand"

	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
)

type naiveContext struct {
	id     nonce
	chain  string
	height uint64
	perms  []basecoin.Actor
	log.Logger
}

// MockContext returns a simple, non-checking context for test cases.
//
// Always use NewContext() for production code to sandbox malicious code better
func MockContext(chain string, height uint64) basecoin.Context {
	return naiveContext{
		id:     nonce(rand.Int63()),
		chain:  chain,
		height: height,
		Logger: log.NewNopLogger(),
	}
}

var _ basecoin.Context = naiveContext{}

func (c naiveContext) ChainID() string {
	return c.chain
}

func (c naiveContext) BlockHeight() uint64 {
	return c.height
}

// WithPermissions will panic if they try to set permission without the proper app
func (c naiveContext) WithPermissions(perms ...basecoin.Actor) basecoin.Context {
	return naiveContext{
		id:     c.id,
		chain:  c.chain,
		height: c.height,
		perms:  append(c.perms, perms...),
		Logger: c.Logger,
	}
}

func (c naiveContext) HasPermission(perm basecoin.Actor) bool {
	for _, p := range c.perms {
		if perm.App == p.App && bytes.Equal(perm.Address, p.Address) {
			return true
		}
	}
	return false
}

func (c naiveContext) GetPermissions(chain, app string) (res []basecoin.Actor) {
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
func (c naiveContext) IsParent(other basecoin.Context) bool {
	nc, ok := other.(naiveContext)
	if !ok {
		return false
	}
	return c.id == nc.id
}

// Reset should clear out all permissions,
// but carry on knowledge that this is a child
func (c naiveContext) Reset() basecoin.Context {
	return naiveContext{
		id:     c.id,
		chain:  c.chain,
		height: c.height,
		Logger: c.Logger,
	}
}
