package state

type Mapping struct {
	base       Base
	start, end []byte
}

func NewMapping(base Base, prefix []byte) Mapping {
	return Mapping{
		base:  base.Prefix(prefix),
		start: []byte{}, // preventing nil key access in store.Last
	}
}

func (m Mapping) store(ctx Context) KVStore {
	return m.base.Store(ctx)
}

func (m Mapping) Value(key []byte) Value {
	return NewValue(m.base, key)
}

func (m Mapping) Get(ctx Context, key []byte, ptr interface{}) {
	m.Value(key).Get(ctx, ptr)
}

func (m Mapping) GetSafe(ctx Context, key []byte, ptr interface{}) error {
	return m.Value(key).GetSafe(ctx, ptr)
}

func (m Mapping) Set(ctx Context, key []byte, o interface{}) {
	if o == nil {
		m.Delete(ctx, key)
		return
	}
	m.Value(key).Set(ctx, o)
}

func (m Mapping) Has(ctx Context, key []byte) bool {
	return m.Value(key).Exists(ctx)
}

func (m Mapping) Delete(ctx Context, key []byte) {
	m.Value(key).Delete(ctx)
}

func (m Mapping) IsEmpty(ctx Context) bool {
	iter := m.store(ctx).Iterator(nil, nil)
	defer iter.Close()
	return iter.Valid()
}

func (m Mapping) Prefix(prefix []byte) Mapping {
	return NewMapping(m.base, prefix)
}
