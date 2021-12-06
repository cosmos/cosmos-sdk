package memkv

import (
	"fmt"

	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
)

type Debugger interface {
	Log(string)
	Decode(storeName string, key, value []byte) string
}

type TestStore struct {
	store    kvstore.Store
	debugger Debugger
	name     string
}

func NewTestStore(store kvstore.Store, debugger Debugger, name string) kvstore.Store {
	return &TestStore{store: store, debugger: debugger, name: name}
}

func (t TestStore) Get(key []byte) ([]byte, error) {
	val, err := t.store.Get(key)
	if err != nil {
		if t.debugger != nil {
			t.debugger.Log(fmt.Sprintf("ERR on GET %s: %v", t.debugger.Decode(t.name, key, nil), err))
		}
		return nil, err
	}
	if t.debugger != nil {
		t.debugger.Log(fmt.Sprintf("GET %s", t.debugger.Decode(t.name, key, val)))
	}
	return val, nil
}

func (t TestStore) Has(key []byte) (bool, error) {
	has, err := t.store.Has(key)
	if err != nil {
		if t.debugger != nil {
			t.debugger.Log(fmt.Sprintf("ERR on HAS %s: %v", t.debugger.Decode(t.name, key, nil), err))
		}
		return has, err
	}
	if t.debugger != nil {
		t.debugger.Log(fmt.Sprintf("HAS %s", t.debugger.Decode(t.name, key, nil)))
	}
	return has, nil
}

func (t TestStore) Iterator(start, end []byte) (kvstore.Iterator, error) {
	if t.debugger != nil {
		t.debugger.Log(fmt.Sprintf("ITERATOR %s -> %s",
			t.debugger.Decode(t.name, start, nil),
			t.debugger.Decode(t.name, end, nil),
		))
	}
	return t.store.Iterator(start, end)
}

func (t TestStore) ReverseIterator(start, end []byte) (kvstore.Iterator, error) {
	if t.debugger != nil {
		t.debugger.Log(fmt.Sprintf("ITERATOR %s <- %s",
			t.debugger.Decode(t.name, start, nil),
			t.debugger.Decode(t.name, end, nil),
		))
	}
	return t.store.ReverseIterator(start, end)
}

func (t TestStore) Set(key, value []byte) error {
	if t.debugger != nil {
		t.debugger.Log(fmt.Sprintf("SET %s", t.debugger.Decode(t.name, key, value)))
	}
	err := t.store.Set(key, value)
	if err != nil {
		if t.debugger != nil {
			t.debugger.Log(fmt.Sprintf("ERR on SET %s: %v", t.debugger.Decode(t.name, key, value), err))
		}
		return err
	}
	return nil
}

func (t TestStore) Delete(key []byte) error {
	if t.debugger != nil {
		t.debugger.Log(fmt.Sprintf("DEL %s", t.debugger.Decode(t.name, key, nil)))
	}
	err := t.store.Delete(key)
	if err != nil {
		if t.debugger != nil {
			t.debugger.Log(fmt.Sprintf("ERR on SET %s: %v", t.debugger.Decode(t.name, key, nil), err))
		}
		return err
	}
	return nil
}

var _ kvstore.Store = &TestStore{}

type IndexCommitmentStore struct {
	commitment kvstore.Store
	index      kvstore.Store
}

func NewMemIndexCommitmentStore() *IndexCommitmentStore {
	return &IndexCommitmentStore{
		commitment: dbm.NewMemDB(),
		index:      dbm.NewMemDB(),
	}
}

func NewDebugIndexCommitmentStore(debugger Debugger) *IndexCommitmentStore {
	return &IndexCommitmentStore{
		commitment: NewTestStore(dbm.NewMemDB(), debugger, "commitment"),
		index:      NewTestStore(dbm.NewMemDB(), debugger, "index"),
	}
}

var _ kvstore.IndexCommitmentStore = &IndexCommitmentStore{}

func (i IndexCommitmentStore) ReadCommitmentStore() kvstore.ReadStore {
	return i.commitment
}

func (i IndexCommitmentStore) ReadIndexStore() kvstore.ReadStore {
	return i.index
}

func (i IndexCommitmentStore) CommitmentStore() kvstore.Store {
	return i.commitment
}

func (i IndexCommitmentStore) IndexStore() kvstore.Store {
	return i.index
}

type EntryCodecDebugger struct {
	EntryCodec ormkv.EntryCodec
	Print      func(string)
}

func (d EntryCodecDebugger) Log(s string) {
	if d.Print != nil {
		d.Print(s)
	} else {
		fmt.Println(s)
	}
}

func (d EntryCodecDebugger) Decode(storeName string, key, value []byte) string {
	entry, err := d.EntryCodec.DecodeEntry(key, value)
	if err != nil {
		return fmt.Sprintf("ERR:%v", err)
	}
	return entry.String()
}
