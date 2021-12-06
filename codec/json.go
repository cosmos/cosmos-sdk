package codec

import (
	"bytes"
	"fmt"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/gogo/protobuf/jsonpb"
	gogoproto "github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

var defaultJM = &jsonpb.Marshaler{OrigName: true, EmitDefaults: true, AnyResolver: nil}

// ProtoMarshalJSON provides an auxiliary function to return Proto3 JSON encoded
// bytes of a message.
func ProtoMarshalJSON(msg interface{}, resolver types.InterfaceRegistry) ([]byte, error) {
	switch msg := msg.(type) {
	case gogoproto.Message:
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
	case *protoreflect.Message:
		return protojson.MarshalOptions{Resolver: resolver}.Marshal((*msg).Interface())
	default:
		return nil, fmt.Errorf("%T is neither a v1 or v2 proto message", msg)
	}

}
