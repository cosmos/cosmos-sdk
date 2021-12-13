package testkv

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
)

type Debugger interface {
	Log(string)
	Decode(storeName string, key, value []byte) string
}

type debugStore struct {
	store     kvstore.Store
	debugger  Debugger
	storeName string
}

func NewDebugStore(store kvstore.Store, debugger Debugger, storeName string) kvstore.Store {
	return &debugStore{store: store, debugger: debugger, storeName: storeName}
}

func (t debugStore) Get(key []byte) ([]byte, error) {
	val, err := t.store.Get(key)
	if err != nil {
		if t.debugger != nil {
			t.debugger.Log(fmt.Sprintf("ERR on GET %s: %v", t.debugger.Decode(t.storeName, key, nil), err))
		}
		return nil, err
	}
	if t.debugger != nil {
		t.debugger.Log(fmt.Sprintf("GET %x %x", key, val))
		t.debugger.Log(fmt.Sprintf("    %s", t.debugger.Decode(t.storeName, key, val)))
	}
	return val, nil
}

func (t debugStore) Has(key []byte) (bool, error) {
	has, err := t.store.Has(key)
	if err != nil {
		if t.debugger != nil {
			t.debugger.Log(fmt.Sprintf("ERR on HAS %s: %v", t.debugger.Decode(t.storeName, key, nil), err))
		}
		return has, err
	}
	if t.debugger != nil {
		t.debugger.Log(fmt.Sprintf("HAS %x", key))
		t.debugger.Log(fmt.Sprintf("    %s", t.debugger.Decode(t.storeName, key, nil)))
	}
	return has, nil
}

func (t debugStore) Iterator(start, end []byte) (kvstore.Iterator, error) {
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

func (t debugStore) ReverseIterator(start, end []byte) (kvstore.Iterator, error) {
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
		t.debugger.Log(fmt.Sprintf("    %s", t.debugger.Decode(t.storeName, key, value)))
	}
	err := t.store.Set(key, value)
	if err != nil {
		if t.debugger != nil {
			t.debugger.Log(fmt.Sprintf("ERR on SET %s: %v", t.debugger.Decode(t.storeName, key, value), err))
		}
		return err
	}
	return nil
}

func (t debugStore) Delete(key []byte) error {
	if t.debugger != nil {
		t.debugger.Log(fmt.Sprintf("DEL %x", key))
		t.debugger.Log(fmt.Sprintf("DEL %s", t.debugger.Decode(t.storeName, key, nil)))
	}
	err := t.store.Delete(key)
	if err != nil {
		if t.debugger != nil {
			t.debugger.Log(fmt.Sprintf("ERR on SET %s: %v", t.debugger.Decode(t.storeName, key, nil), err))
		}
		return err
	}
	return nil
}

var _ kvstore.Store = &debugStore{}

type debugIterator struct {
	iterator  kvstore.Iterator
	storeName string
	debugger  Debugger
}

func (d debugIterator) Domain() (start []byte, end []byte) {
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
	d.debugger.Log(fmt.Sprintf("      %s", d.debugger.Decode(d.storeName, key, value)))
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

var _ kvstore.Iterator = &debugIterator{}

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

func (d *EntryCodecDebugger) Decode(storeName string, key, value []byte) string {
	entry, err := d.EntryCodec.DecodeEntry(key, value)
	if err != nil {
		return fmt.Sprintf("ERR:%v", err)
	}

	return entry.String()
}
