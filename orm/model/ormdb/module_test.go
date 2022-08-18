package ormdb_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	ormv1alpha1 "cosmossdk.io/api/cosmos/orm/v1alpha1"

	"github.com/golang/mock/gomock"

	"github.com/cosmos/cosmos-sdk/orm/testing/ormmocks"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/golden"

	"github.com/cosmos/cosmos-sdk/orm/internal/testkv"
	"github.com/cosmos/cosmos-sdk/orm/internal/testpb"
	"github.com/cosmos/cosmos-sdk/orm/model/ormdb"
	"github.com/cosmos/cosmos-sdk/orm/model/ormtable"
	"github.com/cosmos/cosmos-sdk/orm/testing/ormtest"
	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"
	"github.com/cosmos/cosmos-sdk/orm/types/ormjson"
)

// These tests use a simulated bank keeper. Addresses and balances use
// string and uint64 types respectively for simplicity.

var TestBankSchema = &ormv1alpha1.ModuleSchemaDescriptor{
	SchemaFile: []*ormv1alpha1.ModuleSchemaDescriptor_FileEntry{
		{
			Id:            1,
			ProtoFileName: testpb.File_testpb_bank_proto.Path(),
		},
	},
}

type keeper struct {
	store testpb.BankStore
}

func NewKeeper(db ormdb.ModuleDB) (Keeper, error) {
	store, err := testpb.NewBankStore(db)
	return keeper{store}, err
}

type Keeper interface {
	Send(ctx context.Context, from, to, denom string, amount uint64) error
	Mint(ctx context.Context, acct, denom string, amount uint64) error
	Burn(ctx context.Context, acct, denom string, amount uint64) error
	Balance(ctx context.Context, acct, denom string) (uint64, error)
	Supply(ctx context.Context, denom string) (uint64, error)
}

func (k keeper) Send(ctx context.Context, from, to, denom string, amount uint64) error {
	err := k.safeSubBalance(ctx, from, denom, amount)
	if err != nil {
		return err
	}

	return k.addBalance(ctx, to, denom, amount)
}

func (k keeper) Mint(ctx context.Context, acct, denom string, amount uint64) error {
	supply, err := k.store.SupplyTable().Get(ctx, denom)
	if err != nil && !ormerrors.IsNotFound(err) {
		return err
	}

	if supply == nil {
		supply = &testpb.Supply{Denom: denom, Amount: amount}
	} else {
		supply.Amount = supply.Amount + amount
	}

	err = k.store.SupplyTable().Save(ctx, supply)
	if err != nil {
		return err
	}

	return k.addBalance(ctx, acct, denom, amount)
}

func (k keeper) Burn(ctx context.Context, acct, denom string, amount uint64) error {
	supplyStore := k.store.SupplyTable()
	supply, err := supplyStore.Get(ctx, denom)
	if err != nil {
		return err
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
	balance, err := k.store.BalanceTable().Get(ctx, acct, denom)
	if err != nil {
		if ormerrors.IsNotFound(err) {
			return 0, nil
		}

		return 0, err
	}
	return balance.Amount, err
}

func (k keeper) Supply(ctx context.Context, denom string) (uint64, error) {
	supply, err := k.store.SupplyTable().Get(ctx, denom)
	if supply == nil {
		if ormerrors.IsNotFound(err) {
			return 0, nil
		}

		return 0, err
	}
	return supply.Amount, err
}

func (k keeper) addBalance(ctx context.Context, acct, denom string, amount uint64) error {
	balance, err := k.store.BalanceTable().Get(ctx, acct, denom)
	if err != nil && !ormerrors.IsNotFound(err) {
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

	return k.store.BalanceTable().Save(ctx, balance)
}

func (k keeper) safeSubBalance(ctx context.Context, acct, denom string, amount uint64) error {
	balanceStore := k.store.BalanceTable()
	balance, err := balanceStore.Get(ctx, acct, denom)
	if err != nil {
		return err
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

func TestModuleDB(t *testing.T) {
	// create db & debug context
	db, err := ormdb.NewModuleDB(TestBankSchema, ormdb.ModuleDBOptions{})
	assert.NilError(t, err)
	debugBuf := &strings.Builder{}
	backend := ormtest.NewMemoryBackend()
	ctx := ormtable.WrapContextDefault(testkv.NewDebugBackend(
		backend,
		&testkv.EntryCodecDebugger{
			EntryCodec: db,
			Print:      func(s string) { debugBuf.WriteString(s + "\n") },
		},
	))

	// create keeper
	k, err := NewKeeper(db)
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
	it, err := backend.CommitmentStore().Iterator(nil, nil)
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

	// check JSON
	target := ormjson.NewRawMessageTarget()
	assert.NilError(t, db.DefaultJSON(target))
	rawJson, err := target.JSON()
	assert.NilError(t, err)
	golden.Assert(t, string(rawJson), "default_json.golden")

	target = ormjson.NewRawMessageTarget()
	assert.NilError(t, db.ExportJSON(ctx, target))
	rawJson, err = target.JSON()
	assert.NilError(t, err)

	goodJSON := `{
  "testpb.Supply": []
}`
	source, err := ormjson.NewRawMessageSource(json.RawMessage(goodJSON))
	assert.NilError(t, err)
	assert.NilError(t, db.ValidateJSON(source))
	assert.NilError(t, db.ImportJSON(ormtable.WrapContextDefault(ormtest.NewMemoryBackend()), source))

	badJSON := `{
  "testpb.Balance": 5,
  "testpb.Supply": {}
}
`
	source, err = ormjson.NewRawMessageSource(json.RawMessage(badJSON))
	assert.NilError(t, err)
	assert.ErrorIs(t, db.ValidateJSON(source), ormerrors.JSONValidationError)

	backend2 := ormtest.NewMemoryBackend()
	ctx2 := ormtable.WrapContextDefault(backend2)
	source, err = ormjson.NewRawMessageSource(rawJson)
	assert.NilError(t, err)
	assert.NilError(t, db.ValidateJSON(source))
	assert.NilError(t, db.ImportJSON(ctx2, source))
	testkv.AssertBackendsEqual(t, backend, backend2)
}

func TestHooks(t *testing.T) {
	ctrl := gomock.NewController(t)
	db, err := ormdb.NewModuleDB(TestBankSchema, ormdb.ModuleDBOptions{})
	assert.NilError(t, err)
	validateHooks := ormmocks.NewMockValidateHooks(ctrl)
	writeHooks := ormmocks.NewMockWriteHooks(ctrl)
	ctx := ormtable.WrapContextDefault(ormtest.NewMemoryBackend().
		WithValidateHooks(validateHooks).
		WithWriteHooks(writeHooks))
	k, err := NewKeeper(db)
	assert.NilError(t, err)

	denom := "foo"
	acct1 := "bob"
	acct2 := "sally"

	validateHooks.EXPECT().ValidateInsert(gomock.Any(), ormmocks.Eq(&testpb.Balance{Address: acct1, Denom: denom, Amount: 10}))
	validateHooks.EXPECT().ValidateInsert(gomock.Any(), ormmocks.Eq(&testpb.Supply{Denom: denom, Amount: 10}))
	writeHooks.EXPECT().OnInsert(gomock.Any(), ormmocks.Eq(&testpb.Balance{Address: acct1, Denom: denom, Amount: 10}))
	writeHooks.EXPECT().OnInsert(gomock.Any(), ormmocks.Eq(&testpb.Supply{Denom: denom, Amount: 10}))
	assert.NilError(t, k.Mint(ctx, acct1, denom, 10))

	validateHooks.EXPECT().ValidateUpdate(
		gomock.Any(),
		ormmocks.Eq(&testpb.Balance{Address: acct1, Denom: denom, Amount: 10}),
		ormmocks.Eq(&testpb.Balance{Address: acct1, Denom: denom, Amount: 5}),
	)
	validateHooks.EXPECT().ValidateInsert(
		gomock.Any(),
		ormmocks.Eq(&testpb.Balance{Address: acct2, Denom: denom, Amount: 5}),
	)
	writeHooks.EXPECT().OnUpdate(
		gomock.Any(),
		ormmocks.Eq(&testpb.Balance{Address: acct1, Denom: denom, Amount: 10}),
		ormmocks.Eq(&testpb.Balance{Address: acct1, Denom: denom, Amount: 5}),
	)
	writeHooks.EXPECT().OnInsert(
		gomock.Any(),
		ormmocks.Eq(&testpb.Balance{Address: acct2, Denom: denom, Amount: 5}),
	)
	assert.NilError(t, k.Send(ctx, acct1, acct2, denom, 5))

	validateHooks.EXPECT().ValidateUpdate(
		gomock.Any(),
		ormmocks.Eq(&testpb.Supply{Denom: denom, Amount: 10}),
		ormmocks.Eq(&testpb.Supply{Denom: denom, Amount: 5}),
	)
	validateHooks.EXPECT().ValidateDelete(
		gomock.Any(),
		ormmocks.Eq(&testpb.Balance{Address: acct1, Denom: denom, Amount: 5}),
	)
	writeHooks.EXPECT().OnUpdate(
		gomock.Any(),
		ormmocks.Eq(&testpb.Supply{Denom: denom, Amount: 10}),
		ormmocks.Eq(&testpb.Supply{Denom: denom, Amount: 5}),
	)
	writeHooks.EXPECT().OnDelete(
		gomock.Any(),
		ormmocks.Eq(&testpb.Balance{Address: acct1, Denom: denom, Amount: 5}),
	)
	assert.NilError(t, k.Burn(ctx, acct1, denom, 5))
}

func TestGetBackendResolver(t *testing.T) {
	backend := ormtest.NewMemoryBackend()
	getResolver := func(storageType ormv1alpha1.StorageType) (ormtable.BackendResolver, error) {
		switch storageType {
		case ormv1alpha1.StorageType_STORAGE_TYPE_MEMORY:
			return func(ctx context.Context) (ormtable.ReadBackend, error) {
				return backend, nil
			}, nil
		default:
			return nil, fmt.Errorf("storage type %s unsupported", storageType)
		}
	}
	_, err := ormdb.NewModuleDB(TestBankSchema, ormdb.ModuleDBOptions{
		GetBackendResolver: getResolver,
	})
	assert.ErrorContains(t, err, "unsupported")

	_, err = ormdb.NewModuleDB(&ormv1alpha1.ModuleSchemaDescriptor{
		SchemaFile: []*ormv1alpha1.ModuleSchemaDescriptor_FileEntry{
			{
				Id:            1,
				ProtoFileName: testpb.File_testpb_bank_proto.Path(),
				StorageType:   ormv1alpha1.StorageType_STORAGE_TYPE_MEMORY,
			},
		},
	}, ormdb.ModuleDBOptions{
		GetBackendResolver: getResolver,
	})
	assert.NilError(t, err)
}
