package state

// Mapping is key []byte -> value []byte mapping using a base(possibly prefixed).
// All store accessing operations are redirected to the Value corresponding to the key argument
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

// Get() unmarshales and sets the stored value to the pointer if it exists.
// It will panic if the value exists but not unmarshalable.
func (m Mapping) Get(ctx Context, key []byte, ptr interface{}) {
	m.Value(key).Get(ctx, ptr)
}

// GetSafe() unmarshales and sets the stored value to the pointer.
// It will return an error if the value does not exists or unmarshalable.
func (m Mapping) GetSafe(ctx Context, key []byte, ptr interface{}) error {
	return m.Value(key).GetSafe(ctx, ptr)
}

// Set() marshales and sets the argument to the state.
func (m Mapping) Set(ctx Context, key []byte, o interface{}) {
	if o == nil {
		m.Delete(ctx, key)
		return
	}
	m.Value(key).Set(ctx, o)
}

// Has() returns true if the stored value is not nil
func (m Mapping) Has(ctx Context, key []byte) bool {
	return m.Value(key).Exists(ctx)
}

// Delete() deletes the stored value.
func (m Mapping) Delete(ctx Context, key []byte) {
	m.Value(key).Delete(ctx)
}

// Prefix() returns a new mapping with the updated prefix.
func (m Mapping) Prefix(prefix []byte) Mapping {
	return NewMapping(m.base, prefix)
}
