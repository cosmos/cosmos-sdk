package ffi

import (
	"context"
	"sync"
)

type contextWrapper struct {
	ctx context.Context
}

func resolveContext(ctxId uint32) *contextWrapper {
	c, ok := contextMap.Load(ctxId)
	if !ok {
		panic("invalid context")
	}
	return c.(*contextWrapper)
}

var contextMap = &sync.Map{}
