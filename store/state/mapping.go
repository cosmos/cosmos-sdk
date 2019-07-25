package state

// Mapping is key []byte -> value []byte mapping using a base(possibly prefixed)
type Mapping struct {
	base Base
}

// NewMapping() constructs a Mapping with a provided prefix
func NewMapping(base Base, prefix []byte) Mapping {
	return Mapping{
		base: base.Prefix(prefix),
	}
}

// Value() returns the Value corresponding to the provided key
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

func (m Mapping) Prefix(prefix []byte) Mapping {
	return NewMapping(m.base, prefix)
}
