// nolint: dupl
// dupl is reading this as the same file as crypto/merkle/result.go
package kv

import (
	"bytes"
	"encoding/json"

	"github.com/gogo/protobuf/jsonpb"
)

//---------------------------------------------------------------------------
// override JSON marshalling so we emit defaults (ie. disable omitempty)

var (
	jsonpbMarshaller = jsonpb.Marshaler{
		EnumsAsInts:  true,
		EmitDefaults: true,
	}
	jsonpbUnmarshaller = jsonpb.Unmarshaler{}
)

func (r *Pair) MarshalJSON() ([]byte, error) {
	s, err := jsonpbMarshaller.MarshalToString(r)
	return []byte(s), err
}

func (r *Pair) UnmarshalJSON(b []byte) error {
	reader := bytes.NewBuffer(b)
	return jsonpbUnmarshaller.Unmarshal(reader, r)
}

// Some compile time assertions to ensure we don't
// have accidental runtime surprises later on.
// jsonEncodingRoundTripper ensures that asserted
// interfaces implement both MarshalJSON and UnmarshalJSON

type jsonRoundTripper interface {
	json.Marshaler
	json.Unmarshaler
}

var _ jsonRoundTripper = (*Pair)(nil)
