package proto

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
)

type Codec struct{}

var _ codec.CodecI = &Codec{}

func (c Codec) MustUnmarshalBinaryLengthPrefixed(bz []byte, t Marshaler) {
	// TODO actually length prefix
	err := t.Unmarshal(bz)
	if err != nil {
		panic(err)
	}
}

func (c Codec) MustMarshalBinaryLengthPrefixed(t Marshaler) []byte {
	bz, err := t.Marshal()
	if err != nil {
		panic(err)
	}
	return bz
}

func (c Codec) MarshalJSON(obj interface{}) ([]byte, error) {
	msg, ok := obj.(proto.Message)
	if !ok {
		return json.Marshal(obj)
	}
	m := jsonpb.Marshaler{}
	str, err := m.MarshalToString(msg)
	if err != nil {
		return nil, err
	}
	return []byte(str), nil
}

func (c Codec) UnmarshalJSON(data []byte, t interface{}) error {
	msg, ok := t.(proto.Message)
	if !ok {
		return json.Unmarshal(data, t)
	}
	return jsonpb.UnmarshalString(string(data), msg)
}

type Marshaler interface {
	Marshal() ([]byte, error)
	MarshalTo(data []byte) (n int, err error)
	MarshalToSizedBuffer(dAtA []byte) (int, error)
	Size() int
	Unmarshal(data []byte) error
}
