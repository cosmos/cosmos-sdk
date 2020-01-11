package proto

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

type Codec struct{}

var _ codec.CodecI = &Codec{}

func (c Codec) MustUnmarshalBinaryLengthPrefixed(bz []byte, t Marshaler) {
	panic("TODO")
}

func (c Codec) MustMarshalBinaryLengthPrefixed(t Marshaler) []byte {
	panic("TODO")
}

func (c Codec) MarshalJSON(obj interface{}) ([]byte, error) {
	panic("implement me")
}

func (c Codec) UnmarshalJSON(data []byte, t interface{}) error {
	panic("implement me")
}

type Marshaler interface {
	Marshal() ([]byte, error)
	MarshalTo(data []byte) (n int, err error)
	MarshalToSizedBuffer(dAtA []byte) (int, error)
	Size() int
	Unmarshal(data []byte) error
}
