package mapping

type Boolean struct {
	Enum
}

func NewBoolean(v Value) Boolean {
	return Boolean{NewEnum(v)}
}

func (v Boolean) Get(ctx Context) bool {
	return v.Enum.Get(ctx) != 0x00
}

func (v Boolean) GetIfExists(ctx Context) bool {
	return v.Enum.GetIfExists(ctx) != 0x00
}

func (v Boolean) GetSafe(ctx Context) (bool, error) {
	res, err := v.Enum.GetSafe(ctx)
	return res != 0x00, err
}

func (v Boolean) Set(ctx Context, value bool) {
	if value {
		v.Enum.Set(ctx, 0x01)
	} else {
		v.Enum.Set(ctx, 0x00)
	}
}
