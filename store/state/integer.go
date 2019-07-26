package state

// Integer is a uint64 types wrapper for Value.
// Except for the type checking, it does not alter the behaviour.
type Integer struct {
	Value
}

// NewInteger() wraps the argument value as Integer
func NewInteger(v Value) Integer {
	return Integer{
		Value: v,
	}
}

// Get() unmarshales and returns the stored uint64 value if it exists.
// If will panic if the value exists but is not uint64 type.
func (v Integer) Get(ctx Context) (res uint64) {
	v.Value.Get(ctx, &res)
	return res
}

// GetSafe() unmarshales and returns the stored uint64 value.
// It will return an error if the value does not exist or not uint64.
func (v Integer) GetSafe(ctx Context) (res uint64, err error) {
	err = v.Value.GetSafe(ctx, &res)
	return
}

// Set() marshales and sets the uint64 argument to the state.
func (v Integer) Set(ctx Context, value uint64) {
	v.Value.Set(ctx, value)
}

// Incr() increments the stored value, and returns the updated value.
func (v Integer) Incr(ctx Context) (res uint64) {
	res = v.Get(ctx) + 1
	v.Set(ctx, res)
	return
}
