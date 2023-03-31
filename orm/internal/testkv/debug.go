package testkv

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/internal/stablejson"
	"github.com/cosmos/cosmos-sdk/orm/model/ormtable"
	"github.com/cosmos/cosmos-sdk/orm/types/kv"
)

// Debugger is an interface that handles debug info from the debug store wrapper.
type Debugger interface {
	// Log logs a single log message.
	Log(string)

	// Decode decodes a key-value entry into a debug string.
	Decode(key, value []byte) string
}

// NewDebugBackend wraps both stores from a Backend with a debugger.
func NewDebugBackend(backend ormtable.Backend, debugger Debugger) ormtable.Backend {
	hooks := debugHooks{
		debugger:      debugger,
		validateHooks: backend.ValidateHooks(),
		writeHooks:    backend.WriteHooks(),
	}
	return ormtable.NewBackend(ormtable.BackendOptions{
		CommitmentStore: NewDebugStore(backend.CommitmentStore(), debugger, "commit"),
		IndexStore:      NewDebugStore(backend.IndexStore(), debugger, "index"),
		ValidateHooks:   hooks,
		WriteHooks:      hooks,
	})
}

type debugStore struct {
	store     kv.Store
	debugger  Debugger
	storeName string
}

// NewDebugStore wraps the store with the debugger instance returning a debug store wrapper.
func NewDebugStore(store kv.Store, debugger Debugger, storeName string) kv.Store {
	return &debugStore{store: store, debugger: debugger, storeName: storeName}
}

func (t debugStore) Get(key []byte) ([]byte, error) {
	val, err := t.store.Get(key)
	if err != nil {
		if t.debugger != nil {
			t.debugger.Log(fmt.Sprintf("ERR on GET %s: %v", t.debugger.Decode(key, nil), err))
		}
		return nil, err
	}
	if t.debugger != nil {
		t.debugger.Log(fmt.Sprintf("GET %x %x", key, val))
		t.debugger.Log(fmt.Sprintf("    %s", t.debugger.Decode(key, val)))
	}
	return val, nil
}

func (t debugStore) Has(key []byte) (bool, error) {
	has, err := t.store.Has(key)
	if err != nil {
		if t.debugger != nil {
			t.debugger.Log(fmt.Sprintf("ERR on HAS %s: %v", t.debugger.Decode(key, nil), err))
		}
		return has, err
	}
	if t.debugger != nil {
		t.debugger.Log(fmt.Sprintf("HAS %x", key))
		t.debugger.Log(fmt.Sprintf("    %s", t.debugger.Decode(key, nil)))
	}
	return has, nil
}

func (t debugStore) Iterator(start, end []byte) (kv.Iterator, error) {
	if t.debugger != nil {
		t.debugger.Log(fmt.Sprintf("ITERATOR %x -> %x", start, end))
	}
	it, err := t.store.Iterator(start, end)
	if err != nil {
		return nil, err
	}
	return &debugIterator{
		iterator:  it,
		storeName: t.storeName,
		debugger:  t.debugger,
	}, nil
}

func (t debugStore) ReverseIterator(start, end []byte) (kv.Iterator, error) {
	if t.debugger != nil {
		t.debugger.Log(fmt.Sprintf("ITERATOR %x <- %x", start, end))
	}
	it, err := t.store.ReverseIterator(start, end)
	if err != nil {
		return nil, err
	}
	return &debugIterator{
		iterator:  it,
		storeName: t.storeName,
		debugger:  t.debugger,
	}, nil
}

func (t debugStore) Set(key, value []byte) error {
	if t.debugger != nil {
		t.debugger.Log(fmt.Sprintf("SET %x %x", key, value))
		t.debugger.Log(fmt.Sprintf("    %s", t.debugger.Decode(key, value)))
	}
	err := t.store.Set(key, value)
	if err != nil {
		if t.debugger != nil {
			t.debugger.Log(fmt.Sprintf("ERR on SET %s: %v", t.debugger.Decode(key, value), err))
		}
		return err
	}
	return nil
}

func (t debugStore) Delete(key []byte) error {
	if t.debugger != nil {
		t.debugger.Log(fmt.Sprintf("DEL %x", key))
		t.debugger.Log(fmt.Sprintf("DEL %s", t.debugger.Decode(key, nil)))
	}
	err := t.store.Delete(key)
	if err != nil {
		if t.debugger != nil {
			t.debugger.Log(fmt.Sprintf("ERR on SET %s: %v", t.debugger.Decode(key, nil), err))
		}
		return err
	}
	return nil
}

var _ kv.Store = &debugStore{}

type debugIterator struct {
	iterator  kv.Iterator
	storeName string
	debugger  Debugger
}

func (d debugIterator) Domain() (start, end []byte) {
	start, end = d.iterator.Domain()
	d.debugger.Log(fmt.Sprintf("  DOMAIN %x -> %x", start, end))
	return start, end
}

func (d debugIterator) Valid() bool {
	valid := d.iterator.Valid()
	d.debugger.Log(fmt.Sprintf("  VALID %t", valid))
	return valid
}

func (d debugIterator) Next() {
	d.debugger.Log("  NEXT")
	d.iterator.Next()
}

func (d debugIterator) Key() (key []byte) {
	key = d.iterator.Key()
	value := d.iterator.Value()
	d.debugger.Log(fmt.Sprintf("  KEY %x %x", key, value))
	d.debugger.Log(fmt.Sprintf("      %s", d.debugger.Decode(key, value)))
	return key
}

func (d debugIterator) Value() (value []byte) {
	return d.iterator.Value()
}

func (d debugIterator) Error() error {
	err := d.iterator.Error()
	d.debugger.Log(fmt.Sprintf("  ERR %+v", err))
	return err
}

func (d debugIterator) Close() error {
	d.debugger.Log("  CLOSE")
	return d.iterator.Close()
}

var _ kv.Iterator = &debugIterator{}

// EntryCodecDebugger is a Debugger instance that uses an EntryCodec and Print
// function for debugging.
type EntryCodecDebugger struct {
	EntryCodec ormkv.EntryCodec
	Print      func(string)
}

func (d *EntryCodecDebugger) Log(s string) {
	if d.Print != nil {
		d.Print(s)
	} else {
		fmt.Println(s)
	}
}

func (d *EntryCodecDebugger) Decode(key, value []byte) string {
	entry, err := d.EntryCodec.DecodeEntry(key, value)
	if err != nil {
		return fmt.Sprintf("ERR:%v", err)
	}

	return entry.String()
}

type debugHooks struct {
	debugger      Debugger
	validateHooks ormtable.ValidateHooks
	writeHooks    ormtable.WriteHooks
}

func (d debugHooks) ValidateInsert(context context.Context, message proto.Message) error {
	jsonBz, err := stablejson.Marshal(message)
	if err != nil {
		return err
	}

	d.debugger.Log(fmt.Sprintf(
		"ORM BEFORE INSERT %s %s",
		message.ProtoReflect().Descriptor().FullName(),
		jsonBz,
	))
	if d.validateHooks != nil {
		return d.validateHooks.ValidateInsert(context, message)
	}
	return nil
}

func (d debugHooks) ValidateUpdate(ctx context.Context, existing, new proto.Message) error {
	existingJson, err := stablejson.Marshal(existing)
	if err != nil {
		return err
	}

	newJson, err := stablejson.Marshal(new)
	if err != nil {
		return err
	}

	d.debugger.Log(fmt.Sprintf(
		"ORM BEFORE UPDATE %s %s -> %s",
		existing.ProtoReflect().Descriptor().FullName(),
		existingJson,
		newJson,
	))
	if d.validateHooks != nil {
		return d.validateHooks.ValidateUpdate(ctx, existing, new)
	}
	return nil
}

func (d debugHooks) ValidateDelete(ctx context.Context, message proto.Message) error {
	jsonBz, err := stablejson.Marshal(message)
	if err != nil {
		return err
	}

	d.debugger.Log(fmt.Sprintf(
		"ORM BEFORE DELETE %s %s",
		message.ProtoReflect().Descriptor().FullName(),
		jsonBz,
	))
	if d.validateHooks != nil {
		return d.validateHooks.ValidateDelete(ctx, message)
	}
	return nil
}

func (d debugHooks) OnInsert(ctx context.Context, message proto.Message) {
	jsonBz, err := stablejson.Marshal(message)
	if err != nil {
		panic(err)
	}

	d.debugger.Log(fmt.Sprintf(
		"ORM AFTER INSERT %s %s",
		message.ProtoReflect().Descriptor().FullName(),
		jsonBz,
	))
	if d.writeHooks != nil {
		d.writeHooks.OnInsert(ctx, message)
	}
}

func (d debugHooks) OnUpdate(ctx context.Context, existing, new proto.Message) {
	existingJson, err := stablejson.Marshal(existing)
	if err != nil {
		panic(err)
	}

	newJson, err := stablejson.Marshal(new)
	if err != nil {
		panic(err)
	}

	d.debugger.Log(fmt.Sprintf(
		"ORM AFTER UPDATE %s %s -> %s",
		existing.ProtoReflect().Descriptor().FullName(),
		existingJson,
		newJson,
	))
	if d.writeHooks != nil {
		d.writeHooks.OnUpdate(ctx, existing, new)
	}
}

func (d debugHooks) OnDelete(ctx context.Context, message proto.Message) {
	jsonBz, err := stablejson.Marshal(message)
	if err != nil {
		panic(err)
	}

	d.debugger.Log(fmt.Sprintf(
		"ORM AFTER DELETE %s %s",
		message.ProtoReflect().Descriptor().FullName(),
		jsonBz,
	))
	if d.writeHooks != nil {
		d.writeHooks.OnDelete(ctx, message)
	}
}
