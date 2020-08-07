package codec

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
)

// ProtoMarshalJSON provides an auxiliary function to return Proto3 JSON encoded
// bytes of a message.
func ProtoMarshalJSON(msg proto.Message) ([]byte, error) {
	// we use the original proto name because camel case just doesn't make sense
	// EmitDefaults is also the more expected behavior for CLI users.
	jm := &jsonpb.Marshaler{OrigName: true, EmitDefaults: true}
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
