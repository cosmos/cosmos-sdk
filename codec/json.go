package codec

import (
	"bytes"
	"encoding/json"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
)

// MarshalIndentFromJSON returns indented JSON-encoded bytes from already encoded
// JSON bytes. The output encoding will adhere to the original input's encoding
// (e.g. Proto3).
func MarshalIndentFromJSON(bz []byte) ([]byte, error) {
	var generic interface{}

	if err := json.Unmarshal(bz, &generic); err != nil {
		return nil, err
	}

	return json.MarshalIndent(generic, "", "  ")
}

// ProtoMarshalJSON provides an auxiliary function to return Proto3 JSON encoded
// bytes of a message.
func ProtoMarshalJSON(msg proto.Message) ([]byte, error) {
	jm := &jsonpb.Marshaler{EmitDefaults: false, OrigName: false}
	buf := new(bytes.Buffer)

	if err := jm.Marshal(buf, msg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ProtoMarshalJSONIndent provides an auxiliary function to return Proto3 indented
// JSON encoded bytes of a message.
func ProtoMarshalJSONIndent(msg proto.Message) ([]byte, error) {
	jm := &jsonpb.Marshaler{EmitDefaults: false, OrigName: false, Indent: "  "}
	buf := new(bytes.Buffer)

	if err := jm.Marshal(buf, msg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
