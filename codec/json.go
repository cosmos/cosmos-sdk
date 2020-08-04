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
	jm := &jsonpb.Marshaler{
		EnumsAsInts:  false,
		EmitDefaults: true, // emit defaults by default - this should not be used for sign bytes
		Indent:       "",
		OrigName:     true, // don't camel case names
		AnyResolver:  nil,
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
