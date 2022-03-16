package ormtable_test

import (
	"fmt"
	"testing"

	"github.com/gogo/protobuf/proto"
	dbm "github.com/tendermint/tm-db"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/orm/internal/testpb"
	"github.com/cosmos/cosmos-sdk/orm/model/ormtable"
	"github.com/cosmos/cosmos-sdk/orm/testing/ormtest"
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

func BenchmarkInsert(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ctx := ormtable.WrapContextDefault(ormtest.NewMemoryBackend())
		b.StartTimer()
		assert.NilError(b, balanceTable.Insert(ctx, &testpb.Balance{
			Address: "foo",
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

func BenchmarkManualInsert(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		store := dbm.NewMemDB()
		b.StartTimer()
		assert.NilError(b, insertBalance(store, &testpb.Balance{
			Address: "foo",
			Denom:   "ba",
			Amount:  10,
		}))
	}
}
