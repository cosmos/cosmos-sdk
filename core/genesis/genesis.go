package genesis

import (
	"io"
)

// Source is a source for genesis data in JSON format. It may abstract over a
// single JSON object or separate files for each field in a JSON object that can
// be streamed over. Modules should open a separate reader for each field that
// is required. When fields represent arrays they can efficiently be streamed
// over.
type Source func (field string) (io.ReadCloser, error)
	// OpenReader returns an io.ReadCloser for the named field. If there
	// is data for this field, this method will return nil, nil. It is
	// important that the caller closes the reader when done with it.
	// It is expected that the reader points to a stream of JSON data.
	OpenReader(field string) (io.ReadCloser, error)
}

// Target is a target for writing genesis data in JSON format or. It may
// abstract over a single JSON object or JSON in separate files that can be
// streamed over. Modules should prefer writing fields as arrays when possible
// to support efficient iteration.
type Target interface {
	// OpenWriter returns an io.WriteCloser for the named field. It is
	// important the caller closers the writer AND checks the error
	// when done with it. It is expected that a steam of JSON data is written
	// to the writer.
	OpenWriter(field string) (io.WriteCloser, error)
}
