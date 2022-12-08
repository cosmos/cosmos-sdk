package genesis

import (
	"encoding/json"
	"io"

	"google.golang.org/protobuf/runtime/protoiface"
)

// Source is a source for reading genesis files in JSON format or reading
// from the protobuf messages. It may abstract over a single JSON object or JSON
// in separate files that can be streamed over.
type Source interface {
	// ReadMessage reads the source as a protobuf message.
	ReadMessage(protoiface.MessageV1) error

	// OpenReader returns an io.ReadCloser for the named field. If there
	// is no genesis file, this method will return nil. It is
	// important the caller closes the reader when done with it.
	OpenReader(field string) (io.ReadCloser, error)

	// ReadRawJSON reads an encoded JSON data from the source file.
	ReadRawJSON() (json.RawMessage, error)
}

// Target is a target for writing genesis files in JSON format or
// writing the genesis to protobuf messages. It may abstract over a single JSON
// object or JSON in separate files that can be streamed over.
type Target interface {

	// WriteMessage writes a protobuf message to the target.
	WriteMessage(protoiface.MessageV1) error

	// OpenWriter returns an io.WriteCloser for the named field. It is
	// important the caller closers the writer AND checks the error
	// when done with it.
	OpenWriter(field string) (io.WriteCloser, error)

	// WriteRawJSON writes the encoded JSON data to the target file.
	WriteRawJSON(json.RawMessage) error
}
