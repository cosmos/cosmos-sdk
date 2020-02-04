package codec

// HybridCodec defines a codec that utilizes Protobuf for binary encoding
// and Amino for JSON encoding.
type HybridCodec struct {
	proto Marshaler
	amino Marshaler
}

func NewHybridCodec(amino *Codec) Marshaler {
	return &HybridCodec{
		proto: NewProtoCodec(),
		amino: NewAminoCodec(amino),
	}
}

func (hc *HybridCodec) MarshalBinaryBare(o ProtoMarshaler) ([]byte, error) {
	return hc.proto.MarshalBinaryBare(o)
}

func (hc *HybridCodec) MustMarshalBinaryBare(o ProtoMarshaler) []byte {
	return hc.proto.MustMarshalBinaryBare(o)
}

func (hc *HybridCodec) MarshalBinaryLengthPrefixed(o ProtoMarshaler) ([]byte, error) {
	return hc.proto.MarshalBinaryLengthPrefixed(o)
}

func (hc *HybridCodec) MustMarshalBinaryLengthPrefixed(o ProtoMarshaler) []byte {
	return hc.proto.MustMarshalBinaryLengthPrefixed(o)
}

func (hc *HybridCodec) UnmarshalBinaryBare(bz []byte, ptr ProtoMarshaler) error {
	return hc.proto.UnmarshalBinaryBare(bz, ptr)
}

func (hc *HybridCodec) MustUnmarshalBinaryBare(bz []byte, ptr ProtoMarshaler) {
	hc.proto.MustUnmarshalBinaryBare(bz, ptr)
}

func (hc *HybridCodec) UnmarshalBinaryLengthPrefixed(bz []byte, ptr ProtoMarshaler) error {
	return hc.proto.UnmarshalBinaryLengthPrefixed(bz, ptr)
}

func (hc *HybridCodec) MustUnmarshalBinaryLengthPrefixed(bz []byte, ptr ProtoMarshaler) {
	hc.proto.MustUnmarshalBinaryLengthPrefixed(bz, ptr)
}

func (hc *HybridCodec) MarshalJSON(o interface{}) ([]byte, error) { // nolint: stdmethods
	return hc.amino.MarshalJSON(o)
}

func (hc *HybridCodec) MustMarshalJSON(o interface{}) []byte {
	return hc.amino.MustMarshalJSON(o)
}

func (hc *HybridCodec) UnmarshalJSON(bz []byte, ptr interface{}) error { // nolint: stdmethods
	return hc.amino.UnmarshalJSON(bz, ptr)
}

func (hc *HybridCodec) MustUnmarshalJSON(bz []byte, ptr interface{}) {
	hc.amino.MustUnmarshalJSON(bz, ptr)
}
