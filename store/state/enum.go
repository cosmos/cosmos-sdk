package state

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
	v.Value.GetSafe(ctx, &res)
	return
}

func (v Enum) Set(ctx Context, value byte) {
	v.Value.Set(ctx, value)
}

func (v Enum) Incr(ctx Context) (res byte) {
	res = v.Get(ctx) + 1
	v.Set(ctx, res)
	return
}

func (v Enum) Transit(ctx Context, from, to byte) bool {
	if v.Get(ctx) != from {
		return false
	}
	v.Set(ctx, to)
	return true
}
