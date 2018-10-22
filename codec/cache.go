package codec

import (
	"container/list"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"sync"
)

/*
*codec.Cache provides LRU-cached amino codec

It is recommended to use cache when there is a few value unserialized
frequently. Adjust the cache size to make cache hit rate at least 60%.
*/

type kvpair struct {
	key   string
	value reflect.Value
}

type lru struct {
	m map[string]*list.Element
	l *list.List

	size int

	resize  bool
	hitrate float64

	mtx *sync.Mutex
}

type Cache struct {
	*Amino

	json *lru
	bin  *lru
}

func newLRU() *lru {
	return &lru{
		m: make(map[string]*list.Element),
		l: list.New(),

		size: 1,

		resize:  true,
		hitrate: 0,

		mtx: &sync.Mutex{},
	}
}

func newCache(cdc *Amino) *Cache {
	return &Cache{
		Amino: cdc,

		json: newLRU(),
		bin:  newLRU(),
	}
}

func deref(rv reflect.Value) reflect.Value {
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			newPtr := reflect.New(rv.Type().Elem())
			rv.Set(newPtr)
		}
		rv = rv.Elem()
	}
	return rv
}

func (lru *lru) read(bz []byte, ptr interface{}) (ok bool) {
	var e *list.Element
	if e, ok = lru.m[string(bz)]; ok {
		if lru.resize {
			lru.hitrate = (lru.hitrate*(window-1) + 1) / window
		}

		lru.l.MoveToFront(e)

		rv := deref(reflect.ValueOf(ptr))
		rv.Set(e.Value.(*kvpair).value)
	}
	return
}

const window = float64(10000)
const inc = 1 / window
const requirement = 0.6

func (lru *lru) write(bz []byte, ptr interface{}) {
	if len(lru.m) >= lru.size {
		if lru.resize {
			fmt.Println(lru.hitrate)
			if lru.hitrate < requirement {
				lru.size *= 2
			} else {
				lru.resize = false
			}
		}
		e := lru.l.Back()
		lru.l.Remove(e)
		delete(lru.m, e.Value.(*kvpair).key)
	}
	strbz := string(bz)
	kvp := &kvpair{
		key:   strbz,
		value: deref(reflect.ValueOf(ptr)),
	}
	e := lru.l.PushFront(kvp)
	lru.m[strbz] = e
}

func (lru *lru) lock() {
	lru.mtx.Lock()
}

func (lru *lru) unlock() {
	lru.mtx.Unlock()
}

func (c *Cache) UnmarshalJSON(bz []byte, ptr interface{}) (err error) {
	lru := c.json
	lru.lock()
	defer lru.unlock()
	if lru.read(bz, ptr) {
		return
	}
	err = c.Amino.UnmarshalJSON(bz, ptr)
	if err != nil {
		return
	}
	lru.write(bz, ptr)
	return
}

func bareBytes(bz []byte) ([]byte, error) {
	// validity checking logic is from go-amino/amino.go/UnmarshalBinary
	if len(bz) == 0 {
		return nil, errors.New("UnmarshalBinary cannot decode empty bytes")
	}

	// Read byte-length prefix
	u64, n := binary.Uvarint(bz)
	if n < 0 {
		return nil, fmt.Errorf("Error reading msg byte-length prefix")
	}
	if u64 > uint64(len(bz)-n) {
		return nil, fmt.Errorf("Not enough bytes to read in UnmarshalBinary, want %v more bytes but only have %v", u64, len(bz)-n)
	} else if u64 > uint64(len(bz)-n) {
		return nil, fmt.Errorf("Bytes left over in UnmarshalBinary, should read %v more bytes but only have %v", u64, len(bz)-n)
	}
	return bz[n:], nil
}

func (c *Cache) UnmarshalBinary(bz []byte, ptr interface{}) (err error) {
	bz, err = bareBytes(bz)
	if err != nil {
		return
	}
	return c.UnmarshalBinaryBare(bz, ptr)
}

func (c *Cache) MustUnmarshalBinary(bz []byte, ptr interface{}) {
	err := c.UnmarshalBinary(bz, ptr)
	if err != nil {
		panic(err)
	}
}

func (c *Cache) UnmarshalBinaryBare(bz []byte, ptr interface{}) (err error) {
	lru := c.bin
	lru.lock()
	defer lru.unlock()
	if lru.read(bz, ptr) {
		return
	}
	err = c.Amino.UnmarshalBinaryBare(bz, ptr)
	if err != nil {
		return
	}
	lru.write(bz, ptr)
	return
}

func (c *Cache) MustUnmarshalBinaryBare(bz []byte, ptr interface{}) {
	err := c.UnmarshalBinaryBare(bz, ptr)
	if err != nil {
		panic(err)
	}
}
