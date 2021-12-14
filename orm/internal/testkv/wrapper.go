package testkv

import (
	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
)

type indexCommitmentStoreWrapper struct {
	commitment debugReader
	index      debugReader
	store      kvstore.IndexCommitmentStore
}

type icsWriterWrapper struct {
	commitment debugWriter
	index      debugWriter
	writer     kvstore.IndexCommitmentStoreWriter
}

func (i icsWriterWrapper) CommitmentStoreReader() kvstore.Reader {
	return i.writer.CommitmentStoreReader()
}

func (i icsWriterWrapper) IndexStoreReader() kvstore.Reader {
	return i.writer.IndexStoreReader()
}

func (i icsWriterWrapper) CommitmentStoreWriter() kvstore.Writer {
	return i.commitment
}

func (i icsWriterWrapper) IndexStoreWriter() kvstore.Writer {
	return i.index
}

func (i icsWriterWrapper) Write() error {
	return i.writer.Write()
}

func (i icsWriterWrapper) Close() {
	i.writer.Close()
}

func (i indexCommitmentStoreWrapper) NewWriter() kvstore.IndexCommitmentStoreWriter {
	writer := i.store.NewWriter()
	return &icsWriterWrapper{
		debugWriter{
			debugReader: i.commitment,
			writer:      writer.CommitmentStoreWriter(),
		},
		debugWriter{
			debugReader: i.index,
			writer:      writer.IndexStoreWriter(),
		},
		writer,
	}
}

// NewDebugIndexCommitmentStore wraps both stores from an IndexCommitmentStore
// with a debugger.
func NewDebugIndexCommitmentStore(store kvstore.IndexCommitmentStore, debugger Debugger) kvstore.IndexCommitmentStore {
	return &indexCommitmentStoreWrapper{
		store:      store,
		commitment: debugReader{store.CommitmentStoreReader(), debugger, "commit"},
		index:      debugReader{store.IndexStoreReader(), debugger, "index"},
	}

}

var _ kvstore.IndexCommitmentStore = &indexCommitmentStoreWrapper{}

func (i indexCommitmentStoreWrapper) CommitmentStoreReader() kvstore.Reader {
	return i.commitment
}

func (i indexCommitmentStoreWrapper) IndexStoreReader() kvstore.Reader {
	return i.index
}
