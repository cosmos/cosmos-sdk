package store

import "context"

type Key interface {
	Open(context.Context) KVStore
}

type KVStoreKey struct {
	name string
}

type IndexStoreKey struct {
	name string
}

type MemoryStoreKey struct {
	name string
}

type TransientStoreKey struct {
	name string
}
