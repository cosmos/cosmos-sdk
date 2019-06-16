package state

import "fmt"

type Enum struct {
	Value
}

func NewEnum(v Value) Enum {
	return Enum{v}
}

func (v Enum) Get(ctx Context) byte {
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
	res = v.Get(ctx) + 1
	v.Set(ctx, res)
	return
}

func (v Enum) Transit(ctx Context, from, to byte) bool {
	if v.Get(ctx) != from {
		fmt.Println("nnnnnnn", from, to, v.Get(ctx))
		return false
	}
	v.Set(ctx, to)
	return true
}
