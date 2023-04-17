package ormtable_test

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/orm/internal/testkv"
	"github.com/cosmos/cosmos-sdk/orm/testing/ormtest"

	dbm "github.com/cosmos/cosmos-db"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/orm/internal/testpb"
	"github.com/cosmos/cosmos-sdk/orm/model/ormtable"
	"github.com/cosmos/cosmos-sdk/orm/types/kv"
)

func initBalanceTable(t testing.TB) testpb.BalanceTable {
	table, err := ormtable.Build(ormtable.Options{
		MessageType: (&testpb.Balance{}).ProtoReflect().Type(),
	})
	assert.NilError(t, err)

	balanceTable, err := testpb.NewBalanceTable(table)
	assert.NilError(t, err)

	return balanceTable
}

func BenchmarkMemory(b *testing.B) {
	bench(b, func(tb testing.TB) ormtable.Backend {
		return ormtest.NewMemoryBackend()
	})
}

func BenchmarkLevelDB(b *testing.B) {
	bench(b, testkv.NewGoLevelDBBackend)
}

func bench(b *testing.B, newBackend func(testing.TB) ormtable.Backend) {
	b.Run("insert", func(b *testing.B) {
		b.StopTimer()
		ctx := ormtable.WrapContextDefault(newBackend(b))
		b.StartTimer()
		benchInsert(b, ctx)
	})
	b.Run("update", func(b *testing.B) {
		b.StopTimer()
		ctx := ormtable.WrapContextDefault(newBackend(b))
		benchInsert(b, ctx)
		b.StartTimer()
		benchUpdate(b, ctx)
	})
	b.Run("get", func(b *testing.B) {
		b.StopTimer()
		ctx := ormtable.WrapContextDefault(newBackend(b))
		benchInsert(b, ctx)
		b.StartTimer()
		benchGet(b, ctx)
	})
	b.Run("delete", func(b *testing.B) {
		b.StopTimer()
		ctx := ormtable.WrapContextDefault(newBackend(b))
		benchInsert(b, ctx)
		b.StartTimer()
		benchDelete(b, ctx)
	})
}

func benchInsert(b *testing.B, ctx context.Context) { //nolint:revive // ignore for benchmark
	balanceTable := initBalanceTable(b)
	for i := 0; i < b.N; i++ {
		assert.NilError(b, balanceTable.Insert(ctx, &testpb.Balance{
			Address: fmt.Sprintf("acct%d", i),
			Denom:   "bar",
			Amount:  10,
		}))
	}
}

func benchUpdate(b *testing.B, ctx context.Context) { //nolint:revive // ignore for benchmark
	balanceTable := initBalanceTable(b)
	for i := 0; i < b.N; i++ {
		assert.NilError(b, balanceTable.Update(ctx, &testpb.Balance{
			Address: fmt.Sprintf("acct%d", i),
			Denom:   "bar",
			Amount:  11,
		}))
	}
}

func benchGet(b *testing.B, ctx context.Context) { //nolint:revive // ignore for benchmark
	balanceTable := initBalanceTable(b)
	for i := 0; i < b.N; i++ {
		balance, err := balanceTable.Get(ctx, fmt.Sprintf("acct%d", i), "bar")
		assert.NilError(b, err)
		assert.Equal(b, uint64(10), balance.Amount)
	}
}

func benchDelete(b *testing.B, ctx context.Context) { //nolint:revive // ignore for benchmark
	balanceTable := initBalanceTable(b)
	for i := 0; i < b.N; i++ {
		assert.NilError(b, balanceTable.Delete(ctx, &testpb.Balance{
			Address: fmt.Sprintf("acct%d", i),
			Denom:   "bar",
		}))
	}
}

//
// Manually written versions of insert, update, delete and get for testpb.Balance
//

const (
	addressDenomPrefix byte = iota
	denomAddressPrefix
)

func insertBalance(store kv.Store, balance *testpb.Balance) error {
	denom := balance.Denom
	balance.Denom = ""
	addr := balance.Address
	balance.Address = ""

	addressDenomKey := []byte{addressDenomPrefix}
	addressDenomKey = append(addressDenomKey, []byte(addr)...)
	addressDenomKey = append(addressDenomKey, 0x0)
	addressDenomKey = append(addressDenomKey, []byte(denom)...)
	has, err := store.Has(addressDenomKey)
	if err != nil {
		return err
	}

	if has {
		return fmt.Errorf("already exists")
	}

	bz, err := proto.Marshal(balance)
	if err != nil {
		return err
	}
	balance.Denom = denom
	balance.Address = addr

	err = store.Set(addressDenomKey, bz)
	if err != nil {
		return err
	}

	// set denom address index
	denomAddressKey := []byte{denomAddressPrefix}
	denomAddressKey = append(denomAddressKey, []byte(balance.Denom)...)
	denomAddressKey = append(denomAddressKey, 0x0)
	denomAddressKey = append(denomAddressKey, []byte(balance.Address)...)
	err = store.Set(denomAddressKey, []byte{})
	if err != nil {
		return err
	}

	return nil
}

func updateBalance(store kv.Store, balance *testpb.Balance) error {
	denom := balance.Denom
	balance.Denom = ""
	addr := balance.Address
	balance.Address = ""
	bz, err := proto.Marshal(balance)
	if err != nil {
		return err
	}
	balance.Denom = denom
	balance.Address = addr

	addressDenomKey := []byte{addressDenomPrefix}
	addressDenomKey = append(addressDenomKey, []byte(addr)...)
	addressDenomKey = append(addressDenomKey, 0x0)
	addressDenomKey = append(addressDenomKey, []byte(denom)...)

	return store.Set(addressDenomKey, bz)
}

func deleteBalance(store kv.Store, balance *testpb.Balance) error {
	denom := balance.Denom
	addr := balance.Address

	addressDenomKey := []byte{addressDenomPrefix}
	addressDenomKey = append(addressDenomKey, []byte(addr)...)
	addressDenomKey = append(addressDenomKey, 0x0)
	addressDenomKey = append(addressDenomKey, []byte(denom)...)
	err := store.Delete(addressDenomKey)
	if err != nil {
		return err
	}

	denomAddressKey := []byte{denomAddressPrefix}
	denomAddressKey = append(denomAddressKey, []byte(balance.Denom)...)
	denomAddressKey = append(denomAddressKey, 0x0)
	denomAddressKey = append(denomAddressKey, []byte(balance.Address)...)
	return store.Delete(denomAddressKey)
}

func getBalance(store kv.Store, address, denom string) (*testpb.Balance, error) {
	addressDenomKey := []byte{addressDenomPrefix}
	addressDenomKey = append(addressDenomKey, []byte(address)...)
	addressDenomKey = append(addressDenomKey, 0x0)
	addressDenomKey = append(addressDenomKey, []byte(denom)...)

	bz, err := store.Get(addressDenomKey)
	if err != nil {
		return nil, err
	}

	if bz == nil {
		return nil, fmt.Errorf("not found")
	}

	balance := testpb.Balance{}
	err = proto.Unmarshal(bz, &balance)
	if err != nil {
		return nil, err
	}

	balance.Address = address
	balance.Denom = denom

	return &balance, nil
}

func BenchmarkManualInsertMemory(b *testing.B) {
	benchManual(b, func() (dbm.DB, error) {
		return dbm.NewMemDB(), nil
	})
}

func BenchmarkManualInsertLevelDB(b *testing.B) {
	benchManual(b, func() (dbm.DB, error) {
		return dbm.NewGoLevelDB("test", b.TempDir(), nil)
	})
}

func benchManual(b *testing.B, newStore func() (dbm.DB, error)) {
	b.Run("insert", func(b *testing.B) {
		b.StopTimer()
		store, err := newStore()
		assert.NilError(b, err)
		b.StartTimer()
		benchManualInsert(b, store)
	})
	b.Run("update", func(b *testing.B) {
		b.StopTimer()
		store, err := newStore()
		assert.NilError(b, err)
		benchManualInsert(b, store)
		b.StartTimer()
		benchManualUpdate(b, store)
	})
	b.Run("get", func(b *testing.B) {
		b.StopTimer()
		store, err := newStore()
		assert.NilError(b, err)
		benchManualInsert(b, store)
		b.StartTimer()
		benchManualGet(b, store)
	})
	b.Run("delete", func(b *testing.B) {
		b.StopTimer()
		store, err := newStore()
		assert.NilError(b, err)
		benchManualInsert(b, store)
		b.StartTimer()
		benchManualDelete(b, store)
	})
}

func benchManualInsert(b *testing.B, store kv.Store) {
	for i := 0; i < b.N; i++ {
		assert.NilError(b, insertBalance(store, &testpb.Balance{
			Address: fmt.Sprintf("acct%d", i),
			Denom:   "bar",
			Amount:  10,
		}))
	}
}

func benchManualUpdate(b *testing.B, store kv.Store) {
	for i := 0; i < b.N; i++ {
		assert.NilError(b, updateBalance(store, &testpb.Balance{
			Address: fmt.Sprintf("acct%d", i),
			Denom:   "bar",
			Amount:  11,
		}))
	}
}

func benchManualDelete(b *testing.B, store kv.Store) {
	for i := 0; i < b.N; i++ {
		assert.NilError(b, deleteBalance(store, &testpb.Balance{
			Address: fmt.Sprintf("acct%d", i),
			Denom:   "bar",
		}))
	}
}

func benchManualGet(b *testing.B, store kv.Store) {
	for i := 0; i < b.N; i++ {
		balance, err := getBalance(store, fmt.Sprintf("acct%d", i), "bar")
		assert.NilError(b, err)
		assert.Equal(b, uint64(10), balance.Amount)
	}
}
