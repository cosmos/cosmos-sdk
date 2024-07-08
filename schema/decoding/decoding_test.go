package decoding

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"cosmossdk.io/schema"
	"cosmossdk.io/schema/appdata"
)

func TestMiddleware(t *testing.T) {
	appMods := map[string]interface{}{}
	resolver := ModuleSetDecoderResolver(appMods)
	tl := newTestBankListener(t)
	listener, err := Middleware(tl.Listener, resolver, MiddlewareOptions{})
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	store := newTestStore(t, "bank", listener)
	bankMod := exampleBankModule{
		store: store,
	}
	appMods["bank"] = bankMod

	bankMod.Mint("bob", "foo", 100)
	err = bankMod.Send("bob", "alice", "foo", 50)
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	expected := []schema.ObjectUpdate{
		{
			TypeName: "supply",
			Key:      []interface{}{"foo"},
			Value:    uint64(100),
		},
		{
			TypeName: "balances",
			Key:      []interface{}{"bob", "foo"},
			Value:    uint64(100),
		},
		{
			TypeName: "balances",
			Key:      []interface{}{"bob", "foo"},
			Value:    uint64(50),
		},
		{
			TypeName: "balances",
			Key:      []interface{}{"alice", "foo"},
			Value:    uint64(50),
		},
	}

	if !reflect.DeepEqual(tl.updates, expected) {
		t.Fatalf("expected %v, got %v", expected, tl.updates)
	}
}

func TestSync(t *testing.T) {
	tl := newTestBankListener(t)
	store := newTestStore(t, "bank", tl.Listener)
	bankMod := exampleBankModule{
		store: store,
	}

	bankMod.Mint("bob", "foo", 100)
	err := bankMod.Send("bob", "alice", "foo", 50)
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	expected := []schema.ObjectUpdate{
		{
			TypeName: "balances",
			Key:      []interface{}{"alice", "foo"},
			Value:    uint64(50),
		},
		{
			TypeName: "balances",
			Key:      []interface{}{"bob", "foo"},
			Value:    uint64(50),
		},
		{
			TypeName: "supply",
			Key:      []interface{}{"foo"},
			Value:    uint64(100),
		},
	}

	appMods := map[string]interface{}{}
	appMods["bank"] = bankMod
	resolver := ModuleSetDecoderResolver(appMods)
	err = Sync(tl.Listener, store, resolver, SyncOptions{})
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	if !reflect.DeepEqual(tl.updates, expected) {
		t.Fatalf("expected %v, got %v", expected, tl.updates)
	}
}

type testListener struct {
	appdata.Listener
	updates []schema.ObjectUpdate
}

func newTestBankListener(t *testing.T) *testListener {
	res := &testListener{}
	res.Listener = appdata.Listener{
		InitializeModuleData: func(data appdata.ModuleInitializationData) error {
			if data.ModuleName != "bank" {
				t.Errorf("expected bank module, got %s", data.ModuleName)
			}
			if !reflect.DeepEqual(data.Schema, exampleBankSchema) {
				t.Errorf("expected %v, got %v", exampleBankSchema, data.Schema)
			}
			return nil
		},
		OnObjectUpdate: func(data appdata.ObjectUpdateData) error {
			if data.ModuleName != "bank" {
				t.Errorf("expected bank module, got %s", data.ModuleName)
			}
			res.updates = append(res.updates, data.Updates...)
			return nil
		},
	}
	return res
}

type testStore struct {
	t        *testing.T
	modName  string
	store    map[string][]byte
	listener appdata.Listener
}

func newTestStore(t *testing.T, modName string, listener appdata.Listener) *testStore {
	return &testStore{
		t:        t,
		modName:  modName,
		listener: listener,
		store:    map[string][]byte{},
	}
}

func (t testStore) Get(key []byte) []byte {
	return t.store[string(key)]
}

func (t testStore) GetUInt64(key []byte) uint64 {
	bz := t.store[string(key)]
	if len(bz) == 0 {
		return 0
	}
	x, err := strconv.ParseUint(string(bz), 10, 64)
	if err != nil {
		t.t.Fatalf("unexpected error: %v", err)
	}
	return x
}

func (t testStore) Set(key, value []byte) {
	if t.listener.OnKVPair != nil {
		err := t.listener.OnKVPair(appdata.KVPairData{Updates: []appdata.ModuleKVPairUpdate{
			{
				ModuleName: t.modName,
				Update: schema.KVPairUpdate{
					Key:   key,
					Value: value,
				},
			},
		}})
		if err != nil {
			t.t.Fatalf("unexpected error: %v", err)
		}
	}
	t.store[string(key)] = value
}

func (t testStore) SetUInt64(key []byte, value uint64) {
	t.Set(key, []byte(strconv.FormatUint(value, 10)))
}

var _ SyncSource = &testStore{}

func (t testStore) IterateAllKVPairs(moduleName string, fn func(key []byte, value []byte) error) error {
	if t.modName != moduleName {
		return fmt.Errorf("don't have state for module %s", moduleName)
	}

	var keys []string
	for key := range t.store {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		err := fn([]byte(key), t.store[key])
		if err != nil {
			return err
		}
	}
	return nil
}

type exampleBankModule struct {
	store *testStore
}

func (e exampleBankModule) Mint(acct, denom string, amount uint64) {
	key := supplyKey(denom)
	e.store.SetUInt64(key, e.store.GetUInt64(key)+amount)
	e.addBalance(acct, denom, amount)
}

func (e exampleBankModule) Send(from, to, denom string, amount uint64) error {
	err := e.subBalance(from, denom, amount)
	if err != nil {
		return nil
	}
	e.addBalance(to, denom, amount)
	return nil
}

func (e exampleBankModule) GetBalance(acct, denom string) uint64 {
	return e.store.GetUInt64(balanceKey(acct, denom))
}

func (e exampleBankModule) GetSupply(denom string) uint64 {
	return e.store.GetUInt64(supplyKey(denom))
}

func balanceKey(acct, denom string) []byte {
	return []byte(fmt.Sprintf("balance/%s/%s", acct, denom))
}

func supplyKey(denom string) []byte {
	return []byte(fmt.Sprintf("supply/%s", denom))
}

func (e exampleBankModule) addBalance(acct, denom string, amount uint64) {
	key := balanceKey(acct, denom)
	e.store.SetUInt64(key, e.store.GetUInt64(key)+amount)
}

func (e exampleBankModule) subBalance(acct, denom string, amount uint64) error {
	key := balanceKey(acct, denom)
	cur := e.store.GetUInt64(key)
	if cur < amount {
		return fmt.Errorf("insufficient balance")
	}
	e.store.SetUInt64(key, cur-amount)
	return nil
}

var exampleBankSchema = schema.ModuleSchema{
	ObjectTypes: []schema.ObjectType{
		{
			Name: "balances",
			KeyFields: []schema.Field{
				{
					Name: "account",
					Kind: schema.StringKind,
				},
				{
					Name: "denom",
					Kind: schema.StringKind,
				},
			},
			ValueFields: []schema.Field{
				{
					Name: "amount",
					Kind: schema.Uint64Kind,
				},
			},
		},
	},
}

func (e exampleBankModule) ModuleCodec() (schema.ModuleCodec, error) {
	return schema.ModuleCodec{
		Schema: exampleBankSchema,
		KVDecoder: func(update schema.KVPairUpdate) ([]schema.ObjectUpdate, error) {
			key := string(update.Key)
			value, err := strconv.ParseUint(string(update.Value), 10, 64)
			if err != nil {
				return nil, err
			}
			if strings.HasPrefix(key, "balance/") {
				parts := strings.Split(key, "/")
				return []schema.ObjectUpdate{{
					TypeName: "balances",
					Key:      []interface{}{parts[1], parts[2]},
					Value:    value,
				}}, nil
			} else if strings.HasPrefix(key, "supply/") {
				parts := strings.Split(key, "/")
				return []schema.ObjectUpdate{{
					TypeName: "supply",
					Key:      []interface{}{parts[1]},
					Value:    value,
				}}, nil
			} else {
				return nil, fmt.Errorf("unexpected key: %s", key)
			}
		},
	}, nil
}

var _ schema.HasModuleCodec = exampleBankModule{}
