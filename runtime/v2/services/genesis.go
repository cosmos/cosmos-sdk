package services

import (
	"context"
	"fmt"

	"cosmossdk.io/core/header"
	"cosmossdk.io/core/store"
)

var (
	_ store.KVStoreService = (*GenesisKVStoreService)(nil)
	_ header.Service       = (*GenesisHeaderService)(nil)
	_ store.KVStore        = (*readonlyKVStore)(nil)
)

type genesisContextKeyType struct{}

var genesisContextKey = genesisContextKeyType{}

// genesisContext is a context that is used during genesis initialization.
// it backs the store.KVStoreService and header.Service interface implementations
// defined in this file.
type genesisContext struct {
	state store.ReaderMap
}

// NewGenesisContext creates a new genesis context.
func NewGenesisContext(state store.ReaderMap) genesisContext {
	return genesisContext{
		state: state,
	}
}

// Mutate runs the provided function within the genesis context and returns an
// updated store.WriterMap containing the state modifications made during InitGenesis.
func (g genesisContext) Mutate(
	ctx context.Context,
	fn func(ctx context.Context) error,
) (store.WriterMap, error) {
	writerMap, ok := g.state.(store.WriterMap)
	if !ok {
		return nil, fmt.Errorf("mutate requires a store.WriterMap, got a %T", g.state)
	}
	ctx = context.WithValue(ctx, genesisContextKey, g)
	err := fn(ctx)
	if err != nil {
		return nil, err
	}
	return writerMap, nil
}

// Read runs the provided function within the genesis context.
func (g genesisContext) Read(
	ctx context.Context,
	fn func(ctx context.Context) error,
) error {
	ctx = context.WithValue(ctx, genesisContextKey, g)
	return fn(ctx)
}

// GenesisKVStoreService is a store.KVStoreService implementation that is used during
// genesis initialization.  It wraps an inner execution context store.KVStoreService.
type GenesisKVStoreService struct {
	actor            []byte
	executionService store.KVStoreService
}

// NewGenesisKVService creates a new GenesisKVStoreService.
// - actor is the module store key.
// - executionService is the store.KVStoreService to use when the genesis context is not active.
func NewGenesisKVService(
	actor []byte,
	executionService store.KVStoreService,
) *GenesisKVStoreService {
	return &GenesisKVStoreService{
		actor:            actor,
		executionService: executionService,
	}
}

// OpenKVStore implements store.KVStoreService.
func (g *GenesisKVStoreService) OpenKVStore(ctx context.Context) store.KVStore {
	v := ctx.Value(genesisContextKey)
	if v == nil {
		return g.executionService.OpenKVStore(ctx)
	}
	genCtx, ok := v.(genesisContext)
	if !ok {
		panic(fmt.Errorf("unexpected genesis context type: %T", v))
	}
	writerMap, ok := genCtx.state.(store.WriterMap)
	if ok {
		state, err := writerMap.GetWriter(g.actor)
		if err != nil {
			panic(err)
		}
		return state

	}
	state, err := genCtx.state.GetReader(g.actor)
	if err != nil {
		panic(err)
	}
	return readonlyKVStore{state}
}

// GenesisHeaderService is a header.Service implementation that is used during
// genesis initialization.  It wraps an inner execution context header.Service.
type GenesisHeaderService struct {
	executionService header.Service
}

// HeaderInfo implements header.Service.
// During genesis initialization, it returns an empty header.Info.
func (g *GenesisHeaderService) HeaderInfo(ctx context.Context) header.Info {
	v := ctx.Value(genesisContextKey)
	if v == nil {
		return g.executionService.HeaderInfo(ctx)
	}
	return header.Info{}
}

// NewGenesisHeaderService creates a new GenesisHeaderService.
// - executionService is the header.Service to use when the genesis context is not active.
func NewGenesisHeaderService(executionService header.Service) *GenesisHeaderService {
	return &GenesisHeaderService{
		executionService: executionService,
	}
}

type readonlyKVStore struct {
	store.Reader
}

func (r readonlyKVStore) Set(key, value []byte) error {
	panic("tried to call Set on a readonly store")
}

func (r readonlyKVStore) Delete(key []byte) error {
	panic("tried to call Delete on a readonly store")
}
