package ormdb_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/golden"

	"github.com/cosmos/cosmos-sdk/orm/internal/testkv"
	"github.com/cosmos/cosmos-sdk/orm/internal/testpb"
	"github.com/cosmos/cosmos-sdk/orm/model/ormdb"
	"github.com/cosmos/cosmos-sdk/orm/model/ormtable"
)

// These tests use a simulated bank keeper. Addresses and balances use
// string and uint64 types respectively for simplicity.

var TestBankSchema = ormdb.ModuleSchema{
	FileDescriptors: map[uint32]protoreflect.FileDescriptor{
		1: testpb.File_testpb_bank_proto,
	},
}

type keeper struct {
	store testpb.BankStore
}

func (k keeper) Send(ctx context.Context, from, to, denom string, amount uint64) error {
	err := k.safeSubBalance(ctx, from, denom, amount)
	if err != nil {
		return err
	}

	return k.addBalance(ctx, to, denom, amount)
}

func (k keeper) Mint(ctx context.Context, acct, denom string, amount uint64) error {
	supply, err := k.store.SupplyStore().Get(ctx, denom)
	if err != nil {
		return err
	}

	if supply == nil {
		supply = &testpb.Supply{Denom: denom, Amount: amount}
	} else {
		supply.Amount = supply.Amount + amount
	}

	err = k.store.SupplyStore().Save(ctx, supply)
	if err != nil {
		return err
	}

	return k.addBalance(ctx, acct, denom, amount)
}

func (k keeper) Burn(ctx context.Context, acct, denom string, amount uint64) error {
	supplyStore := k.store.SupplyStore()
	supply, err := supplyStore.Get(ctx, denom)
	if err != nil {
		return err
	}

	if supply == nil {
		return fmt.Errorf("no supply for %s", denom)
	}

	if amount > supply.Amount {
		return fmt.Errorf("insufficient supply")
	}

	supply.Amount = supply.Amount - amount

	if supply.Amount == 0 {
		err = supplyStore.Delete(ctx, supply)
	} else {
		err = supplyStore.Save(ctx, supply)
	}
	if err != nil {
		return err
	}

	return k.safeSubBalance(ctx, acct, denom, amount)
}

func (k keeper) Balance(ctx context.Context, acct, denom string) (uint64, error) {
	balance, err := k.store.BalanceStore().Get(ctx, acct, denom)
	if balance == nil {
		return 0, err
	}
	return balance.Amount, err
}

func (k keeper) Supply(ctx context.Context, denom string) (uint64, error) {
	supply, err := k.store.SupplyStore().Get(ctx, denom)
	if supply == nil {
		return 0, err
	}
	return supply.Amount, err
}

func (k keeper) addBalance(ctx context.Context, acct, denom string, amount uint64) error {
	balance, err := k.store.BalanceStore().Get(ctx, acct, denom)
	if err != nil {
		return err
	}

	if balance == nil {
		balance = &testpb.Balance{
			Address: acct,
			Denom:   denom,
			Amount:  amount,
		}
	} else {
		balance.Amount = balance.Amount + amount
	}

	return k.store.BalanceStore().Save(ctx, balance)
}

func (k keeper) safeSubBalance(ctx context.Context, acct, denom string, amount uint64) error {
	balanceStore := k.store.BalanceStore()
	balance, err := balanceStore.Get(ctx, acct, denom)
	if err != nil {
		return err
	}

	if balance == nil {
		return fmt.Errorf("acct %x has no balance for %s", acct, denom)
	}

	if amount > balance.Amount {
		return fmt.Errorf("insufficient funds")
	}

	balance.Amount = balance.Amount - amount

	if balance.Amount == 0 {
		return balanceStore.Delete(ctx, balance)
	} else {
		return balanceStore.Save(ctx, balance)
	}
}

func newKeeper(db ormdb.ModuleDB) (keeper, error) {
	store, err := testpb.NewBankStore(db)
	return keeper{store}, err
}

func TestModuleDB(t *testing.T) {
	// create db & debug context
	db, err := ormdb.NewModuleDB(TestBankSchema, ormdb.ModuleDBOptions{})
	assert.NilError(t, err)
	debugBuf := &strings.Builder{}
	store := testkv.NewDebugBackend(
		testkv.NewSharedMemBackend(),
		&testkv.EntryCodecDebugger{
			EntryCodec: db,
			Print:      func(s string) { debugBuf.WriteString(s + "\n") },
		},
	)
	ctx := ormtable.WrapContextDefault(store)

	// create keeper
	k, err := newKeeper(db)
	assert.NilError(t, err)

	// mint coins
	denom := "foo"
	acct1 := "bob"
	err = k.Mint(ctx, acct1, denom, 100)
	assert.NilError(t, err)
	bal, err := k.Balance(ctx, acct1, denom)
	assert.NilError(t, err)
	assert.Equal(t, uint64(100), bal)
	supply, err := k.Supply(ctx, denom)
	assert.NilError(t, err)
	assert.Equal(t, uint64(100), supply)

	// send coins
	acct2 := "sally"
	err = k.Send(ctx, acct1, acct2, denom, 30)
	bal, err = k.Balance(ctx, acct1, denom)
	assert.NilError(t, err)
	assert.Equal(t, uint64(70), bal)
	bal, err = k.Balance(ctx, acct2, denom)
	assert.NilError(t, err)
	assert.Equal(t, uint64(30), bal)

	// burn coins
	err = k.Burn(ctx, acct2, denom, 3)
	bal, err = k.Balance(ctx, acct2, denom)
	assert.NilError(t, err)
	assert.Equal(t, uint64(27), bal)
	supply, err = k.Supply(ctx, denom)
	assert.NilError(t, err)
	assert.Equal(t, uint64(97), supply)

	// check debug output
	golden.Assert(t, debugBuf.String(), "bank_scenario.golden")

	// check decode & encode
	it, err := store.CommitmentStore().Iterator(nil, nil)
	assert.NilError(t, err)
	for it.Valid() {
		entry, err := db.DecodeEntry(it.Key(), it.Value())
		assert.NilError(t, err)
		k, v, err := db.EncodeEntry(entry)
		assert.NilError(t, err)
		assert.Assert(t, bytes.Equal(k, it.Key()))
		assert.Assert(t, bytes.Equal(v, it.Value()))
		it.Next()
	}
}
