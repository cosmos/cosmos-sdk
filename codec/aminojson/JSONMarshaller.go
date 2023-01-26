package aminojson

import (
	"bytes"
	"github.com/cosmos/gogoproto/proto"
)

type JSONMarshaller interface {
	MarshalAmino(proto.Message) ([]byte, error)
}

func MarshalAmino(message proto.Message) ([]byte, error) {
	buf := &bytes.Buffer{}
	return buf.Bytes(), nil
}
