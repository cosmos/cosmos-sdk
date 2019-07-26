package state

// String is a string types wrapper for Value.
// Except for the type checking, it does not alter the behaviour.
type String struct {
	Value
}

// NewString() wraps the argument value as String
func NewString(v Value) String {
	return String{v}
}

// Get() unmarshales and returns the stored string value if it exists.
// It will panic if the value exists but is not strin type.
func (v String) Get(ctx Context) (res string) {
	v.Value.Get(ctx, &res)
	return
}

// GetSafe() unmarshales and returns the stored string value.
// It will return an error if the value does not exist or not string
func (v String) GetSafe(ctx Context) (res string, err error) {
	err = v.Value.GetSafe(ctx, &res)
	return
}

// Set() marshales and sets the string argument to the state.
func (v String) Set(ctx Context, value string) {
	v.Value.Set(ctx, value)
}
