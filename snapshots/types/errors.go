package types

import "errors"

var (
	// ErrUnknownFormat is returned when an unknown format is used.
	ErrUnknownFormat = errors.New("unknown snapshot format")

	// ErrChunkHashMismatch is returned when chunk hash verification failed
	ErrChunkHashMismatch = errors.New("chunk hash verification failed")
)
