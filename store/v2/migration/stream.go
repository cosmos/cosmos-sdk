package migration

import (
	"fmt"
	"io"
	"sync/atomic"

	protoio "github.com/cosmos/gogoproto/io"
	"github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/store/v2/snapshots"
	snapshotstypes "cosmossdk.io/store/v2/snapshots/types"
)

var (
	_ snapshots.WriteCloser = (*MigrationStream)(nil)
	_ protoio.ReadCloser    = (*MigrationStream)(nil)
)

// MigrationStream is a stream for migrating the whole IAVL state as a snapshot.
// It's used to sync the whole state from the store/v1 to store/v2.
// The main idea is to use the same snapshotter interface without writing to disk.
type MigrationStream struct {
	chBuffer chan proto.Message
	err      atomic.Value // atomic error
}

// NewMigrationStream returns a new MigrationStream.
func NewMigrationStream(chBufferSize int) *MigrationStream {
	return &MigrationStream{
		chBuffer: make(chan proto.Message, chBufferSize),
	}
}

// WriteMsg implements protoio.Write interface.
func (ms *MigrationStream) WriteMsg(msg proto.Message) error {
	ms.chBuffer <- msg
	return nil
}

// CloseWithError implements snapshots.WriteCloser interface.
func (ms *MigrationStream) CloseWithError(err error) {
	ms.err.Store(err)
	close(ms.chBuffer)
}

// ReadMsg implements the protoio.Read interface.
//
// NOTE: It we follow the pattern of snapshot.Restore, however, the migration is done in memory.
// It doesn't require any deserialization -- just passing the pointer to the <msg>.
func (ms *MigrationStream) ReadMsg(msg proto.Message) error {
	// msg should be a pointer to the same type as the one written to the stream
	snapshotsItem, ok := msg.(*snapshotstypes.SnapshotItem)
	if !ok {
		return fmt.Errorf("unexpected message type: %T", msg)
	}

	// It doesn't require any deserialization, just a type assertion.
	item := <-ms.chBuffer
	if item == nil {
		return io.EOF
	}

	*snapshotsItem = *(item.(*snapshotstypes.SnapshotItem))

	// check if there is an error from the writer.
	err := ms.err.Load()
	if err != nil {
		return err.(error)
	}

	return nil
}

// Close implements io.Closer interface.
func (ms *MigrationStream) Close() error {
	close(ms.chBuffer)
	return nil
}
