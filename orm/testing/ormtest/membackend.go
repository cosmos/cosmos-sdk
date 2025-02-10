// Package ormtest contains utilities for testing modules built with the ORM.
package ormtest

import (
	"cosmossdk.io/orm/internal/testkv"
	"cosmossdk.io/orm/model/ormtable"
)

// NewMemoryBackend returns a new ORM memory backend which can be used for
// testing purposes independent of any storage layer.
//
// Example:
//
//	backend := ormtest.NewMemoryBackend()
//	ctx := ormtable.WrapContextDefault()
//	...
func NewMemoryBackend() ormtable.Backend {
	return testkv.NewSplitMemBackend()
}
