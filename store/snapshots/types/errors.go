package types

import (
	"errors"
)

var (
	// ErrUnknownFormat is returned when an unknown format is used.
	ErrUnknownFormat = errors.New("unknown snapshot format")

	// ErrChunkHashMismatch is returned when chunk hash verification failed.
	ErrChunkHashMismatch = errors.New("chunk hash verification failed")

	// ErrInvalidMetadata is returned when the snapshot metadata is invalid.
	ErrInvalidMetadata = errors.New("invalid snapshot metadata")

	// ErrInvalidSnapshotVersion is returned when the snapshot version is invalid
	ErrInvalidSnapshotVersion = errors.New("invalid snapshot version")

	// ErrDecompressedChunkTooLarge is returned when a snapshot chunk decompresses into more
	// bytes than allowed, guarding against a malicious peer forcing unbounded decompression
	// work (a decompression bomb) before the resulting state root can be verified.
	ErrDecompressedChunkTooLarge = errors.New("decompressed snapshot chunk exceeds maximum allowed size: possible decompression bomb")
)
