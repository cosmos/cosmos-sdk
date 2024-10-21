package ormdb_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/golden"

	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	ormmodulev1alpha1 "cosmossdk.io/api/cosmos/orm/module/v1alpha1"
	ormv1alpha1 "cosmossdk.io/api/cosmos/orm/v1alpha1"
	"cosmossdk.io/core/genesis"
	corestore "cosmossdk.io/core/store"
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	_ "cosmossdk.io/orm" // required for ORM module registration
	"cosmossdk.io/orm/internal/testkv"
	"cosmossdk.io/orm/internal/testpb"
	"cosmossdk.io/orm/model/ormdb"
	"cosmossdk.io/orm/model/ormtable"
	"cosmossdk.io/orm/testing/ormmocks"
	"cosmossdk.io/orm/testing/ormtest"
	"cosmossdk.io/orm/types/ormerrors"
)

// These tests use a simulated bank keeper. Addresses and balances use
// string and uint64 types respectively for simplicity.

func init() {
	// this registers the test module with the module registry
	appconfig.RegisterModule(&testpb.Module{},
		appconfig.Provide(NewKeeper),
	)
}

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
	bankStore, err := testpb.NewBankStore(db)
	return keeper{bankStore}, err
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
		supply.Amount += amount
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
		return errors.New("insufficient supply")
	}

	supply.Amount -= amount

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
		balance.Amount += amount
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
		return errors.New("insufficient funds")
	}

	balance.Amount -= amount

	if balance.Amount == 0 {
		return balanceStore.Delete(ctx, balance)
	}

	return balanceStore.Save(ctx, balance)
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

	runSimpleBankTests(t, k, ctx)

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
	target := genesis.RawJSONTarget{}
	assert.NilError(t, db.GenesisHandler().DefaultGenesis(target.Target()))
	rawJSON, err := target.JSON()
	assert.NilError(t, err)
	golden.Assert(t, string(rawJSON), "default_json.golden")

	target = genesis.RawJSONTarget{}
	assert.NilError(t, db.GenesisHandler().ExportGenesis(ctx, target.Target()))
	rawJSON, err = target.JSON()
	assert.NilError(t, err)

	goodJSON := `{
  "testpb.Supply": []
}`
	source, err := genesis.SourceFromRawJSON(json.RawMessage(goodJSON))
	assert.NilError(t, err)
	assert.NilError(t, db.GenesisHandler().ValidateGenesis(source))
	assert.NilError(t, db.GenesisHandler().InitGenesis(ormtable.WrapContextDefault(ormtest.NewMemoryBackend()), source))

	badJSON := `{
  "testpb.Balance": 5,
  "testpb.Supply": {}
}
`
	source, err = genesis.SourceFromRawJSON(json.RawMessage(badJSON))
	assert.NilError(t, err)
	assert.ErrorIs(t, db.GenesisHandler().ValidateGenesis(source), ormerrors.JSONValidationError)

	backend2 := ormtest.NewMemoryBackend()
	ctx2 := ormtable.WrapContextDefault(backend2)
	source, err = genesis.SourceFromRawJSON(rawJSON)
	assert.NilError(t, err)
	assert.NilError(t, db.GenesisHandler().ValidateGenesis(source))
	assert.NilError(t, db.GenesisHandler().InitGenesis(ctx2, source))
	testkv.AssertBackendsEqual(t, backend, backend2)
}

func runSimpleBankTests(t *testing.T, k Keeper, ctx context.Context) {
	t.Helper()
	// mint coins
	denom := "foo"
	acct1 := "bob"
	err := k.Mint(ctx, acct1, denom, 100)
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
	assert.NilError(t, err)
	bal, err = k.Balance(ctx, acct1, denom)
	assert.NilError(t, err)
	assert.Equal(t, uint64(70), bal)
	bal, err = k.Balance(ctx, acct2, denom)
	assert.NilError(t, err)
	assert.Equal(t, uint64(30), bal)

	// burn coins
	err = k.Burn(ctx, acct2, denom, 3)
	assert.NilError(t, err)
	bal, err = k.Balance(ctx, acct2, denom)
	assert.NilError(t, err)
	assert.Equal(t, uint64(27), bal)
	supply, err = k.Supply(ctx, denom)
	assert.NilError(t, err)
	assert.Equal(t, uint64(97), supply)
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

type testStoreService struct {
	db corestore.KVStoreWithBatch
}

func (t testStoreService) OpenKVStore(context.Context) corestore.KVStore {
	return testkv.TestStore{Db: t.db}
}

func (t testStoreService) OpenMemoryStore(context.Context) corestore.KVStore {
	return testkv.TestStore{Db: t.db}
}

func TestGetBackendResolver(t *testing.T) {
	_, err := ormdb.NewModuleDB(&ormv1alpha1.ModuleSchemaDescriptor{
		SchemaFile: []*ormv1alpha1.ModuleSchemaDescriptor_FileEntry{
			{
				Id:            1,
				ProtoFileName: testpb.File_testpb_bank_proto.Path(),
				StorageType:   ormv1alpha1.StorageType_STORAGE_TYPE_MEMORY,
			},
		},
	}, ormdb.ModuleDBOptions{})
	assert.ErrorContains(t, err, "missing MemoryStoreService")

	_, err = ormdb.NewModuleDB(&ormv1alpha1.ModuleSchemaDescriptor{
		SchemaFile: []*ormv1alpha1.ModuleSchemaDescriptor_FileEntry{
			{
				Id:            1,
				ProtoFileName: testpb.File_testpb_bank_proto.Path(),
				StorageType:   ormv1alpha1.StorageType_STORAGE_TYPE_MEMORY,
			},
		},
	}, ormdb.ModuleDBOptions{
		MemoryStoreService: testStoreService{db: coretesting.NewMemDB()},
	})
	assert.NilError(t, err)
}

func ProvideTestRuntime() corestore.KVStoreService {
	return testStoreService{db: coretesting.NewMemDB()}
}

func TestAppConfigModule(t *testing.T) {
	appCfg := appconfig.Compose(&appv1alpha1.Config{
		Modules: []*appv1alpha1.ModuleConfig{
			{Name: "bank", Config: appconfig.WrapAny(&testpb.Module{})},
			{Name: "orm", Config: appconfig.WrapAny(&ormmodulev1alpha1.Module{})},
		},
	})
	var k Keeper
	err := depinject.Inject(depinject.Configs(
		appCfg, depinject.Provide(ProvideTestRuntime),
	), &k)
	assert.NilError(t, err)

	runSimpleBankTests(t, k, context.Background())
}
