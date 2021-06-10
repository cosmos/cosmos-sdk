package orm

import (
	"fmt"
	"io"

	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/store/gaskv"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MockContext struct {
	db    *dbm.MemDB
	store types.CommitMultiStore
}

func NewMockContext() *MockContext {
	db := dbm.NewMemDB()
	return &MockContext{
		db:    dbm.NewMemDB(),
		store: store.NewCommitMultiStore(db),
	}
}

func (m MockContext) KVStore(key sdk.StoreKey) sdk.KVStore {
	if s := m.store.GetCommitKVStore(key); s != nil {
		return s
	}
	m.store.MountStoreWithDB(key, sdk.StoreTypeIAVL, m.db)
	if err := m.store.LoadLatestVersion(); err != nil {
		panic(err)
	}
	return m.store.GetCommitKVStore(key)
}

type debuggingGasMeter struct {
	g types.GasMeter
}

func (d debuggingGasMeter) GasConsumed() types.Gas {
	return d.g.GasConsumed()
}

func (d debuggingGasMeter) GasConsumedToLimit() types.Gas {
	return d.g.GasConsumedToLimit()
}

func (d debuggingGasMeter) Limit() types.Gas {
	return d.g.Limit()
}

func (d debuggingGasMeter) ConsumeGas(amount types.Gas, descriptor string) {
	fmt.Printf("++ Consuming gas: %q :%d\n", descriptor, amount)
	d.g.ConsumeGas(amount, descriptor)
}

func (d debuggingGasMeter) RefundGas(amount types.Gas, descriptor string) {
	fmt.Printf("-- Refunding gas: %q :%d\n", descriptor, amount)
	d.g.RefundGas(amount, descriptor)
}

func (d debuggingGasMeter) IsPastLimit() bool {
	return d.g.IsPastLimit()
}

func (d debuggingGasMeter) IsOutOfGas() bool {
	return d.g.IsOutOfGas()
}

func (d debuggingGasMeter) String() string {
	return d.g.String()
}

type GasCountingMockContext struct {
	parent   HasKVStore
	GasMeter sdk.GasMeter
}

func NewGasCountingMockContext(parent HasKVStore) *GasCountingMockContext {
	return &GasCountingMockContext{
		parent:   parent,
		GasMeter: &debuggingGasMeter{sdk.NewInfiniteGasMeter()},
	}
}

func (g GasCountingMockContext) KVStore(key sdk.StoreKey) sdk.KVStore {
	return gaskv.NewStore(g.parent.KVStore(key), g.GasMeter, types.KVGasConfig())
}

func (g GasCountingMockContext) GasConsumed() types.Gas {
	return g.GasMeter.GasConsumed()
}

func (g *GasCountingMockContext) ResetGasMeter() {
	g.GasMeter = sdk.NewInfiniteGasMeter()
}

type AlwaysPanicKVStore struct{}

func (a AlwaysPanicKVStore) GetStoreType() types.StoreType {
	panic("Not implemented")
}

func (a AlwaysPanicKVStore) CacheWrap() types.CacheWrap {
	panic("Not implemented")
}

func (a AlwaysPanicKVStore) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	panic("Not implemented")
}

func (a AlwaysPanicKVStore) Get(key []byte) []byte {
	panic("Not implemented")
}

func (a AlwaysPanicKVStore) Has(key []byte) bool {
	panic("Not implemented")
}

func (a AlwaysPanicKVStore) Set(key, value []byte) {
	panic("Not implemented")
}

func (a AlwaysPanicKVStore) Delete(key []byte) {
	panic("Not implemented")
}

func (a AlwaysPanicKVStore) Iterator(start, end []byte) types.Iterator {
	panic("Not implemented")
}

func (a AlwaysPanicKVStore) ReverseIterator(start, end []byte) types.Iterator {
	panic("Not implemented")
}
