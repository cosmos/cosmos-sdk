package protocdc

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/jsonpb"
	"google.golang.org/protobuf/proto"
)

// MarshalJSONI same as codec.ProtoMarshalJSON, but does msg type inspection to assert
// that it implements `proto.Message` and return an error if it doesn't.
func MarshalJSONI(msg interface{}, resolver jsonpb.AnyResolver) ([]byte, error) {
	msgProto, err := AssertMsg(msg)
	if err != nil {
		return nil, err
	}
	return codec.ProtoMarshalJSON(msgProto, resolver)
}

// AssertMsg casts i to a proto.Message. Returns an error if it's not possible.
func AssertMsg(i interface{}) (proto.Message, error) {
	pm, ok := i.(proto.Message)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "Expecting protobuf Message type, got %T", i)
	}
	return pm, nil
}
