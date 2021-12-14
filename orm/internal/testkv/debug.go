package testkv

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
)

// Debugger is an interface that handles debug info from the debug reader wrapper.
type Debugger interface {

	// Log logs a single log message.
	Log(string)

	// Decode decodes a key-value entry into a debug string.
	Decode(storeName string, key, value []byte) string
}

type debugReader struct {
	reader    kvstore.Reader
	debugger  Debugger
	storeName string
}

type debugWriter struct {
	debugReader
	writer kvstore.Writer
}

func (t debugReader) Get(key []byte) ([]byte, error) {
	val, err := t.reader.Get(key)
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

func (t debugReader) Has(key []byte) (bool, error) {
	has, err := t.reader.Has(key)
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

func (t debugReader) Iterator(start, end []byte) (kvstore.Iterator, error) {
	if t.debugger != nil {
		t.debugger.Log(fmt.Sprintf("ITERATOR %x -> %x", start, end))
	}
	it, err := t.reader.Iterator(start, end)
	if err != nil {
		return nil, err
	}
	return &debugIterator{
		iterator:  it,
		storeName: t.storeName,
		debugger:  t.debugger,
	}, nil
}

func (t debugReader) ReverseIterator(start, end []byte) (kvstore.Iterator, error) {
	if t.debugger != nil {
		t.debugger.Log(fmt.Sprintf("ITERATOR %x <- %x", start, end))
	}
	it, err := t.reader.ReverseIterator(start, end)
	if err != nil {
		return nil, err
	}
	return &debugIterator{
		iterator:  it,
		storeName: t.storeName,
		debugger:  t.debugger,
	}, nil
}

func (t debugWriter) Set(key, value []byte) error {
	if t.debugger != nil {
		t.debugger.Log(fmt.Sprintf("SET %x %x", key, value))
		t.debugger.Log(fmt.Sprintf("    %s", t.debugger.Decode(t.storeName, key, value)))
	}
	err := t.writer.Set(key, value)
	if err != nil {
		if t.debugger != nil {
			t.debugger.Log(fmt.Sprintf("ERR on SET %s: %v", t.debugger.Decode(t.storeName, key, value), err))
		}
		return err
	}
	return nil
}

func (t debugWriter) Delete(key []byte) error {
	if t.debugger != nil {
		t.debugger.Log(fmt.Sprintf("DEL %x", key))
		t.debugger.Log(fmt.Sprintf("DEL %s", t.debugger.Decode(t.storeName, key, nil)))
	}
	err := t.writer.Delete(key)
	if err != nil {
		if t.debugger != nil {
			t.debugger.Log(fmt.Sprintf("ERR on SET %s: %v", t.debugger.Decode(t.storeName, key, nil), err))
		}
		return err
	}
	return nil
}

var _ kvstore.Reader = &debugReader{}
var _ kvstore.Writer = &debugWriter{}

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

func (d *EntryCodecDebugger) Decode(storeName string, key, value []byte) string {
	entry, err := d.EntryCodec.DecodeEntry(key, value)
	if err != nil {
		return fmt.Sprintf("ERR:%v", err)
	}

	return entry.String()
}
