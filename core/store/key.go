package store

import "context"

// Key represents a unique, non-forgeable handle to a KVStore.
type Key interface {
	// Open retrieves the KVStore from the context.
	Open(context.Context) KVStore
}

//
type KVStoreKey struct {
	name string
}

func (K KVStoreKey) Open(ctx context.Context) KVStore {
	//TODO implement me
	panic("implement me")
}

type IndexStoreKey struct {
	name string
}

func (i IndexStoreKey) Open(ctx context.Context) KVStore {
	//TODO implement me
	panic("implement me")
}

type MemoryStoreKey struct {
	name string
}

func (m MemoryStoreKey) Open(ctx context.Context) KVStore {
	//TODO implement me
	panic("implement me")
}

type TransientStoreKey struct {
	name string
}

func (t TransientStoreKey) Open(ctx context.Context) KVStore {
	//TODO implement me
	panic("implement me")
}

var _, _, _, _ Key = &KVStoreKey{}, &IndexStoreKey{}, &MemoryStoreKey{}, &TransientStoreKey{}
