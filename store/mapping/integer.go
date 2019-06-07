package mapping

type Integer interface {
	coreValue
	Get(Context) uint64
	GetIfExists(Context) uint64
	GetSafe(Context) (uint64, error)
	Set(Context, uint64)
	Incr(Context) uint64
	Equal(Context, uint64) bool
}

var _ Integer = integer{}

type integer struct {
	Value

	enc IntEncoding
}

func NewInteger(v Value, enc IntEncoding) integer {
	return integer{
		Value: v,
		enc:   enc,
	}
}

/*
func (v Value) integer() integer {
	return integer{v}
}
*/
func (v integer) Get(ctx Context) uint64 {
	res, err := DecodeInt(v.GetRaw(ctx), v.enc)
	if err != nil {
		panic(err)
	}
	return res
}

func (v integer) GetIfExists(ctx Context) (res uint64) {
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

func (v integer) GetSafe(ctx Context) (uint64, error) {
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

func (v integer) Set(ctx Context, value uint64) {
	v.SetRaw(ctx, EncodeInt(value, v.enc))
}

func (v integer) Incr(ctx Context) (res uint64) {
	res = v.GetIfExists(ctx) + 1
	v.Set(ctx, res)
	return
}

func (v integer) Equal(ctx Context, value uint64) bool {
	return v.Get(ctx) == value
}

/*
func (v integer) Key() []byte {
	return v.base.key(v.key)
}
*/
