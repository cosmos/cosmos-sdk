package runtime

import (
	"context"
	"fmt"

	"cosmossdk.io/core/header"
	"cosmossdk.io/core/store"
)

var _ store.KVStoreService = (*GenesisKVStoreServie)(nil)

type genesisContextKeyType struct{}

var genesisContextKey = genesisContextKeyType{}

type genesisContext struct {
	state      store.WriterMap
	headerInfo header.Info
	didRun     bool
}

func makeGenesisContext(state store.WriterMap) genesisContext {
	return genesisContext{
		state: state,
	}
}

func (g *genesisContext) Run(
	ctx context.Context,
	fn func(ctx context.Context) error,
) (store.WriterMap, error) {
	ctx = context.WithValue(ctx, genesisContextKey, g)
	err := fn(ctx)
	if err != nil {
		return nil, err
	}
	g.didRun = true
	return g.state, nil
}

type GenesisKVStoreServie struct {
	genesisCapable bool
	actor          []byte
	execution      store.KVStoreService
}

func NewGenesisCapableKVStoreService(
	actor []byte,
	execution store.KVStoreService,
) *GenesisKVStoreServie {
	return &GenesisKVStoreServie{
		genesisCapable: true,
		actor:          actor,
		execution:      execution,
	}
}

// OpenKVStore implements store.KVStoreService.
func (g *GenesisKVStoreServie) OpenKVStore(ctx context.Context) store.KVStore {
	if !g.genesisCapable {
		return g.execution.OpenKVStore(ctx)
	}
	v := ctx.Value(genesisContextKey)
	if v == nil {
		panic("genesis context not found")
	}
	genCtx, ok := v.(*genesisContext)
	if !ok {
		panic(fmt.Errorf("unexpected genesis context type: %T", v))
	}
	if genCtx.didRun {
		g.genesisCapable = false
		return g.execution.OpenKVStore(ctx)
	}
	state, err := genCtx.state.GetWriter(g.actor)
	if err != nil {
		panic(err)
	}
	return state
}

type GenesisHeaderService struct {
	genesisCapable bool
}

// HeaderInfo implements header.Service.
func (g *GenesisHeaderService) HeaderInfo(ctx context.Context) header.Info {
	v := ctx.Value(genesisContextKey)
	if v == nil {
		panic("genesis context not found")
	}
	genCtx, ok := v.(*genesisContext)
	if !ok {
		panic(fmt.Errorf("unexpected genesis context type: %T", v))
	}
	return genCtx.headerInfo
}

var _ header.Service = (*GenesisHeaderService)(nil)
