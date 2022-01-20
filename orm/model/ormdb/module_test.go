package ormdb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/orm/internal/testkv"

	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/orm/model/ormtable"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/internal/testpb"
	"github.com/cosmos/cosmos-sdk/orm/model/ormdb"
)

// These tests use a simulated bank keeper. Addresses and balances use
//  []byte and uint64 types respectively for simplicity.

var TestBankSchema = ormdb.ModuleSchema{
	FileDescriptors: map[uint32]protoreflect.FileDescriptor{
		1: testpb.File_testpb_bank_proto,
	},
}

type keeper struct {
	balanceTable             ormtable.Table
	balanceAddressDenomIndex ormtable.UniqueIndex
	balanceDenomIndex        ormtable.Index

	supplyTable      ormtable.Table
	supplyDenomIndex ormtable.UniqueIndex
}

func (k keeper) Send(ctx context.Context, from, to []byte, denom string, amount uint64) error {
	err := k.safeSubBalance(ctx, from, denom, amount)
	if err != nil {
		return err
	}

	return k.addBalance(ctx, to, denom, amount)
}

func (k keeper) Mint(ctx context.Context, acct []byte, denom string, amount uint64) error {
	var supply testpb.Supply
	found, err := k.supplyDenomIndex.Get(ctx, &supply, denom)
	if err != nil {
		return err
	}

	if !found {
		supply.Denom = denom
		supply.Amount = amount
	} else {
		supply.Amount = supply.Amount + amount
	}

	err = k.supplyTable.Save(ctx, &supply)
	if err != nil {
		return err
	}

	return k.addBalance(ctx, acct, denom, amount)
}

func (k keeper) Burn(ctx context.Context, acct []byte, denom string, amount uint64) error {
	var supply testpb.Supply
	found, err := k.supplyDenomIndex.Get(ctx, &supply, denom)
	if err != nil {
		return err
	}

	if !found {
		return fmt.Errorf("no supply for %s", denom)
	}

	if amount > supply.Amount {
		return fmt.Errorf("insufficient supply")
	}

	supply.Amount = supply.Amount - amount

	if supply.Amount == 0 {
		err = k.supplyTable.Delete(ctx, &supply)
	} else {
		err = k.supplyTable.Save(ctx, &supply)
	}
	if err != nil {
		return err
	}

	return k.safeSubBalance(ctx, acct, denom, amount)
}

func (k keeper) Balance(ctx context.Context, acct []byte, denom string) (uint64, error) {
	var balance testpb.Balance
	found, err := k.balanceAddressDenomIndex.Get(ctx, &balance, acct, denom)
	if err != nil || !found {
		return 0, err
	}

	return balance.Amount, nil
}

func (k keeper) Supply(ctx context.Context, denom string) (uint64, error) {
	var supply testpb.Supply
	found, err := k.supplyDenomIndex.Get(ctx, &supply, denom)
	if err != nil || !found {
		return 0, err
	}

	return supply.Amount, nil
}

func (k keeper) addBalance(ctx context.Context, acct []byte, denom string, amount uint64) error {
	var balance testpb.Balance
	found, err := k.balanceAddressDenomIndex.Get(ctx, &balance, acct, denom)
	if err != nil {
		return err
	}

	if !found {
		balance.Address = acct
		balance.Denom = denom
		balance.Amount = amount
	} else {
		balance.Amount = balance.Amount + amount
	}

	return k.balanceTable.Save(ctx, &balance)
}

func (k keeper) safeSubBalance(ctx context.Context, acct []byte, denom string, amount uint64) error {
	var balance testpb.Balance
	found, err := k.balanceAddressDenomIndex.Get(ctx, &balance, acct, denom)
	if err != nil {
		return err
	}

	if !found {
		return fmt.Errorf("acct %x has no balance for %s", acct, denom)
	}

	if amount > balance.Amount {
		return fmt.Errorf("insufficient funds")
	}

	balance.Amount = balance.Amount - amount

	if balance.Amount == 0 {
		return k.balanceTable.Delete(ctx, &balance)
	} else {
		return k.balanceTable.Save(ctx, &balance)
	}
}

func newKeeper(db ormdb.ModuleDB) keeper {
	k := keeper{
		balanceTable: db.GetTable(&testpb.Balance{}),
		supplyTable:  db.GetTable(&testpb.Supply{}),
	}

	k.balanceAddressDenomIndex = k.balanceTable.GetUniqueIndex("address,denom")
	k.balanceDenomIndex = k.balanceTable.GetIndex("denom")
	k.supplyDenomIndex = k.supplyTable.GetUniqueIndex("denom")

	return k
}

func TestModuleDB(t *testing.T) {
	// create db & ctx
	db, err := ormdb.NewModuleDB(TestBankSchema, ormdb.ModuleDBOptions{})
	assert.NilError(t, err)
	ctx := ormtable.WrapContextDefault(testkv.NewSplitMemBackend())

	// create keeper
	k := newKeeper(db)
	assert.Assert(t, k.balanceTable != nil)
	assert.Assert(t, k.balanceAddressDenomIndex != nil)
	assert.Assert(t, k.balanceDenomIndex != nil)
	assert.Assert(t, k.supplyTable != nil)
	assert.Assert(t, k.supplyDenomIndex != nil)

	// mint coins
	denom := "foo"
	acct1 := []byte{0, 1, 2, 3}
	err = k.Mint(ctx, acct1, denom, 100)
	assert.NilError(t, err)
	bal, err := k.Balance(ctx, acct1, denom)
	assert.NilError(t, err)
	assert.Equal(t, uint64(100), bal)
	supply, err := k.Supply(ctx, denom)
	assert.NilError(t, err)
	assert.Equal(t, uint64(100), supply)

	// send coins
	acct2 := []byte{3, 2, 1, 0}
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
}
