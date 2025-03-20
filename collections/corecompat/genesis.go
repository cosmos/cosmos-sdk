package corecompat

import "io"

// GenesisSource is a source for genesis data in JSON format. It may abstract over a
// single JSON object or separate files for each field in a JSON object that can
// be streamed over. Modules should open a separate io.ReadCloser for each field that
// is required. When fields represent arrays they can efficiently be streamed
// over. If there is no data for a field, this function should return nil, nil. It is
// important that the caller closes the reader when done with it.
type GenesisSource = func(field string) (io.ReadCloser, error)

// GenesisTarget is a target for writing genesis data in JSON format. It may
// abstract over a single JSON object or JSON in separate files that can be
// streamed over. Modules should open a separate io.WriteCloser for each field
// and should prefer writing fields as arrays when possible to support efficient
// iteration. It is important the caller closers the writer AND checks the error
// when done with it. It is expected that a stream of JSON data is written
// to the writer.
type GenesisTarget = func(field string) (io.WriteCloser, error)
