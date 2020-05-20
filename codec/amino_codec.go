package codec

// AminoCodec defines a codec that utilizes Codec for both binary and JSON
// encoding.
type AminoCodec struct {
	*Codec
}

var _ Marshaler = &AminoCodec{}

func NewAminoCodec(codec *Codec) *AminoCodec {
	return &AminoCodec{Codec: codec}
}

func (ac *AminoCodec) MarshalBinaryBare(o ProtoMarshaler) ([]byte, error) {
	return ac.Codec.MarshalBinaryBare(o)
}

func (ac *AminoCodec) MustMarshalBinaryBare(o ProtoMarshaler) []byte {
	return ac.Codec.MustMarshalBinaryBare(o)
}

func (ac *AminoCodec) MarshalBinaryLengthPrefixed(o ProtoMarshaler) ([]byte, error) {
	return ac.Codec.MarshalBinaryLengthPrefixed(o)
}

func (ac *AminoCodec) MustMarshalBinaryLengthPrefixed(o ProtoMarshaler) []byte {
	return ac.Codec.MustMarshalBinaryLengthPrefixed(o)
}

func (ac *AminoCodec) UnmarshalBinaryBare(bz []byte, ptr ProtoMarshaler) error {
	return ac.Codec.UnmarshalBinaryBare(bz, ptr)
}

func (ac *AminoCodec) MustUnmarshalBinaryBare(bz []byte, ptr ProtoMarshaler) {
	ac.Codec.MustUnmarshalBinaryBare(bz, ptr)
}

func (ac *AminoCodec) UnmarshalBinaryLengthPrefixed(bz []byte, ptr ProtoMarshaler) error {
	return ac.Codec.UnmarshalBinaryLengthPrefixed(bz, ptr)
}

func (ac *AminoCodec) MustUnmarshalBinaryLengthPrefixed(bz []byte, ptr ProtoMarshaler) {
	ac.Codec.MustUnmarshalBinaryLengthPrefixed(bz, ptr)
}
