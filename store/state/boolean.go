package state

// Boolean is a bool typed wrapper for Value.
// Except for the type checking, it does not alter the behaviour.
type Boolean struct {
	Value
}

func NewBoolean(v Value) Boolean {
	return Boolean{v}
}

func (v Boolean) Get(ctx Context) (res bool) {
	v.Value.Get(ctx, &res)
	return
}

func (v Boolean) GetSafe(ctx Context) (res bool, err error) {
	err = v.Value.GetSafe(ctx, &res)
	return
}

func (v Boolean) Set(ctx Context, value bool) {
	v.Value.Set(ctx, value)
}
