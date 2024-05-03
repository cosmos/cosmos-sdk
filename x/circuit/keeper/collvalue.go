package keeper

import (
	"bytes"

	collcodec "cosmossdk.io/collections/codec"
	"github.com/cosmos/gogoproto/proto"
	"github.com/gogo/protobuf/jsonpb"
)

type protoMessage[T any] interface {
	*T
	proto.Message
}

// CollValue inits a collections.ValueCodec for a generic gogo protobuf message.
func CollValue[T any, PT protoMessage[T]]() collcodec.ValueCodec[T] {
	return &collValue[T, PT]{proto.MessageName(PT(new(T)))}
}

type collValue[T any, PT protoMessage[T]] struct {
	messageName string
}

func (c collValue[T, PT]) Encode(value T) ([]byte, error) {
	return proto.Marshal(PT(&value))
}

func (c collValue[T, PT]) Decode(b []byte) (value T, err error) {
	err = proto.Unmarshal(b, PT(&value))
	return value, err
}

func (c collValue[T, PT]) EncodeJSON(value T) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := (&jsonpb.Marshaler{}).Marshal(buf, PT(&value))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c collValue[T, PT]) DecodeJSON(b []byte) (value T, err error) {
	err = jsonpb.Unmarshal(bytes.NewBuffer(b), PT(&value))
	return
}

func (c collValue[T, PT]) Stringify(value T) string {
	return PT(&value).String()
}

func (c collValue[T, PT]) ValueType() string {
	return "github.com/cosmos/gogoproto/" + c.messageName
}
