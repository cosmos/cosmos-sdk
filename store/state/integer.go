package state

// Integer is a uint64 types wrapper for Value.
// Except for the type checking, it does not alter the behaviour.
type Integer struct {
	Value

	enc IntEncoding
}

func NewInteger(v Value, enc IntEncoding) Integer {
	return Integer{
		Value: v,
		enc:   enc,
	}
}

func (v Integer) Get(ctx Context) (res uint64) {
	v.Value.Get(ctx, &res)
	return res
}

func (v Integer) GetSafe(ctx Context) (res uint64, err error) {
	err = v.Value.GetSafe(ctx, &res)
	return
}

func (v Integer) Set(ctx Context, value uint64) {
	v.Value.Set(ctx, value)
}

// Incr() increments the stored value, and returns the updated value.
func (v Integer) Incr(ctx Context) (res uint64) {
	res = v.Get(ctx) + 1
	v.Set(ctx, res)
	return
}
