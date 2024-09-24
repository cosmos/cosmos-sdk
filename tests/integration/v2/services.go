package integration

import (
	"context"
	"cosmossdk.io/core/comet"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/server"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/runtime/v2"
	"fmt"
)

func (c cometServiceImpl) CometInfo(context.Context) comet.Info {
	return comet.Info{}
}

// Services

var _ server.DynamicConfig = &dynamicConfigImpl{}

type dynamicConfigImpl struct {
	homeDir string
}

func (d *dynamicConfigImpl) Get(key string) any {
	return d.GetString(key)
}

func (d *dynamicConfigImpl) GetString(key string) string {
	switch key {
	case runtime.FlagHome:
		return d.homeDir
	case "store.app-db-backend":
		return "goleveldb"
	case "server.minimum-gas-prices":
		return "0stake"
	default:
		panic(fmt.Sprintf("unknown key: %s", key))
	}
}

func (d *dynamicConfigImpl) UnmarshalSub(string, any) (bool, error) {
	return false, nil
}

var _ comet.Service = &cometServiceImpl{}

type cometServiceImpl struct{}

type storeService struct {
	actor            []byte
	executionService corestore.KVStoreService
}

type contextKeyType struct{}

var contextKey = contextKeyType{}

type integrationContext struct {
	state corestore.WriterMap
}

func (s storeService) OpenKVStore(ctx context.Context) corestore.KVStore {
	iCtx, ok := ctx.Value(contextKey).(integrationContext)
	if !ok {
		return s.executionService.OpenKVStore(ctx)
	}

	state, err := iCtx.state.GetWriter(s.actor)
	if err != nil {
		panic(err)
	}
	return state
}

var (
	_ event.Service = &eventService{}
	_ event.Manager = &eventManager{}
)

type eventService struct{}

// EventManager implements event.Service.
func (e *eventService) EventManager(context.Context) event.Manager {
	return &eventManager{}
}

type eventManager struct{}

// Emit implements event.Manager.
func (e *eventManager) Emit(event transaction.Msg) error {
	return nil
}

// EmitKV implements event.Manager.
func (e *eventManager) EmitKV(eventType string, attrs ...event.Attribute) error {
	return nil
}
