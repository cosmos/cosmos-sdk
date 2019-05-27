package mapping

type Boolean struct {
	Value
}

func NewBoolean(base Base, key []byte) Boolean {
	return Boolean{
		Value: NewValue(base, key),
	}
}

func (v Value) Boolean() Boolean {
	return Boolean{v}
}

func (v Boolean) Get(ctx Context) (res bool) {
	v.Value.Get(ctx, &res)
	return
}

func (v Boolean) GetIfExists(ctx Context) (res bool) {
	v.Value.GetIfExists(ctx, &res)
	return
}

func (v Boolean) GetSafe(ctx Context) (res bool, err error) {
	err = v.Value.GetSafe(ctx, &res)
	return
}

func (v Boolean) Set(ctx Context, value bool) {
	v.Value.Set(ctx, value)
}

func (v Boolean) Flip(ctx Context) (res bool) {
	res = !v.GetIfExists(ctx)
	v.Set(ctx, res)
	return
}

func (v Boolean) Key() []byte {
	return v.base.key(v.key)
}
