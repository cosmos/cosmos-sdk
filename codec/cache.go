package codec

import (
	"container/list"
	"reflect"
	//"sync"
)

type kvpair struct {
	key   string
	value reflect.Value
}

type lru struct {
	m map[string]*list.Element
	l *list.List

	size int

	//	mtx *sync.Mutex
}

func newLRU(size int) *lru {
	return &lru{
		m: make(map[string]*list.Element),
		l: list.New(),

		size: size,

		//	mtx: &sync.Mutex{},
	}
}

type cache struct {
	cdc Codec

	json *lru
	bin  *lru
}

func newCache(cdc Codec, size int) *cache {
	if size < 0 {
		panic("negative cache size")
	}

	return &cache{
		cdc: cdc,

		json: newLRU(size),
		bin:  newLRU(size),
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
		lru.l.MoveToFront(e)

		rv := deref(reflect.ValueOf(ptr))
		rv.Set(e.Value.(*kvpair).value)
	}
	return
}

func (lru *lru) write(bz []byte, ptr interface{}) {
	if len(lru.m) >= lru.size {
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
	//lru.mtx.Lock()
}

func (lru *lru) unlock() {
	//lru.mtx.Unlock()
}

func (c *cache) MarshalJSON(o interface{}) ([]byte, error) {
	return c.cdc.MarshalJSON(o)
}

func (c *cache) UnmarshalJSON(bz []byte, ptr interface{}) (err error) {
	lru := c.json
	lru.lock()
	defer lru.unlock()
	if lru.read(bz, ptr) {
		return
	}
	err = c.cdc.UnmarshalJSON(bz, ptr)
	if err != nil {
		return
	}
	lru.write(bz, ptr)
	return
}

func (c *cache) MarshalBinary(o interface{}) ([]byte, error) {
	return c.cdc.MarshalBinary(o)
}

func (c *cache) UnmarshalBinary(bz []byte, ptr interface{}) (err error) {
	lru := c.bin
	lru.lock()
	defer lru.unlock()
	if lru.read(bz, ptr) {
		return
	}
	err = c.cdc.UnmarshalBinary(bz, ptr)
	if err != nil {
		return
	}
	lru.write(bz, ptr)
	return
}

func (c *cache) MustMarshalBinary(o interface{}) []byte {
	return c.cdc.MustMarshalBinary(o)
}

func (c *cache) MustUnmarshalBinary(bz []byte, ptr interface{}) {
	lru := c.bin
	lru.lock()
	defer lru.unlock()
	if lru.read(bz, ptr) {
		return
	}
	c.cdc.MustUnmarshalBinary(bz, ptr)
	lru.write(bz, ptr)
}
