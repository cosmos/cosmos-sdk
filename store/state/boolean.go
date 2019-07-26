package state

// Boolean is a bool typed wrapper for Value.
// Except for the type checking, it does not alter the behaviour.
type Boolean struct {
	Value
}

// NewBoolean() wraps the argument Value as Boolean
func NewBoolean(v Value) Boolean {
	return Boolean{v}
}

// Get() unmarshales and returns the stored boolean value if it exists.
// It will panic if the value exists but is not boolean type.
func (v Boolean) Get(ctx Context) (res bool) {
	v.Value.Get(ctx, &res)
	return
}

// GetSafe() unmarshales and returns the stored boolean value.
// It will return an error if the value does not exist or not boolean.
func (v Boolean) GetSafe(ctx Context) (res bool, err error) {
	err = v.Value.GetSafe(ctx, &res)
	return
}

// Set() marshales and sets the boolean argument to the state.
func (v Boolean) Set(ctx Context, value bool) {
	v.Value.Set(ctx, value)
}
