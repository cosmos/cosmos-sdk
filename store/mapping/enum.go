package mapping

type Enum interface {
	coreValue
	Get(Context) byte
	GetIfExists(Context) byte
	GetSafe(Context) (byte, error)
	Set(Context, byte)
	Transit(Context, byte, byte) bool
	Equal(Context, byte) bool
}

var _ Enum = enum{}

type enum struct {
	Value
}

func NewEnum(v Value) enum {
	return enum{v}
}

/*
func (v Value) enum() enum {
	return enum{v}
}
*/
func (v enum) Get(ctx Context) byte {
	return v.Value.GetRaw(ctx)[0]
}

func (v enum) GetIfExists(ctx Context) byte {
	res := v.Value.GetRaw(ctx)
	if res != nil {
		return res[0]
	}
	return 0x00
}

func (v enum) GetSafe(ctx Context) (byte, error) {
	res := v.Value.GetRaw(ctx)
	if res == nil {
		return 0x00, &GetSafeError{}
	}
	return res[0], nil
}

func (v enum) Set(ctx Context, value byte) {
	v.Value.SetRaw(ctx, []byte{value})
}

func (v enum) Incr(ctx Context) (res byte) {
	res = v.GetIfExists(ctx) + 1
	v.Set(ctx, res)
	return
}

func (v enum) Transit(ctx Context, from, to byte) bool {
	if v.GetIfExists(ctx) != from {
		return false
	}
	v.Set(ctx, to)
	return true
}

func (v enum) Equal(ctx Context, value byte) bool {
	return v.Get(ctx) == value
}

/*
func (v enum) Key() []byte {
	return v.base.key(v.key)
}
*/
