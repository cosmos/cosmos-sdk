package stack

import (
	"bytes"

	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
)

type mockContext struct {
	perms []basecoin.Actor
	chain string
	log.Logger
}

func MockContext(chain string) basecoin.Context {
	return mockContext{
		chain:  chain,
		Logger: log.NewNopLogger(),
	}
}

var _ basecoin.Context = mockContext{}

func (c mockContext) ChainID() string {
	return c.chain
}

// WithPermissions will panic if they try to set permission without the proper app
func (c mockContext) WithPermissions(perms ...basecoin.Actor) basecoin.Context {
	return mockContext{
		perms:  append(c.perms, perms...),
		Logger: c.Logger,
	}
}

func (c mockContext) HasPermission(perm basecoin.Actor) bool {
	for _, p := range c.perms {
		if perm.App == p.App && bytes.Equal(perm.Address, p.Address) {
			return true
		}
	}
	return false
}

func (c mockContext) GetPermissions(chain, app string) (res []basecoin.Actor) {
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
func (c mockContext) IsParent(other basecoin.Context) bool {
	_, ok := other.(mockContext)
	return ok
}

// Reset should clear out all permissions,
// but carry on knowledge that this is a child
func (c mockContext) Reset() basecoin.Context {
	return mockContext{
		Logger: c.Logger,
	}
}
