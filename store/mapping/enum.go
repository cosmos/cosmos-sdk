package mapping

type Enum struct {
	Value
}

func NewEnum(base Base, key []byte) Enum {
	return Enum{
		Value: NewValue(base, key),
	}
}

/*
func (v Value) Enum() Enum {
	return Enum{v}
}
*/
func (v Enum) Get(ctx Context) byte {
	return v.Value.GetRaw(ctx)[0]
}

func (v Enum) GetIfExists(ctx Context) byte {
	res := v.Value.GetRaw(ctx)
	if res != nil {
		return res[0]
	}
	return 0x00
}

func (v Enum) GetSafe(ctx Context) (byte, error) {
	res := v.Value.GetRaw(ctx)
	if res == nil {
		return 0x00, &GetSafeError{}
	}
	return res[0], nil
}

func (v Enum) Set(ctx Context, value byte) {
	v.Value.SetRaw(ctx, []byte{value})
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

/*
func (v Enum) Key() []byte {
	return v.base.key(v.key)
}
*/
