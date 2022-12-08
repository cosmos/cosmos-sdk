package genesis

import (
	"encoding/json"
	"io"

	"google.golang.org/protobuf/runtime/protoiface"
)

// Source is a source for genesis data in JSON format. It may abstract over a
// single JSON object or JSON in separate files that can be streamed over. Those
// details are left to the implementation. Generally modules should either
// consume a single JSON blob using the ReadMessage or ReadRawJSON methods OR
// ideally, read individual fields using the OpenReader method. Using the
// OpenReader method with array data will allow that data to be streamed which
// can result in lower memory usage than reading a single large JSON blob into
// memory.
type Source interface {
	// ReadRawJSON reads the all the genesis data as a single JSON blob.
	ReadRawJSON() (json.RawMessage, error)

	// ReadMessage reads the all the genesis data and attempts to unmarshal
	// it as the provided protobuf message.
	ReadMessage(protoiface.MessageV1) error

	// OpenReader returns an io.ReadCloser for the named field. If there
	// is data for this field, this method will return nil, nil. It is
	// important that the caller closes the reader when done with it.
	// It is expected that the reader points to a stream of JSON data.
	OpenReader(field string) (io.ReadCloser, error)
}

// Target is a target for writing genesis data in JSON format or. It may
// abstract over a single JSON object or JSON in separate files that can be
// streamed over. Modules should prefer writing fields separately if possible
// using the OpenWriter method to take advantage of streaming.
type Target interface {
	// WriteRawJSON writes the encoded JSON data to the target.
	WriteRawJSON(json.RawMessage) error

	// WriteMessage writes the provided protobuf message to the target.
	WriteMessage(protoiface.MessageV1) error

	// OpenWriter returns an io.WriteCloser for the named field. It is
	// important the caller closers the writer AND checks the error
	// when done with it. It is expected that a steam of JSON data is written
	// to the writer.
	OpenWriter(field string) (io.WriteCloser, error)
}
