// Package kv should eventually be removed once KV API in the sdk stabilizes
package kv

import (
	"context"
)

type KV interface {
	Get(ctx context.Context, key []byte) (value []byte)
	Set(ctx context.Context, key, value []byte)
	Has(ctx context.Context, key []byte) (exists bool)
	Delete(ctx context.Context, key []byte)
	Iterate(ctx context.Context, start, end []byte) Iterator
	IteratePrefix(ctx context.Context, prefix []byte) Iterator
}

type Iterator interface {
	Key() []byte
	Next()
	Valid() bool
	Close()
	Context() context.Context
}
