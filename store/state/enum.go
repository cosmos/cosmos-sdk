package state

// Enum is a byte typed wrapper for Value.
// Except for the type checking, it does not alter the behaviour.
type Enum struct {
	Value
}

func NewEnum(v Value) Enum {
	return Enum{v}
}

func (v Enum) Get(ctx Context) (res byte) {
	v.Value.Get(ctx, &res)
	return
}

func (v Enum) GetSafe(ctx Context) (res byte, err error) {
	err = v.Value.GetSafe(ctx, &res)
	return
}

func (v Enum) Set(ctx Context, value byte) {
	v.Value.Set(ctx, value)
}

// Incr() increments the stored value, and returns the updated value.
func (v Enum) Incr(ctx Context) (res byte) {
	res = v.Get(ctx) + 1
	v.Set(ctx, res)
	return
}

// Transit() checks whether the stored value matching with the "from" argument.
// If it matches, it stores the "to" argument to the state and returns true.
func (v Enum) Transit(ctx Context, from, to byte) bool {
	if v.Get(ctx) != from {
		return false
	}
	v.Set(ctx, to)
	return true
}
