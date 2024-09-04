package unorderedtx

import (
	"encoding/binary"
	"errors"
	"io"
	"time"
)

const (
	txHashSize  = 32
	timeoutSize = 8
	chunkSize   = txHashSize + timeoutSize
)

// ErrUnknownFormat is returned when an unknown format is used.
var ErrUnknownFormat = errors.New("unknown snapshot format")

// ExtensionPayloadReader read extension payloads,
// it returns io.EOF when reached either end of stream or the extension boundaries.
type ExtensionPayloadReader = func() ([]byte, error)

// ExtensionPayloadWriter is a helper to write extension payloads to underlying stream.
type ExtensionPayloadWriter = func([]byte) error

// ExtensionSnapshotter is an extension Snapshotter that is appended to the snapshot stream.
// ExtensionSnapshotter has an unique name and manages it's own internal formats.
type ExtensionSnapshotter interface {
	// SnapshotName returns the name of snapshotter, it should be unique in the manager.
	SnapshotName() string

	// SnapshotFormat returns the default format the extension snapshotter use to encode the
	// payloads when taking a snapshot.
	// It's defined within the extension, different from the global format for the whole state-sync snapshot.
	SnapshotFormat() uint32

	// SupportedFormats returns a list of formats it can restore from.
	SupportedFormats() []uint32

	// SnapshotExtension writes extension payloads into the underlying protobuf stream.
	SnapshotExtension(height uint64, payloadWriter ExtensionPayloadWriter) error

	// RestoreExtension restores an extension state snapshot,
	// the payload reader returns `io.EOF` when reached the extension boundaries.
	RestoreExtension(height uint64, format uint32, payloadReader ExtensionPayloadReader) error
}

var _ ExtensionSnapshotter = &Snapshotter{}

const (
	// SnapshotFormat defines the snapshot format of exported unordered transactions.
	// No protobuf envelope, no metadata.
	SnapshotFormat = 1

	// SnapshotName defines the snapshot name of exported unordered transactions.
	SnapshotName = "unordered_txs"
)

type Snapshotter struct {
	m *Manager
}

func NewSnapshotter(m *Manager) *Snapshotter {
	return &Snapshotter{m: m}
}

func (s *Snapshotter) SnapshotName() string {
	return SnapshotName
}

func (s *Snapshotter) SnapshotFormat() uint32 {
	return SnapshotFormat
}

func (s *Snapshotter) SupportedFormats() []uint32 {
	return []uint32{SnapshotFormat}
}

func (s *Snapshotter) SnapshotExtension(height uint64, payloadWriter ExtensionPayloadWriter) error {
	// export all unordered transactions as a single blob
	return s.m.exportSnapshot(height, payloadWriter)
}

func (s *Snapshotter) RestoreExtension(height uint64, format uint32, payloadReader ExtensionPayloadReader) error {
	if format == SnapshotFormat {
		return s.restore(height, payloadReader)
	}

	return ErrUnknownFormat
}

func (s *Snapshotter) restore(height uint64, payloadReader ExtensionPayloadReader) error {
	// the payload should be the entire set of unordered transactions
	payload, err := payloadReader()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return io.ErrUnexpectedEOF
		}

		return err
	}

	if len(payload)%chunkSize != 0 {
		return errors.New("invalid unordered txs payload length")
	}

	var i int
	for i < len(payload) {
		var txHash TxHash
		copy(txHash[:], payload[i:i+txHashSize])

		timestamp := binary.BigEndian.Uint64(payload[i+txHashSize : i+chunkSize])

		// purge any expired txs
		if timestamp != 0 && timestamp > height {
			s.m.Add(txHash, time.Unix(int64(timestamp), 0))
		}

		i += chunkSize
	}

	return nil
}
