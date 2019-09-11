package state

// String is a string types wrapper for Value.
// x <-> []byte(x)
type String struct {
	Value
}

func (v Value) String() String {
	return String{v}
}

// Get decodes and returns the stored string value if it exists. It will panic
// if the value exists but is not string type.
func (v String) Get(ctx Context) (res string) {
	return string(v.Value.GetRaw(ctx))
}

// GetSafe decodes and returns the stored string value. It will return an error
// if the value does not exist or not string.
func (v String) GetSafe(ctx Context) (res string, err error) {
	bz := v.Value.GetRaw(ctx)
	if bz == nil {
		return res, ErrEmptyValue()
	}
	return string(bz), nil
}

// Set encodes and sets the string argument to the state.
func (v String) Set(ctx Context, value string) {
	v.Value.SetRaw(ctx, []byte(value))
}

// Query() retrives state value and proof from a queryable reference
func (v String) Query(q ABCIQuerier) (res string, proof *Proof, err error) {
	value, proof, err := v.Value.QueryRaw(q)
	return string(value), proof, err
}
