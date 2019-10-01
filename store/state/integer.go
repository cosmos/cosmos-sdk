package state

// Integer is a uint64 types wrapper for Value.
// The serialization follows the @IntEncoding@ format provided to the NewInteger.
type Integer struct {
	Value

	encoding IntEncoding
}

func (v Value) Integer(enc IntEncoding) Integer {
	return Integer{v, enc}
}

// Get() unmarshales and returns the stored uint64 value if it exists.
// If will panic if the value exists but not decodable.
func (v Integer) Get(ctx Context) uint64 {
	bz := v.Value.GetRaw(ctx)
	if bz == nil {
		return 0
	}
	res, err := DecodeInt(bz, v.enc)
	if err != nil {
		panic(err)
	}
	return res
}

// GetSafe() unmarshales and returns the stored uint64 value.
// It will return an error if the value does not exist or not uint64.
func (v Integer) GetSafe(ctx Context) (res uint64, err error) {
	bz := v.Value.GetRaw(ctx)
	if bz == nil {
		return 0, ErrEmptyValue()
	}
	res, err = DecodeInt(bz, v.enc)
	if err != nil {
		err = ErrUnmarshal(err)
	}
	return
}

// Set() marshales and sets the uint64 argument to the state.
func (v Integer) Set(ctx Context, value uint64) {
	v.Value.SetRaw(ctx, EncodeInt(value, v.enc))
}

// Increment increments the stored value and returns it.
func (v Integer) Increment(ctx Context) (res uint64) {
	res = v.Get(ctx) + 1
	v.Set(ctx, res)
	return
}

// Query() retrives state value and proof from a queryable reference
func (v Integer) Query(q ABCIQuerier) (res uint64, proof *Proof, err error) {
	value, proof, err := v.Value.QueryRaw(q)
	if err != nil {
		return
	}
	if value == nil {
		res = 0
		return
	}
	res, err = DecodeInt(value, v.enc)
	return
}
