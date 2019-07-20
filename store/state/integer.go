package state

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
	/*
		bz := v.GetRaw(ctx)
		if bz == nil {
			return 0
		}
		res, err := DecodeInt(bz, v.enc)
		if err != nil {
			panic(err)
		}
		return res
	*/
	v.Value.Get(ctx, &res)
	return
}

func (v Integer) GetSafe(ctx Context) (res uint64, err error) {
	/*
		bz := v.GetRaw(ctx)
		if bz == nil {
			return 0, &GetSafeError{}
		}
		res, err := DecodeInt(bz, v.enc)
		if err != nil {
			panic(err)
		}
		return res, nil
	*/
	err = v.Value.GetSafe(ctx, &res)
	return
}

func (v Integer) Set(ctx Context, value uint64) {
	//	v.SetRaw(ctx, EncodeInt(value, v.enc))
	v.Value.Set(ctx, value)
}

func (v Integer) Incr(ctx Context) (res uint64) {
	res = v.Get(ctx) + 1
	v.Set(ctx, res)
	return
}
