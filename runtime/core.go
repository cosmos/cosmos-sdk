package runtime

import (
	"context"
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/runtime/protoiface"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/blockinfo"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/gas"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type service struct {
	moduleKey         depinject.ModuleKey
	rootAddress       []byte
	kvStoreKey        *storetypes.KVStoreKey
	memoryStoreKey    *storetypes.MemoryStoreKey
	transientStoreKey *storetypes.TransientStoreKey
}

func (s service) DerivedClient(key []byte) appmodule.InterModuleClient {
	panic("ADR-033 isn't implemented yet")
}

func (s service) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	return fmt.Errorf("ADR-033 is not implemented yet")
}

func (s service) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("streaming clients are not supported")
}

func (s service) Address() []byte {
	return s.rootAddress
}

func (s service) OpenKVStore(ctx context.Context) store.KVStore {
	return sdk.UnwrapSDKContext(ctx).KVStore(s.kvStoreKey)
}

func (s service) OpenMemoryStore(ctx context.Context) store.KVStore {
	return sdk.UnwrapSDKContext(ctx).KVStore(s.memoryStoreKey)
}

func (s service) OpenTransientStore(ctx context.Context) store.KVStore {
	return sdk.UnwrapSDKContext(ctx).KVStore(s.transientStoreKey)
}

func (s service) GetEventManager(ctx context.Context) event.Manager {
	return &eventManager{legacyMgr: sdk.UnwrapSDKContext(ctx).EventManager()}
}

func (s service) GetBlockInfo(ctx context.Context) blockinfo.BlockInfo {
	header := sdk.UnwrapSDKContext(ctx).BlockHeader()
	return blockinfo.BlockInfo{
		ChainID: header.ChainID,
		Height:  header.Height,
		Time:    header.Time,
		Hash:    header.LastCommitHash,
	}
}

func (s service) GetGasMeter(ctx context.Context) gas.Meter {
	return sdk.UnwrapSDKContext(ctx).GasMeter()
}

func (s service) GetBlockGasMeter(ctx context.Context) gas.Meter {
	return sdk.UnwrapSDKContext(ctx).BlockGasMeter()
}

func (s service) WithGasMeter(ctx context.Context, meter gas.Meter) context.Context {
	return sdk.UnwrapSDKContext(ctx).WithGasMeter(meter)
}

func (s service) WithBlockGasMeter(ctx context.Context, meter gas.Meter) context.Context {
	return sdk.UnwrapSDKContext(ctx).WithBlockGasMeter(meter)
}

var _ appmodule.Service = service{}

type eventManager struct {
	legacyMgr *sdk.EventManager
}

func (e eventManager) Emit(msg protoiface.MessageV1) error {
	return e.legacyMgr.EmitTypedEvent(msg)
}

func (e eventManager) EmitLegacy(eventType string, attrs ...event.LegacyEventAttribute) error {
	legacyAttrs := make([]abci.EventAttribute, len(attrs))
	for i, attr := range attrs {
		legacyAttrs[i] = abci.EventAttribute{
			Key:   attr.Key,
			Value: attr.Value,
		}
	}
	e.legacyMgr.EmitEvent(sdk.Event{
		Type:       eventType,
		Attributes: legacyAttrs,
	})
	return nil
}

func (e eventManager) EmitClientOnly(msg protoiface.MessageV1) error {
	return e.legacyMgr.EmitTypedEvent(msg)
}

var _ event.Manager = eventManager{}
