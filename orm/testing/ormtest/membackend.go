package ormtest

import (
	"github.com/cosmos/cosmos-sdk/orm/internal/testkv"
	"github.com/cosmos/cosmos-sdk/orm/model/ormtable"
)

func NewMemoryBackend() ormtable.Backend {
	return testkv.NewSplitMemBackend()
}
