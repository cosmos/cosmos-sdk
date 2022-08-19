package appmodule

import (
	"encoding/json"
	"io"

	"google.golang.org/protobuf/runtime/protoiface"
)

type GenesisSource interface {
	ReadMessage(protoiface.MessageV1) error
	OpenReader(field string) (io.ReadCloser, error)
	ReadRawJSON() (json.RawMessage, error)
}

type GenesisTarget interface {
	WriteMessage(protoiface.MessageV1) error
	OpenWriter(field string) (io.WriteCloser, error)
	WriteRawJSON(json.RawMessage) error
}
