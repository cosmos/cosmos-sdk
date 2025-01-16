package integration

import (
	"context"
	"fmt"

	"cosmossdk.io/core/comet"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/gas"
<<<<<<< HEAD
=======
	"cosmossdk.io/core/header"
>>>>>>> 952db2b32 (chore: remove baseapp from `x/accounts` (#23355))
	"cosmossdk.io/core/server"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	stfgas "cosmossdk.io/server/v2/stf/gas"
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
	case "home":
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
	state    corestore.WriterMap
	gasMeter gas.Meter
}

func GasMeterFromContext(ctx context.Context) gas.Meter {
	iCtx, ok := ctx.Value(contextKey).(*integrationContext)
	if !ok {
		return nil
	}
	return iCtx.gasMeter
}

func GasMeterFactory(ctx context.Context) func() gas.Meter {
	return func() gas.Meter {
		return GasMeterFromContext(ctx)
	}
}

func (s storeService) OpenKVStore(ctx context.Context) corestore.KVStore {
	const gasLimit = 100_000
	iCtx, ok := ctx.Value(contextKey).(*integrationContext)
	if !ok {
		return s.executionService.OpenKVStore(ctx)
	}

	iCtx.gasMeter = stfgas.NewMeter(gasLimit)
	writerMap := stfgas.NewMeteredWriterMap(stfgas.DefaultConfig, iCtx.gasMeter, iCtx.state)
	state, err := writerMap.GetWriter(s.actor)
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
<<<<<<< HEAD
=======

var _ branch.Service = &BranchService{}

// custom branch service for integration tests
type BranchService struct{}

func (bs *BranchService) Execute(ctx context.Context, f func(ctx context.Context) error) error {
	_, ok := ctx.Value(contextKey).(*integrationContext)
	if !ok {
		return errors.New("context is not an integration context")
	}

	return f(ctx)
}

func (bs *BranchService) ExecuteWithGasLimit(
	ctx context.Context,
	gasLimit uint64,
	f func(ctx context.Context) error,
) (gasUsed uint64, err error) {
	iCtx, ok := ctx.Value(contextKey).(*integrationContext)
	if !ok {
		return 0, errors.New("context is not an integration context")
	}

	originalGasMeter := iCtx.gasMeter

	iCtx.gasMeter = stfgas.DefaultGasMeter(gasLimit)

	// execute branched, with predefined gas limit.
	err = bs.execute(ctx, iCtx, f)

	// restore original context
	gasUsed = iCtx.gasMeter.Limit() - iCtx.gasMeter.Remaining()
	_ = originalGasMeter.Consume(gasUsed, "execute-with-gas-limit")
	iCtx.gasMeter = stfgas.DefaultGasMeter(originalGasMeter.Remaining())

	return gasUsed, err
}

func (bs BranchService) execute(ctx context.Context, ictx *integrationContext, f func(ctx context.Context) error) error {
	branchedState := stfbranch.DefaultNewWriterMap(ictx.state)
	meteredBranchedState := stfgas.DefaultWrapWithGasMeter(ictx.gasMeter, branchedState)

	branchedCtx := &integrationContext{
		state:    meteredBranchedState,
		gasMeter: ictx.gasMeter,
		header:   ictx.header,
		events:   ictx.events,
	}

	newCtx := context.WithValue(ctx, contextKey, branchedCtx)

	err := f(newCtx)
	if err != nil {
		return err
	}

	err = applyStateChanges(ictx.state, branchedCtx.state)
	if err != nil {
		return err
	}

	return nil
}

func applyStateChanges(dst, src corestore.WriterMap) error {
	changes, err := src.GetStateChanges()
	if err != nil {
		return err
	}
	return dst.ApplyStateChanges(changes)
}

var _ header.Service = &HeaderService{}

type HeaderService struct{}

func (h *HeaderService) HeaderInfo(ctx context.Context) header.Info {
	iCtx, ok := ctx.Value(contextKey).(*integrationContext)
	if !ok {
		return header.Info{}
	}
	return iCtx.header
}

var _ gas.Service = &GasService{}

type GasService struct{}

func (g *GasService) GasMeter(ctx context.Context) gas.Meter {
	return GasMeterFromContext(ctx)
}

func (g *GasService) GasConfig(ctx context.Context) gas.GasConfig {
	return gas.GasConfig{}
}
>>>>>>> 952db2b32 (chore: remove baseapp from `x/accounts` (#23355))
