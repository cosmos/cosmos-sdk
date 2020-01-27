package codec

// AminoCodec defines a codec that utilizes Amino for both binary and JSON
// encoding.
type AminoCodec struct {
	amino *Codec
}

func NewAminoCodec(amino *Codec) Marshaler {
	return &AminoCodec{amino}
}

func (ac *AminoCodec) MarshalBinaryBare(o ProtoMarshaler) ([]byte, error) {
	return ac.amino.MarshalBinaryBare(o)
}

func (ac *AminoCodec) MustMarshalBinaryBare(o ProtoMarshaler) []byte {
	return ac.amino.MustMarshalBinaryBare(o)
}

func (ac *AminoCodec) MarshalBinaryLengthPrefixed(o ProtoMarshaler) ([]byte, error) {
	return ac.amino.MarshalBinaryLengthPrefixed(o)
}

func (ac *AminoCodec) MustMarshalBinaryLengthPrefixed(o ProtoMarshaler) []byte {
	return ac.amino.MustMarshalBinaryLengthPrefixed(o)
}

func (ac *AminoCodec) UnmarshalBinaryBare(bz []byte, ptr ProtoMarshaler) error {
	return ac.amino.UnmarshalBinaryBare(bz, ptr)
}

func (ac *AminoCodec) MustUnmarshalBinaryBare(bz []byte, ptr ProtoMarshaler) {
	ac.amino.MustUnmarshalBinaryBare(bz, ptr)
}

func (ac *AminoCodec) UnmarshalBinaryLengthPrefixed(bz []byte, ptr ProtoMarshaler) error {
	return ac.amino.UnmarshalBinaryLengthPrefixed(bz, ptr)
}

func (ac *AminoCodec) MustUnmarshalBinaryLengthPrefixed(bz []byte, ptr ProtoMarshaler) {
	ac.amino.MustUnmarshalBinaryLengthPrefixed(bz, ptr)
}

func (ac *AminoCodec) MarshalJSON(o interface{}) ([]byte, error) { // nolint: stdmethods
	return ac.amino.MarshalJSON(o)
}

func (ac *AminoCodec) MustMarshalJSON(o interface{}) []byte {
	return ac.amino.MustMarshalJSON(o)
}

func (ac *AminoCodec) UnmarshalJSON(bz []byte, ptr interface{}) error { // nolint: stdmethods
	return ac.amino.UnmarshalJSON(bz, ptr)
}

func (ac *AminoCodec) MustUnmarshalJSON(bz []byte, ptr interface{}) {
	ac.amino.MustUnmarshalJSON(bz, ptr)
}
