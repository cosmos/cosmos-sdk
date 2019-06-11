package mapping

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

/*
func (v Value) Integer() Integer {
	return Integer{v}
}
*/
func (v Integer) Get(ctx Context) uint64 {
	res, err := DecodeInt(v.GetRaw(ctx), v.enc)
	if err != nil {
		panic(err)
	}
	return res
}

func (v Integer) GetIfExists(ctx Context) (res uint64) {
	bz := v.GetRaw(ctx)
	if bz == nil {
		return 0
	}
	res, err := DecodeInt(bz, v.enc)
	if err != nil {
		panic(err)
	}
	return res
}

func (v Integer) GetSafe(ctx Context) (uint64, error) {
	bz := v.GetRaw(ctx)
	if bz == nil {
		return 0, &GetSafeError{}
	}
	res, err := DecodeInt(bz, v.enc)
	if err != nil {
		panic(err)
	}
	return res, nil
}

func (v Integer) Set(ctx Context, value uint64) {
	v.SetRaw(ctx, EncodeInt(value, v.enc))
}

func (v Integer) Incr(ctx Context) (res uint64) {
	res = v.GetIfExists(ctx) + 1
	v.Set(ctx, res)
	return
}

func (v Integer) Is(ctx Context, value uint64) bool {
	return v.Get(ctx) == value
}

/*
func (v Integer) Key() []byte {
	return v.base.key(v.key)
}
*/
