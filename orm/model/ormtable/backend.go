package ormtable

import (
	"context"

	"github.com/cosmos/cosmos-sdk/orm/types/kv"
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
	CommitmentStore() kv.Store

	// IndexStore returns the index store if a separate one exists,
	// otherwise it the commitment store.
	IndexStore() kv.Store

	// Hooks returns a Hooks instance or nil.
	Hooks() Hooks

	// WithHooks returns a copy of this backend with the provided hooks.
	WithHooks(Hooks) Backend
}

type backend struct {
	commitmentStore kv.ReadonlyStore
	indexStore      kv.ReadonlyStore
	hooks           Hooks
}

func (c backend) WithHooks(hooks Hooks) Backend {
	c.hooks = hooks
	return c
}

func (backend) private() {}

func (c backend) CommitmentStoreReader() kv.ReadonlyStore {
	return c.commitmentStore
}

func (c backend) IndexStoreReader() kv.ReadonlyStore {
	return c.indexStore
}

func (c backend) CommitmentStore() kv.Store {
	return c.commitmentStore.(kv.Store)
}

func (c backend) IndexStore() kv.Store {
	return c.indexStore.(kv.Store)
}

func (c backend) Hooks() Hooks {
	return c.hooks
}

// BackendOptions defines options for creating a Backend.
// Context can optionally define two stores - a commitment store
// that is backed by a merkle tree and an index store that isn't.
// If the index store is not defined, the commitment store will be
// used for all operations.
type BackendOptions struct {

	// CommitmentStore is the commitment store.
	CommitmentStore kv.ReadonlyStore

	// IndexStore is the optional index store.
	// If it is nil the CommitmentStore will be used.
	IndexStore kv.ReadonlyStore

	// Hooks are optional hooks into ORM insert, update and delete operations.
	Hooks Hooks
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
		hooks:           options.Hooks,
	}
}

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
	return ctx.Value(defaultContextKey).(Backend), nil
}
