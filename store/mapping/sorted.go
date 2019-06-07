package mapping

type Sorted struct {
	m Mapping

	enc IntEncoding
}

func NewSorted(base Base, prefix []byte, enc IntEncoding) Sorted {
	return Sorted{
		m: NewMapping(base, prefix),

		enc: enc,
	}
}

func (s Sorted) key(power uint64, key []byte) []byte {
	return append(EncodeInt(power, s.enc), key...)
}

func (s Sorted) Value(power uint64, key []byte) Value {
	return s.m.Value(s.key(power, key))
}

func (s Sorted) Get(ctx Context, power uint64, key []byte, ptr interface{}) {
	s.Value(power, key).Get(ctx, ptr)
}

func (s Sorted) GetIfExists(ctx Context, power uint64, key []byte, ptr interface{}) {
	s.Value(power, key).GetIfExists(ctx, ptr)
}

func (s Sorted) Set(ctx Context, power uint64, key []byte, o interface{}) {
	s.Value(power, key).Set(ctx, o)
}

func (s Sorted) Has(ctx Context, power uint64, key []byte) bool {
	return s.Value(power, key).Exists(ctx)
}

func (s Sorted) Delete(ctx Context, power uint64, key []byte) {
	s.Value(power, key).Delete(ctx)
}

func (s Sorted) IsEmpty(ctx Context) bool {
	return s.m.IsEmpty(ctx)
}

func (s Sorted) Prefix(prefix []byte) Sorted {
	return Sorted{
		m:   s.m.Prefix(prefix),
		enc: s.enc,
	}
}

func (s Sorted) IterateAscending(ctx Context, ptr interface{}, fn func([]byte) bool) {
	s.m.Iterate(ctx, ptr, fn)
}

func (s Sorted) IterateDescending(ctx Context, ptr interface{}, fn func([]byte) bool) {
	s.m.ReverseIterate(ctx, ptr, fn)
}

func (s Sorted) Clear(ctx Context) {
	s.m.Clear(ctx)
}
