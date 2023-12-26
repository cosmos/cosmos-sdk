package ormtable

import (
	"context"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/orm/types/kv"
)

// ReadBackend defines the type used for read-only ORM operations.
type ReadBackend interface {
	// CommitmentStoreReader returns the reader for the commitment store.
	CommitmentStoreReader() kv.ReadonlyStore

	// IndexStoreReader returns the reader for the index store.
	IndexStoreReader() kv.ReadonlyStore

	private()
}

// Backend defines the type used for read-write ORM operations.
// Unlike ReadBackend, write access to the underlying kv-store
// is hidden so that this can be fully encapsulated by the ORM.
type Backend interface {
	ReadBackend

	// CommitmentStore returns the merklized commitment store.
	CommitmentStore() store.KVStore

	// IndexStore returns the index store if a separate one exists,
	// otherwise it the commitment store.
	IndexStore() store.KVStore

	// ValidateHooks returns a ValidateHooks instance or nil.
	ValidateHooks() ValidateHooks

	// WithValidateHooks returns a copy of this backend with the provided validate hooks.
	WithValidateHooks(ValidateHooks) Backend

	// WriteHooks returns a WriteHooks instance of nil.
	WriteHooks() WriteHooks

	// WithWriteHooks returns a copy of this backend with the provided write hooks.
	WithWriteHooks(WriteHooks) Backend
}

// ReadBackendOptions defines options for creating a ReadBackend.
// Read context can optionally define two stores - a commitment store
// that is backed by a merkle tree and an index store that isn't.
// If the index store is not defined, the commitment store will be
// used for all operations.
type ReadBackendOptions struct {
	// CommitmentStoreReader is a reader for the commitment store.
	CommitmentStoreReader kv.ReadonlyStore

	// IndexStoreReader is an optional reader for the index store.
	// If it is nil the CommitmentStoreReader will be used.
	IndexStoreReader kv.ReadonlyStore
}

type readBackend struct {
	commitmentReader kv.ReadonlyStore
	indexReader      kv.ReadonlyStore
}

func (r readBackend) CommitmentStoreReader() kv.ReadonlyStore {
	return r.commitmentReader
}

func (r readBackend) IndexStoreReader() kv.ReadonlyStore {
	return r.indexReader
}

func (readBackend) private() {}

// NewReadBackend creates a new ReadBackend.
func NewReadBackend(options ReadBackendOptions) ReadBackend {
	indexReader := options.IndexStoreReader
	if indexReader == nil {
		indexReader = options.CommitmentStoreReader
	}
	return &readBackend{
		commitmentReader: options.CommitmentStoreReader,
		indexReader:      indexReader,
	}
}

type backend struct {
	commitmentStore store.KVStore
	indexStore      store.KVStore
	validateHooks   ValidateHooks
	writeHooks      WriteHooks
}

func (c backend) ValidateHooks() ValidateHooks {
	return c.validateHooks
}

func (c backend) WithValidateHooks(hooks ValidateHooks) Backend {
	c.validateHooks = hooks
	return c
}

func (c backend) WriteHooks() WriteHooks {
	return c.writeHooks
}

func (c backend) WithWriteHooks(hooks WriteHooks) Backend {
	c.writeHooks = hooks
	return c
}

func (backend) private() {}

func (c backend) CommitmentStoreReader() kv.ReadonlyStore {
	return c.commitmentStore
}

func (c backend) IndexStoreReader() kv.ReadonlyStore {
	return c.indexStore
}

func (c backend) CommitmentStore() store.KVStore {
	return c.commitmentStore
}

func (c backend) IndexStore() store.KVStore {
	return c.indexStore
}

// BackendOptions defines options for creating a Backend.
// Context can optionally define two stores - a commitment store
// that is backed by a merkle tree and an index store that isn't.
// If the index store is not defined, the commitment store will be
// used for all operations.
type BackendOptions struct {
	// CommitmentStore is the commitment store.
	CommitmentStore store.KVStore

	// IndexStore is the optional index store.
	// If it is nil the CommitmentStore will be used.
	IndexStore store.KVStore

	// ValidateHooks are optional hooks into ORM insert, update and delete operations.
	ValidateHooks ValidateHooks

	WriteHooks WriteHooks
}

// NewBackend creates a new Backend.
func NewBackend(options BackendOptions) Backend {
	indexStore := options.IndexStore
	if indexStore == nil {
		indexStore = options.CommitmentStore
	}
	return &backend{
		commitmentStore: options.CommitmentStore,
		indexStore:      indexStore,
		validateHooks:   options.ValidateHooks,
		writeHooks:      options.WriteHooks,
	}
}

// BackendResolver resolves a backend from the context or returns an error.
// Callers should type cast the returned ReadBackend to Backend to test whether
// the backend is writable.
type BackendResolver func(context.Context) (ReadBackend, error)

// WrapContextDefault performs the default wrapping of a backend in a context.
// This should be used primarily for testing purposes and production code
// should use some other framework specific wrapping (for instance using
// "store keys").
func WrapContextDefault(backend ReadBackend) context.Context {
	return context.WithValue(context.Background(), defaultContextKey, backend)
}

type contextKeyType string

var defaultContextKey = contextKeyType("backend")

func getBackendDefault(ctx context.Context) (ReadBackend, error) {
	value := ctx.Value(defaultContextKey)
	if value == nil {
		return nil, fmt.Errorf("can't resolve backend")
	}

	backend, ok := value.(ReadBackend)
	if !ok {
		return nil, fmt.Errorf("expected value of type %T, instead got %T", backend, value)
	}

	return backend, nil
}
