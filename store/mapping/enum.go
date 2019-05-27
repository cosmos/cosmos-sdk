package mapping

type Enum struct {
	Value
}

func NewEnum(base Base, key []byte) Enum {
	return Enum{
		Value: NewValue(base, key),
	}
}

func (v Value) Enum() Enum {
	return Enum{v}
}

func (v Enum) Get(ctx Context) (res byte) {
	v.Value.Get(ctx, &res)
	return
}

func (v Enum) GetIfExists(ctx Context) (res byte) {
	v.Value.GetIfExists(ctx, &res)
	return
}

func (v Enum) GetSafe(ctx Context) (res byte, err error) {
	err = v.Value.GetSafe(ctx, &res)
	return
}

func (v Enum) Set(ctx Context, value byte) {
	v.Value.Set(ctx, value)
}

func (v Enum) Incr(ctx Context) (res byte) {
	res = v.GetIfExists(ctx) + 1
	v.Set(ctx, res)
	return
}

func (v Enum) Transit(ctx Context, from, to byte) bool {
	if v.GetIfExists(ctx) != from {
		return false
	}
	v.Set(ctx, to)
	return true
}

func (v Enum) Key() []byte {
	return v.base.key(v.key)
}
