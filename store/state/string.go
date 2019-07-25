package state

// String is a string types wrapper for Value.
// Except for the type checking, it does not alter the behaviour.
type String struct {
	Value
}

func NewString(v Value) String {
	return String{v}
}

func (v String) Get(ctx Context) (res string) {
	v.Value.Get(ctx, &res)
	return
}

func (v String) GetSafe(ctx Context) (res string, err error) {
	err = v.Value.GetSafe(ctx, &res)
	return
}

func (v String) Set(ctx Context, value string) {
	v.Value.Set(ctx, value)
}
