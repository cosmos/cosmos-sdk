package codec

import (
	"bytes"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/gogo/protobuf/jsonpb"
	gogoproto "github.com/gogo/protobuf/proto"
	protopb "github.com/golang/protobuf/jsonpb"
	"google.golang.org/protobuf/encoding/protojson"
	protov2 "google.golang.org/protobuf/proto"
)

var defaultJM = &jsonpb.Marshaler{OrigName: true, EmitDefaults: true, AnyResolver: nil}
var protoJM = &protopb.Marshaler{OrigName: true, EmitDefaults: true, AnyResolver: nil}

// ProtoMarshalJSON provides an auxiliary function to return Proto3 JSON encoded
// bytes of a message.
func ProtoMarshalJSON(msg interface{}, resolver types.InterfaceRegistry) ([]byte, error) {
	switch msg := msg.(type) {
	case protov2.Message:
		return protojson.MarshalOptions{Resolver: resolver}.Marshal(msg.ProtoReflect().Interface())
	case gogoproto.Message:
		// We use the OrigName because camel casing fields just doesn't make sense.
		// EmitDefaults is also often the more expected behavior for CLI users
		jm := defaultJM
		if resolver != nil {
			jm = &jsonpb.Marshaler{OrigName: true, EmitDefaults: true, AnyResolver: resolver}
		}
		err := types.UnpackInterfaces(msg, types.ProtoJSONPacker{JSONPBMarshaler: jm, V2MarshalOptions: protojson.MarshalOptions{
			NoUnkeyedLiterals: struct{}{},
			Multiline:         false,
			Indent:            " ",
			AllowPartial:      false,
			UseProtoNames:     false,
			UseEnumNumbers:    false,
			EmitUnpopulated:   false,
			Resolver:          resolver,
		}})
		if err != nil {
			return nil, err
		}

		buf := new(bytes.Buffer)
		if err := protoJM.Marshal(buf, msg); err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	default:
		return nil, fmt.Errorf("%T is neither a v1 or v2 proto message", msg)
	}

}
