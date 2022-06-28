package orm

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/db/memdb"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/store/gaskv"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/store/v2alpha1/multi"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MockContext struct {
	db     *memdb.MemDB
	store  store.CommitMultiStore
	params multi.StoreParams
}

func NewMockContext() *MockContext {
	db := memdb.NewDB()
	params := multi.DefaultStoreParams()
	cms, err := multi.NewStore(db, params)
	if err != nil {
		panic(err)
	}
	return &MockContext{
		db:     db,
		store:  cms,
		params: params,
	}
}

func (m *MockContext) KVStore(key store.Key) sdk.KVStore {
	if m.store.HasKVStore(key) {
		return m.store.GetKVStore(key)
	}
	err := m.store.Close()
	if err != nil {
		panic(err)
	}
	if err = m.params.RegisterSubstore(key, storetypes.StoreTypePersistent); err != nil {
		panic(err)
	}
	if m.store, err = multi.NewStore(m.db, m.params); err != nil {
		panic(err)
	}
	return m.store.GetKVStore(key)
}

type debuggingGasMeter struct {
	g store.GasMeter
}

func (d debuggingGasMeter) GasConsumed() store.Gas {
	return d.g.GasConsumed()
}

func (d debuggingGasMeter) GasRemaining() store.Gas {
	return d.g.GasRemaining()
}

func (d debuggingGasMeter) GasConsumedToLimit() store.Gas {
	return d.g.GasConsumedToLimit()
}

func (d debuggingGasMeter) RefundGas(amount uint64, descriptor string) {
	d.g.RefundGas(amount, descriptor)
}

func (d debuggingGasMeter) Limit() store.Gas {
	return d.g.Limit()
}

func (d debuggingGasMeter) ConsumeGas(amount store.Gas, descriptor string) {
	fmt.Printf("++ Consuming gas: %q :%d\n", descriptor, amount)
	d.g.ConsumeGas(amount, descriptor)
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
	GasMeter sdk.GasMeter
}

func NewGasCountingMockContext() *GasCountingMockContext {
	return &GasCountingMockContext{
		GasMeter: &debuggingGasMeter{sdk.NewInfiniteGasMeter()},
	}
}

func (g GasCountingMockContext) KVStore(store sdk.KVStore) sdk.KVStore {
	return gaskv.NewStore(store, g.GasMeter, storetypes.KVGasConfig())
}

func (g GasCountingMockContext) GasConsumed() store.Gas {
	return g.GasMeter.GasConsumed()
}

func (g GasCountingMockContext) GasRemaining() store.Gas {
	return g.GasMeter.GasRemaining()
}

func (g *GasCountingMockContext) ResetGasMeter() {
	g.GasMeter = sdk.NewInfiniteGasMeter()
}
