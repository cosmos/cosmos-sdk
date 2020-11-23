package protocdc

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec"
)

// MarshalJSON is a wrapper for codec.ProtoMarshalJSON. It asserts that msg
// implements `proto.Message` and calls codec.ProtoMarshalJSON.
// This function should be used only with concrete types. For interface serialization
// you need to wrap the interface into Any or generally use MarshalIfcJSON.
func MarshalJSON(msg interface{}, resolver jsonpb.AnyResolver) ([]byte, error) {
	msgProto, err := AssertMsg(msg)
	if err != nil {
		return nil, err
	}
	return codec.ProtoMarshalJSON(msgProto, resolver)
}

// AssertMsg casts i to a proto.Message. Returns an error if it's not possible.
// TODO remove
func AssertMsg(i interface{}) (proto.Message, error) {
	pm, ok := i.(proto.Message)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "Expecting protobuf Message type, got %T", i)
	}
	return pm, nil
}
