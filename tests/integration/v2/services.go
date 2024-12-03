package integration

import (
	"context"
	"errors"
	"fmt"

	gogoproto "github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/core/branch"
	"cosmossdk.io/core/comet"
	corecontext "cosmossdk.io/core/context"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/gas"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/router"
	"cosmossdk.io/core/server"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	stfgas "cosmossdk.io/server/v2/stf/gas"
)

var ErrInvalidMsgType = fmt.Errorf("invalid message type")

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
	header   header.Info
}

func SetHeaderInfo(ctx context.Context, h header.Info) context.Context {
	iCtx, ok := ctx.Value(contextKey).(*integrationContext)
	if ok {
		iCtx.header = h
	}
	return ctx
}

func HeaderInfoFromContext(ctx context.Context) header.Info {
	iCtx, ok := ctx.Value(contextKey).(*integrationContext)
	if ok {
		return iCtx.header
	}
	return header.Info{}
}

func SetCometInfo(ctx context.Context, c comet.Info) context.Context {
	return context.WithValue(ctx, corecontext.CometInfoKey, c)
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

	// execute branched, with predefined gas limit.
	err = f(ctx)
	// restore original context
	gasUsed = iCtx.gasMeter.Limit() - iCtx.gasMeter.Remaining()
	_ = iCtx.gasMeter.Consume(gasUsed, "execute-with-gas-limit")

	return gasUsed, err
}

// msgTypeURL returns the TypeURL of a proto message.
func msgTypeURL(msg gogoproto.Message) string {
	return gogoproto.MessageName(msg)
}

type routerHandler func(context.Context, transaction.Msg) (transaction.Msg, error)

var _ router.Service = &RouterService{}

// custom router service for integration tests
type RouterService struct {
	handlers map[string]routerHandler
}

func NewRouterService() *RouterService {
	return &RouterService{
		handlers: make(map[string]routerHandler),
	}
}

func (rs *RouterService) RegisterHandler(handler routerHandler, typeUrl string) {
	rs.handlers[typeUrl] = handler
}

func (rs RouterService) CanInvoke(ctx context.Context, typeUrl string) error {
	if rs.handlers[typeUrl] == nil {
		return fmt.Errorf("no handler for typeURL %s", typeUrl)
	}
	return nil
}

func (rs RouterService) Invoke(ctx context.Context, req transaction.Msg) (transaction.Msg, error) {
	typeUrl := msgTypeURL(req)
	if rs.handlers[typeUrl] == nil {
		return nil, fmt.Errorf("no handler for typeURL %s", typeUrl)
	}
	return rs.handlers[typeUrl](ctx, req)
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
