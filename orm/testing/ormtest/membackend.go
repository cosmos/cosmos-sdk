package ormtest

import (
	"github.com/cosmos/cosmos-sdk/orm/internal/testkv"
	"github.com/cosmos/cosmos-sdk/orm/model/ormtable"
)

// NewMemoryBackend returns a new ORM memory backend which can be used for
// testing purposes.
func NewMemoryBackend() ormtable.Backend {
	return testkv.NewSplitMemBackend()
}
