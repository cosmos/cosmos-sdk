package codec

import (
	"bytes"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var defaultJM = &jsonpb.Marshaler{OrigName: true, EmitDefaults: true, AnyResolver: nil}

// ProtoMarshalJSON provides an auxiliary function to return Proto3 JSON encoded
// bytes of a message.
func ProtoMarshalJSON(msg proto.Message, resolver jsonpb.AnyResolver) ([]byte, error) {
	// We use the OrigName because camel casing fields just doesn't make sense.
	// EmitDefaults is also often the more expected behavior for CLI users
	jm := defaultJM
	if resolver != nil {
		jm = &jsonpb.Marshaler{OrigName: true, EmitDefaults: true, AnyResolver: resolver}
	}
	err := types.UnpackInterfaces(msg, types.ProtoJSONPacker{JSONPBMarshaler: jm})
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)

	if err := jm.Marshal(buf, msg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ProtoMarshalJSONI same as ProtoMarshalJSON, but does msg type inspection to assert
// that it implements `proto.Message` and return an error if it doesn't.
func ProtoMarshalJSONI(msg interface{}, resolver jsonpb.AnyResolver) ([]byte, error) {
	msgProto, err := AssertProtoMsg(msg)
	if err != nil {
		return nil, err
	}
	return ProtoMarshalJSON(msgProto, resolver)
}

// AssertProtoMsg casts i to a proto.Message. Returns an error if it's not possible.
func AssertProtoMsg(i interface{}) (proto.Message, error) {
	pm, ok := i.(proto.Message)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "Expecting protobuf Message type, got %T", i)
	}
	return pm, nil
}
