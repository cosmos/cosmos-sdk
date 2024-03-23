package cgo

import (
	"context"
	"sync"

	"cosmossdk.io/core/intermodule"
)

type contextWrapper struct {
	client intermodule.Client
	ctx    context.Context
}

func resolveContext(ctxId uint32) *contextWrapper {
	c, ok := contextMap.Load(ctxId)
	if !ok {
		panic("invalid context")
	}
	return c.(*contextWrapper)
}

var contextMap = &sync.Map{}
