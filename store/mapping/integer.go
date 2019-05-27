package mapping

type Integer struct {
	Value
}

func NewInteger(base Base, key []byte) Integer {
	return Integer{
		Value: NewValue(base, key),
	}
}

func (v Value) Integer() Integer {
	return Integer{v}
}

func (v Integer) Get(ctx Context) (res int64) {
	v.Value.Get(ctx, &res)
	return
}

func (v Integer) GetIfExists(ctx Context) (res int64) {
	v.Value.GetIfExists(ctx, &res)
	return
}

func (v Integer) GetSafe(ctx Context) (res int64, err error) {
	err = v.Value.GetSafe(ctx, &res)
	return
}

func (v Integer) Set(ctx Context, value int64) {
	v.Value.Set(ctx, value)
}

func (v Integer) Incr(ctx Context) (res int64) {
	res = v.GetIfExists(ctx) + 1
	v.Set(ctx, res)
	return
}

func (v Integer) Key() []byte {
	return v.base.key(v.key)
}
