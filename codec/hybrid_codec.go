package codec

import "github.com/cosmos/cosmos-sdk/codec/types"

// HybridCodec defines a codec that utilizes Protobuf for binary encoding
// and Amino for JSON encoding.
type HybridCodec struct {
	proto Marshaler
	amino Marshaler
}

func NewHybridCodec(amino *LegacyAmino, unpacker types.AnyUnpacker) Marshaler {
	return &HybridCodec{
		proto: NewProtoCodec(unpacker),
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

func (hc *HybridCodec) MarshalJSON(o interface{}) ([]byte, error) {
	return hc.amino.MarshalJSON(o)
}

func (hc *HybridCodec) MustMarshalJSON(o interface{}) []byte {
	return hc.amino.MustMarshalJSON(o)
}

func (hc *HybridCodec) UnmarshalJSON(bz []byte, ptr interface{}) error {
	return hc.amino.UnmarshalJSON(bz, ptr)
}

func (hc *HybridCodec) MustUnmarshalJSON(bz []byte, ptr interface{}) {
	hc.amino.MustUnmarshalJSON(bz, ptr)
}

func (hc *HybridCodec) UnpackAny(any *types.Any, iface interface{}) error {
	return hc.proto.UnpackAny(any, iface)
}
