package ormjson

import (
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// ReadSource is a source for reading tables in JSON format. It
// may abstract over a single JSON object or JSON in separate files that
// can be streamed over.
type ReadSource interface {
	// OpenReader returns an io.ReadCloser for the named table. If there
	// is no JSON for this table, this method will return nil. It is
	// important the caller closes the reader when done with it.
	OpenReader(tableName protoreflect.FullName) (io.ReadCloser, error)
}

// WriteTarget is a target for writing tables in JSON format. It
// may abstract over a single JSON object or JSON in separate files that
// can be written incrementally.
type WriteTarget interface {
	// OpenWriter returns an io.WriteCloser for the named table. It is
	// important the caller closers the writer AND checks the error
	// when done with it.
	OpenWriter(tableName protoreflect.FullName) (io.WriteCloser, error)
}
