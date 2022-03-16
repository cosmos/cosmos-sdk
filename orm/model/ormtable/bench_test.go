package ormtable_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/orm/internal/testkv"
	"github.com/cosmos/cosmos-sdk/orm/testing/ormtest"

	"github.com/gogo/protobuf/proto"
	dbm "github.com/tendermint/tm-db"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/orm/internal/testpb"
	"github.com/cosmos/cosmos-sdk/orm/model/ormtable"
	"github.com/cosmos/cosmos-sdk/orm/types/kv"
)

var balanceTable testpb.BalanceTable

func init() {
	table, err := ormtable.Build(ormtable.Options{
		MessageType: (&testpb.Balance{}).ProtoReflect().Type(),
	})
	if err != nil {
		panic(err)
	}

	balanceTable, err = testpb.NewBalanceTable(table)
	if err != nil {
		panic(err)
	}
}

func BenchmarkInsertMemory(b *testing.B) {
	b.StopTimer()
	ctx := ormtable.WrapContextDefault(ormtest.NewMemoryBackend())
	b.StartTimer()
	benchInsert(b, ctx)
}

func BenchmarkInsertLevelDB(b *testing.B) {
	b.StopTimer()
	ctx := ormtable.WrapContextDefault(testkv.NewGoLevelDBBackend(b))
	b.StartTimer()
	benchInsert(b, ctx)
}

func benchInsert(b *testing.B, ctx context.Context) {
	for i := 0; i < b.N; i++ {
		assert.NilError(b, balanceTable.Insert(ctx, &testpb.Balance{
			Address: fmt.Sprintf("acct%d", i),
			Denom:   "ba",
			Amount:  10,
		}))
	}
}

const (
	addressDenomPrefix byte = iota
	defnomAddressPrefix
)

func insertBalance(store kv.Store, balance *testpb.Balance) error {
	addressDenomKey := []byte{addressDenomPrefix}
	addressDenomKey = append(addressDenomKey, []byte(balance.Address)...)
	addressDenomKey = append(addressDenomKey, 0x0)
	addressDenomKey = append(addressDenomKey, []byte(balance.Denom)...)
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

	err = store.Set(addressDenomKey, bz)
	if err != nil {
		return err
	}

	denomAddressKey := []byte{defnomAddressPrefix}
	denomAddressKey = append(denomAddressKey, []byte(balance.Denom)...)
	denomAddressKey = append(denomAddressKey, 0x0)
	denomAddressKey = append(denomAddressKey, []byte(balance.Address)...)
	err = store.Set(denomAddressKey, []byte{})
	if err != nil {
		return err
	}

	return nil
}

func BenchmarkManualInsertMemory(b *testing.B) {
	b.StopTimer()
	store := dbm.NewMemDB()
	b.StartTimer()
	benchManualInsert(b, store)
}

func BenchmarkManualInsertLevelDB(b *testing.B) {
	b.StopTimer()
	store, err := dbm.NewGoLevelDB("test", b.TempDir())
	assert.NilError(b, err)
	b.StartTimer()
	benchManualInsert(b, store)
}

func benchManualInsert(b *testing.B, store kv.Store) {
	for i := 0; i < b.N; i++ {
		assert.NilError(b, insertBalance(store, &testpb.Balance{
			Address: fmt.Sprintf("acct%d", i),
			Denom:   "ba",
			Amount:  10,
		}))
	}
}
