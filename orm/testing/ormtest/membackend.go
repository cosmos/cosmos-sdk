// Package ormtest contains utilities for testing modules built with the ORM.
package ormtest

import (
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/orm/model/ormtable"
)

// NewMemoryBackend returns a new ORM memory backend which can be used for
// testing purposes independent of any storage layer.
//
// Example:
//  backend := ormtest.NewMemoryBackend()
//  ctx := ormtable.WrapContextDefault()
//  ...
func NewMemoryBackend() ormtable.Backend {
	return NewMemoryBackendWithHooks(nil)
}

// NewMemoryBackendWithHooks returns an ORM memory backend which can
// be used for testing with the provided hooks.
func NewMemoryBackendWithHooks(hooks ormtable.Hooks) ormtable.Backend {
	return ormtable.NewBackend(ormtable.BackendOptions{
		CommitmentStore: dbm.NewMemDB(),
		IndexStore:      dbm.NewMemDB(),
		Hooks:           hooks,
	})
}
