package state

// Enum is a byte typed wrapper for Value.
// x <-> []byte{x}
type Enum struct {
	Value
}

func (v Value) Enum() Enum {
	return Enum{v}
}

// Get() unmarshales and returns the stored byte value if it exists.
// It will panic if the value exists but is not byte type.
func (v Enum) Get(ctx Context) (res byte) {
	return v.Value.GetRaw(ctx)[0]
}

// GetSafe() unmarshales and returns the stored byte value.
// It will returns an error if the value does not exists or not byte.
func (v Enum) GetSafe(ctx Context) (res byte, err error) {
	bz := v.Value.GetRaw(ctx)
	if bz == nil {
		return res, ErrEmptyValue()
	}
	return bz[0], nil // TODO: check length
}

// Set() marshales and sets the byte argument to the state.
func (v Enum) Set(ctx Context, value byte) {
	v.Value.SetRaw(ctx, []byte{value})
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

func (v Enum) Query(ctx CLIContext) (res byte, proof *Proof, err error) {
	value, proof, err := v.Value.QueryRaw(ctx)
	return value[0], proof, err
}
