package codec

// AminoCodec defines a codec that utilizes Codec for both binary and JSON
// encoding.
type AminoCodec struct {
	*LegacyAmino
}

var _ Marshaler = &AminoCodec{}

func NewAminoCodec(codec *LegacyAmino) *AminoCodec {
	return &AminoCodec{LegacyAmino: codec}
}

func (ac *AminoCodec) MarshalBinaryBare(o ProtoMarshaler) ([]byte, error) {
	return ac.LegacyAmino.MarshalBinaryBare(o)
}

func (ac *AminoCodec) MustMarshalBinaryBare(o ProtoMarshaler) []byte {
	return ac.LegacyAmino.MustMarshalBinaryBare(o)
}

func (ac *AminoCodec) MarshalBinaryLengthPrefixed(o ProtoMarshaler) ([]byte, error) {
	return ac.LegacyAmino.MarshalBinaryLengthPrefixed(o)
}

func (ac *AminoCodec) MustMarshalBinaryLengthPrefixed(o ProtoMarshaler) []byte {
	return ac.LegacyAmino.MustMarshalBinaryLengthPrefixed(o)
}

func (ac *AminoCodec) UnmarshalBinaryBare(bz []byte, ptr ProtoMarshaler) error {
	return ac.LegacyAmino.UnmarshalBinaryBare(bz, ptr)
}

func (ac *AminoCodec) MustUnmarshalBinaryBare(bz []byte, ptr ProtoMarshaler) {
	ac.LegacyAmino.MustUnmarshalBinaryBare(bz, ptr)
}

func (ac *AminoCodec) UnmarshalBinaryLengthPrefixed(bz []byte, ptr ProtoMarshaler) error {
	return ac.LegacyAmino.UnmarshalBinaryLengthPrefixed(bz, ptr)
}

func (ac *AminoCodec) MustUnmarshalBinaryLengthPrefixed(bz []byte, ptr ProtoMarshaler) {
	ac.LegacyAmino.MustUnmarshalBinaryLengthPrefixed(bz, ptr)
}
